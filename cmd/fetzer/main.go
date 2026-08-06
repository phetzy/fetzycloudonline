// Command fetzer serves the personal site over SSH as a Bubble Tea TUI.
//
// This is an unauthenticated listener intended for the public internet:
// every public key is accepted (there is no notion of "logging in"), and
// every session is handled exclusively by the Bubble Tea middleware. There
// is deliberately no shell, no exec, no subsystem — a client cannot run a
// command, open SFTP, or forward a port, because no handler exists for any
// of those request types. A per-connection session gate additionally caps
// each TCP connection to a single granted session channel, an unrecovered
// panic in the session-handling path is caught before it can take down the
// whole process, and every timeout and limit is validated to be strictly
// positive at startup so a config typo cannot silently disable a control.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	wishbubbletea "github.com/charmbracelet/wish/bubbletea"
	wishrecover "github.com/charmbracelet/wish/recover"
	gossh "golang.org/x/crypto/ssh"

	site "github.com/phetzy/fetzycloudonline"
	"github.com/phetzy/fetzycloudonline/internal/ratelimit"
	"github.com/phetzy/fetzycloudonline/internal/tui"
)

// shutdownWindow bounds how long graceful shutdown waits for live sessions
// to finish on their own before the server is force-closed.
const shutdownWindow = 10 * time.Second

// maxLoggedUserLen bounds how much of the client-supplied SSH username is
// ever written to the log. sess.User() is attacker-controlled and
// unbounded in length; slog's TextHandler escapes it so log injection is
// mitigated, but an unbounded field is still an unforced way to bloat logs.
const maxLoggedUserLen = 32

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg := parseConfig()
	validateConfig(cfg)

	content := site.MustLoad()
	limiter := ratelimit.New(cfg.maxConnsPerIP, cfg.ratePerMin)

	srv, err := wish.NewServer(
		wish.WithAddress(cfg.addr),
		wish.WithHostKeyPath(cfg.hostKey),
		wish.WithIdleTimeout(cfg.idleTimeout),
		wish.WithMaxTimeout(cfg.maxSession),
		// Accept every public key: this server is public by design and has
		// no concept of authorized users. This is not a placeholder for
		// "add real auth later" — it is the intended, permanent behavior.
		wish.WithPublicKeyAuth(func(ssh.Context, ssh.PublicKey) bool { return true }),
		ssh.WrapConn(rateLimitConnCallback(logger, limiter)),
		// Caps each TCP connection to a single granted session channel (see
		// sessionRequestCallback doc comment for why this is needed at
		// all: the per-IP connection limiter above bounds TCP connections,
		// not SSH session channels, and neither charmbracelet/ssh nor
		// x/crypto/ssh cap the latter on their own).
		withSessionRequestCallback(sessionRequestCallback(logger)),
		// Caps how many "session" channels (not just granted session
		// channels — see sessionChannelCounterContextKey's doc comment) a
		// single TCP connection can open, closing the channel itself
		// rather than merely refusing what happens inside it.
		withSessionChannelLimit(logger),
		// Wish composes middleware so that the LAST one listed here runs
		// FIRST (each wraps the next; last-in is outermost). The list below
		// is [recover-wrapped(bubbletea, activeterm), logging], which puts:
		//
		//   logging (outermost) -> recover -> activeterm (gate) -> bubbletea (innermost)
		//
		// in actual execution order:
		//
		//   - activeterm still runs before bubbletea ever sees the session
		//     (rejecting any session without a PTY) — recover.MiddlewareWithLogger
		//     composes its own variadic middleware list the same way Wish
		//     does (last-in is outermost), so passing (bubbletea, activeterm)
		//     to it puts activeterm outside bubbletea, same as before.
		//   - Both bubbletea's and activeterm's handler bodies now run under
		//     recover(): charmbracelet/ssh runs the session handler in a bare
		//     goroutine with no recover of its own, and while Bubble Tea
		//     recovers panics inside Run/Update/View, it does NOT recover a
		//     panic in tui.New itself or in wish middleware bodies. Without
		//     this, one bad input coercing a panic anywhere in that path
		//     would be an unrecovered goroutine panic — the whole process
		//     dies, taking every other live session down with it.
		//   - logging wraps the ENTIRE session lifetime (including the
		//     recover boundary), so its connect line is emitted at the real
		//     start of the session and its disconnect line at the real end,
		//     with an accurate duration, and disconnects are logged even for
		//     a session that panicked.
		//
		// The naive reading of "logging -> bubbletea -> activeterm" as the
		// literal call order instead makes logging innermost: its own
		// function body only runs after bubbletea's handler has already
		// returned (bubbletea does all its work before calling next, not
		// after), so both the connect and disconnect lines fire together,
		// post-hoc, once the whole session is already over, with a
		// duration of a few microseconds regardless of how long the
		// session actually ran. Verified empirically against a live
		// session before settling on this order.
		//
		// There is no ssh.WithSubsystem, no shell handler, and no exec
		// handler registered anywhere: the only way a session produces
		// output is through this middleware chain.
		wish.WithMiddleware(
			wishrecover.MiddlewareWithLogger(printfLogger{logger},
				wishbubbletea.Middleware(teaHandler(content)),
				activeterm.Middleware(),
			),
			loggingMiddleware(logger),
		),
	)
	if err != nil {
		logger.Error("failed to configure server", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", cfg.addr)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		if err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		stop()
		logger.Info("shutdown signal received, stopping new connections", "window", shutdownWindow)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownWindow)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Warn("graceful shutdown window exceeded, forcing close", "error", err)
			if cerr := srv.Close(); cerr != nil {
				logger.Error("forced close failed", "error", cerr)
			}
		} else {
			logger.Info("shutdown complete")
		}
	}
}

// teaHandler adapts the site content into a wish/bubbletea Handler. Each
// session gets its own Model, wired to write OSC 52 sequences (link reveal)
// back to that same session so the "copy to clipboard" escape reaches the
// connected client rather than the server's own terminal.
//
// tea.WithoutSignalHandler() is required here: by default every tea.Program
// installs its own process-wide SIGINT/SIGTERM handler (see bubbletea's
// Program.handleSignals). Since Go delivers a signal to every channel
// registered via signal.Notify, the process-level SIGTERM this server
// listens for to trigger graceful shutdown would otherwise also land in
// every live session's Bubble Tea program simultaneously, quitting all of
// them immediately and defeating the whole point of the bounded shutdown
// window below. The per-session program is instead torn down by the
// wish/bubbletea middleware itself, which cancels the session's own context
// (derived from ssh.Session.Context(), independent of OS signals) and calls
// program.Quit() when that session's connection actually closes.
func teaHandler(content site.Content) wishbubbletea.Handler {
	return func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
		// wishbubbletea.MakeRenderer binds a lipgloss renderer to this
		// session's terminal, so color detection reflects what the
		// connecting client supports. The package-level lipgloss default
		// renderer instead detects color support from this process's own
		// stdout, which under systemd is a journald socket, not a TTY —
		// that would strip color for every visitor regardless of their
		// terminal.
		renderer := wishbubbletea.MakeRenderer(sess)
		m := tui.New(content, renderer, sess)
		return m, []tea.ProgramOption{tea.WithAltScreen(), tea.WithoutSignalHandler()}
	}
}

// loggingMiddleware records connect and disconnect events, with the remote
// address, using structured logging.
func loggingMiddleware(logger *slog.Logger) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			addr := sess.RemoteAddr().String()
			user := truncateUTF8(sess.User(), maxLoggedUserLen)
			start := time.Now()
			logger.Info("session connect", "remote_addr", addr, "user", user)
			next(sess)
			logger.Info("session disconnect", "remote_addr", addr, "duration", time.Since(start))
		}
	}
}

// printfLogger adapts a *slog.Logger to wish/recover's Logger interface
// (a single Printf method), so panic reports from the recover middleware
// go through the same structured logger as everything else.
type printfLogger struct {
	logger *slog.Logger
}

func (p printfLogger) Printf(format string, v ...interface{}) {
	p.logger.Error(fmt.Sprintf(format, v...))
}

// sessionGateContextKey is the ssh.Context key under which each
// connection's *sessionGate is stored. It is set once per TCP connection,
// in rateLimitConnCallback, and read by sessionRequestCallback for every
// session channel request on that connection — the same ssh.Context is
// shared by every channel multiplexed over one connection, which is what
// makes a per-connection cap possible here.
type sessionGateContextKey struct{}

// sessionGate caps a single TCP connection to one granted session channel.
type sessionGate struct {
	count atomic.Int32

	// rejectionLogged ensures sessionRequestCallback logs at most one
	// "session rejected" line per connection, no matter how many requests
	// the client sends on however many channels it opens. Without this, a
	// client that gets its one legitimate channel and then sends channel
	// requests on it at packet rate — no reconnect, no new channel, so
	// neither the per-IP limiter nor the channel-open cap ever sees it —
	// could log at whatever rate it sends requests.
	rejectionLogged atomic.Bool
}

// allow reports whether this call is the first to succeed for this gate.
// Every subsequent call, including concurrent ones, returns false. Safe
// for concurrent use.
func (g *sessionGate) allow() bool {
	return g.count.Add(1) == 1
}

// sessionRequestCallback caps each TCP connection to a single granted
// session channel (a "shell", "exec", or "subsystem" request).
//
// Without this, the per-IP connection limiter in rateLimitConnCallback only
// bounds how many TCP connections an IP can hold open — it does not bound
// how many SSH session channels a single connection can multiplex. Neither
// charmbracelet/ssh nor golang.org/x/crypto/ssh impose a cap of their own:
// the server spawns a new goroutine, with its own tea.Program, Model, and
// renderer once bubbletea's middleware runs, for every incoming session
// channel. A handful of TCP connections from one IP could otherwise open
// an unbounded number of session channels and exhaust memory — the
// cheapest resource-exhaustion path available against this server, and one
// that would otherwise silently void the concurrency cap the rate limiter
// exists to enforce.
func sessionRequestCallback(logger *slog.Logger) ssh.SessionRequestCallback {
	return func(sess ssh.Session, requestType string) bool {
		gate, ok := sess.Context().Value(sessionGateContextKey{}).(*sessionGate)
		if !ok {
			// Unreachable in normal operation: rateLimitConnCallback always
			// installs a gate before any channel on the connection can be
			// requested. Fail closed rather than silently allowing an
			// unbounded number of sessions if that invariant is ever broken.
			logger.Warn("session rejected",
				"remote_addr", sess.RemoteAddr().String(),
				"reason", "no session gate on connection context",
			)
			return false
		}

		if !gate.allow() {
			// Every request is still rejected; only the logging is capped —
			// CompareAndSwap(false, true) succeeds for exactly one goroutine
			// per gate, so this connection logs at most one line here
			// regardless of how many rejected requests it sends.
			if gate.rejectionLogged.CompareAndSwap(false, true) {
				logger.Warn("session rejected",
					"remote_addr", sess.RemoteAddr().String(),
					"reason", "connection already has a granted session channel",
					"request_type", requestType,
				)
			}
			return false
		}

		return true
	}
}

// withSessionRequestCallback is an ssh.Option setting Server.SessionRequestCallback.
// Wish has no built-in wrapper for this option (unlike WithPublicKeyAuth,
// WithIdleTimeout, etc.), so it is set directly here.
func withSessionRequestCallback(cb ssh.SessionRequestCallback) ssh.Option {
	return func(s *ssh.Server) error {
		s.SessionRequestCallback = cb
		return nil
	}
}

// sessionChannelCounterContextKey is the ssh.Context key under which each
// connection's session-channel open counter is stored. Installed once per
// connection in rateLimitConnCallback, alongside the sessionGate, and read
// by every "session" channel handler wrapped with limitSessionChannels on
// that connection.
type sessionChannelCounterContextKey struct{}

// maxSessionChannelsPerConn caps how many "session" channels a single TCP
// connection may open.
//
// sessionRequestCallback (above) only gates what happens *inside* an
// already-open channel — a "shell", "exec", or "subsystem" request.
// charmbracelet/ssh's DefaultSessionHandler accepts the raw channel itself
// (newChan.Accept(), one goroutine, one unbounded env slice) before any of
// that gate is ever consulted. Without a separate cap here, a single
// connection inside the per-IP concurrency budget could open session
// channels without limit: unbounded goroutine and memory growth, and log
// amplification too, since every request rejected inside a channel logs a
// Warn. Four is generous headroom for a single-screen TUI, which only ever
// opens one.
const maxSessionChannelsPerConn = 4

// limitSessionChannels wraps a ChannelHandler (ssh.DefaultSessionHandler in
// practice) so that "session" channel opens past
// maxSessionChannelsPerConn on one connection are rejected outright — the
// channel itself is never accepted, so no goroutine, session, or env slice
// is ever created for it. The rejection is logged once per connection, on
// the first channel open that crosses the limit, rather than once per
// request, so a client that keeps trying cannot amplify the log path either.
func limitSessionChannels(logger *slog.Logger, next ssh.ChannelHandler) ssh.ChannelHandler {
	return func(srv *ssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx ssh.Context) {
		counter, ok := ctx.Value(sessionChannelCounterContextKey{}).(*atomic.Int32)
		if !ok {
			// Unreachable in normal operation, mirroring the analogous
			// fallback in sessionRequestCallback: rateLimitConnCallback
			// always installs the counter before any channel on the
			// connection can be requested. Fail closed rather than
			// silently allowing unlimited channels if that invariant is
			// ever broken.
			_ = newChan.Reject(gossh.ResourceShortage, "no channel counter on connection context")
			return
		}

		if n := counter.Add(1); n > maxSessionChannelsPerConn {
			if n == maxSessionChannelsPerConn+1 {
				logger.Warn("session channel rejected",
					"remote_addr", conn.RemoteAddr().String(),
					"reason", "connection exceeded per-connection session channel limit",
					"limit", maxSessionChannelsPerConn,
				)
			}
			_ = newChan.Reject(gossh.ResourceShortage, "too many session channels on this connection")
			return
		}

		next(srv, conn, newChan, ctx)
	}
}

// withSessionChannelLimit is an ssh.Option installing limitSessionChannels
// around whatever "session" channel handler the server would otherwise use
// (ssh.DefaultSessionHandler, since nothing else in this program overrides
// it). It must run after any option that might set ChannelHandlers itself,
// though nothing here does; it is defensive against that changing later.
func withSessionChannelLimit(logger *slog.Logger) ssh.Option {
	return func(s *ssh.Server) error {
		handlers := make(map[string]ssh.ChannelHandler, len(ssh.DefaultChannelHandlers)+len(s.ChannelHandlers))
		for k, v := range ssh.DefaultChannelHandlers {
			handlers[k] = v
		}
		for k, v := range s.ChannelHandlers {
			handlers[k] = v
		}
		handlers["session"] = limitSessionChannels(logger, handlers["session"])
		s.ChannelHandlers = handlers
		return nil
	}
}

// rateLimitConnCallback wraps every accepted net.Conn, before the SSH
// handshake begins, with the per-IP concurrency and rate limiter. A
// rejected connection is closed immediately with a logged reason and never
// reaches the SSH protocol layer at all. The returned wrapper releases the
// limiter slot when the connection is closed, however that happens
// (session end, idle timeout, max-session timeout, or shutdown) — the
// limiter's release is idempotent, so double-release from overlapping
// close paths is safe.
//
// It also installs a fresh *sessionGate and a fresh session-channel counter
// (*atomic.Int32, under sessionChannelCounterContextKey) on the connection's
// ssh.Context: the gate is what sessionRequestCallback uses to cap the
// connection to a single granted session channel, and the counter is what
// limitSessionChannels uses to cap how many session channels the connection
// can open in the first place. This is the only point in the code that runs
// exactly once per TCP connection with access to that connection's
// ssh.Context, before any channel on it can be requested — installing
// either one anywhere else would race multiple concurrent channel requests
// against each other to initialize it.
func rateLimitConnCallback(logger *slog.Logger, limiter *ratelimit.Limiter) ssh.ConnCallback {
	return func(ctx ssh.Context, conn net.Conn) net.Conn {
		ip := hostOnly(conn.RemoteAddr())

		release, ok := limiter.Acquire(ip)
		if !ok {
			logger.Warn("connection rejected",
				"remote_addr", conn.RemoteAddr().String(),
				"reason", "per-IP concurrency or rate limit exceeded",
			)
			_ = conn.Close()
			return nil
		}

		ctx.SetValue(sessionGateContextKey{}, &sessionGate{})
		ctx.SetValue(sessionChannelCounterContextKey{}, &atomic.Int32{})

		return &releasingConn{Conn: conn, release: release}
	}
}

// releasingConn frees its limiter slot when the connection is closed.
type releasingConn struct {
	net.Conn
	release func()
}

func (c *releasingConn) Close() error {
	c.release()
	return c.Conn.Close()
}

// hostOnly strips the port from addr, falling back to the full address if
// it cannot be split (e.g. it has no port) or is not a parseable IP. The
// rate limiter keys on the result so that multiple connections from the
// same client share a budget.
//
// IPv4 addresses are kept as-is. IPv6 addresses are masked to their /64:
// every consumer IPv6 allocation routed to an end user is a /64, so keying
// at the full /128 address would let anyone with such an allocation bypass
// the per-IP cap entirely just by cycling addresses within their own /64.
// Masking to /64 makes the limiter mean the same thing — one real-world
// client, one budget — on both protocols.
func hostOnly(addr net.Addr) string {
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		host = addr.String()
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return host
	}

	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}

	return ip.Mask(net.CIDRMask(64, 128)).String()
}

// truncateUTF8 returns s, or if s is longer than maxBytes, a valid-UTF8
// prefix of it (never splitting a multi-byte rune) with a trailing marker.
func truncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	b := s[:maxBytes]
	for len(b) > 0 && !utf8.ValidString(b) {
		b = b[:len(b)-1]
	}
	return b + "…"
}

// config holds the server's runtime configuration, populated from flags
// with environment variable fallbacks.
type config struct {
	addr          string
	hostKey       string
	idleTimeout   time.Duration
	maxSession    time.Duration
	maxConnsPerIP int
	ratePerMin    int
}

// parseConfig reads flags, falling back to environment variables and then
// hardcoded defaults when a flag is not set on the command line.
func parseConfig() config {
	var cfg config

	flag.StringVar(&cfg.addr, "addr", envString("FETZER_ADDR", ":23234"),
		"address to listen on")
	flag.StringVar(&cfg.hostKey, "host-key", envString("FETZER_HOST_KEY", "./.ssh/fetzer_ed25519"),
		"path to the SSH host key (created if it does not exist)")
	flag.DurationVar(&cfg.idleTimeout, "idle-timeout", envDuration("FETZER_IDLE_TIMEOUT", 5*time.Minute),
		"disconnect a session after this long with no activity")
	flag.DurationVar(&cfg.maxSession, "max-session", envDuration("FETZER_MAX_SESSION", 30*time.Minute),
		"hard cap on a single session's total duration, active or not")
	flag.IntVar(&cfg.maxConnsPerIP, "max-conns-per-ip", envInt("FETZER_MAX_CONNS_PER_IP", 3),
		"maximum concurrent connections allowed from a single IP")
	flag.IntVar(&cfg.ratePerMin, "rate-per-min", envInt("FETZER_RATE_PER_MIN", 20),
		"maximum new connections allowed from a single IP per minute")

	flag.Parse()

	return cfg
}

// validateConfig rejects any configuration where a security-relevant
// timeout or limit is non-positive, printing a clear diagnostic and
// exiting(2) rather than starting with a control silently disabled.
//
// This matters specifically because zero and negative values do not fail
// loudly downstream: an idle-timeout or max-session of 0 makes the
// underlying ssh library's serverConn.updateDeadline fall through to
// SetDeadline(maxDeadline); if max-session is ALSO 0, that deadline is the
// zero time, which clears the deadline entirely — no idle timeout, no
// session cap, connections held open forever. A non-positive
// max-conns-per-ip or rate-per-min hits internal/ratelimit's documented
// fail-closed behavior instead (rejects all traffic), which is safer but
// still not something a config typo should be able to trigger silently.
func validateConfig(cfg config) {
	problems := configProblems(cfg)
	if len(problems) == 0 {
		return
	}
	for _, p := range problems {
		fmt.Fprintln(os.Stderr, "error:", p)
	}
	fmt.Fprintln(os.Stderr, "refusing to start: a non-positive value for a security-relevant "+
		"timeout or limit would silently disable the control it names")
	os.Exit(2)
}

// configProblems returns a human-readable problem description for every
// non-positive security-relevant field in cfg. Separated from
// validateConfig so it can be unit tested without invoking os.Exit.
func configProblems(cfg config) []string {
	var problems []string
	if cfg.idleTimeout <= 0 {
		problems = append(problems, fmt.Sprintf("-idle-timeout must be positive, got %s", cfg.idleTimeout))
	}
	if cfg.maxSession <= 0 {
		problems = append(problems, fmt.Sprintf("-max-session must be positive, got %s", cfg.maxSession))
	}
	if cfg.maxConnsPerIP <= 0 {
		problems = append(problems, fmt.Sprintf("-max-conns-per-ip must be positive, got %d", cfg.maxConnsPerIP))
	}
	if cfg.ratePerMin <= 0 {
		problems = append(problems, fmt.Sprintf("-rate-per-min must be positive, got %d", cfg.ratePerMin))
	}
	return problems
}

func envString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// envDuration reads key as a duration, falling back to def (with a
// diagnostic on stderr) if the variable is unset, unparsable, or
// non-positive. A non-positive duration is rejected here rather than only
// caught later by validateConfig so that a bad *environment* value is
// pinpointed by name instead of surfacing as an opaque flag-level error.
func envDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	switch {
	case err != nil:
		fmt.Fprintf(os.Stderr, "warning: invalid duration in %s (%q), using default %s\n", key, v, def)
		return def
	case d <= 0:
		fmt.Fprintf(os.Stderr, "warning: %s must be positive, got %s, using default %s\n", key, d, def)
		return def
	default:
		return d
	}
}

// envInt reads key as an integer, falling back to def (with a diagnostic on
// stderr) if the variable is unset, unparsable, or non-positive. See
// envDuration for why non-positive values are rejected here too.
func envInt(key string, def int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	switch {
	case err != nil:
		fmt.Fprintf(os.Stderr, "warning: invalid integer in %s (%q), using default %d\n", key, v, def)
		return def
	case n <= 0:
		fmt.Fprintf(os.Stderr, "warning: %s must be positive, got %d, using default %d\n", key, n, def)
		return def
	default:
		return n
	}
}

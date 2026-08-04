// Command fetzer serves the personal site over SSH as a Bubble Tea TUI.
//
// This is an unauthenticated listener intended for the public internet:
// every public key is accepted (there is no notion of "logging in"), and
// every session is handled exclusively by the Bubble Tea middleware. There
// is deliberately no shell, no exec, no subsystem — a client cannot run a
// command, open SFTP, or forward a port, because no handler exists for any
// of those request types.
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
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	wishbubbletea "github.com/charmbracelet/wish/bubbletea"

	site "github.com/phetzy/fetzycloudonline"
	"github.com/phetzy/fetzycloudonline/internal/ratelimit"
	"github.com/phetzy/fetzycloudonline/internal/tui"
)

// shutdownWindow bounds how long graceful shutdown waits for live sessions
// to finish on their own before the server is force-closed.
const shutdownWindow = 10 * time.Second

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg := parseConfig()

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
		// Wish composes middleware so that the LAST one listed here runs
		// FIRST (each wraps the next; last-in is outermost). The list below
		// is bubbletea, activeterm, logging — which puts:
		//
		//   logging (outermost) -> activeterm (gate) -> bubbletea (innermost)
		//
		// in actual execution order. This gives both required properties
		// at once: activeterm still runs before bubbletea ever sees the
		// session (rejecting any session without a PTY), and logging wraps
		// the ENTIRE session lifetime, so its connect line is emitted at
		// the real start of the session and its disconnect line at the
		// real end, with an accurate duration.
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
			wishbubbletea.Middleware(teaHandler(content)),
			activeterm.Middleware(),
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
		m := tui.New(content, sess)
		return m, []tea.ProgramOption{tea.WithAltScreen(), tea.WithoutSignalHandler()}
	}
}

// loggingMiddleware records connect and disconnect events, with the remote
// address, using structured logging.
func loggingMiddleware(logger *slog.Logger) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			addr := sess.RemoteAddr().String()
			start := time.Now()
			logger.Info("session connect", "remote_addr", addr, "user", sess.User())
			next(sess)
			logger.Info("session disconnect", "remote_addr", addr, "duration", time.Since(start))
		}
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
func rateLimitConnCallback(logger *slog.Logger, limiter *ratelimit.Limiter) ssh.ConnCallback {
	return func(_ ssh.Context, conn net.Conn) net.Conn {
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
// it cannot be split (e.g. it has no port). The rate limiter keys on host
// only so that multiple connections from the same client share a budget.
func hostOnly(addr net.Addr) string {
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String()
	}
	return host
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

func envString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		fmt.Fprintf(os.Stderr, "warning: invalid duration in %s, using default\n", key)
	}
	return def
}

func envInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		fmt.Fprintf(os.Stderr, "warning: invalid integer in %s, using default\n", key)
	}
	return def
}

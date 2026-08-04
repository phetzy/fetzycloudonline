package main

import (
	"bytes"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/ssh"
	wishrecover "github.com/charmbracelet/wish/recover"
)

// --- panic recovery: the exact wish/recover wiring used in main() ---
//
// This does not, and cannot practically, exercise a panic inside the real
// tui/bubbletea integration over a live network connection — that would
// require deliberately forcing bad state into internal/tui or forging
// malicious protocol input, which isn't something to do against the real
// session-handling path. What IS tested here is the actual composition
// used in main.go: printfLogger as the Logger adapter, and
// wishrecover.MiddlewareWithLogger wrapping a handler that panics. This
// pins two things a refactor could otherwise silently break: that a panic
// inside the wrapped chain does not escape (the test process itself would
// crash if it did), and that execution still reaches whatever comes after
// the recover middleware in the outer chain (loggingMiddleware's
// disconnect line, in the real server) rather than getting stuck.
func TestRecoverMiddlewareWiringSwallowsPanicAndContinues(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	panicking := func(ssh.Handler) ssh.Handler {
		return func(ssh.Session) {
			panic("boom: simulated panic inside the session-handling path")
		}
	}

	mw := wishrecover.MiddlewareWithLogger(printfLogger{logger}, panicking)

	var reachedNext bool
	next := func(ssh.Session) { reachedNext = true }

	handler := mw(next)

	// The important assertion is implicit: if the panic escaped, this call
	// itself would panic and fail the test (and, in the real server, take
	// the whole process down with it).
	handler(nil)

	if !reachedNext {
		t.Fatal("expected the outer chain's next() to still run after the panic was recovered, so disconnect logging is not skipped")
	}
	if !strings.Contains(buf.String(), "boom") {
		t.Fatalf("expected the panic message to reach the structured logger via printfLogger, got: %s", buf.String())
	}
}

// --- sessionGate: the per-connection session-channel cap ---

func TestSessionGateAllowsExactlyOne(t *testing.T) {
	g := &sessionGate{}

	if !g.allow() {
		t.Fatal("first call to allow() must return true")
	}
	if g.allow() {
		t.Fatal("second call to allow() must return false")
	}
	if g.allow() {
		t.Fatal("third call to allow() must return false")
	}
}

// TestSessionGateConcurrentAllowGrantsExactlyOne is the important case: the
// real attack this gate defends against is many session-channel requests
// arriving on one TCP connection at once. If allow() were not safe for
// concurrent use, a race could grant more than one.
func TestSessionGateConcurrentAllowGrantsExactlyOne(t *testing.T) {
	g := &sessionGate{}

	const n = 200
	var wg sync.WaitGroup
	var granted int32Counter
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if g.allow() {
				granted.add(1)
			}
		}()
	}
	wg.Wait()

	if got := granted.load(); got != 1 {
		t.Fatalf("expected exactly 1 of %d concurrent allow() calls to succeed, got %d", n, got)
	}
}

// int32Counter is a tiny race-safe counter for the test above, kept out of
// main.go since it exists only to make the concurrent test assertion safe
// under -race without importing sync/atomic twice for a test-only need.
type int32Counter struct {
	mu sync.Mutex
	n  int
}

func (c *int32Counter) add(d int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n += d
}

func (c *int32Counter) load() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// --- hostOnly: IPv4 unchanged, IPv6 masked to /64 ---

type fakeAddr string

func (f fakeAddr) Network() string { return "tcp" }
func (f fakeAddr) String() string  { return string(f) }

func TestHostOnlyIPv4KeepsFullAddress(t *testing.T) {
	got := hostOnly(fakeAddr("203.0.113.5:54321"))
	want := "203.0.113.5"
	if got != want {
		t.Fatalf("hostOnly(%q) = %q, want %q", "203.0.113.5:54321", got, want)
	}
}

func TestHostOnlyIPv6MasksToSlash64(t *testing.T) {
	got := hostOnly(fakeAddr("[2001:db8::1]:54321"))
	want := "2001:db8::"
	if got != want {
		t.Fatalf("hostOnly(%q) = %q, want %q", "[2001:db8::1]:54321", got, want)
	}
}

// TestHostOnlyIPv6SameSlash64CollidesOnPurpose pins the actual security
// property: two different IPv6 addresses within the same routed /64 must
// key identically, or the whole point of masking is lost.
func TestHostOnlyIPv6SameSlash64CollidesOnPurpose(t *testing.T) {
	a := hostOnly(fakeAddr("[2001:db8:abcd:1234::1]:1"))
	b := hostOnly(fakeAddr("[2001:db8:abcd:1234:ffff:ffff:ffff:ffff]:2"))
	if a != b {
		t.Fatalf("two addresses in the same /64 keyed differently: %q vs %q", a, b)
	}
}

func TestHostOnlyIPv6DifferentSlash64DoNotCollide(t *testing.T) {
	a := hostOnly(fakeAddr("[2001:db8:0000::1]:1"))
	b := hostOnly(fakeAddr("[2001:db8:0001::1]:1"))
	if a == b {
		t.Fatalf("two addresses in different /64s keyed the same: %q", a)
	}
}

func TestHostOnlyFallsBackWhenNoPort(t *testing.T) {
	got := hostOnly(fakeAddr("not-an-address"))
	want := "not-an-address"
	if got != want {
		t.Fatalf("hostOnly(%q) = %q, want %q", "not-an-address", got, want)
	}
}

// net.TCPAddr also satisfies net.Addr and is what the real server actually
// sees; confirm hostOnly behaves the same way against it.
func TestHostOnlyAgainstRealTCPAddr(t *testing.T) {
	addr := &net.TCPAddr{IP: net.ParseIP("2001:db8::abcd"), Port: 1234}
	got := hostOnly(addr)
	want := "2001:db8::"
	if got != want {
		t.Fatalf("hostOnly(%v) = %q, want %q", addr, got, want)
	}
}

// --- truncateUTF8: bounding the attacker-controlled username ---

func TestTruncateUTF8LeavesShortStringsAlone(t *testing.T) {
	if got := truncateUTF8("dfetz", 32); got != "dfetz" {
		t.Fatalf("got %q, want %q", got, "dfetz")
	}
}

func TestTruncateUTF8CutsLongStrings(t *testing.T) {
	long := make([]byte, 100)
	for i := range long {
		long[i] = 'a'
	}
	got := truncateUTF8(string(long), 32)
	if len(got) > 32+len("…") {
		t.Fatalf("truncated result too long: %d bytes: %q", len(got), got)
	}
	if got[len(got)-len("…"):] != "…" {
		t.Fatalf("truncated result missing marker: %q", got)
	}
}

func TestTruncateUTF8NeverSplitsARune(t *testing.T) {
	// Every rune here is multi-byte, and 32 does not fall on a rune
	// boundary if truncated naively byte-for-byte.
	s := ""
	for i := 0; i < 20; i++ {
		s += "€" // 3 bytes each
	}
	got := truncateUTF8(s, 32)
	for i, r := range got {
		if r == 0xFFFD && got[i:i+3] != "€" {
			t.Fatalf("truncation produced an invalid rune: %q", got)
		}
	}
}

// --- configProblems: refusing to silently disable a control ---

func validConfig() config {
	return config{
		addr:          ":23234",
		hostKey:       "./.ssh/fetzer_ed25519",
		idleTimeout:   5 * time.Minute,
		maxSession:    30 * time.Minute,
		maxConnsPerIP: 3,
		ratePerMin:    20,
	}
}

func TestConfigProblemsAcceptsValidConfig(t *testing.T) {
	if got := configProblems(validConfig()); len(got) != 0 {
		t.Fatalf("expected no problems, got %v", got)
	}
}

func TestConfigProblemsRejectsZeroIdleTimeout(t *testing.T) {
	cfg := validConfig()
	cfg.idleTimeout = 0
	if got := configProblems(cfg); len(got) != 1 {
		t.Fatalf("expected exactly 1 problem, got %v", got)
	}
}

func TestConfigProblemsRejectsZeroMaxSession(t *testing.T) {
	cfg := validConfig()
	cfg.maxSession = 0
	if got := configProblems(cfg); len(got) != 1 {
		t.Fatalf("expected exactly 1 problem, got %v", got)
	}
}

// TestConfigProblemsRejectsBothTimeoutsZero pins the specific failure mode
// called out in review: idle-timeout=0 AND max-session=0 together make the
// underlying library's deadline the zero time, clearing it entirely and
// holding connections open forever. Both must be flagged.
func TestConfigProblemsRejectsBothTimeoutsZero(t *testing.T) {
	cfg := validConfig()
	cfg.idleTimeout = 0
	cfg.maxSession = 0
	got := configProblems(cfg)
	if len(got) != 2 {
		t.Fatalf("expected exactly 2 problems, got %v", got)
	}
}

func TestConfigProblemsRejectsNegativeDurations(t *testing.T) {
	cfg := validConfig()
	cfg.idleTimeout = -5 * time.Minute
	cfg.maxSession = -1
	got := configProblems(cfg)
	if len(got) != 2 {
		t.Fatalf("expected exactly 2 problems, got %v", got)
	}
}

func TestConfigProblemsRejectsNonPositiveLimits(t *testing.T) {
	cfg := validConfig()
	cfg.maxConnsPerIP = 0
	cfg.ratePerMin = -1
	got := configProblems(cfg)
	if len(got) != 2 {
		t.Fatalf("expected exactly 2 problems, got %v", got)
	}
}

// --- envDuration / envInt: bad env values fall back, loudly ---

func TestEnvDurationUsesDefaultWhenUnset(t *testing.T) {
	key := "FETZER_TEST_UNSET_DURATION"
	os.Unsetenv(key)
	got := envDuration(key, 7*time.Second)
	if got != 7*time.Second {
		t.Fatalf("got %s, want 7s", got)
	}
}

func TestEnvDurationUsesValueWhenValidAndPositive(t *testing.T) {
	key := "FETZER_TEST_VALID_DURATION"
	t.Setenv(key, "42s")
	got := envDuration(key, 7*time.Second)
	if got != 42*time.Second {
		t.Fatalf("got %s, want 42s", got)
	}
}

func TestEnvDurationFallsBackOnUnparsable(t *testing.T) {
	key := "FETZER_TEST_BAD_DURATION"
	t.Setenv(key, "not-a-duration")
	got := envDuration(key, 7*time.Second)
	if got != 7*time.Second {
		t.Fatalf("got %s, want default 7s", got)
	}
}

func TestEnvDurationFallsBackOnZero(t *testing.T) {
	key := "FETZER_TEST_ZERO_DURATION"
	t.Setenv(key, "0s")
	got := envDuration(key, 7*time.Second)
	if got != 7*time.Second {
		t.Fatalf("got %s, want default 7s (zero must not pass through)", got)
	}
}

func TestEnvDurationFallsBackOnNegative(t *testing.T) {
	key := "FETZER_TEST_NEGATIVE_DURATION"
	t.Setenv(key, "-5m")
	got := envDuration(key, 7*time.Second)
	if got != 7*time.Second {
		t.Fatalf("got %s, want default 7s (negative must not pass through)", got)
	}
}

func TestEnvIntUsesDefaultWhenUnset(t *testing.T) {
	key := "FETZER_TEST_UNSET_INT"
	os.Unsetenv(key)
	if got := envInt(key, 9); got != 9 {
		t.Fatalf("got %d, want 9", got)
	}
}

func TestEnvIntFallsBackOnZero(t *testing.T) {
	key := "FETZER_TEST_ZERO_INT"
	t.Setenv(key, "0")
	if got := envInt(key, 9); got != 9 {
		t.Fatalf("got %d, want default 9 (zero must not pass through)", got)
	}
}

func TestEnvIntFallsBackOnNegative(t *testing.T) {
	key := "FETZER_TEST_NEGATIVE_INT"
	t.Setenv(key, "-3")
	if got := envInt(key, 9); got != 9 {
		t.Fatalf("got %d, want default 9 (negative must not pass through)", got)
	}
}

func TestEnvIntFallsBackOnUnparsable(t *testing.T) {
	key := "FETZER_TEST_BAD_INT"
	t.Setenv(key, "not-an-int")
	if got := envInt(key, 9); got != 9 {
		t.Fatalf("got %d, want default 9", got)
	}
}

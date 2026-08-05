// Package ratelimit provides per-IP concurrent connection and rate limiting
// for an unauthenticated network listener. It exists to stop a single
// source from exhausting server resources (file descriptors, memory) by
// opening unbounded connections or reconnecting in a tight loop.
package ratelimit

import (
	"sync"
	"time"
)

// window is the fixed rate-limit window: an IP may make at most
// ratePerMinute Acquire calls within any given window-sized bucket of time.
//
// Algorithm choice: fixed window counter, not a sliding window or token
// bucket. A fixed window is the simplest correct implementation and its
// worst case (a burst at the boundary between two windows allowing close to
// 2x ratePerMinute) is an acceptable trade for a personal site's abuse
// protection, where the goal is "stop unbounded abuse", not "meter traffic
// precisely". The concurrent-connection cap (enforced independently) bounds
// the damage any such boundary burst can do, since bursts still can't hold
// more than maxConcurrentPerIP connections open at once.
const window = time.Minute

// staleAfter is how long an IP's state may sit with zero active connections
// and no new activity before it is evicted from the map. It is set equal to
// the rate window: once a full window has elapsed with no activity, the rate
// counter would reset on next use anyway, so there is no information lost by
// dropping the entry, and no reason to keep it around.
const staleAfter = window

// ipState tracks per-IP concurrency and rate-limit bookkeeping.
type ipState struct {
	concurrent  int
	windowStart time.Time
	count       int
	lastSeen    time.Time
}

// Limiter enforces a per-IP concurrent connection cap and a per-IP rate
// limit (requests per minute, fixed window). It is safe for concurrent use.
type Limiter struct {
	mu sync.Mutex

	maxConcurrentPerIP int
	ratePerMinute      int
	clock              func() time.Time

	states map[string]*ipState
}

// New creates a Limiter that allows at most maxConcurrentPerIP simultaneous
// acquisitions and ratePerMinute acquisitions per fixed one-minute window,
// per IP.
func New(maxConcurrentPerIP int, ratePerMinute int) *Limiter {
	return &Limiter{
		maxConcurrentPerIP: maxConcurrentPerIP,
		ratePerMinute:      ratePerMinute,
		clock:              time.Now,
		states:             make(map[string]*ipState),
	}
}

// Acquire attempts to reserve a connection slot for ip. It returns
// ok == false if ip is already at its concurrent connection cap or has
// exceeded its rate limit for the current window; no slot is reserved in
// that case and release is nil.
//
// When ok is true, the caller must eventually call release to free the
// concurrent-connection slot. release is idempotent: calling it more than
// once has no additional effect.
func (l *Limiter) Acquire(ip string) (release func(), ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock()
	l.evictStaleLocked(now)

	s, exists := l.states[ip]
	if !exists {
		s = &ipState{windowStart: now}
		l.states[ip] = s
	}

	if now.Sub(s.windowStart) >= window {
		s.windowStart = now
		s.count = 0
	}

	if s.concurrent >= l.maxConcurrentPerIP {
		return nil, false
	}
	if s.count >= l.ratePerMinute {
		return nil, false
	}

	s.concurrent++
	s.count++
	s.lastSeen = now

	var once sync.Once
	release = func() {
		once.Do(func() {
			l.mu.Lock()
			defer l.mu.Unlock()
			s.concurrent--
			s.lastSeen = l.clock()
		})
	}

	return release, true
}

// evictStaleLocked removes entries for IPs with no active connections and no
// activity within staleAfter. It must be called with l.mu held.
//
// Without this, a naive per-IP map grows without bound as an attacker (or
// just churn of legitimate clients over time) cycles through source
// addresses, which is itself a slow denial-of-service against the server's
// memory. Eviction is swept opportunistically on every Acquire rather than
// via a background goroutine/ticker, so there is no extra lifecycle to
// manage and no eviction work happens while the limiter is idle.
func (l *Limiter) evictStaleLocked(now time.Time) {
	for ip, s := range l.states {
		if s.concurrent == 0 && now.Sub(s.lastSeen) > staleAfter {
			delete(l.states, ip)
		}
	}
}

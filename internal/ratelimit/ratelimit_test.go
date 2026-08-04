package ratelimit

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeClock lets tests move time deterministically without sleeping.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock(start time.Time) *fakeClock {
	return &fakeClock{now: start}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func TestConcurrentCapRejectsNPlus1(t *testing.T) {
	l := New(2, 100)
	l.clock = newFakeClock(time.Now()).Now

	r1, ok1 := l.Acquire("1.2.3.4")
	if !ok1 {
		t.Fatal("expected first acquire to succeed")
	}
	r2, ok2 := l.Acquire("1.2.3.4")
	if !ok2 {
		t.Fatal("expected second acquire to succeed")
	}
	_, ok3 := l.Acquire("1.2.3.4")
	if ok3 {
		t.Fatal("expected third acquire (over concurrent cap) to fail")
	}

	r1()
	r2()
}

func TestReleaseFreesSlot(t *testing.T) {
	l := New(1, 100)
	l.clock = newFakeClock(time.Now()).Now

	r1, ok1 := l.Acquire("5.6.7.8")
	if !ok1 {
		t.Fatal("expected first acquire to succeed")
	}
	if _, ok := l.Acquire("5.6.7.8"); ok {
		t.Fatal("expected second acquire to fail while first is held")
	}
	r1()
	if _, ok := l.Acquire("5.6.7.8"); !ok {
		t.Fatal("expected acquire to succeed after release")
	}
}

func TestReleaseIsIdempotent(t *testing.T) {
	l := New(1, 100)
	l.clock = newFakeClock(time.Now()).Now

	r1, ok1 := l.Acquire("9.9.9.9")
	if !ok1 {
		t.Fatal("expected first acquire to succeed")
	}
	r1()
	r1() // calling twice must not free two slots

	r2, ok2 := l.Acquire("9.9.9.9")
	if !ok2 {
		t.Fatal("expected acquire to succeed after single effective release")
	}
	// A second acquire should now fail since the cap is 1 and r2 holds the slot.
	if _, ok := l.Acquire("9.9.9.9"); ok {
		t.Fatal("double release must not have granted an extra slot")
	}
	r2()
}

func TestRateLimitRejectsBurst(t *testing.T) {
	l := New(100, 3)
	clock := newFakeClock(time.Now())
	l.clock = clock.Now

	ip := "10.0.0.1"
	for i := 0; i < 3; i++ {
		release, ok := l.Acquire(ip)
		if !ok {
			t.Fatalf("expected acquire %d to succeed within rate", i)
		}
		release()
	}

	if _, ok := l.Acquire(ip); ok {
		t.Fatal("expected 4th acquire within the same window to be rate limited")
	}
}

func TestRateLimitResetsAfterWindow(t *testing.T) {
	l := New(100, 2)
	clock := newFakeClock(time.Now())
	l.clock = clock.Now

	ip := "10.0.0.2"
	for i := 0; i < 2; i++ {
		release, ok := l.Acquire(ip)
		if !ok {
			t.Fatalf("expected acquire %d to succeed within rate", i)
		}
		release()
	}
	if _, ok := l.Acquire(ip); ok {
		t.Fatal("expected acquire to be rate limited before window advance")
	}

	clock.Advance(time.Minute + time.Second)

	release, ok := l.Acquire(ip)
	if !ok {
		t.Fatal("expected acquire to succeed after window advanced")
	}
	release()
}

func TestIndependentIPBudgets(t *testing.T) {
	l := New(1, 1)
	l.clock = newFakeClock(time.Now()).Now

	r1, ok1 := l.Acquire("1.1.1.1")
	if !ok1 {
		t.Fatal("expected first IP to succeed")
	}
	r2, ok2 := l.Acquire("2.2.2.2")
	if !ok2 {
		t.Fatal("expected second, independent IP to succeed")
	}
	r1()
	r2()
}

func TestConcurrentAcquireRelease(t *testing.T) {
	l := New(5, 1000)
	l.clock = newFakeClock(time.Now()).Now

	const goroutines = 50
	const itersPerG = 100
	var wg sync.WaitGroup
	var successes int64

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ip := "192.168.0.1"
			if id%2 == 0 {
				ip = "192.168.0.2"
			}
			for i := 0; i < itersPerG; i++ {
				release, ok := l.Acquire(ip)
				if ok {
					atomic.AddInt64(&successes, 1)
					release()
					release() // exercise idempotency under race too
				}
			}
		}(g)
	}
	wg.Wait()

	if atomic.LoadInt64(&successes) == 0 {
		t.Fatal("expected at least some successful acquisitions under concurrency")
	}
}

// TestEvictionRemovesStaleEntries pins the eviction behaviour: once an IP has
// no active connections and its rate window has fully elapsed with no new
// activity, its entry must be removed from the internal map so that an
// attacker cycling through source addresses cannot grow it without bound.
func TestEvictionRemovesStaleEntries(t *testing.T) {
	l := New(2, 5)
	clock := newFakeClock(time.Now())
	l.clock = clock.Now

	ip := "172.16.0.1"
	release, ok := l.Acquire(ip)
	if !ok {
		t.Fatal("expected acquire to succeed")
	}
	release()

	l.mu.Lock()
	if _, present := l.states[ip]; !present {
		l.mu.Unlock()
		t.Fatal("expected state to be present immediately after use")
	}
	l.mu.Unlock()

	// Advance well past the eviction threshold and trigger a sweep via any
	// Acquire call (from an unrelated IP so it doesn't resurrect ip's entry).
	clock.Advance(10 * time.Minute)
	other, ok := l.Acquire("172.16.0.2")
	if !ok {
		t.Fatal("expected unrelated acquire to succeed")
	}
	other()

	l.mu.Lock()
	defer l.mu.Unlock()
	if _, present := l.states[ip]; present {
		t.Fatal("expected stale IP entry to be evicted")
	}
}

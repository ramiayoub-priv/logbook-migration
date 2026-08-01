package ratelimit_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/ratelimit"
)

// clock is a hand-wound clock. Rate limiting is entirely about the passage of
// time, and a test that actually slept would be both slow and flaky.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func newClock() *clock {
	return &clock{t: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
}

func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

func newLimiter(c *clock) *ratelimit.Limiter {
	return ratelimit.New(ratelimit.Config{Clock: c.now})
}

// TestFreeAttemptsThenThrottled is the control from app/docs/security.md: the
// Nth failed login is throttled. A few free attempts because the owner will
// fat-finger his own password on a phone in a cockpit, and it must not lock him
// out of his own logbook for a typo.
func TestFreeAttemptsThenThrottled(t *testing.T) {
	c := newClock()
	l := newLimiter(c)

	for i := range ratelimit.DefaultThreshold {
		if ok, _ := l.Allow("rami"); !ok {
			t.Fatalf("attempt %d was throttled; the first %d must be free",
				i+1, ratelimit.DefaultThreshold)
		}
		l.Fail("rami")
	}

	ok, retry := l.Allow("rami")
	if ok {
		t.Fatalf("attempt %d was allowed; it must be throttled", ratelimit.DefaultThreshold+1)
	}
	if retry <= 0 {
		t.Errorf("retry-after = %v, want a positive wait the caller can report", retry)
	}
}

// TestBackoffIsExponentialAndCapped. Doubling makes a sustained guessing run
// cost more than it can pay; the cap stops a burst from locking the real user
// out for a week.
func TestBackoffIsExponentialAndCapped(t *testing.T) {
	c := newClock()
	l := newLimiter(c)

	for range ratelimit.DefaultThreshold {
		l.Fail("rami")
	}

	var waits []time.Duration
	for range 12 {
		_, retry := l.Allow("rami")
		waits = append(waits, retry)
		// Sit out the penalty, then fail once more.
		c.advance(retry)
		if ok, _ := l.Allow("rami"); !ok {
			t.Fatalf("still throttled after waiting the full retry-after of %v", retry)
		}
		l.Fail("rami")
	}

	if waits[0] != ratelimit.DefaultBase {
		t.Errorf("the first penalty is %v, want the base of %v", waits[0], ratelimit.DefaultBase)
	}
	for i := 1; i < len(waits); i++ {
		switch {
		case waits[i-1] >= ratelimit.DefaultMax:
			if waits[i] != ratelimit.DefaultMax {
				t.Errorf("penalty %d is %v, want it pinned at the cap %v",
					i, waits[i], ratelimit.DefaultMax)
			}
		case waits[i] != min(waits[i-1]*2, ratelimit.DefaultMax):
			t.Errorf("penalty %d is %v, want %v (double the previous %v, capped)",
				i, waits[i], min(waits[i-1]*2, ratelimit.DefaultMax), waits[i-1])
		}
	}
	if waits[len(waits)-1] != ratelimit.DefaultMax {
		t.Errorf("the penalty never reached the cap; last was %v", waits[len(waits)-1])
	}
}

// TestWaitingOutThePenaltyWorks. The lockout must actually end, or a
// rate limiter becomes a denial of service against the only user.
func TestWaitingOutThePenaltyWorks(t *testing.T) {
	c := newClock()
	l := newLimiter(c)
	for range ratelimit.DefaultThreshold {
		l.Fail("rami")
	}

	_, retry := l.Allow("rami")
	c.advance(retry - time.Nanosecond)
	if ok, _ := l.Allow("rami"); ok {
		t.Error("allowed one nanosecond before the penalty expired")
	}
	c.advance(time.Nanosecond)
	if ok, _ := l.Allow("rami"); !ok {
		t.Error("still throttled after the penalty expired")
	}
}

// TestSuccessResets is the second half of the security.md control: a successful
// login clears the counter, so yesterday's typos do not accumulate into a
// lockout.
func TestSuccessResets(t *testing.T) {
	c := newClock()
	l := newLimiter(c)
	for range ratelimit.DefaultThreshold {
		l.Fail("rami")
	}
	if ok, _ := l.Allow("rami"); ok {
		t.Fatal("should be throttled before the reset")
	}

	l.Succeed("rami")

	if ok, retry := l.Allow("rami"); !ok {
		t.Errorf("still throttled after a successful login (retry %v)", retry)
	}
	// And the backoff restarts from the base rather than resuming where it was.
	for range ratelimit.DefaultThreshold {
		l.Fail("rami")
	}
	if _, retry := l.Allow("rami"); retry != ratelimit.DefaultBase {
		t.Errorf("penalty after a reset is %v, want the base %v", retry, ratelimit.DefaultBase)
	}
}

// TestKeysAreIndependent is what makes per-IP and per-account limiting two
// separate defences rather than one. Throttling one account must not lock out
// another, and one hostile IP must not lock out the owner's phone.
func TestKeysAreIndependent(t *testing.T) {
	c := newClock()
	l := newLimiter(c)
	for range ratelimit.DefaultThreshold + 3 {
		l.Fail("ip:198.51.100.7")
	}
	if ok, _ := l.Allow("ip:198.51.100.7"); ok {
		t.Fatal("the hostile key should be throttled")
	}
	if ok, _ := l.Allow("ip:192.0.2.1"); !ok {
		t.Error("a different IP was caught by another IP's penalty")
	}
	if ok, _ := l.Allow("user:rami"); !ok {
		t.Error("an account was caught by an IP's penalty")
	}
}

// TestIdleKeysAreForgotten keeps the table from being a memory leak, and means
// a single bad day does not follow the user around forever.
func TestIdleKeysAreForgotten(t *testing.T) {
	c := newClock()
	l := newLimiter(c)
	for range ratelimit.DefaultThreshold {
		l.Fail("rami")
	}
	if l.Len() != 1 {
		t.Fatalf("tracking %d keys, want 1", l.Len())
	}

	c.advance(ratelimit.DefaultForget + time.Second)
	if ok, _ := l.Allow("rami"); !ok {
		t.Error("a key idle past the forget window is still penalised")
	}
	l.Prune()
	if l.Len() != 0 {
		t.Errorf("tracking %d keys after the forget window, want 0", l.Len())
	}
}

// TestTheTableIsBoundedAndEvictsTheStalest. An attacker rotating source
// addresses must not be able to grow this map without limit -- that is a way to
// exhaust the memory of a box shared with the owner's other sites (rule 0.3).
//
// Eviction takes the least recently active entry, which is deliberately the one
// an attacker cannot arrange to be their own: the key they are hammering was
// just touched, so it is the newest. Flooding the table cannot flush your own
// penalty.
func TestTheTableIsBoundedAndEvictsTheStalest(t *testing.T) {
	c := newClock()
	l := ratelimit.New(ratelimit.Config{Clock: c.now, MaxEntries: 8})

	// The key under attack, failed first and then kept warm.
	l.Fail("victim")

	for i := range 50 {
		c.advance(time.Second)
		l.Fail(fmt.Sprintf("throwaway-%d", i))
		// Keep the victim's key the most recently touched.
		c.advance(time.Second)
		l.Fail("victim")
	}

	if l.Len() > 8 {
		t.Errorf("tracking %d keys with MaxEntries=8", l.Len())
	}
	if ok, _ := l.Allow("victim"); ok {
		t.Error("flooding the table with fresh keys cleared the attacked key's penalty")
	}
}

// TestPruneDropsIdleKeysWithoutBeingAsked covers the periodic sweep the server
// runs. TestIdleKeysAreForgotten only proves the lazy path in Allow; if nobody
// ever attempts that key again, Prune is the only thing that reclaims it.
func TestPruneDropsIdleKeysWithoutBeingAsked(t *testing.T) {
	c := newClock()
	l := newLimiter(c)
	for _, k := range []string{"a", "b", "c"} {
		l.Fail(k)
	}
	c.advance(ratelimit.DefaultForget + time.Second)
	l.Fail("fresh")

	l.Prune()

	if l.Len() != 1 {
		t.Errorf("tracking %d keys after the prune, want only the fresh one", l.Len())
	}
}

// TestAKeyThatAgesOutBetweenAttemptsStartsOver. Failing, disappearing for an
// hour and failing again must not resume an old escalation -- otherwise a
// penalty from another era ambushes the user weeks later.
func TestAKeyThatAgesOutBetweenAttemptsStartsOver(t *testing.T) {
	c := newClock()
	l := newLimiter(c)
	for range ratelimit.DefaultThreshold + 5 {
		l.Fail("rami")
	}

	c.advance(ratelimit.DefaultForget + time.Second)

	// The count restarts, so the first few failures are free again.
	for range ratelimit.DefaultThreshold - 1 {
		l.Fail("rami")
		if ok, _ := l.Allow("rami"); !ok {
			t.Fatal("the escalation resumed instead of starting over")
		}
	}
	l.Fail("rami")
	if _, retry := l.Allow("rami"); retry != ratelimit.DefaultBase {
		t.Errorf("penalty is %v, want the base %v", retry, ratelimit.DefaultBase)
	}
}

// TestTheBackoffShiftCannotOverflow. A time.Duration is an int64 of
// nanoseconds; at 63 doublings the shift wraps to a negative number, which
// would read as "no penalty" -- the precise inversion of this package's job.
// Sustained failure must stay pinned at the cap however long it runs.
func TestTheBackoffShiftCannotOverflow(t *testing.T) {
	c := newClock()
	l := newLimiter(c)
	for range 100 {
		l.Fail("rami")
	}
	ok, retry := l.Allow("rami")
	if ok {
		t.Fatal("100 consecutive failures left the key unthrottled")
	}
	if retry != ratelimit.DefaultMax {
		t.Errorf("penalty after 100 failures is %v, want the cap %v", retry, ratelimit.DefaultMax)
	}
}

func TestZeroConfigUsesTheDefaults(t *testing.T) {
	// The server constructs this with no configuration; the defaults must be
	// the documented ones rather than a zero-value limiter that lets
	// everything through.
	l := ratelimit.New(ratelimit.Config{})
	for range ratelimit.DefaultThreshold {
		l.Fail("rami")
	}
	if ok, _ := l.Allow("rami"); ok {
		t.Error("a default-constructed limiter does not throttle")
	}
}

// TestConcurrentUseIsSafe. Every request touches this, and `go test -race`
// is the only thing that will ever prove the locking is right.
func TestConcurrentUseIsSafe(t *testing.T) {
	c := newClock()
	l := newLimiter(c)

	var wg sync.WaitGroup
	for i := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i%4)
			for range 50 {
				l.Allow(key)
				l.Fail(key)
				l.Succeed(key)
				l.Len()
			}
		}()
	}
	wg.Wait()
}

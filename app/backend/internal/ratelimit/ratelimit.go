// Package ratelimit throttles repeated failures against a key, with
// exponential backoff. It exists for exactly one caller: the login endpoint,
// where app/docs/security.md requires per-IP and per-account limiting.
//
// It is in-memory rather than in SQLite. The server is a single process, the
// state is worthless across restarts, and writing a row per failed login would
// let an attacker drive disk I/O for free. The trade-off is that a restart
// clears the penalties -- acceptable, because an attacker cannot cause one.
//
// Two properties matter more than the throttling itself:
//
//   - The table is bounded. An attacker rotating source addresses must not be
//     able to grow it without limit; that would be a memory-exhaustion lever
//     against a box shared with the owner's other sites (CLAUDE.md rule 0.3).
//   - Eviction takes the least recently active key, which is never the key an
//     attacker is hammering -- theirs was just touched. Flooding the table
//     cannot flush your own penalty.
package ratelimit

import (
	"sync"
	"time"
)

// The defaults, all reachable from the tests so an assertion cannot drift from
// the value it is asserting about.
const (
	// DefaultThreshold is how many failures are free. The owner types his
	// password on a phone, sometimes in a cockpit; locking him out of his own
	// logbook over a typo is a worse outcome than three cheap guesses.
	DefaultThreshold = 5
	// DefaultBase is the first penalty, doubling with each further failure.
	DefaultBase = 5 * time.Second
	// DefaultMax caps the penalty. Without a cap a burst of failures locks the
	// real user out for days, which turns the defence into the attack.
	DefaultMax = 15 * time.Minute
	// DefaultForget is how long an idle key keeps its record.
	DefaultForget = time.Hour
	// DefaultMaxEntries bounds the table. Far above what one user generates
	// and far below what would matter on a 2 GB box.
	DefaultMaxEntries = 4096
)

// Config parameterises a Limiter. The zero value gives the defaults above,
// which is what the server uses.
type Config struct {
	Threshold  int
	Base       time.Duration
	Max        time.Duration
	Forget     time.Duration
	MaxEntries int
	// Clock is the time source, replaced in tests. Defaults to time.Now.
	Clock func() time.Time
}

type entry struct {
	failures int
	// lastSeen drives both forgetting and eviction.
	lastSeen time.Time
	// until is when the current penalty ends; zero if there is none.
	until time.Time
}

// Limiter tracks failures per key. It is safe for concurrent use: every
// authenticated request may touch it.
type Limiter struct {
	cfg Config

	mu      sync.Mutex
	entries map[string]*entry
}

// New builds a Limiter, filling in the defaults for any zero field.
func New(cfg Config) *Limiter {
	if cfg.Threshold <= 0 {
		cfg.Threshold = DefaultThreshold
	}
	if cfg.Base <= 0 {
		cfg.Base = DefaultBase
	}
	if cfg.Max <= 0 {
		cfg.Max = DefaultMax
	}
	if cfg.Forget <= 0 {
		cfg.Forget = DefaultForget
	}
	if cfg.MaxEntries <= 0 {
		cfg.MaxEntries = DefaultMaxEntries
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	return &Limiter{cfg: cfg, entries: map[string]*entry{}}
}

// Allow reports whether an attempt against key may proceed. When it may not, it
// returns how long the caller must wait, which the login handler passes on as a
// Retry-After header.
//
// Allow does not itself count anything. Counting is Fail's job, so an attempt
// that succeeds costs the key nothing.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	now := l.cfg.Clock()

	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[key]
	if !ok {
		return true, 0
	}
	// An idle key is forgotten rather than carried forever: a bad day should
	// not follow the user around.
	if now.Sub(e.lastSeen) >= l.cfg.Forget {
		delete(l.entries, key)
		return true, 0
	}
	if wait := e.until.Sub(now); wait > 0 {
		return false, wait
	}
	return true, 0
}

// Fail records a failed attempt and extends the penalty.
//
// The penalty starts only after Threshold failures, then doubles each time up
// to Max. It is computed from the failure count rather than from the previous
// penalty, so waiting one out and failing again resumes the escalation instead
// of restarting it.
func (l *Limiter) Fail(key string) {
	now := l.cfg.Clock()

	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[key]
	if !ok {
		l.evictIfFullLocked(now)
		e = &entry{}
		l.entries[key] = e
	} else if now.Sub(e.lastSeen) >= l.cfg.Forget {
		// The record aged out between attempts; start again rather than
		// resuming a penalty from another era.
		*e = entry{}
	}

	e.failures++
	e.lastSeen = now
	if e.failures >= l.cfg.Threshold {
		e.until = now.Add(l.penalty(e.failures))
	}
}

// Succeed clears a key. This is the reset half of the security.md control: a
// correct password wipes the record, so accumulated typos cannot add up to a
// lockout later.
func (l *Limiter) Succeed(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

// Len is how many keys are being tracked. Used by the tests to prove the table
// stays bounded, and available for a metric.
func (l *Limiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}

// Prune drops every key idle past the forget window. The server calls it on the
// same timer as the expired-session sweep.
func (l *Limiter) Prune() {
	now := l.cfg.Clock()

	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked(now)
}

// penalty is Base doubled once per failure past the threshold, capped at Max.
// Threshold failures earn Base; each one after that doubles it.
//
// The shift is bounded before it is taken: at 63 doublings a time.Duration
// overflows into a negative number, which would read as "no penalty" -- the
// exact inversion of what this function is for.
// Fail is the only caller and only calls it once failures has reached the
// threshold, so n is never negative.
func (l *Limiter) penalty(failures int) time.Duration {
	n := failures - l.cfg.Threshold
	if n > 62 {
		return l.cfg.Max
	}
	d := l.cfg.Base << uint(n)
	if d <= 0 || d > l.cfg.Max {
		return l.cfg.Max
	}
	return d
}

func (l *Limiter) pruneLocked(now time.Time) {
	for k, e := range l.entries {
		if now.Sub(e.lastSeen) >= l.cfg.Forget {
			delete(l.entries, k)
		}
	}
}

// evictIfFullLocked makes room for a new key. It prunes the expired first, and
// only if that is not enough does it evict the least recently active entry.
//
// Least-recently-active is the safe choice: the key an attacker is hammering is
// by definition the most recently touched, so no amount of flooding lets them
// evict their own penalty.
func (l *Limiter) evictIfFullLocked(now time.Time) {
	if len(l.entries) < l.cfg.MaxEntries {
		return
	}
	l.pruneLocked(now)
	for len(l.entries) >= l.cfg.MaxEntries {
		var oldestKey string
		var oldest time.Time
		for k, e := range l.entries {
			if oldestKey == "" || e.lastSeen.Before(oldest) {
				oldestKey, oldest = k, e.lastSeen
			}
		}
		delete(l.entries, oldestKey)
	}
}

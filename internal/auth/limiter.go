package auth

import (
	"sync"
	"time"
)

const (
	// LoginAttemptLimit is how many failed logins one client may make inside
	// LoginAttemptWindow before the endpoint answers 429.
	LoginAttemptLimit = 10

	// LoginAttemptWindow is the sliding window those attempts are counted in.
	LoginAttemptWindow = 5 * time.Minute

	// maxLimiterEntries keeps a burst from many addresses from growing the map
	// without bound; expired buckets are swept once it is reached.
	maxLimiterEntries = 4096
)

// Limiter counts failed logins per client. It lives in memory on purpose: with
// one shared credential the real risk is guessing, and forgetting the counts on
// restart costs an attacker nothing they cannot regain anyway.
type Limiter struct {
	mu      sync.Mutex
	entries map[string][]time.Time
	limit   int
	window  time.Duration
	now     func() time.Time
}

// NewLimiter builds a limiter. A non-positive limit disables it, matching how
// the other limits in this codebase are switched off.
func NewLimiter(limit int, window time.Duration) *Limiter {
	return &Limiter{
		entries: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
		now:     time.Now,
	}
}

// Allowed reports whether key may try to log in again.
func (l *Limiter) Allowed(key string) bool {
	if l.limit <= 0 {
		return true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	return len(l.pruneLocked(key)) < l.limit
}

// Fail records one failed attempt for key.
func (l *Limiter) Fail(key string) {
	if l.limit <= 0 {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	attempts := append(l.pruneLocked(key), l.now())
	// Keep only the newest limit attempts: anything older cannot change the
	// answer and would only grow the bucket.
	if len(attempts) > l.limit {
		attempts = attempts[len(attempts)-l.limit:]
	}
	l.entries[key] = attempts

	l.sweepLocked()
}

// Reset forgets the failures of key: a successful login proves the client is
// not guessing.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.entries, key)
}

// pruneLocked drops attempts that fell out of the window and returns what is
// left. The caller must hold the lock.
func (l *Limiter) pruneLocked(key string) []time.Time {
	attempts, ok := l.entries[key]
	if !ok {
		return nil
	}

	cutoff := l.now().Add(-l.window)
	kept := make([]time.Time, 0, len(attempts))
	for _, at := range attempts {
		if at.After(cutoff) {
			kept = append(kept, at)
		}
	}

	if len(kept) == 0 {
		delete(l.entries, key)
		return nil
	}

	l.entries[key] = kept
	return kept
}

// sweepLocked removes buckets that have gone quiet. It only runs once the map
// is already large, so the normal path stays O(1).
func (l *Limiter) sweepLocked() {
	if len(l.entries) < maxLimiterEntries {
		return
	}

	cutoff := l.now().Add(-l.window)
	for key, attempts := range l.entries {
		if len(attempts) == 0 || attempts[len(attempts)-1].Before(cutoff) {
			delete(l.entries, key)
		}
	}
}

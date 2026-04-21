package ratelimit

import (
	"sync"
	"time"
)

// Limiter suppresses repeated alerts for the same port within a cooldown window.
type Limiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[string]time.Time
	now      func() time.Time
}

// New creates a Limiter with the given cooldown duration.
func New(cooldown time.Duration) *Limiter {
	return &Limiter{
		cooldown: cooldown,
		last:     make(map[string]time.Time),
		now:      time.Now,
	}
}

// Allow returns true if an alert for the given key should be emitted.
// It returns false if the same key was allowed within the cooldown window.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	t, seen := l.last[key]
	if !seen || l.now().Sub(t) >= l.cooldown {
		l.last[key] = l.now()
		return true
	}
	return false
}

// Reset clears the rate-limit state for a specific key.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.last, key)
}

// ResetAll clears all rate-limit state.
func (l *Limiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.last = make(map[string]time.Time)
}

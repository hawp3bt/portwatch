// Package suppress provides a mechanism to temporarily silence alerts
// for specific ports during maintenance windows or known activity.
package suppress

import (
	"sync"
	"time"
)

// Entry holds suppression metadata for a single port.
type Entry struct {
	Port      int
	Reason    string
	ExpiresAt time.Time
}

// Suppressor tracks which ports are currently suppressed.
type Suppressor struct {
	mu      sync.Mutex
	entries map[int]Entry
}

// New returns an initialised Suppressor.
func New() *Suppressor {
	return &Suppressor{
		entries: make(map[int]Entry),
	}
}

// Suppress silences alerts for port for the given duration.
func (s *Suppressor) Suppress(port int, reason string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[port] = Entry{
		Port:      port,
		Reason:    reason,
		ExpiresAt: time.Now().Add(duration),
	}
}

// IsSuppressed reports whether port is currently suppressed.
// Expired entries are removed lazily on access.
func (s *Suppressor) IsSuppressed(port int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[port]
	if !ok {
		return false
	}
	if time.Now().After(e.ExpiresAt) {
		delete(s.entries, port)
		return false
	}
	return true
}

// Lift removes a suppression entry for port before it expires.
func (s *Suppressor) Lift(port int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, port)
}

// List returns all currently active suppression entries.
func (s *Suppressor) List() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	out := make([]Entry, 0, len(s.entries))
	for port, e := range s.entries {
		if now.After(e.ExpiresAt) {
			delete(s.entries, port)
			continue
		}
		out = append(out, e)
	}
	return out
}

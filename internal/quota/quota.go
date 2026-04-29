// Package quota enforces per-port alert frequency limits, preventing
// notification floods when a port flaps rapidly.
package quota

import (
	"sync"
	"time"
)

// Entry tracks how many alerts have been emitted for a single port within
// the current window.
type Entry struct {
	Count     int
	WindowEnd time.Time
}

// Quota enforces a maximum number of alerts per port per time window.
type Quota struct {
	mu      sync.Mutex
	entries map[int]*Entry
	max     int
	window  time.Duration
	now     func() time.Time
}

// New creates a Quota that allows at most maxAlerts notifications per port
// within the given window duration.
func New(maxAlerts int, window time.Duration) *Quota {
	if maxAlerts < 1 {
		maxAlerts = 1
	}
	if window < time.Second {
		window = time.Second
	}
	return &Quota{
		entries: make(map[int]*Entry),
		max:     maxAlerts,
		window:  window,
		now:     time.Now,
	}
}

// Allow returns true if an alert for the given port is permitted under the
// current quota. It increments the counter when permitted.
func (q *Quota) Allow(port int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.now()
	e, ok := q.entries[port]
	if !ok || now.After(e.WindowEnd) {
		q.entries[port] = &Entry{Count: 1, WindowEnd: now.Add(q.window)}
		return true
	}
	if e.Count >= q.max {
		return false
	}
	e.Count++
	return true
}

// Reset clears quota state for a specific port.
func (q *Quota) Reset(port int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.entries, port)
}

// Remaining returns how many alerts are still permitted for the given port
// in the current window. Returns max if no window is active.
func (q *Quota) Remaining(port int) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	e, ok := q.entries[port]
	if !ok || q.now().After(e.WindowEnd) {
		return q.max
	}
	remaining := q.max - e.Count
	if remaining < 0 {
		return 0
	}
	return remaining
}

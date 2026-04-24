// Package debounce provides a mechanism to suppress rapid repeated port
// change events, emitting only after a quiet period has elapsed.
package debounce

import (
	"sync"
	"time"
)

// Event represents a pending debounced port event.
type Event struct {
	Port   int
	Opened bool
}

// Debouncer holds pending events and fires them after a configurable delay.
type Debouncer struct {
	mu      sync.Mutex
	delay   time.Duration
	pending map[int]*time.Timer
	output  chan Event
}

// New creates a Debouncer that waits delay before forwarding an event.
// The returned channel receives de-duplicated events.
func New(delay time.Duration) (*Debouncer, <-chan Event) {
	ch := make(chan Event, 64)
	return &Debouncer{
		delay:   delay,
		pending: make(map[int]*time.Timer),
		output:  ch,
	}, ch
}

// Push schedules an event for port. If an event for that port is already
// pending, the timer is reset so the quiet window restarts.
func (d *Debouncer) Push(port int, opened bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if t, exists := d.pending[port]; exists {
		t.Stop()
	}

	event := Event{Port: port, Opened: opened}
	d.pending[port] = time.AfterFunc(d.delay, func() {
		d.mu.Lock()
		delete(d.pending, port)
		d.mu.Unlock()
		d.output <- event
	})
}

// Flush cancels all pending timers and emits them immediately. Useful on
// shutdown to avoid losing events.
func (d *Debouncer) Flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for port, t := range d.pending {
		t.Stop()
		e := Event{Port: port}
		delete(d.pending, port)
		d.output <- e
	}
}

// PendingCount returns the number of events currently waiting to fire.
func (d *Debouncer) PendingCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.pending)
}

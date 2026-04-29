// Package rollup batches multiple port-change events into a single
// summary notification, reducing alert noise during large network shifts.
package rollup

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Event represents a single port-change event to be rolled up.
type Event struct {
	Port   int
	Opened bool // true = opened, false = closed
}

// Summary is the result of flushing a rollup window.
type Summary struct {
	Opened []int
	Closed []int
	At     time.Time
}

// String returns a human-readable one-liner for the summary.
func (s Summary) String() string {
	parts := []string{}
	if len(s.Opened) > 0 {
		parts = append(parts, fmt.Sprintf("%d opened", len(s.Opened)))
	}
	if len(s.Closed) > 0 {
		parts = append(parts, fmt.Sprintf("%d closed", len(s.Closed)))
	}
	if len(parts) == 0 {
		return "no changes"
	}
	return strings.Join(parts, ", ")
}

// Rollup collects events within a time window and emits summaries.
type Rollup struct {
	mu      sync.Mutex
	opened  map[int]struct{}
	closed  map[int]struct{}
	window  time.Duration
	timer   *time.Timer
	emit    func(Summary)
}

// New creates a Rollup that flushes accumulated events after window
// duration of inactivity and calls emit with the resulting Summary.
func New(window time.Duration, emit func(Summary)) *Rollup {
	if window < time.Second {
		window = time.Second
	}
	return &Rollup{
		opened: make(map[int]struct{}),
		closed: make(map[int]struct{}),
		window: window,
		emit:   emit,
	}
}

// Push adds an event to the current window, resetting the flush timer.
func (r *Rollup) Push(e Event) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if e.Opened {
		r.opened[e.Port] = struct{}{}
		delete(r.closed, e.Port)
	} else {
		r.closed[e.Port] = struct{}{}
		delete(r.opened, e.Port)
	}

	if r.timer != nil {
		r.timer.Reset(r.window)
	} else {
		r.timer = time.AfterFunc(r.window, r.flush)
	}
}

// Flush emits any pending events immediately, regardless of the window.
func (r *Rollup) Flush() {
	if r.timer != nil {
		r.timer.Stop()
	}
	r.flush()
}

func (r *Rollup) flush() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.opened) == 0 && len(r.closed) == 0 {
		r.timer = nil
		return
	}

	s := Summary{At: time.Now()}
	for p := range r.opened {
		s.Opened = append(s.Opened, p)
	}
	for p := range r.closed {
		s.Closed = append(s.Closed, p)
	}

	r.opened = make(map[int]struct{})
	r.closed = make(map[int]struct{})
	r.timer = nil

	go r.emit(s)
}

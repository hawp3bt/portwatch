package schedule

import (
	"fmt"
	"sync"
	"time"
)

// Entry represents a named scan schedule with a custom interval.
type Entry struct {
	Name     string        `json:"name"`
	Interval time.Duration `json:"interval"`
	Enabled  bool          `json:"enabled"`
}

// Schedule manages a collection of named scan intervals.
type Schedule struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// New returns a new Schedule with no entries.
func New() *Schedule {
	return &Schedule{
		entries: make(map[string]*Entry),
	}
}

// Set adds or replaces a named schedule entry.
func (s *Schedule) Set(name string, interval time.Duration, enabled bool) error {
	if name == "" {
		return fmt.Errorf("schedule name must not be empty")
	}
	if interval < time.Second {
		return fmt.Errorf("interval must be at least 1s, got %s", interval)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[name] = &Entry{Name: name, Interval: interval, Enabled: enabled}
	return nil
}

// Get returns the entry for the given name, or false if not found.
func (s *Schedule) Get(name string) (*Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[name]
	return e, ok
}

// Remove deletes a schedule entry by name.
func (s *Schedule) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, name)
}

// List returns all entries sorted by name.
func (s *Schedule) List() []*Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}

// ActiveIntervals returns intervals for all enabled entries.
func (s *Schedule) ActiveIntervals() []time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var intervals []time.Duration
	for _, e := range s.entries {
		if e.Enabled {
			intervals = append(intervals, e.Interval)
		}
	}
	return intervals
}

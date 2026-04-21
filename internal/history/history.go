package history

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// EventKind describes whether a port was opened or closed.
type EventKind string

const (
	Opened EventKind = "opened"
	Closed EventKind = "closed"
)

// Event records a single port change detected by the monitor.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Port      int       `json:"port"`
	Kind      EventKind `json:"kind"`
}

// History stores a bounded list of port-change events and persists them to disk.
type History struct {
	mu     sync.Mutex
	events []Event
	path   string
	limit  int
}

// New creates a History that persists to path and keeps at most limit events.
func New(path string, limit int) *History {
	return &History{path: path, limit: limit}
}

// Record appends a new event and persists the history to disk.
func (h *History) Record(port int, kind EventKind) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.events = append(h.events, Event{
		Timestamp: time.Now().UTC(),
		Port:      port,
		Kind:      kind,
	})

	if h.limit > 0 && len(h.events) > h.limit {
		h.events = h.events[len(h.events)-h.limit:]
	}

	return h.save()
}

// All returns a copy of all recorded events.
func (h *History) All() []Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make([]Event, len(h.events))
	copy(result, h.events)
	return result
}

// Load reads previously persisted events from disk.
func (h *History) Load() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	data, err := os.ReadFile(h.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &h.events)
}

func (h *History) save() error {
	data, err := json.MarshalIndent(h.events, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(h.path, data, 0o644)
}

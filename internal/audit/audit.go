package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// EventKind classifies the type of audit event.
type EventKind string

const (
	KindScan      EventKind = "scan"
	KindAlert     EventKind = "alert"
	KindSuppress  EventKind = "suppress"
	KindBaseline  EventKind = "baseline"
	KindProfile   EventKind = "profile"
	KindThreshold EventKind = "threshold"
)

// Entry is a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Kind      EventKind `json:"kind"`
	Actor     string    `json:"actor"`
	Message   string    `json:"message"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// Logger appends audit entries to a file.
type Logger struct {
	mu   sync.Mutex
	path string
}

// New returns a Logger that writes to path.
func New(path string) *Logger {
	return &Logger{path: path}
}

// Record appends an entry to the audit log.
func (l *Logger) Record(kind EventKind, actor, message string, meta map[string]string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	e := Entry{
		Timestamp: time.Now().UTC(),
		Kind:      kind,
		Actor:     actor,
		Message:   message,
		Meta:      meta,
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("audit: open %s: %w", l.path, err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(e); err != nil {
		return fmt.Errorf("audit: encode: %w", err)
	}
	return nil
}

// Load reads all entries from the audit log file.
func Load(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return fmt.Errorf("audit: open %s: %w", path, err), nil
	}
	defer f.Close()

	var entries []Entry
	dec := json.NewDecoder(f)
	for dec.More() {
		var e Entry
		if err := dec.Decode(&e); err != nil {
			return nil, fmt.Errorf("audit: decode: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

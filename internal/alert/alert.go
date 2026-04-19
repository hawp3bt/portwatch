package alert

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelAlert Level = "ALERT"
)

// Event describes a port change event.
type Event struct {
	Timestamp time.Time
	Level     Level
	Port      int
	Message   string
}

// Notifier sends alerts for port change events.
type Notifier struct {
	out io.Writer
}

// New creates a Notifier that writes to the given writer.
// Pass nil to use os.Stdout.
func New(w io.Writer) *Notifier {
	if w == nil {
		w = os.Stdout
	}
	return &Notifier{out: w}
}

// PortOpened emits an alert for a newly opened port.
func (n *Notifier) PortOpened(port int) {
	n.emit(Event{
		Timestamp: time.Now(),
		Level:     LevelAlert,
		Port:      port,
		Message:   fmt.Sprintf("port %d opened unexpectedly", port),
	})
}

// PortClosed emits an alert for a port that disappeared.
func (n *Notifier) PortClosed(port int) {
	n.emit(Event{
		Timestamp: time.Now(),
		Level:     LevelWarn,
		Port:      port,
		Message:   fmt.Sprintf("port %d closed unexpectedly", port),
	})
}

func (n *Notifier) emit(e Event) {
	fmt.Fprintf(n.out, "[%s] %s | port=%d msg=%q\n",
		e.Timestamp.Format(time.RFC3339),
		e.Level,
		e.Port,
		e.Message,
	)
}

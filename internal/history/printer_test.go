package history_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/history"
)

func makeEvents() []history.Event {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	return []history.Event{
		{Timestamp: base, Port: 80, Kind: history.Opened},
		{Timestamp: base.Add(time.Minute), Port: 443, Kind: history.Opened},
		{Timestamp: base.Add(2 * time.Minute), Port: 80, Kind: history.Closed},
	}
}

func TestPrint_ContainsHeader(t *testing.T) {
	var buf bytes.Buffer
	opts := history.PrintOptions{Out: &buf}
	history.Print(makeEvents(), opts)
	if !strings.Contains(buf.String(), "TIMESTAMP") {
		t.Error("expected header row in output")
	}
}

func TestPrint_ContainsAllPorts(t *testing.T) {
	var buf bytes.Buffer
	opts := history.PrintOptions{Out: &buf}
	history.Print(makeEvents(), opts)
	out := buf.String()
	for _, want := range []string{"80", "443"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected port %s in output", want)
		}
	}
}

func TestPrint_LimitReducesRows(t *testing.T) {
	var buf bytes.Buffer
	opts := history.PrintOptions{Out: &buf, Limit: 1}
	history.Print(makeEvents(), opts)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	// header + 1 data row
	if len(lines) != 2 {
		t.Errorf("expected 2 lines with limit=1, got %d", len(lines))
	}
}

func TestPrint_SinceFiltersOldEvents(t *testing.T) {
	var buf bytes.Buffer
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	opts := history.PrintOptions{
		Out:   &buf,
		Since: base.Add(90 * time.Second),
	}
	history.Print(makeEvents(), opts)
	out := buf.String()
	// Only the third event (12:02) should appear
	if strings.Count(out, "port") > 1 {
		t.Error("expected only events after Since threshold")
	}
	if !strings.Contains(out, "closed") {
		t.Error("expected closed event to appear after Since filter")
	}
}

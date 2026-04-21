package history_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/portwatch/internal/history"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "history.json")
}

func TestRecord_AppendsEvent(t *testing.T) {
	h := history.New(tempPath(t), 100)
	if err := h.Record(8080, history.Opened); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	events := h.All()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Port != 8080 || events[0].Kind != history.Opened {
		t.Errorf("unexpected event: %+v", events[0])
	}
}

func TestRecord_EnforcesLimit(t *testing.T) {
	h := history.New(tempPath(t), 3)
	for i := 0; i < 5; i++ {
		if err := h.Record(8000+i, history.Opened); err != nil {
			t.Fatalf("record error: %v", err)
		}
	}
	if got := len(h.All()); got != 3 {
		t.Errorf("expected 3 events after limit, got %d", got)
	}
}

func TestLoad_RoundTrip(t *testing.T) {
	path := tempPath(t)
	h1 := history.New(path, 100)
	_ = h1.Record(9090, history.Closed)
	_ = h1.Record(443, history.Opened)

	h2 := history.New(path, 100)
	if err := h2.Load(); err != nil {
		t.Fatalf("load error: %v", err)
	}
	events := h2.All()
	if len(events) != 2 {
		t.Fatalf("expected 2 events after reload, got %d", len(events))
	}
	if events[0].Port != 9090 || events[0].Kind != history.Closed {
		t.Errorf("unexpected first event: %+v", events[0])
	}
}

func TestLoad_MissingFile_ReturnsNil(t *testing.T) {
	h := history.New(filepath.Join(t.TempDir(), "no-such-file.json"), 100)
	if err := h.Load(); err != nil {
		t.Errorf("expected nil for missing file, got %v", err)
	}
	if len(h.All()) != 0 {
		t.Error("expected empty events for missing file")
	}
}

func TestRecord_PersistsToDisk(t *testing.T) {
	path := tempPath(t)
	h := history.New(path, 100)
	_ = h.Record(22, history.Opened)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected history file to exist after Record")
	}
}

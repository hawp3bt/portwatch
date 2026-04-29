package audit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/portwatch/internal/audit"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "audit.log")
}

func TestRecord_AppendsEntry(t *testing.T) {
	p := tempPath(t)
	l := audit.New(p)

	if err := l.Record(audit.KindScan, "daemon", "scan complete", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := audit.Load(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Kind != audit.KindScan {
		t.Errorf("kind: got %q, want %q", entries[0].Kind, audit.KindScan)
	}
	if entries[0].Message != "scan complete" {
		t.Errorf("message: got %q", entries[0].Message)
	}
}

func TestRecord_MultipleEntries(t *testing.T) {
	p := tempPath(t)
	l := audit.New(p)

	kinds := []audit.EventKind{audit.KindAlert, audit.KindSuppress, audit.KindBaseline}
	for _, k := range kinds {
		if err := l.Record(k, "cli", "msg", nil); err != nil {
			t.Fatalf("record %s: %v", k, err)
		}
	}

	entries, err := audit.Load(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for i, k := range kinds {
		if entries[i].Kind != k {
			t.Errorf("entry %d: kind %q != %q", i, entries[i].Kind, k)
		}
	}
}

func TestRecord_MetaIsPreserved(t *testing.T) {
	p := tempPath(t)
	l := audit.New(p)

	meta := map[string]string{"port": "8080", "proto": "tcp"}
	if err := l.Record(audit.KindAlert, "daemon", "port opened", meta); err != nil {
		t.Fatalf("record: %v", err)
	}

	entries, err := audit.Load(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if entries[0].Meta["port"] != "8080" {
		t.Errorf("meta port: got %q", entries[0].Meta["port"])
	}
}

func TestLoad_MissingFile_ReturnsNil(t *testing.T) {
	p := filepath.Join(t.TempDir(), "missing.log")
	entries, err := audit.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil, got %v", entries)
	}
}

func TestRecord_TimestampIsRecent(t *testing.T) {
	p := tempPath(t)
	l := audit.New(p)
	before := time.Now().UTC()

	if err := l.Record(audit.KindProfile, "cli", "profile saved", nil); err != nil {
		t.Fatalf("record: %v", err)
	}

	after := time.Now().UTC()
	entries, _ := audit.Load(p)
	ts := entries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in range [%v, %v]", ts, before, after)
	}
}

func TestRecord_CreatesFileIfMissing(t *testing.T) {
	p := tempPath(t)
	l := audit.New(p)
	_ = l.Record(audit.KindThreshold, "daemon", "threshold violated", nil)

	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Error("expected file to be created")
	}
}

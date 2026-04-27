package schedule

import (
	"testing"
	"time"
)

func TestSet_AndGet_RoundTrip(t *testing.T) {
	s := New()
	err := s.Set("default", 10*time.Second, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("default")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Interval != 10*time.Second {
		t.Errorf("expected 10s, got %s", e.Interval)
	}
	if !e.Enabled {
		t.Error("expected entry to be enabled")
	}
}

func TestSet_EmptyName_ReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("", 5*time.Second, true); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSet_IntervalTooShort_ReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("fast", 500*time.Millisecond, true); err == nil {
		t.Fatal("expected error for sub-second interval")
	}
}

func TestRemove_DeletesEntry(t *testing.T) {
	s := New()
	_ = s.Set("temp", 5*time.Second, true)
	s.Remove("temp")
	_, ok := s.Get("temp")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestList_ReturnsAllEntries(t *testing.T) {
	s := New()
	_ = s.Set("a", 5*time.Second, true)
	_ = s.Set("b", 10*time.Second, false)
	list := s.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
}

func TestActiveIntervals_OnlyEnabled(t *testing.T) {
	s := New()
	_ = s.Set("enabled", 5*time.Second, true)
	_ = s.Set("disabled", 10*time.Second, false)
	intervals := s.ActiveIntervals()
	if len(intervals) != 1 {
		t.Fatalf("expected 1 active interval, got %d", len(intervals))
	}
	if intervals[0] != 5*time.Second {
		t.Errorf("expected 5s, got %s", intervals[0])
	}
}

func TestGet_MissingEntry_ReturnsFalse(t *testing.T) {
	s := New()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected false for missing entry")
	}
}

package suppress_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/suppress"
)

func TestIsSuppressed_ReturnsFalseForUnknownPort(t *testing.T) {
	s := suppress.New()
	if s.IsSuppressed(8080) {
		t.Fatal("expected false for unknown port")
	}
}

func TestSuppress_MakesPortSuppressed(t *testing.T) {
	s := suppress.New()
	s.Suppress(9090, "maintenance", 5*time.Minute)
	if !s.IsSuppressed(9090) {
		t.Fatal("expected port 9090 to be suppressed")
	}
}

func TestIsSuppressed_ReturnsFalseAfterExpiry(t *testing.T) {
	s := suppress.New()
	s.Suppress(3000, "test", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	if s.IsSuppressed(3000) {
		t.Fatal("expected suppression to have expired")
	}
}

func TestLift_RemovesActiveSuppression(t *testing.T) {
	s := suppress.New()
	s.Suppress(4000, "manual", 10*time.Minute)
	s.Lift(4000)
	if s.IsSuppressed(4000) {
		t.Fatal("expected suppression to be lifted")
	}
}

func TestList_ReturnsActiveEntries(t *testing.T) {
	s := suppress.New()
	s.Suppress(5000, "deploy", 10*time.Minute)
	s.Suppress(6000, "test", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	list := s.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 active entry, got %d", len(list))
	}
	if list[0].Port != 5000 {
		t.Errorf("expected port 5000, got %d", list[0].Port)
	}
}

func TestList_ReturnsEmptyWhenNoneSuppressed(t *testing.T) {
	s := suppress.New()
	if got := s.List(); len(got) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(got))
	}
}

func TestSuppress_OverwritesExistingEntry(t *testing.T) {
	s := suppress.New()
	s.Suppress(7000, "first", 1*time.Millisecond)
	s.Suppress(7000, "second", 10*time.Minute)
	time.Sleep(10 * time.Millisecond)
	// second suppression should still be active
	if !s.IsSuppressed(7000) {
		t.Fatal("expected overwritten suppression to still be active")
	}
}

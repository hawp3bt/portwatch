package quota

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_FirstCallAlwaysPermitted(t *testing.T) {
	q := New(3, time.Minute)
	if !q.Allow(8080) {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_RespectsMaxWithinWindow(t *testing.T) {
	base := time.Now()
	q := New(2, time.Minute)
	q.now = fixedNow(base)

	if !q.Allow(9000) {
		t.Fatal("first allow should succeed")
	}
	if !q.Allow(9000) {
		t.Fatal("second allow should succeed")
	}
	if q.Allow(9000) {
		t.Fatal("third allow should be denied (max=2)")
	}
}

func TestAllow_WindowExpiryResetsCount(t *testing.T) {
	base := time.Now()
	q := New(1, time.Minute)
	q.now = fixedNow(base)

	q.Allow(443)
	if q.Allow(443) {
		t.Fatal("second call within window should be denied")
	}

	// Advance past the window.
	q.now = fixedNow(base.Add(2 * time.Minute))
	if !q.Allow(443) {
		t.Fatal("first call after window expiry should be allowed")
	}
}

func TestAllow_DifferentPortsAreIndependent(t *testing.T) {
	q := New(1, time.Minute)
	q.Allow(80)
	if !q.Allow(443) {
		t.Fatal("different port should have its own quota")
	}
}

func TestReset_ClearsPort(t *testing.T) {
	base := time.Now()
	q := New(1, time.Minute)
	q.now = fixedNow(base)

	q.Allow(8080)
	if q.Allow(8080) {
		t.Fatal("should be denied before reset")
	}
	q.Reset(8080)
	if !q.Allow(8080) {
		t.Fatal("should be allowed after reset")
	}
}

func TestRemaining_ReflectsUsage(t *testing.T) {
	base := time.Now()
	q := New(3, time.Minute)
	q.now = fixedNow(base)

	if q.Remaining(22) != 3 {
		t.Fatalf("expected 3 remaining before any calls, got %d", q.Remaining(22))
	}
	q.Allow(22)
	if q.Remaining(22) != 2 {
		t.Fatalf("expected 2 remaining after one call, got %d", q.Remaining(22))
	}
}

func TestNew_ClampsInvalidParams(t *testing.T) {
	q := New(0, 0)
	if q.max < 1 {
		t.Errorf("max should be clamped to at least 1, got %d", q.max)
	}
	if q.window < time.Second {
		t.Errorf("window should be clamped to at least 1s, got %v", q.window)
	}
}

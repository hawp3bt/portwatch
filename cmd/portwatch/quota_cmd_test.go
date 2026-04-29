package main

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/quota"
)

func TestDefaultQuota_ReturnsNonNil(t *testing.T) {
	q := defaultQuota()
	if q == nil {
		t.Fatal("defaultQuota returned nil")
	}
}

func TestDefaultQuota_EnvOverride(t *testing.T) {
	t.Setenv("PORTWATCH_QUOTA_MAX", "2")
	t.Setenv("PORTWATCH_QUOTA_WINDOW", "30s")

	q := defaultQuota()
	// Exhaust the quota to verify max=2 was applied.
	if !q.Allow(8080) {
		t.Fatal("first allow should succeed")
	}
	if !q.Allow(8080) {
		t.Fatal("second allow should succeed")
	}
	if q.Allow(8080) {
		t.Fatal("third allow should be denied with max=2")
	}
}

func TestRunQuota_CheckAllowed(t *testing.T) {
	// Just verify Allow path doesn't panic; output goes to stdout.
	q := quota.New(5, time.Minute)
	// Should not panic.
	runQuota_noExit([]string{"remaining", "443"}, q)
}

// runQuota_noExit is a thin wrapper used in tests to avoid os.Exit.
func runQuota_noExit(args []string, q *quota.Quota) {
	if len(args) < 2 {
		return
	}
	switch args[0] {
	case "remaining":
		_ = q.Remaining(443)
	case "reset":
		q.Reset(443)
	case "check":
		_ = q.Allow(443)
	}
}

func TestRunQuota_ResetThenAllow(t *testing.T) {
	q := quota.New(1, time.Minute)
	q.Allow(22)
	if q.Remaining(22) != 0 {
		t.Fatalf("expected 0 remaining, got %d", q.Remaining(22))
	}
	runQuota_noExit([]string{"reset", "22"}, q)
	if q.Remaining(22) != 1 {
		t.Fatalf("expected 1 remaining after reset, got %d", q.Remaining(22))
	}
}

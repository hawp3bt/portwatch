package watchdog

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatchdog_CallsOnUnhealthyAfterMaxFails(t *testing.T) {
	var callCount int32
	onUnhealthy := func(_ string) {
		atomic.AddInt32(&callCount, 1)
	}

	alwaysUnhealthy := func() bool { return false }

	w := New("test", 10*time.Millisecond, 2, alwaysUnhealthy, onUnhealthy)

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	w.Run(ctx)

	if atomic.LoadInt32(&callCount) < 1 {
		t.Errorf("expected onUnhealthy to be called at least once, got %d", callCount)
	}
}

func TestWatchdog_DoesNotCallOnUnhealthyWhenHealthy(t *testing.T) {
	var callCount int32
	onUnhealthy := func(_ string) {
		atomic.AddInt32(&callCount, 1)
	}

	alwaysHealthy := func() bool { return true }

	w := New("test", 10*time.Millisecond, 1, alwaysHealthy, onUnhealthy)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	w.Run(ctx)

	if atomic.LoadInt32(&callCount) != 0 {
		t.Errorf("expected onUnhealthy to never be called, got %d", callCount)
	}
}

func TestWatchdog_ResetsFailCountOnRecovery(t *testing.T) {
	var callCount int32
	onUnhealthy := func(_ string) {
		atomic.AddInt32(&callCount, 1)
	}

	// Healthy on first tick, unhealthy on subsequent — never reaches maxFails=3 in time.
	tick := 0
	healthFn := func() bool {
		tick++
		return tick%2 == 1 // alternates healthy/unhealthy
	}

	w := New("test", 10*time.Millisecond, 3, healthFn, onUnhealthy)

	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Millisecond)
	defer cancel()

	w.Run(ctx)

	if atomic.LoadInt32(&callCount) != 0 {
		t.Errorf("expected no unhealthy trigger with alternating health, got %d", callCount)
	}
}

func TestNew_MaxFailsClampedToOne(t *testing.T) {
	w := New("test", time.Second, 0, func() bool { return true }, func(_ string) {})
	if w.maxFails != 1 {
		t.Errorf("expected maxFails to be clamped to 1, got %d", w.maxFails)
	}
}

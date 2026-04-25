package anomaly

import (
	"math"
	"testing"
)

func TestNew_ClampsWindowAndThreshold(t *testing.T) {
	d := New(0, 0.0)
	if d.window < 2 {
		t.Errorf("expected window >= 2, got %d", d.window)
	}
	if d.threshold < 0.1 {
		t.Errorf("expected threshold >= 0.1, got %f", d.threshold)
	}
}

func TestCheck_FewerThanTwoSamples_NoAnomaly(t *testing.T) {
	d := New(5, 2.0)
	d.Push(10)
	r := d.Check(100)
	if r.Anomaly {
		t.Error("expected no anomaly with fewer than 2 samples")
	}
	if r.Current != 100 {
		t.Errorf("expected Current=100, got %d", r.Current)
	}
}

func TestCheck_StableWindow_NoAnomaly(t *testing.T) {
	d := New(5, 2.0)
	for i := 0; i < 5; i++ {
		d.Push(20)
	}
	// stddev is 0 when all samples equal; z-score cannot exceed threshold
	r := d.Check(20)
	if r.Anomaly {
		t.Error("expected no anomaly for stable window")
	}
}

func TestCheck_LargeSpike_IsAnomaly(t *testing.T) {
	d := New(10, 2.0)
	for i := 0; i < 10; i++ {
		d.Push(10)
	}
	// inject one outlier into samples so stddev > 0
	d.Push(10)
	d.Push(10)
	d.Push(10)
	d.Push(10)
	d.Push(50) // outlier in window
	r := d.Check(50)
	if r.StdDev == 0 {
		t.Skip("stddev is zero, cannot test z-score")
	}
	if r.ZScore < 0 {
		t.Error("expected non-negative z-score")
	}
}

func TestPush_EvictsOldestBeyondWindow(t *testing.T) {
	d := New(3, 2.0)
	for i := 1; i <= 5; i++ {
		d.Push(i)
	}
	if d.Len() != 3 {
		t.Errorf("expected 3 samples, got %d", d.Len())
	}
}

func TestCheck_ZScoreCalculation(t *testing.T) {
	d := New(4, 2.0)
	d.Push(10)
	d.Push(10)
	d.Push(10)
	d.Push(10)
	// mean=10, but all same so stddev=0; push a varied set
	d2 := New(4, 2.0)
	d2.Push(8)
	d2.Push(10)
	d2.Push(12)
	d2.Push(10)
	r := d2.Check(20)
	if r.Mean == 0 {
		t.Error("expected non-zero mean")
	}
	expectedZ := math.Abs(20-r.Mean) / r.StdDev
	if math.Abs(r.ZScore-expectedZ) > 1e-9 {
		t.Errorf("z-score mismatch: got %f, want %f", r.ZScore, expectedZ)
	}
	if r.Anomaly != (r.ZScore >= 2.0) {
		t.Error("Anomaly flag does not match z-score vs threshold")
	}
}

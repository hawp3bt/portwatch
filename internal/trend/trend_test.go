package trend

import (
	"testing"
)

func TestNew_WindowClampedToTwo(t *testing.T) {
	tr := New(0)
	if tr.window != 2 {
		t.Fatalf("expected window=2, got %d", tr.window)
	}
}

func TestDirection_StableWhenFewerThanTwoSamples(t *testing.T) {
	tr := New(5)
	if tr.Direction() != Stable {
		t.Fatal("expected Stable with no samples")
	}
	tr.Record(3)
	if tr.Direction() != Stable {
		t.Fatal("expected Stable with one sample")
	}
}

func TestDirection_Rising(t *testing.T) {
	tr := New(5)
	tr.Record(2)
	tr.Record(5)
	if got := tr.Direction(); got != Rising {
		t.Fatalf("expected Rising, got %s", got)
	}
}

func TestDirection_Falling(t *testing.T) {
	tr := New(5)
	tr.Record(10)
	tr.Record(4)
	if got := tr.Direction(); got != Falling {
		t.Fatalf("expected Falling, got %s", got)
	}
}

func TestDirection_Stable(t *testing.T) {
	tr := New(5)
	tr.Record(7)
	tr.Record(7)
	if got := tr.Direction(); got != Stable {
		t.Fatalf("expected Stable, got %s", got)
	}
}

func TestRecord_EnforcesWindow(t *testing.T) {
	tr := New(3)
	for i := 1; i <= 6; i++ {
		tr.Record(i)
	}
	samples := tr.Samples()
	if len(samples) != 3 {
		t.Fatalf("expected 3 samples, got %d", len(samples))
	}
	if samples[0].Count != 4 {
		t.Fatalf("expected oldest sample count=4, got %d", samples[0].Count)
	}
}

func TestReset_ClearsSamples(t *testing.T) {
	tr := New(5)
	tr.Record(1)
	tr.Record(2)
	tr.Reset()
	if len(tr.Samples()) != 0 {
		t.Fatal("expected empty samples after Reset")
	}
	if tr.Direction() != Stable {
		t.Fatal("expected Stable after Reset")
	}
}

func TestSamples_ReturnsCopy(t *testing.T) {
	tr := New(5)
	tr.Record(3)
	s := tr.Samples()
	s[0].Count = 999
	original := tr.Samples()
	if original[0].Count == 999 {
		t.Fatal("Samples should return a copy, not a reference")
	}
}

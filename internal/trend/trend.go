package trend

import (
	"sync"
	"time"
)

// Direction indicates whether port activity is increasing or decreasing.
type Direction string

const (
	Rising  Direction = "rising"
	Falling Direction = "falling"
	Stable  Direction = "stable"
)

// Sample records how many ports were open at a given time.
type Sample struct {
	At    time.Time
	Count int
}

// Tracker accumulates port-count samples and reports trend direction.
type Tracker struct {
	mu      sync.Mutex
	window  int
	samples []Sample
}

// New returns a Tracker that keeps at most window samples.
func New(window int) *Tracker {
	if window < 2 {
		window = 2
	}
	return &Tracker{window: window}
}

// Record appends a new sample, dropping the oldest when the window is full.
func (t *Tracker) Record(count int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.samples = append(t.samples, Sample{At: time.Now(), Count: count})
	if len(t.samples) > t.window {
		t.samples = t.samples[len(t.samples)-t.window:]
	}
}

// Direction returns the overall trend across all retained samples.
func (t *Tracker) Direction() Direction {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.samples) < 2 {
		return Stable
	}
	first := t.samples[0].Count
	last := t.samples[len(t.samples)-1].Count
	switch {
	case last > first:
		return Rising
	case last < first:
		return Falling
	default:
		return Stable
	}
}

// Samples returns a copy of the current sample slice.
func (t *Tracker) Samples() []Sample {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Sample, len(t.samples))
	copy(out, t.samples)
	return out
}

// Reset clears all recorded samples.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.samples = nil
}

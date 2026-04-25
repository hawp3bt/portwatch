// Package anomaly detects statistically unusual port activity
// by comparing current open port counts against a rolling baseline mean.
package anomaly

import "math"

// Detector tracks rolling port count samples and flags anomalies
// when the current count deviates beyond a configurable threshold.
type Detector struct {
	window    int
	threshold float64 // standard deviations
	samples   []int
}

// Result describes the outcome of an anomaly check.
type Result struct {
	Anomaly  bool
	Current  int
	Mean     float64
	StdDev   float64
	ZScore   float64
}

// New creates a Detector with the given rolling window size and z-score threshold.
// window is clamped to a minimum of 2; threshold is clamped to a minimum of 0.1.
func New(window int, threshold float64) *Detector {
	if window < 2 {
		window = 2
	}
	if threshold < 0.1 {
		threshold = 0.1
	}
	return &Detector{window: window, threshold: threshold}
}

// Push records a new port count sample, evicting the oldest if the window is full.
func (d *Detector) Push(count int) {
	d.samples = append(d.samples, count)
	if len(d.samples) > d.window {
		d.samples = d.samples[len(d.samples)-d.window:]
	}
}

// Check evaluates whether count is anomalous relative to the current sample window.
// It returns a Result with Anomaly=false and zero statistics when fewer than 2
// samples have been recorded.
func (d *Detector) Check(count int) Result {
	if len(d.samples) < 2 {
		return Result{Current: count}
	}
	mean := d.mean()
	std := d.stddev(mean)
	var z float64
	if std > 0 {
		z = math.Abs(float64(count)-mean) / std
	}
	return Result{
		Anomaly: std > 0 && z >= d.threshold,
		Current: count,
		Mean:    mean,
		StdDev:  std,
		ZScore:  z,
	}
}

// Len returns the number of samples currently held.
func (d *Detector) Len() int { return len(d.samples) }

func (d *Detector) mean() float64 {
	sum := 0
	for _, s := range d.samples {
		sum += s
	}
	return float64(sum) / float64(len(d.samples))
}

func (d *Detector) stddev(mean float64) float64 {
	var variance float64
	for _, s := range d.samples {
		d := float64(s) - mean
		variance += d * d
	}
	return math.Sqrt(variance / float64(len(d.samples)))
}

package watchdog

import (
	"context"
	"log"
	"time"
)

// HealthFunc is a function that returns true if the monitored component is healthy.
type HealthFunc func() bool

// Watchdog periodically checks a health function and calls a restart callback
// if the component appears unhealthy for a configurable number of consecutive failures.
type Watchdog struct {
	name      string
	interval  time.Duration
	maxFails  int
	healthFn  HealthFunc
	onUnhealthy func(name string)
}

// New creates a new Watchdog.
// name is a label used in log output.
// interval is how often the health check runs.
// maxFails is how many consecutive failures trigger the onUnhealthy callback.
func New(name string, interval time.Duration, maxFails int, healthFn HealthFunc, onUnhealthy func(string)) *Watchdog {
	if maxFails < 1 {
		maxFails = 1
	}
	return &Watchdog{
		name:        name,
		interval:    interval,
		maxFails:    maxFails,
		healthFn:    healthFn,
		onUnhealthy: onUnhealthy,
	}
}

// Run starts the watchdog loop. It blocks until ctx is cancelled.
func (w *Watchdog) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	consecutiveFails := 0

	for {
		select {
		case <-ctx.Done():
			log.Printf("[watchdog] %s: stopped", w.name)
			return
		case <-ticker.C:
			if w.healthFn() {
				if consecutiveFails > 0 {
					log.Printf("[watchdog] %s: recovered after %d failure(s)", w.name, consecutiveFails)
				}
				consecutiveFails = 0
			} else {
				consecutiveFails++
				log.Printf("[watchdog] %s: unhealthy (%d/%d)", w.name, consecutiveFails, w.maxFails)
				if consecutiveFails >= w.maxFails {
					w.onUnhealthy(w.name)
					consecutiveFails = 0
				}
			}
		}
	}
}

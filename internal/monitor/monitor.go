package monitor

import (
	"fmt"
	"sort"
	"time"

	"github.com/user/portwatch/internal/scanner"
)

// Alert represents a change detected in open ports.
type Alert struct {
	Time    time.Time
	Message string
	Added   []int
	Removed []int
}

// Monitor watches ports at a given interval and sends alerts on changes.
type Monitor struct {
	Ports    []int
	Interval time.Duration
	Alerts   chan Alert
	stop     chan struct{}
}

// New creates a new Monitor for the given ports and poll interval.
func New(ports []int, interval time.Duration) *Monitor {
	return &Monitor{
		Ports:    ports,
		Interval: interval,
		Alerts:   make(chan Alert, 16),
		stop:     make(chan struct{}),
	}
}

// Start begins polling in a background goroutine.
func (m *Monitor) Start() {
	go m.run()
}

// Stop signals the monitor to cease polling.
func (m *Monitor) Stop() {
	close(m.stop)
}

func (m *Monitor) run() {
	previous := scanner.OpenPorts(scanner.Scan(m.Ports))
	ticker := time.NewTicker(m.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-m.stop:
			close(m.Alerts)
			return
		case <-ticker.C:
			current := scanner.OpenPorts(scanner.Scan(m.Ports))
			added, removed := diff(previous, current)
			if len(added) > 0 || len(removed) > 0 {
				m.Alerts <- Alert{
					Time:    time.Now(),
					Message: fmt.Sprintf("ports changed: +%v -%v", added, removed),
					Added:   added,
					Removed: removed,
				}
			}
			previous = current
		}
	}
}

func diff(prev, curr []int) (added, removed []int) {
	prevSet := toSet(prev)
	currSet := toSet(curr)
	for p := range currSet {
		if !prevSet[p] {
			added = append(added, p)
		}
	}
	for p := range prevSet {
		if !currSet[p] {
			removed = append(removed, p)
		}
	}
	sort.Ints(added)
	sort.Ints(removed)
	return
}

func toSet(ports []int) map[int]bool {
	s := make(map[int]bool, len(ports))
	for _, p := range ports {
		s[p] = true
	}
	return s
}

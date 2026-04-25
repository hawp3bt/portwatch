package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/user/portwatch/internal/anomaly"
	"github.com/user/portwatch/internal/history"
)

// runAnomaly evaluates recent port-count history for statistical anomalies.
// Usage: portwatch anomaly [--window N] [--threshold Z] [--history path]
func runAnomaly(args []string) {
	window := 10
	threshold := 2.0
	hisPath := defaultHistoryPath()

	for i := 0; i < len(args)-1; i++ {
		switch args[i] {
		case "--window":
			if n, err := strconv.Atoi(args[i+1]); err == nil {
				window = n
			}
			i++
		case "--threshold":
			if f, err := strconv.ParseFloat(args[i+1], 64); err == nil {
				threshold = f
			}
			i++
		case "--history":
			hisPath = args[i+1]
			i++
		}
	}

	events, err := history.Load(hisPath)
	if err != nil || len(events) == 0 {
		fmt.Fprintln(os.Stderr, "portwatch anomaly: no history data found")
		os.Exit(1)
	}

	// Bucket events by scan cycle into per-minute open-port counts.
	counts := aggregateEventCounts(events) // reuse from trend_cmd.go
	if len(counts) < 2 {
		fmt.Println("Not enough data points for anomaly detection.")
		return
	}

	det := anomaly.New(window, threshold)
	// Feed all but the last sample as history.
	for _, c := range counts[:len(counts)-1] {
		det.Push(c)
	}

	current := counts[len(counts)-1]
	result := det.Check(current)

	fmt.Printf("Anomaly Detection Report\n")
	fmt.Printf("  Window   : %d samples\n", det.Len())
	fmt.Printf("  Threshold: %.2f σ\n", threshold)
	fmt.Printf("  Mean     : %.2f\n", result.Mean)
	fmt.Printf("  Std Dev  : %.2f\n", result.StdDev)
	fmt.Printf("  Current  : %d\n", result.Current)
	fmt.Printf("  Z-Score  : %.2f\n", result.ZScore)

	if result.Anomaly {
		fmt.Printf("  Status   : ⚠ ANOMALY DETECTED (z=%.2f >= %.2f)\n", result.ZScore, threshold)
		os.Exit(2)
	} else {
		fmt.Printf("  Status   : ✓ Normal\n")
	}
}

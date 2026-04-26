// Package report generates periodic summaries of port activity,
// combining trend, anomaly, digest, and history data into a single
// human-readable or JSON-serialisable structure.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/user/portwatch/internal/anomaly"
	"github.com/user/portwatch/internal/digest"
	"github.com/user/portwatch/internal/history"
	"github.com/user/portwatch/internal/trend"
)

// Summary is a point-in-time snapshot of port-activity metrics.
type Summary struct {
	GeneratedAt  time.Time        `json:"generated_at"`
	Period       string           `json:"period"`
	Digest       string           `json:"digest"`
	DigestChange bool             `json:"digest_changed"`
	OpenCount    int              `json:"open_port_count"`
	Events       []history.Event  `json:"events"`
	Anomalies    []AnomalyEntry   `json:"anomalies,omitempty"`
	Trend        string           `json:"trend_direction"`
}

// AnomalyEntry pairs a port number with its anomaly score for reporting.
type AnomalyEntry struct {
	Port  int     `json:"port"`
	Score float64 `json:"score"`
}

// Options controls what is included in the generated summary.
type Options struct {
	// Since restricts history events to those after this time.
	Since time.Time
	// PreviousDigest is compared against the current digest to flag changes.
	PreviousDigest string
	// AnomalyThreshold is the minimum score to include an anomaly entry.
	AnomalyThreshold float64
}

// Build constructs a Summary from the provided data sources.
func Build(
	currentPorts []int,
	events []history.Event,
	tr *trend.Tracker,
	an *anomaly.Detector,
	opts Options,
) Summary {
	d := digest.Compute(currentPorts)

	// Filter events to the requested time window.
	var filtered []history.Event
	for _, e := range events {
		if e.Time.After(opts.Since) {
			filtered = append(filtered, e)
		}
	}

	// Collect anomalies above the threshold.
	var anomalies []AnomalyEntry
	if an != nil {
		for _, p := range currentPorts {
			if score, ok := an.Score(p); ok && score >= opts.AnomalyThreshold {
				anomalies = append(anomalies, AnomalyEntry{Port: p, Score: score})
			}
		}
	}

	// Derive period label from opts.Since.
	period := "all time"
	if !opts.Since.IsZero() {
		period = fmt.Sprintf("since %s", opts.Since.Format(time.RFC3339))
	}

	direction := "stable"
	if tr != nil {
		direction = string(tr.Direction())
	}

	return Summary{
		GeneratedAt:  time.Now().UTC(),
		Period:       period,
		Digest:       d,
		DigestChange: opts.PreviousDigest != "" && !digest.Equal(d, opts.PreviousDigest),
		OpenCount:    len(currentPorts),
		Events:       filtered,
		Anomalies:    anomalies,
		Trend:        direction,
	}
}

// PrintText writes a human-readable summary to w.
func PrintText(w io.Writer, s Summary) {
	fmt.Fprintf(w, "Port Activity Report\n")
	fmt.Fprintf(w, "Generated : %s\n", s.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Period    : %s\n", s.Period)
	fmt.Fprintf(w, "Open ports: %d  (digest: %s", s.OpenCount, s.Digest)
	if s.DigestChange {
		fmt.Fprintf(w, ", CHANGED")
	}
	fmt.Fprintf(w, ")\n")
	fmt.Fprintf(w, "Trend     : %s\n", s.Trend)

	if len(s.Anomalies) > 0 {
		fmt.Fprintf(w, "\nAnomalies (%d):\n", len(s.Anomalies))
		for _, a := range s.Anomalies {
			fmt.Fprintf(w, "  port %-6d  score %.2f\n", a.Port, a.Score)
		}
	}

	fmt.Fprintf(w, "\nEvents (%d):\n", len(s.Events))
	for _, e := range s.Events {
		fmt.Fprintf(w, "  [%s] %-8s port %d\n",
			e.Time.Format("15:04:05"), e.Kind, e.Port)
	}
}

// PrintJSON writes the summary as indented JSON to w.
func PrintJSON(w io.Writer, s Summary) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

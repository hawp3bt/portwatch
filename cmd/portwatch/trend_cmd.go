package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/user/portwatch/internal/history"
	"github.com/user/portwatch/internal/trend"
)

const defaultTrendWindow = 20

// runTrend loads historical events and renders a port-count trend report.
func runTrend(args []string) {
	window := defaultTrendWindow
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 2 {
			fmt.Fprintln(os.Stderr, "trend: window must be an integer >= 2")
			os.Exit(1)
		}
		window = n
	}

	hiPath := defaultHistoryPath()
	events, err := history.Load(hiPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "trend: failed to load history: %v\n", err)
		os.Exit(1)
	}
	if len(events) == 0 {
		fmt.Println("No history available to compute trend.")
		return
	}

	// Build a per-scan port count by grouping events on the same second.
	tr := trend.New(window)
	counts := aggregateEventCounts(events)
	for _, c := range counts {
		tr.Record(c)
	}

	trend.Print(tr, trend.PrintOptions{Out: os.Stdout, Header: true})
}

// aggregateEventCounts collapses history events into a slice of open-port
// counts, one entry per distinct scan timestamp (truncated to the second).
func aggregateEventCounts(events []history.Event) []int {
	type key = int64
	counted := make(map[key]int)
	var order []key
	seen := make(map[key]bool)

	for _, e := range events {
		k := e.At.Unix()
		if !seen[k] {
			seen[k] = true
			order = append(order, k)
		}
		if e.Type == "opened" {
			counted[k]++
		}
	}

	out := make([]int, 0, len(order))
	for _, k := range order {
		out = append(out, counted[k])
	}
	return out
}

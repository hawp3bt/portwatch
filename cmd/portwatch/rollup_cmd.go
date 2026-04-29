package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/user/portwatch/internal/history"
	"github.com/user/portwatch/internal/rollup"
)

// runRollup replays recent history events through a rollup window and prints
// a consolidated summary — useful for reviewing noisy scan periods.
func runRollup(args []string) {
	fs := flag.NewFlagSet("rollup", flag.ExitOnError)
	windowSec := fs.Int("window", 30, "rollup window in seconds")
	limit := fs.Int("limit", 200, "max history events to consider")
	histPath := fs.String("history", defaultHistoryPath(), "path to history file")
	_ = fs.Parse(args)

	h := history.New(*histPath, 10_000)
	events, err := h.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "rollup: load history: %v\n", err)
		os.Exit(1)
	}

	if len(events) == 0 {
		fmt.Println("no history events found")
		return
	}

	if len(events) > *limit {
		events = events[len(events)-*limit:]
	}

	window := time.Duration(*windowSec) * time.Second

	type bucket struct {
		opened map[int]struct{}
		closed map[int]struct{}
	}

	var summaries []rollup.Summary
	r := rollup.New(window, func(s rollup.Summary) {
		summaries = append(summaries, s)
	})

	for _, e := range events {
		if e.Type == "opened" {
			r.Push(rollup.Event{Port: e.Port, Opened: true})
		} else if e.Type == "closed" {
			r.Push(rollup.Event{Port: e.Port, Opened: false})
		}
	}
	r.Flush()
	time.Sleep(20 * time.Millisecond) // allow async emit

	if len(summaries) == 0 {
		fmt.Println("rollup: no changes to report")
		return
	}

	fmt.Printf("%-26s  %-8s  %s\n", "time", "opened", "closed")
	fmt.Println("--------------------------------------------------------------")
	for _, s := range summaries {
		sort.Ints(s.Opened)
		sort.Ints(s.Closed)
		fmt.Printf("%-26s  %-8d  %d\n", s.At.Format(time.RFC3339), len(s.Opened), len(s.Closed))
	}
	fmt.Printf("\ntotal summaries: %d\n", len(summaries))
}

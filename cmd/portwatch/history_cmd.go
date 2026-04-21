package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/user/portwatch/internal/history"
)

// runHistory handles the "history" sub-command.
// Usage: portwatch history [-n <limit>] [-since <duration>]
func runHistory(args []string) {
	fs := flag.NewFlagSet("history", flag.ExitOnError)
	limit := fs.Int("n", 50, "maximum number of events to display")
	sinceStr := fs.String("since", "", "show events newer than this duration (e.g. 24h)")
	histPath := fs.String("file", defaultHistoryPath(), "path to history file")

	_ = fs.Parse(args)

	var since time.Time
	if *sinceStr != "" {
		d, err := time.ParseDuration(*sinceStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid -since value: %v\n", err)
			os.Exit(1)
		}
		since = time.Now().Add(-d)
	}

	h := history.New(*histPath, 0)
	if err := h.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load history: %v\n", err)
		os.Exit(1)
	}

	history.Print(h.All(), history.PrintOptions{
		Out:   os.Stdout,
		Limit: *limit,
		Since: since,
	})
}

func defaultHistoryPath() string {
	if p := os.Getenv("PORTWATCH_HISTORY"); p != "" {
		return p
	}
	return "/var/lib/portwatch/history.json"
}

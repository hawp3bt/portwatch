package history

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

// PrintOptions controls how history is rendered.
type PrintOptions struct {
	Out    io.Writer
	Limit  int
	Since  time.Time
}

// DefaultPrintOptions returns options that write to stdout with no limit.
func DefaultPrintOptions() PrintOptions {
	return PrintOptions{Out: os.Stdout}
}

// Print writes a formatted table of events to the configured writer.
func Print(events []Event, opts PrintOptions) {
	out := opts.Out
	if out == nil {
		out = os.Stdout
	}

	filtered := filterEvents(events, opts)

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tPORT\tEVENT")
	for _, e := range filtered {
		fmt.Fprintf(w, "%s\t%d\t%s\n",
			e.Timestamp.Format(time.RFC3339),
			e.Port,
			e.Kind,
		)
	}
	w.Flush()
}

func filterEvents(events []Event, opts PrintOptions) []Event {
	var out []Event
	for _, e := range events {
		if !opts.Since.IsZero() && e.Timestamp.Before(opts.Since) {
			continue
		}
		out = append(out, e)
	}
	if opts.Limit > 0 && len(out) > opts.Limit {
		out = out[len(out)-opts.Limit:]
	}
	return out
}

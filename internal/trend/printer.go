package trend

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// PrintOptions controls how the trend report is rendered.
type PrintOptions struct {
	Out    io.Writer
	Header bool
}

// Print writes a human-readable trend summary to opts.Out.
func Print(tr *Tracker, opts PrintOptions) {
	samples := tr.Samples()
	dir := tr.Direction()

	if opts.Header {
		fmt.Fprintln(opts.Out, "PORT COUNT TREND")
		fmt.Fprintln(opts.Out, strings.Repeat("-", 40))
	}

	if len(samples) == 0 {
		fmt.Fprintln(opts.Out, "No data recorded yet.")
		return
	}

	for _, s := range samples {
		fmt.Fprintf(opts.Out, "  %s  %d ports\n",
			s.At.Format(time.RFC3339), s.Count)
	}

	icon := directionIcon(dir)
	fmt.Fprintf(opts.Out, "\nOverall trend: %s %s\n", icon, dir)
}

func directionIcon(d Direction) string {
	switch d {
	case Rising:
		return "↑"
	case Falling:
		return "↓"
	default:
		return "→"
	}
}

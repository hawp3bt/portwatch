package silence

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05"
	headerLine  = "NAME\tSTARTS\tENDS\tSTATUS\n"
)

// Print writes a formatted table of silence windows to w.
func Print(out io.Writer, windows []Window, now time.Time) {
	if len(windows) == 0 {
		fmt.Fprintln(out, "no silence windows configured")
		return
	}

	sort.Slice(windows, func(i, j int) bool {
		return windows[i].Start.Before(windows[j].Start)
	})

	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprint(tw, headerLine)
	for _, w := range windows {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			w.Name,
			w.Start.Format(timeFormat),
			w.End.Format(timeFormat),
			statusLabel(w, now),
		)
	}
	_ = tw.Flush()
}

func statusLabel(w Window, now time.Time) string {
	switch {
	case now.Before(w.Start):
		return "pending"
	case w.Active(now):
		return "active"
	default:
		return "expired"
	}
}

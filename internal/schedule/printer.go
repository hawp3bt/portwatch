package schedule

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

// Print writes a formatted table of schedule entries to w.
func Print(w io.Writer, entries []*Entry) {
	if len(entries) == 0 {
		fmt.Fprintln(w, "no schedules configured")
		return
	}

	// Sort by name for deterministic output.
	sorted := make([]*Entry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tINTERVAL\tENABLED")
	fmt.Fprintln(tw, "----\t--------\t-------")
	for _, e := range sorted {
		status := enabledLabel(e.Enabled)
		fmt.Fprintf(tw, "%s\t%s\t%s\n", e.Name, e.Interval, status)
	}
	_ = tw.Flush()
}

func enabledLabel(enabled bool) string {
	if enabled {
		return "yes"
	}
	return "no"
}

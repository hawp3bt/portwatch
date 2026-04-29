package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/user/portwatch/internal/audit"
)

const defaultAuditPath = ".portwatch/audit.log"

func runAudit(args []string) {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	path := fs.String("path", defaultAuditPath, "path to audit log")
	kindFilter := fs.String("kind", "", "filter by event kind (scan, alert, suppress, baseline, profile, threshold)")
	limit := fs.Int("n", 50, "max entries to display (0 = all)")
	_ = fs.Parse(args)

	entries, err := audit.Load(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "audit: %v\n", err)
		os.Exit(1)
	}
	if len(entries) == 0 {
		fmt.Println("No audit entries found.")
		return
	}

	// Apply kind filter.
	if *kindFilter != "" {
		filtered := entries[:0]
		for _, e := range entries {
			if string(e.Kind) == *kindFilter {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	// Apply limit (show most recent).
	if *limit > 0 && len(entries) > *limit {
		entries = entries[len(entries)-*limit:]
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tKIND\tACTOR\tMESSAGE")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			e.Timestamp.Format(time.RFC3339),
			e.Kind,
			e.Actor,
			e.Message,
		)
	}
	w.Flush()
}

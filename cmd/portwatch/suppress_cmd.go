package main

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/user/portwatch/internal/suppress"
)

// sharedSuppressor is the process-wide suppressor instance shared with the
// monitor loop. In a real daemon this would live in the monitor struct;
// here it is package-level for CLI sub-command access.
var sharedSuppressor = suppress.New()

// runSuppress handles the "suppress" sub-command.
//
//	portwatch suppress add <port> <duration> [reason]
//	portwatch suppress lift <port>
//	portwatch suppress list
func runSuppress(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: portwatch suppress <add|lift|list> [args]")
		os.Exit(1)
	}

	switch args[0] {
	case "add":
		runSuppressAdd(args[1:])
	case "lift":
		runSuppressLift(args[1:])
	case "list":
		runSuppressList()
	default:
		fmt.Fprintf(os.Stderr, "unknown suppress sub-command: %s\n", args[0])
		os.Exit(1)
	}
}

func runSuppressAdd(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: portwatch suppress add <port> <duration> [reason]")
		os.Exit(1)
	}
	port, err := strconv.Atoi(args[0])
	if err != nil || port < 1 || port > 65535 {
		fmt.Fprintf(os.Stderr, "invalid port: %s\n", args[0])
		os.Exit(1)
	}
	dur, err := time.ParseDuration(args[1])
	if err != nil || dur <= 0 {
		fmt.Fprintf(os.Stderr, "invalid duration: %s\n", args[1])
		os.Exit(1)
	}
	reason := "manual"
	if len(args) >= 3 {
		reason = args[2]
	}
	sharedSuppressor.Suppress(port, reason, dur)
	fmt.Printf("suppressed port %d for %s (%s)\n", port, dur, reason)
}

func runSuppressLift(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: portwatch suppress lift <port>")
		os.Exit(1)
	}
	port, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid port: %s\n", args[0])
		os.Exit(1)
	}
	sharedSuppressor.Lift(port)
	fmt.Printf("lifted suppression for port %d\n", port)
}

func runSuppressList() {
	entries := sharedSuppressor.List()
	if len(entries) == 0 {
		fmt.Println("no active suppressions")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PORT\tREASON\tEXPIRES")
	for _, e := range entries {
		fmt.Fprintf(w, "%d\t%s\t%s\n", e.Port, e.Reason, e.ExpiresAt.Format(time.RFC3339))
	}
	w.Flush()
}

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/user/portwatch/internal/quota"
)

// runQuota handles the `portwatch quota` sub-command family.
// Usage:
//
//	portwatch quota check <port>
//	portwatch quota reset <port>
//	portwatch quota remaining <port>
func runQuota(args []string, q *quota.Quota) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: portwatch quota <check|reset|remaining> <port>")
		os.Exit(1)
	}

	subcmd := args[0]
	port, err := strconv.Atoi(args[1])
	if err != nil || port < 1 || port > 65535 {
		fmt.Fprintf(os.Stderr, "invalid port: %s\n", args[1])
		os.Exit(1)
	}

	switch subcmd {
	case "check":
		if q.Allow(port) {
			fmt.Printf("port %d: alert permitted\n", port)
		} else {
			fmt.Printf("port %d: quota exceeded, alert suppressed\n", port)
		}

	case "reset":
		q.Reset(port)
		fmt.Printf("port %d: quota reset\n", port)

	case "remaining":
		fmt.Printf("port %d: %d alerts remaining in current window\n", port, q.Remaining(port))

	default:
		fmt.Fprintf(os.Stderr, "unknown quota sub-command: %s\n", subcmd)
		os.Exit(1)
	}
}

// defaultQuota returns a Quota configured from environment variables or
// sensible defaults (5 alerts per 10 minutes per port).
func defaultQuota() *quota.Quota {
	max := 5
	window := 10 * time.Minute

	if v := os.Getenv("PORTWATCH_QUOTA_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			max = n
		}
	}
	if v := os.Getenv("PORTWATCH_QUOTA_WINDOW"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			window = d
		}
	}
	return quota.New(max, window)
}

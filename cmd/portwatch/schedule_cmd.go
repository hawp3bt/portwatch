package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/user/portwatch/internal/schedule"
)

func runSchedule(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: portwatch schedule <list|set|remove> [args]")
		os.Exit(1)
	}

	sched := schedule.New()

	switch args[0] {
	case "list":
		runScheduleList(sched)
	case "set":
		runScheduleSet(sched, args[1:])
	case "remove":
		runScheduleRemove(sched, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown schedule subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func runScheduleList(s *schedule.Schedule) {
	schedule.Print(os.Stdout, s.List())
}

func runScheduleSet(s *schedule.Schedule, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: portwatch schedule set <name> <interval> [--disabled]")
		os.Exit(1)
	}
	name := args[0]
	d, err := time.ParseDuration(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid interval %q: %v\n", args[1], err)
		os.Exit(1)
	}
	enabled := true
	if len(args) >= 3 {
		if b, e := strconv.ParseBool(args[2]); e == nil {
			enabled = b
		}
	}
	if err := s.Set(name, d, enabled); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("schedule %q set to %s (enabled=%v)\n", name, d, enabled)
}

func runScheduleRemove(s *schedule.Schedule, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: portwatch schedule remove <name>")
		os.Exit(1)
	}
	s.Remove(args[0])
	fmt.Printf("schedule %q removed\n", args[0])
}

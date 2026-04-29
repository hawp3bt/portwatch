package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/user/portwatch/internal/silence"
)

const defaultSilencePath = ".portwatch_silence.json"

func runSilence(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: portwatch silence <list|add|remove> [flags]")
		os.Exit(1)
	}
	switch args[0] {
	case "list":
		runSilenceList(args[1:])
	case "add":
		runSilenceAdd(args[1:])
	case "remove":
		runSilenceRemove(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown silence command: %s\n", args[0])
		os.Exit(1)
	}
}

func runSilenceList(args []string) {
	fs := flag.NewFlagSet("silence list", flag.ExitOnError)
	path := fs.String("file", defaultSilencePath, "silence registry file")
	_ = fs.Parse(args)

	reg, err := silence.New(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	silence.Print(os.Stdout, reg.List(), time.Now())
}

func runSilenceAdd(args []string) {
	fs := flag.NewFlagSet("silence add", flag.ExitOnError)
	path := fs.String("file", defaultSilencePath, "silence registry file")
	name := fs.String("name", "", "window name (required)")
	duration := fs.Duration("duration", time.Hour, "silence duration")
	start := fs.String("start", "", "start time RFC3339 (default: now)")
	_ = fs.Parse(args)

	if *name == "" {
		fmt.Fprintln(os.Stderr, "error: --name is required")
		os.Exit(1)
	}

	var startTime time.Time
	if *start == "" {
		startTime = time.Now()
	} else {
		var err error
		startTime, err = time.Parse(time.RFC3339, *start)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing start time: %v\n", err)
			os.Exit(1)
		}
	}

	reg, err := silence.New(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if err := reg.Add(*name, startTime, startTime.Add(*duration)); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("silence window %q added (duration: %s)\n", *name, *duration)
}

func runSilenceRemove(args []string) {
	fs := flag.NewFlagSet("silence remove", flag.ExitOnError)
	path := fs.String("file", defaultSilencePath, "silence registry file")
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: portwatch silence remove <name>")
		os.Exit(1)
	}
	reg, err := silence.New(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if err := reg.Remove(fs.Arg(0)); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("silence window %q removed\n", fs.Arg(0))
}

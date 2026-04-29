package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"github.com/user/portwatch/internal/threshold"
)

func defaultThresholdPath() string {
	if p := os.Getenv("PORTWATCH_THRESHOLD_FILE"); p != "" {
		return p
	}
	return filepath.Join(os.Getenv("HOME"), ".portwatch", "thresholds.json")
}

func runThreshold(args []string) {
	fs := flag.NewFlagSet("threshold", flag.ExitOnError)
	file := fs.String("file", defaultThresholdPath(), "path to threshold rules file")
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: portwatch threshold <list|set|remove|check>")
		os.Exit(1)
	}

	reg, err := threshold.New(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "threshold: %v\n", err)
		os.Exit(1)
	}

	switch fs.Arg(0) {
	case "list":
		runThresholdList(reg)
	case "set":
		runThresholdSet(reg, fs.Args()[1:])
	case "remove":
		runThresholdRemove(reg, fs.Args()[1:])
	case "check":
		runThresholdCheck(reg, fs.Args()[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", fs.Arg(0))
		os.Exit(1)
	}
}

func runThresholdList(reg *threshold.Registry) {
	rules := reg.List()
	if len(rules) == 0 {
		fmt.Println("no threshold rules defined")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PORT\tLABEL\tMIN_OPEN\tMAX_OPEN")
	for _, r := range rules {
		fmt.Fprintf(w, "%d\t%s\t%d\t%d\n", r.Port, r.Label, r.MinOpen, r.MaxOpen)
	}
	_ = w.Flush()
}

func runThresholdSet(reg *threshold.Registry, args []string) {
	fs := flag.NewFlagSet("threshold set", flag.ExitOnError)
	max := fs.Int("max", 0, "max open count (0 = disabled)")
	min := fs.Int("min", -1, "min open count (-1 = disabled)")
	label := fs.String("label", "", "human-readable label")
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: threshold set [flags] <port>")
		os.Exit(1)
	}
	port, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid port: %v\n", err)
		os.Exit(1)
	}
	if err := reg.Set(threshold.Rule{Port: port, MaxOpen: *max, MinOpen: *min, Label: *label}); err != nil {
		fmt.Fprintf(os.Stderr, "set: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("threshold rule set for port %d\n", port)
}

func runThresholdRemove(reg *threshold.Registry, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: threshold remove <port>")
		os.Exit(1)
	}
	port, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid port: %v\n", err)
		os.Exit(1)
	}
	if err := reg.Remove(port); err != nil {
		fmt.Fprintf(os.Stderr, "remove: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("threshold rule removed for port %d\n", port)
}

func runThresholdCheck(reg *threshold.Registry, args []string) {
	counts := make(map[int]int)
	for _, a := range args {
		var port, count int
		if _, err := fmt.Sscanf(a, "%d=%d", &port, &count); err != nil {
			fmt.Fprintf(os.Stderr, "invalid count spec %q (want port=count)\n", a)
			os.Exit(1)
		}
		counts[port] = count
	}
	violations := reg.Check(counts)
	if len(violations) == 0 {
		fmt.Println("no threshold violations")
		return
	}
	for _, v := range violations {
		fmt.Println(v)
	}
	os.Exit(2)
}

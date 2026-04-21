package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/snapshot"
)

const defaultSnapshotDir = ".portwatch/snapshots"

// runSnapshot handles the "snapshot" subcommand.
// Usage:
//
//	portwatch snapshot save <name> [label]
//	portwatch snapshot load <name>
//	portwatch snapshot list
func runSnapshot(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: portwatch snapshot <save|load|list> [name] [label]")
		os.Exit(1)
	}

	m := snapshot.New(defaultSnapshotDir)

	switch args[0] {
	case "save":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch snapshot save <name> [label]")
			os.Exit(1)
		}
		name := args[1]
		label := ""
		if len(args) >= 3 {
			label = strings.Join(args[2:], " ")
		}
		ports, err := scanner.Scan(1, 65535)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
			os.Exit(1)
		}
		open := scanner.OpenPorts(ports)
		if err := m.Save(name, open, label); err != nil {
			fmt.Fprintf(os.Stderr, "save error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("snapshot %q saved (%d open ports)\n", name, len(open))

	case "load":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch snapshot load <name>")
			os.Exit(1)
		}
		e, err := m.Load(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "load error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("snapshot: %s  captured: %s\n", args[1], e.CapturedAt.Format("2006-01-02 15:04:05"))
		if e.Label != "" {
			fmt.Printf("label: %s\n", e.Label)
		}
		fmt.Printf("open ports (%d): %v\n", len(e.Ports), e.Ports)

	case "list":
		names, err := m.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "list error: %v\n", err)
			os.Exit(1)
		}
		if len(names) == 0 {
			fmt.Println("no snapshots saved")
			return
		}
		fmt.Printf("%-20s\n", "NAME")
		for _, n := range names {
			fmt.Printf("%-20s\n", strings.TrimSuffix(n, ".json"))
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown snapshot command: %s\n", args[0])
		os.Exit(1)
	}
}

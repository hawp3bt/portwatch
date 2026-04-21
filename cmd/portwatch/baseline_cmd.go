package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/yourorg/portwatch/internal/baseline"
	"github.com/yourorg/portwatch/internal/scanner"
)

const defaultBaselinePath = ".portwatch_baselines.json"

// runBaseline handles the `portwatch baseline` subcommand.
//
// Usage:
//
//	portwatch baseline save [name]   – capture current open ports as a named baseline
//	portwatch baseline list          – list all saved baselines
//	portwatch baseline diff [name]   – compare current ports against a saved baseline
func runBaseline(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: portwatch baseline <save|list|diff> [name]")
	}

	store := baseline.New(defaultBaselinePath)

	switch args[0] {
	case "save":
		name := "default"
		if len(args) >= 2 {
			name = args[1]
		}
		ports, err := scanner.Scan(1, 65535)
		if err != nil {
			return fmt.Errorf("baseline save: scan: %w", err)
		}
		open := scanner.OpenPorts(ports)
		if err := store.Save(name, open); err != nil {
			return fmt.Errorf("baseline save: %w", err)
		}
		fmt.Fprintf(os.Stdout, "baseline %q saved (%d ports)\n", name, len(open))

	case "list":
		list, err := store.List()
		if err != nil {
			return fmt.Errorf("baseline list: %w", err)
		}
		if len(list) == 0 {
			fmt.Println("no baselines saved")
			return nil
		}
		for _, s := range list {
			fmt.Printf("  %-20s  %d ports  (saved %s)\n", s.Name, len(s.Ports), s.CreatedAt.Format("2006-01-02 15:04:05"))
		}

	case "diff":
		name := "default"
		if len(args) >= 2 {
			name = args[1]
		}
		snap, err := store.Load(name)
		if err != nil {
			return fmt.Errorf("baseline diff: %w", err)
		}
		current, err := scanner.Scan(1, 65535)
		if err != nil {
			return fmt.Errorf("baseline diff: scan: %w", err)
		}
		open := scanner.OpenPorts(current)
		printBaselineDiff(snap.Ports, open)

	default:
		return fmt.Errorf("unknown baseline subcommand %q", args[0])
	}
	return nil
}

func printBaselineDiff(saved, current []int) {
	savedSet := make(map[int]struct{}, len(saved))
	for _, p := range saved {
		savedSet[p] = struct{}{}
	}
	currentSet := make(map[int]struct{}, len(current))
	for _, p := range current {
		currentSet[p] = struct{}{}
	}

	var added, removed []string
	for _, p := range current {
		if _, ok := savedSet[p]; !ok {
			added = append(added, strconv.Itoa(p))
		}
	}
	for _, p := range saved {
		if _, ok := currentSet[p]; !ok {
			removed = append(removed, strconv.Itoa(p))
		}
	}
	sort.Strings(added)
	sort.Strings(removed)

	if len(added) == 0 && len(removed) == 0 {
		fmt.Println("no changes from baseline")
		return
	}
	if len(added) > 0 {
		fmt.Printf("+ opened: %s\n", strings.Join(added, ", "))
	}
	if len(removed) > 0 {
		fmt.Printf("- closed: %s\n", strings.Join(removed, ", "))
	}
}

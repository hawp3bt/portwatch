package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/profile"
	"github.com/user/portwatch/internal/scanner"
)

const defaultProfileDir = ".portwatch/profiles"

func runProfile(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: portwatch profile <save|load|delete|list|check> [name] [ports...]")
	}
	r, err := profile.New(defaultProfileDir)
	if err != nil {
		return err
	}
	switch args[0] {
	case "save":
		return runProfileSave(r, args[1:])
	case "load":
		return runProfileLoad(r, args[1:])
	case "delete":
		return runProfileDelete(r, args[1:])
	case "list":
		return runProfileList(r)
	case "check":
		return runProfileCheck(r, args[1:])
	default:
		return fmt.Errorf("unknown profile subcommand: %q", args[0])
	}
}

func runProfileSave(r *profile.Registry, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: portwatch profile save <name> [port...]")
	}
	name := args[0]
	var ports []int
	if len(args) > 1 {
		for _, s := range args[1:] {
			p, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("invalid port %q: %w", s, err)
			}
			ports = append(ports, p)
		}
	} else {
		// Capture current open ports if none specified.
		var err error
		ports, err = scanner.OpenPorts(scanner.Scan(1, 65535))
		if err != nil {
			return fmt.Errorf("scan: %w", err)
		}
	}
	if err := r.Save(name, ports); err != nil {
		return err
	}
	fmt.Printf("profile %q saved with %d port(s)\n", name, len(ports))
	return nil
}

func runProfileLoad(r *profile.Registry, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: portwatch profile load <name>")
	}
	p, err := r.Load(args[0])
	if err != nil {
		return err
	}
	sort.Ints(p.Ports)
	strs := make([]string, len(p.Ports))
	for i, port := range p.Ports {
		strs[i] = strconv.Itoa(port)
	}
	fmt.Printf("profile: %s\nports:   %s\ncreated: %s\n",
		p.Name, strings.Join(strs, ", "), p.CreatedAt.Format("2006-01-02 15:04:05"))
	return nil
}

func runProfileDelete(r *profile.Registry, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: portwatch profile delete <name>")
	}
	if err := r.Delete(args[0]); err != nil {
		return err
	}
	fmt.Printf("profile %q deleted\n", args[0])
	return nil
}

func runProfileList(r *profile.Registry) error {
	names, err := r.List()
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Println("no profiles saved")
		return nil
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Println(n)
	}
	return nil
}

func runProfileCheck(r *profile.Registry, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: portwatch profile check <name>")
	}
	p, err := r.Load(args[0])
	if err != nil {
		return err
	}
	current, err := scanner.OpenPorts(scanner.Scan(1, 65535))
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}
	missing, extra := profile.Diff(p.Ports, current)
	if len(missing) == 0 && len(extra) == 0 {
		fmt.Printf("profile %q matches current open ports\n", p.Name)
		return nil
	}
	if len(missing) > 0 {
		sort.Ints(missing)
		fmt.Fprintf(os.Stdout, "MISSING (in profile, not open): %v\n", missing)
	}
	if len(extra) > 0 {
		sort.Ints(extra)
		fmt.Fprintf(os.Stdout, "EXTRA   (open, not in profile): %v\n", extra)
	}
	return nil
}

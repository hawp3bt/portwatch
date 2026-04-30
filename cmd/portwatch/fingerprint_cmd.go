package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/user/portwatch/internal/fingerprint"
	"github.com/user/portwatch/internal/scanner"
)

func runFingerprint(args []string) {
	if len(args) > 0 && args[0] == "compare" {
		runFingerprintCompare(args[1:])
		return
	}

	ports, err := scanner.OpenPorts(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fingerprint: scan error: %v\n", err)
		os.Exit(1)
	}

	f := fingerprint.Compute(ports)

	if len(args) > 0 && args[0] == "--json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(f)
		return
	}

	fmt.Printf("fingerprint : %s\n", f.Short())
	fmt.Printf("full hash   : %s\n", f.Hash)
	fmt.Printf("ports       : %d open\n", len(f.Ports))
	fmt.Printf("computed at : %s\n", f.ComputedAt.Format("2006-01-02 15:04:05 UTC"))
}

func runFingerprintCompare(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: fingerprint compare <hash-or-port-list-a> <hash-or-port-list-b>")
		os.Exit(1)
	}

	a := parseFingerprintArg(args[0])
	b := parseFingerprintArg(args[1])

	if fingerprint.Equal(a, b) {
		fmt.Println("MATCH — port sets are identical")
		fmt.Printf("  hash: %s\n", a.Short())
	} else {
		fmt.Println("CHANGED — port sets differ")
		fmt.Printf("  a: %s  (%d ports)\n", a.Short(), len(a.Ports))
		fmt.Printf("  b: %s  (%d ports)\n", b.Short(), len(b.Ports))
	}
}

// parseFingerprintArg treats the argument as a comma-separated list of port
// numbers and computes a fingerprint from them.
func parseFingerprintArg(s string) fingerprint.Fingerprint {
	var ports []int
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			token := s[start:i]
			if p, err := strconv.Atoi(token); err == nil {
				ports = append(ports, p)
			}
			start = i + 1
		}
	}
	return fingerprint.Compute(ports)
}

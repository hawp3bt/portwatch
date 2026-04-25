package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/portwatch/internal/digest"
	"github.com/user/portwatch/internal/snapshot"
)

// runDigest handles the `portwatch digest` sub-command.
// Usage:
//
//	portwatch digest                  – digest of the live scan stored in the default snapshot
//	portwatch digest <snapshot-name>  – digest of a named snapshot
//	portwatch digest --compare <a> <b> – compare two named snapshots
func runDigest(args []string) {
	snapshotDir := filepath.Join(os.TempDir(), "portwatch", "snapshots")
	store := snapshot.New(snapshotDir)

	if len(args) >= 2 && args[0] == "--compare" {
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: portwatch digest --compare <snapshot-a> <snapshot-b>")
			os.Exit(1)
		}
		runDigestCompare(store, args[1], args[2])
		return
	}

	name := "latest"
	if len(args) == 1 {
		name = args[0]
	}

	snap, err := store.Load(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading snapshot %q: %v\n", name, err)
		os.Exit(1)
	}

	d := digest.Compute(snap.Ports)
	fmt.Printf("snapshot : %s\n", name)
	fmt.Printf("ports    : %d\n", len(snap.Ports))
	fmt.Printf("digest   : %s\n", d)
	fmt.Printf("taken at : %s\n", snap.TakenAt.Format(time.RFC3339))
}

func runDigestCompare(store *snapshot.Store, nameA, nameB string) {
	loadOrExit := func(name string) *snapshot.Snapshot {
		s, err := store.Load(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading snapshot %q: %v\n", name, err)
			os.Exit(1)
		}
		return s
	}

	snapA := loadOrExit(nameA)
	snapB := loadOrExit(nameB)

	dA := digest.Compute(snapA.Ports)
	dB := digest.Compute(snapB.Ports)
	changed := !digest.Equal(dA, dB)

	result := map[string]interface{}{
		"snapshot_a": nameA,
		"digest_a":   dA.String(),
		"snapshot_b": nameB,
		"digest_b":   dB.String(),
		"changed":    changed,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "encode error: %v\n", err)
		os.Exit(1)
	}
}

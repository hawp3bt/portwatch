package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/portwatch/internal/snapshot"
)

func tempDir(t *testing.T) string {
	t.Helper()
	d, err := os.MkdirTemp("", "snapshot-test-*")
	if err != nil {
		t.Fatalf("tempDir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(d) })
	return d
}

func TestSave_AndLoad_RoundTrip(t *testing.T) {
	m := snapshot.New(tempDir(t))
	ports := []int{22, 80, 443}
	if err := m.Save("test", ports, "initial"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	e, err := m.Load("test")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(e.Ports) != len(ports) {
		t.Fatalf("expected %d ports, got %d", len(ports), len(e.Ports))
	}
	for i, p := range ports {
		if e.Ports[i] != p {
			t.Errorf("port[%d]: expected %d, got %d", i, p, e.Ports[i])
		}
	}
	if e.Label != "initial" {
		t.Errorf("label: expected %q, got %q", "initial", e.Label)
	}
	if e.CapturedAt.IsZero() {
		t.Error("CapturedAt should not be zero")
	}
}

func TestLoad_MissingSnapshot_ReturnsError(t *testing.T) {
	m := snapshot.New(tempDir(t))
	_, err := m.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing snapshot")
	}
}

func TestSave_OverwritesExisting(t *testing.T) {
	m := snapshot.New(tempDir(t))
	_ = m.Save("snap", []int{80}, "first")
	_ = m.Save("snap", []int{443}, "second")
	e, err := m.Load("snap")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(e.Ports) != 1 || e.Ports[0] != 443 {
		t.Errorf("expected [443], got %v", e.Ports)
	}
	if e.Label != "second" {
		t.Errorf("expected label %q, got %q", "second", e.Label)
	}
}

func TestList_ReturnsAllSnapshots(t *testing.T) {
	dir := tempDir(t)
	m := snapshot.New(dir)
	for _, name := range []string{"alpha", "beta", "gamma"} {
		if err := m.Save(name, []int{8080}, ""); err != nil {
			t.Fatalf("Save %s: %v", name, err)
		}
	}
	names, err := m.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 3 {
		t.Errorf("expected 3 snapshots, got %d: %v", len(names), names)
	}
	// ensure files exist
	for _, n := range names {
		if filepath.Ext(n) != ".json" {
			t.Errorf("expected .json extension, got %q", n)
		}
	}
}

func TestList_EmptyDir_ReturnsNil(t *testing.T) {
	m := snapshot.New(tempDir(t))
	names, err := m.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected empty list, got %v", names)
	}
}

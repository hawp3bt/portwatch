package state_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/portwatch/internal/state"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "state.json")
}

func TestSave_AndLoad_RoundTrip(t *testing.T) {
	store := state.New(tempPath(t))
	ports := []int{22, 80, 443}

	if err := store.Save(ports); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	snap, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(snap.Ports) != len(ports) {
		t.Fatalf("expected %d ports, got %d", len(ports), len(snap.Ports))
	}
	for i, p := range ports {
		if snap.Ports[i] != p {
			t.Errorf("port[%d]: expected %d, got %d", i, p, snap.Ports[i])
		}
	}
	if snap.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestLoad_MissingFile_ReturnsEmptySnapshot(t *testing.T) {
	store := state.New(tempPath(t))

	snap, err := store.Load()
	if err != nil {
		t.Fatalf("Load() on missing file should not error, got: %v", err)
	}
	if len(snap.Ports) != 0 {
		t.Errorf("expected empty ports, got %v", snap.Ports)
	}
}

func TestExists_ReturnsFalseWhenMissing(t *testing.T) {
	store := state.New(tempPath(t))
	if store.Exists() {
		t.Error("Exists() should return false for missing file")
	}
}

func TestExists_ReturnsTrueAfterSave(t *testing.T) {
	store := state.New(tempPath(t))
	_ = store.Save([]int{8080})
	if !store.Exists() {
		t.Error("Exists() should return true after Save()")
	}
}

func TestLoad_CorruptFile_ReturnsError(t *testing.T) {
	path := tempPath(t)
	_ = os.WriteFile(path, []byte("not-json{"), 0o644)
	store := state.New(path)
	_, err := store.Load()
	if err == nil {
		t.Error("expected error for corrupt JSON, got nil")
	}
}

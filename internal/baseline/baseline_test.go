package baseline_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/portwatch/internal/baseline"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "baselines.json")
}

func TestSave_AndLoad_RoundTrip(t *testing.T) {
	s := baseline.New(tempPath(t))
	ports := []int{22, 80, 443}

	if err := s.Save("default", ports); err != nil {
		t.Fatalf("Save: %v", err)
	}

	snap, err := s.Load("default")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if snap.Name != "default" {
		t.Errorf("Name = %q, want %q", snap.Name, "default")
	}
	if len(snap.Ports) != len(ports) {
		t.Fatalf("Ports len = %d, want %d", len(snap.Ports), len(ports))
	}
	for i, p := range ports {
		if snap.Ports[i] != p {
			t.Errorf("Ports[%d] = %d, want %d", i, snap.Ports[i], p)
		}
	}
}

func TestLoad_MissingName_ReturnsError(t *testing.T) {
	s := baseline.New(tempPath(t))
	_, err := s.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing baseline, got nil")
	}
}

func TestSave_OverwritesExisting(t *testing.T) {
	s := baseline.New(tempPath(t))
	_ = s.Save("prod", []int{80, 443})
	_ = s.Save("prod", []int{8080})

	snap, err := s.Load("prod")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(snap.Ports) != 1 || snap.Ports[0] != 8080 {
		t.Errorf("expected [8080], got %v", snap.Ports)
	}
}

func TestList_ReturnsAllBaselines(t *testing.T) {
	s := baseline.New(tempPath(t))
	_ = s.Save("a", []int{22})
	_ = s.Save("b", []int{80})

	list, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List len = %d, want 2", len(list))
	}
}

func TestLoad_MissingFile_ReturnsError(t *testing.T) {
	s := baseline.New("/nonexistent/path/baselines.json")
	_, err := s.Load("anything")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestSave_CreatesFile(t *testing.T) {
	path := tempPath(t)
	s := baseline.New(path)
	if err := s.Save("init", []int{443}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}

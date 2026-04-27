package profile_test

import (
	"os"
	"testing"

	"github.com/user/portwatch/internal/profile"
)

// TestSave_EmptyName_ReturnsError ensures the registry rejects blank names.
func TestSave_EmptyName_ReturnsError(t *testing.T) {
	r, _ := profile.New(tempDir(t))
	if err := r.Save("", []int{80}); err == nil {
		t.Fatal("expected error for empty name")
	}
}

// TestDelete_MissingProfile_ReturnsError ensures Delete surfaces the error.
func TestDelete_MissingProfile_ReturnsError(t *testing.T) {
	r, _ := profile.New(tempDir(t))
	if err := r.Delete("nonexistent"); err == nil {
		t.Fatal("expected error when deleting missing profile")
	}
}

// TestSave_OverwritesExisting verifies that ports are updated on re-save.
func TestSave_OverwritesExisting(t *testing.T) {
	r, _ := profile.New(tempDir(t))
	_ = r.Save("web", []int{80})
	_ = r.Save("web", []int{80, 443})
	p, err := r.Load("web")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(p.Ports) != 2 {
		t.Errorf("expected 2 ports after overwrite, got %d", len(p.Ports))
	}
}

// TestList_IgnoresNonJSONFiles ensures non-profile files are not returned.
func TestList_IgnoresNonJSONFiles(t *testing.T) {
	dir := tempDir(t)
	r, _ := profile.New(dir)
	_ = r.Save("alpha", []int{22})
	// Write a non-JSON file into the profile directory.
	_ = os.WriteFile(dir+"/README.txt", []byte("ignore me"), 0o644)
	names, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 1 || names[0] != "alpha" {
		t.Errorf("expected only [alpha], got %v", names)
	}
}

// TestDiff_EmptyProfile treats an empty profile as having no expected ports.
func TestDiff_EmptyProfile(t *testing.T) {
	missing, extra := profile.Diff([]int{}, []int{80, 443})
	if len(missing) != 0 {
		t.Errorf("expected no missing ports, got %v", missing)
	}
	if len(extra) != 2 {
		t.Errorf("expected 2 extra ports, got %v", extra)
	}
}

// TestDiff_EmptyCurrent treats no open ports as all profile ports missing.
func TestDiff_EmptyCurrent(t *testing.T) {
	missing, extra := profile.Diff([]int{22, 80}, []int{})
	if len(missing) != 2 {
		t.Errorf("expected 2 missing ports, got %v", missing)
	}
	if len(extra) != 0 {
		t.Errorf("expected no extra ports, got %v", extra)
	}
}

package profile_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/user/portwatch/internal/profile"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "profile-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestSave_AndLoad_RoundTrip(t *testing.T) {
	r, _ := profile.New(tempDir(t))
	ports := []int{80, 443, 8080}
	if err := r.Save("prod", ports); err != nil {
		t.Fatalf("Save: %v", err)
	}
	p, err := r.Load("prod")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.Name != "prod" {
		t.Errorf("name: got %q want %q", p.Name, "prod")
	}
	if len(p.Ports) != len(ports) {
		t.Errorf("ports length: got %d want %d", len(p.Ports), len(ports))
	}
}

func TestSave_PreservesCreatedAt(t *testing.T) {
	r, _ := profile.New(tempDir(t))
	_ = r.Save("dev", []int{22})
	first, _ := r.Load("dev")
	_ = r.Save("dev", []int{22, 80})
	second, _ := r.Load("dev")
	if !first.CreatedAt.Equal(second.CreatedAt) {
		t.Errorf("CreatedAt changed on overwrite: %v -> %v", first.CreatedAt, second.CreatedAt)
	}
}

func TestLoad_MissingProfile_ReturnsError(t *testing.T) {
	r, _ := profile.New(tempDir(t))
	_, err := r.Load("ghost")
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
}

func TestDelete_RemovesProfile(t *testing.T) {
	r, _ := profile.New(tempDir(t))
	_ = r.Save("tmp", []int{9000})
	if err := r.Delete("tmp"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := r.Load("tmp")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestList_ReturnsAllProfiles(t *testing.T) {
	r, _ := profile.New(tempDir(t))
	for _, name := range []string{"a", "b", "c"} {
		_ = r.Save(name, []int{1})
	}
	names, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	sort.Strings(names)
	if len(names) != 3 || names[0] != "a" || names[2] != "c" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestList_EmptyDir_ReturnsNil(t *testing.T) {
	r, _ := profile.New(tempDir(t))
	names, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected empty list, got %v", names)
	}
}

func TestNew_CreatesDir(t *testing.T) {
	base := filepath.Join(os.TempDir(), "pw-profile-newdir-test")
	defer os.RemoveAll(base)
	_, err := profile.New(base)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := os.Stat(base); err != nil {
		t.Errorf("directory not created: %v", err)
	}
}

func TestDiff_MissingAndExtra(t *testing.T) {
	profPorts := []int{80, 443, 22}
	current := []int{80, 8080}
	missing, extra := profile.Diff(profPorts, current)
	has := func(s []int, v int) bool {
		for _, x := range s {
			if x == v {
				return true
			}
		}
		return false
	}
	if !has(missing, 443) || !has(missing, 22) {
		t.Errorf("missing should contain 443 and 22, got %v", missing)
	}
	if !has(extra, 8080) {
		t.Errorf("extra should contain 8080, got %v", extra)
	}
	if has(extra, 80) {
		t.Errorf("extra should not contain 80")
	}
}

func TestDiff_NoDifference(t *testing.T) {
	ports := []int{80, 443}
	missing, extra := profile.Diff(ports, ports)
	if len(missing) != 0 || len(extra) != 0 {
		t.Errorf("expected no diff, got missing=%v extra=%v", missing, extra)
	}
}

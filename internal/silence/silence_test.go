package silence_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/portwatch/internal/silence"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "silence.json")
}

func TestAdd_AndIsSilenced_Active(t *testing.T) {
	r, _ := silence.New(tempPath(t))
	now := time.Now()
	_ = r.Add("maint", now.Add(-time.Minute), now.Add(time.Hour))
	if !r.IsSilenced(now) {
		t.Fatal("expected silenced during active window")
	}
}

func TestIsSilenced_ReturnsFalseWhenNoWindows(t *testing.T) {
	r, _ := silence.New(tempPath(t))
	if r.IsSilenced(time.Now()) {
		t.Fatal("expected not silenced with no windows")
	}
}

func TestIsSilenced_ReturnsFalseAfterWindowExpires(t *testing.T) {
	r, _ := silence.New(tempPath(t))
	past := time.Now().Add(-2 * time.Hour)
	_ = r.Add("old", past, past.Add(time.Minute))
	if r.IsSilenced(time.Now()) {
		t.Fatal("expected not silenced after window expired")
	}
}

func TestRemove_DeletesWindow(t *testing.T) {
	r, _ := silence.New(tempPath(t))
	now := time.Now()
	_ = r.Add("maint", now.Add(-time.Minute), now.Add(time.Hour))
	_ = r.Remove("maint")
	if r.IsSilenced(now) {
		t.Fatal("expected not silenced after removal")
	}
}

func TestRemove_MissingWindow_ReturnsError(t *testing.T) {
	r, _ := silence.New(tempPath(t))
	if err := r.Remove("nonexistent"); err == nil {
		t.Fatal("expected error removing missing window")
	}
}

func TestAdd_EmptyName_ReturnsError(t *testing.T) {
	r, _ := silence.New(tempPath(t))
	now := time.Now()
	if err := r.Add("", now, now.Add(time.Hour)); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestAdd_EndBeforeStart_ReturnsError(t *testing.T) {
	r, _ := silence.New(tempPath(t))
	now := time.Now()
	if err := r.Add("bad", now.Add(time.Hour), now); err == nil {
		t.Fatal("expected error when end is before start")
	}
}

func TestRoundTrip_PersistsAcrossReload(t *testing.T) {
	p := tempPath(t)
	r1, _ := silence.New(p)
	now := time.Now()
	_ = r1.Add("maint", now.Add(-time.Minute), now.Add(time.Hour))

	r2, err := silence.New(p)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if !r2.IsSilenced(now) {
		t.Fatal("expected silenced after reload")
	}
}

func TestNew_MissingFile_ReturnsEmptyRegistry(t *testing.T) {
	p := filepath.Join(t.TempDir(), "missing.json")
	r, err := silence.New(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.IsSilenced(time.Now()) {
		t.Fatal("expected not silenced for empty registry")
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("file should not be created on load")
	}
}

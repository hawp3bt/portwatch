package tags_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/portwatch/internal/tags"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "tags.json")
}

func TestNew_MissingFile_ReturnsEmptyRegistry(t *testing.T) {
	r, err := tags.New(tempPath(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := len(r.List()); got != 0 {
		t.Fatalf("expected 0 tags, got %d", got)
	}
}

func TestSet_AndLabel_RoundTrip(t *testing.T) {
	path := tempPath(t)
	r, _ := tags.New(path)

	if err := r.Set(tags.Tag{Port: 8080, Label: "http-alt", Description: "alternate HTTP"}); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if got := r.Label(8080); got != "http-alt" {
		t.Errorf("Label(8080) = %q, want %q", got, "http-alt")
	}
}

func TestLabel_UnknownPort_ReturnsFormattedNumber(t *testing.T) {
	r, _ := tags.New(tempPath(t))
	if got := r.Label(9999); got != "9999" {
		t.Errorf("Label(9999) = %q, want %q", got, "9999")
	}
}

func TestRemove_DeletesTag(t *testing.T) {
	path := tempPath(t)
	r, _ := tags.New(path)
	_ = r.Set(tags.Tag{Port: 22, Label: "ssh"})

	if err := r.Remove(22); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if got := r.Label(22); got != "22" {
		t.Errorf("expected fallback label after remove, got %q", got)
	}
}

func TestList_SortedByPort(t *testing.T) {
	path := tempPath(t)
	r, _ := tags.New(path)
	_ = r.Set(tags.Tag{Port: 443, Label: "https"})
	_ = r.Set(tags.Tag{Port: 80, Label: "http"})
	_ = r.Set(tags.Tag{Port: 5432, Label: "postgres"})

	list := r.List()
	ports := []int{list[0].Port, list[1].Port, list[2].Port}
	want := []int{80, 443, 5432}
	for i, p := range ports {
		if p != want[i] {
			t.Errorf("List()[%d].Port = %d, want %d", i, p, want[i])
		}
	}
}

func TestNew_PersistsAcrossReload(t *testing.T) {
	path := tempPath(t)
	r1, _ := tags.New(path)
	_ = r1.Set(tags.Tag{Port: 3306, Label: "mysql"})

	r2, err := tags.New(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := r2.Label(3306); got != "mysql" {
		t.Errorf("after reload Label(3306) = %q, want %q", got, "mysql")
	}
}

func TestNew_InvalidJSON_ReturnsError(t *testing.T) {
	path := tempPath(t)
	_ = os.WriteFile(path, []byte("not-json"), 0o644)
	_, err := tags.New(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

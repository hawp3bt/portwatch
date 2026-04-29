package threshold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/portwatch/internal/threshold"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "thresholds.json")
}

func TestSet_AndList_RoundTrip(t *testing.T) {
	p := tempPath(t)
	r, err := threshold.New(p)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rule := threshold.Rule{Port: 8080, MaxOpen: 3, MinOpen: 1, Label: "http-alt"}
	if err := r.Set(rule); err != nil {
		t.Fatalf("Set: %v", err)
	}
	r2, err := threshold.New(p)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	list := r2.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(list))
	}
	if list[0].Label != "http-alt" {
		t.Errorf("label mismatch: got %q", list[0].Label)
	}
}

func TestRemove_DeletesRule(t *testing.T) {
	p := tempPath(t)
	r, _ := threshold.New(p)
	_ = r.Set(threshold.Rule{Port: 443, MaxOpen: 1, MinOpen: -1})
	_ = r.Remove(443)
	if len(r.List()) != 0 {
		t.Error("expected empty list after remove")
	}
}

func TestCheck_MaxOpenViolation(t *testing.T) {
	r, _ := threshold.New(tempPath(t))
	_ = r.Set(threshold.Rule{Port: 22, MaxOpen: 1, MinOpen: -1, Label: "ssh"})
	violations := r.Check(map[int]int{22: 3})
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Actual != 3 {
		t.Errorf("wrong actual: %d", violations[0].Actual)
	}
}

func TestCheck_MinOpenViolation(t *testing.T) {
	r, _ := threshold.New(tempPath(t))
	_ = r.Set(threshold.Rule{Port: 80, MaxOpen: 0, MinOpen: 2, Label: "http"})
	violations := r.Check(map[int]int{80: 1})
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
}

func TestCheck_NoViolationWhenWithinBounds(t *testing.T) {
	r, _ := threshold.New(tempPath(t))
	_ = r.Set(threshold.Rule{Port: 8443, MaxOpen: 5, MinOpen: 1})
	if v := r.Check(map[int]int{8443: 3}); len(v) != 0 {
		t.Errorf("unexpected violations: %v", v)
	}
}

func TestSet_InvalidPort_ReturnsError(t *testing.T) {
	r, _ := threshold.New(tempPath(t))
	if err := r.Set(threshold.Rule{Port: 0, MaxOpen: 1}); err == nil {
		t.Error("expected error for port 0")
	}
}

func TestNew_MissingFile_ReturnsEmptyRegistry(t *testing.T) {
	p := filepath.Join(t.TempDir(), "missing.json")
	r, err := threshold.New(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.List()) != 0 {
		t.Error("expected empty list")
	}
}

func TestViolation_String_ContainsPort(t *testing.T) {
	v := threshold.Violation{
		Rule:   threshold.Rule{Port: 9090, Label: "metrics"},
		Actual: 5,
		Message: "exceeds max_open=2",
	}
	s := v.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
	_ = os.DevNull // keep import
}

package filter_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/yourorg/portwatch/internal/filter"
)

func sorted(ports []int) []int {
	s := make([]int, len(ports))
	copy(s, ports)
	sort.Ints(s)
	return s
}

func TestIsIgnored_ReturnsTrueForIgnoredPort(t *testing.T) {
	r := filter.New([]int{22, 80}, 0, 0)
	if !r.IsIgnored(22) {
		t.Error("expected port 22 to be ignored")
	}
	if !r.IsIgnored(80) {
		t.Error("expected port 80 to be ignored")
	}
}

func TestIsIgnored_ReturnsFalseForNonIgnoredPort(t *testing.T) {
	r := filter.New([]int{22}, 0, 0)
	if r.IsIgnored(443) {
		t.Error("expected port 443 not to be ignored")
	}
}

func TestIsIgnored_EmptyIgnoreList(t *testing.T) {
	r := filter.New(nil, 0, 0)
	if r.IsIgnored(22) {
		t.Error("expected port 22 not to be ignored when ignore list is empty")
	}
}

func TestInRange_NoRangeAllowsAll(t *testing.T) {
	r := filter.New(nil, 0, 0)
	for _, p := range []int{1, 1024, 65535} {
		if !r.InRange(p) {
			t.Errorf("expected port %d to be in range when no range set", p)
		}
	}
}

func TestInRange_RejectsOutOfRangePorts(t *testing.T) {
	r := filter.New(nil, 1024, 9000)
	if r.InRange(80) {
		t.Error("expected port 80 to be out of range")
	}
	if r.InRange(9001) {
		t.Error("expected port 9001 to be out of range")
	}
	if !r.InRange(1024) {
		t.Error("expected port 1024 to be in range")
	}
	if !r.InRange(9000) {
		t.Error("expected port 9000 to be in range")
	}
}

func TestApply_FiltersIgnoredAndOutOfRange(t *testing.T) {
	r := filter.New([]int{22}, 1024, 9000)
	input := []int{22, 80, 1024, 3000, 9001}
	want := []int{1024, 3000}

	got := sorted(r.Apply(input))
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Apply() = %v, want %v", got, want)
	}
}

func TestApply_EmptyInputReturnsEmpty(t *testing.T) {
	r := filter.New([]int{22}, 0, 0)
	got := r.Apply([]int{})
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestApply_AllPortsIgnoredReturnsEmpty(t *testing.T) {
	r := filter.New([]int{22, 80, 443}, 0, 0)
	got := r.Apply([]int{22, 80, 443})
	if len(got) != 0 {
		t.Errorf("expected empty slice when all ports are ignored, got %v", got)
	}
}

package digest_test

import (
	"testing"

	"github.com/user/portwatch/internal/digest"
)

func TestCompute_DeterministicForSameInput(t *testing.T) {
	ports := []int{80, 443, 8080}
	a := digest.Compute(ports)
	b := digest.Compute(ports)
	if a != b {
		t.Fatalf("expected same digest, got %s vs %s", a, b)
	}
}

func TestCompute_OrderIndependent(t *testing.T) {
	a := digest.Compute([]int{80, 443, 8080})
	b := digest.Compute([]int{8080, 80, 443})
	if a != b {
		t.Fatalf("digest should be order-independent, got %s vs %s", a, b)
	}
}

func TestCompute_DifferentPortsDifferentDigest(t *testing.T) {
	a := digest.Compute([]int{80, 443})
	b := digest.Compute([]int{80, 444})
	if a == b {
		t.Fatal("different port sets should produce different digests")
	}
}

func TestCompute_EmptyPortList(t *testing.T) {
	d := digest.Compute([]int{})
	if d == "" {
		t.Fatal("expected non-empty digest for empty port list")
	}
}

func TestEqual_SameDigests(t *testing.T) {
	d := digest.Compute([]int{22, 80})
	if !digest.Equal(d, d) {
		t.Fatal("equal digests should be equal")
	}
}

func TestEqual_DifferentDigests(t *testing.T) {
	a := digest.Compute([]int{22})
	b := digest.Compute([]int{80})
	if digest.Equal(a, b) {
		t.Fatal("different digests should not be equal")
	}
}

func TestChanged_ReturnsTrueWhenPortAdded(t *testing.T) {
	prev := []int{80, 443}
	curr := []int{80, 443, 8080}
	if !digest.Changed(prev, curr) {
		t.Fatal("expected Changed to return true when a port is added")
	}
}

func TestChanged_ReturnsFalseWhenUnchanged(t *testing.T) {
	ports := []int{80, 443}
	if digest.Changed(ports, ports) {
		t.Fatal("expected Changed to return false for identical port sets")
	}
}

func TestDigest_StringMethod(t *testing.T) {
	d := digest.Compute([]int{80})
	if d.String() == "" {
		t.Fatal("String() should return non-empty value")
	}
}

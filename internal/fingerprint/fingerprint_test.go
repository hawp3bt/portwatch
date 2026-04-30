package fingerprint_test

import (
	"testing"

	"github.com/user/portwatch/internal/fingerprint"
)

func TestCompute_DeterministicForSameInput(t *testing.T) {
	a := fingerprint.Compute([]int{80, 443, 8080})
	b := fingerprint.Compute([]int{80, 443, 8080})
	if a.Hash != b.Hash {
		t.Fatalf("expected identical hashes, got %q and %q", a.Hash, b.Hash)
	}
}

func TestCompute_OrderIndependent(t *testing.T) {
	a := fingerprint.Compute([]int{443, 80, 8080})
	b := fingerprint.Compute([]int{8080, 443, 80})
	if a.Hash != b.Hash {
		t.Fatalf("expected same hash regardless of input order, got %q and %q", a.Hash, b.Hash)
	}
}

func TestCompute_DifferentPortsDifferentHash(t *testing.T) {
	a := fingerprint.Compute([]int{80})
	b := fingerprint.Compute([]int{443})
	if a.Hash == b.Hash {
		t.Fatal("expected different hashes for different ports")
	}
}

func TestCompute_EmptyPortList(t *testing.T) {
	f := fingerprint.Compute([]int{})
	if f.Hash == "" {
		t.Fatal("expected non-empty hash for empty port list")
	}
}

func TestEqual_SameFingerprints(t *testing.T) {
	a := fingerprint.Compute([]int{22, 80})
	b := fingerprint.Compute([]int{22, 80})
	if !fingerprint.Equal(a, b) {
		t.Fatal("expected Equal to return true")
	}
}

func TestChanged_DetectsPortSetChange(t *testing.T) {
	prev := fingerprint.Compute([]int{80})
	next := fingerprint.Compute([]int{80, 443})
	if !fingerprint.Changed(prev, next) {
		t.Fatal("expected Changed to return true when ports differ")
	}
}

func TestChanged_ReturnsFalseWhenUnchanged(t *testing.T) {
	prev := fingerprint.Compute([]int{80, 443})
	next := fingerprint.Compute([]int{443, 80})
	if fingerprint.Changed(prev, next) {
		t.Fatal("expected Changed to return false for same port set")
	}
}

func TestShort_ReturnsTwelveChars(t *testing.T) {
	f := fingerprint.Compute([]int{80})
	if len(f.Short()) != 12 {
		t.Fatalf("expected Short() length 12, got %d", len(f.Short()))
	}
}

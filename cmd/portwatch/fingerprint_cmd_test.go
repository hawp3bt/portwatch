package main

import (
	"testing"

	"github.com/user/portwatch/internal/fingerprint"
)

func TestParseFingerprintArg_SinglePort(t *testing.T) {
	f := parseFingerprintArg("80")
	if len(f.Ports) != 1 || f.Ports[0] != 80 {
		t.Fatalf("expected [80], got %v", f.Ports)
	}
}

func TestParseFingerprintArg_MultiplePorts(t *testing.T) {
	f := parseFingerprintArg("80,443,8080")
	if len(f.Ports) != 3 {
		t.Fatalf("expected 3 ports, got %d", len(f.Ports))
	}
}

func TestParseFingerprintArg_OrderNormalised(t *testing.T) {
	a := parseFingerprintArg("443,80")
	b := parseFingerprintArg("80,443")
	if !fingerprint.Equal(a, b) {
		t.Fatal("expected equal fingerprints for same ports in different order")
	}
}

func TestParseFingerprintArg_EmptyString(t *testing.T) {
	f := parseFingerprintArg("")
	// An empty string produces no valid port tokens; result should be stable.
	if f.Hash == "" {
		t.Fatal("expected non-empty hash even for empty input")
	}
}

func TestParseFingerprintArg_IgnoresInvalidTokens(t *testing.T) {
	// "80,abc,443" — "abc" is not a number and must be silently skipped.
	f := parseFingerprintArg("80,abc,443")
	if len(f.Ports) != 2 {
		t.Fatalf("expected 2 valid ports, got %d: %v", len(f.Ports), f.Ports)
	}
}

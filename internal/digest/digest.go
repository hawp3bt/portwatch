// Package digest computes and compares port-set fingerprints so callers
// can cheaply detect whether the open-port landscape has changed between
// two scans without walking every individual diff.
package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// Digest is a short hex string that uniquely represents a set of ports.
type Digest string

// Compute returns a deterministic SHA-256 digest for the given port list.
// Port order does not matter; the list is sorted before hashing.
func Compute(ports []int) Digest {
	sorted := make([]int, len(ports))
	copy(sorted, ports)
	sort.Ints(sorted)

	parts := make([]string, len(sorted))
	for i, p := range sorted {
		parts[i] = fmt.Sprintf("%d", p)
	}

	h := sha256.Sum256([]byte(strings.Join(parts, ",")))
	return Digest(hex.EncodeToString(h[:8])) // 16-char prefix is plenty
}

// Equal reports whether two digests are identical.
func Equal(a, b Digest) bool {
	return a == b
}

// Changed reports whether the port set has changed by comparing digests.
// It is a convenience wrapper around Compute + Equal.
func Changed(prev, curr []int) bool {
	return !Equal(Compute(prev), Compute(curr))
}

// String satisfies the fmt.Stringer interface.
func (d Digest) String() string {
	return string(d)
}

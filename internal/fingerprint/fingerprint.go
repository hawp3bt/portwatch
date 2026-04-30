// Package fingerprint builds a stable identity string for a set of open ports,
// enabling quick change detection without a full diff.
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Fingerprint holds a computed port-set identity.
type Fingerprint struct {
	Hash      string    `json:"hash"`
	Ports     []int     `json:"ports"`
	ComputedAt time.Time `json:"computed_at"`
}

// Compute returns a Fingerprint for the given port list.
// The hash is order-independent: [80,443] and [443,80] produce the same value.
func Compute(ports []int) Fingerprint {
	sorted := make([]int, len(ports))
	copy(sorted, ports)
	sort.Ints(sorted)

	parts := make([]string, len(sorted))
	for i, p := range sorted {
		parts[i] = fmt.Sprintf("%d", p)
	}
	raw := strings.Join(parts, ",")

	sum := sha256.Sum256([]byte(raw))
	return Fingerprint{
		Hash:       hex.EncodeToString(sum[:]),
		Ports:      sorted,
		ComputedAt: time.Now().UTC(),
	}
}

// Equal reports whether two fingerprints represent the same port set.
func Equal(a, b Fingerprint) bool {
	return a.Hash == b.Hash
}

// Changed reports whether the port set has changed between prev and next.
func Changed(prev, next Fingerprint) bool {
	return !Equal(prev, next)
}

// Short returns the first 12 characters of the hash for display purposes.
func (f Fingerprint) Short() string {
	if len(f.Hash) >= 12 {
		return f.Hash[:12]
	}
	return f.Hash
}

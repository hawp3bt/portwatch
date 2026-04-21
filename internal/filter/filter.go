package filter

// Rule defines a filtering rule for ports.
type Rule struct {
	// Ignored is a set of ports that should never trigger alerts.
	Ignored map[int]struct{}
	// AllowedRange defines an inclusive port range that is considered
	// "expected". Ports outside this range will be flagged.
	MinPort int
	MaxPort int
}

// New creates a Rule from a slice of ignored ports and an optional port range.
// If minPort and maxPort are both 0, no range filtering is applied.
func New(ignoredPorts []int, minPort, maxPort int) *Rule {
	ignored := make(map[int]struct{}, len(ignoredPorts))
	for _, p := range ignoredPorts {
		ignored[p] = struct{}{}
	}
	return &Rule{
		Ignored:  ignored,
		MinPort:  minPort,
		MaxPort:  maxPort,
	}
}

// IsIgnored reports whether the given port should be silently skipped.
func (r *Rule) IsIgnored(port int) bool {
	_, ok := r.Ignored[port]
	return ok
}

// InRange reports whether the given port falls within the configured range.
// If no range is configured (both min and max are 0) every port is considered
// in-range.
func (r *Rule) InRange(port int) bool {
	if r.MinPort == 0 && r.MaxPort == 0 {
		return true
	}
	return port >= r.MinPort && port <= r.MaxPort
}

// Apply filters a slice of ports according to the rule, returning only those
// ports that are neither ignored nor out-of-range.
func (r *Rule) Apply(ports []int) []int {
	out := make([]int, 0, len(ports))
	for _, p := range ports {
		if r.IsIgnored(p) {
			continue
		}
		if !r.InRange(p) {
			continue
		}
		out = append(out, p)
	}
	return out
}

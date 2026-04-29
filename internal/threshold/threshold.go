package threshold

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

// Rule defines an alert threshold for a specific port or port range.
type Rule struct {
	Port     int    `json:"port"`
	MaxOpen  int    `json:"max_open"`  // alert if open count exceeds this
	MinOpen  int    `json:"min_open"`  // alert if open count drops below this (-1 = disabled)
	Label    string `json:"label,omitempty"`
}

// Violation describes a threshold breach.
type Violation struct {
	Rule    Rule
	Actual  int
	Message string
}

func (v Violation) String() string {
	return fmt.Sprintf("threshold violation [%s port %d]: %s (actual=%d)",
		v.Rule.Label, v.Rule.Port, v.Message, v.Actual)
}

// Registry holds threshold rules persisted to disk.
type Registry struct {
	mu    sync.RWMutex
	path  string
	rules map[int]Rule
}

// New loads a Registry from path, creating an empty one if the file is absent.
func New(path string) (*Registry, error) {
	r := &Registry{path: path, rules: make(map[int]Rule)}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return r, nil
	}
	if err != nil {
		return nil, fmt.Errorf("threshold: read %s: %w", path, err)
	}
	var list []Rule
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("threshold: parse %s: %w", path, err)
	}
	for _, rule := range list {
		r.rules[rule.Port] = rule
	}
	return r, nil
}

// Set adds or replaces the rule for the given port and persists.
func (r *Registry) Set(rule Rule) error {
	if rule.Port <= 0 || rule.Port > 65535 {
		return fmt.Errorf("threshold: invalid port %d", rule.Port)
	}
	r.mu.Lock()
	r.rules[rule.Port] = rule
	r.mu.Unlock()
	return r.save()
}

// Remove deletes the rule for port and persists.
func (r *Registry) Remove(port int) error {
	r.mu.Lock()
	delete(r.rules, port)
	r.mu.Unlock()
	return r.save()
}

// Check evaluates open port counts against all rules and returns violations.
func (r *Registry) Check(counts map[int]int) []Violation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Violation
	for port, rule := range r.rules {
		actual := counts[port]
		if rule.MaxOpen > 0 && actual > rule.MaxOpen {
			out = append(out, Violation{Rule: rule, Actual: actual,
				Message: fmt.Sprintf("exceeds max_open=%d", rule.MaxOpen)})
		}
		if rule.MinOpen >= 0 && actual < rule.MinOpen {
			out = append(out, Violation{Rule: rule, Actual: actual,
				Message: fmt.Sprintf("below min_open=%d", rule.MinOpen)})
		}
	}
	return out
}

// List returns a copy of all rules.
func (r *Registry) List() []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		out = append(out, rule)
	}
	return out
}

func (r *Registry) save() error {
	r.mu.RLock()
	list := make([]Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		list = append(list, rule)
	}
	r.mu.RUnlock()
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("threshold: marshal: %w", err)
	}
	return os.WriteFile(r.path, data, 0o644)
}

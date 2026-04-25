// Package tags provides port tagging — attach human-readable labels to port
// numbers so alerts and reports can display names like "http" or "postgres"
// instead of bare numbers.
package tags

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Tag associates a label and optional description with a port.
type Tag struct {
	Port        int    `json:"port"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// Registry holds the full set of port tags.
type Registry struct {
	path string
	tags map[int]Tag
}

// New loads a Registry from the given JSON file path.
// If the file does not exist an empty Registry is returned without error.
func New(path string) (*Registry, error) {
	r := &Registry{path: path, tags: make(map[int]Tag)}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return r, nil
	}
	if err != nil {
		return nil, fmt.Errorf("tags: read %s: %w", path, err)
	}
	var list []Tag
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("tags: parse %s: %w", path, err)
	}
	for _, t := range list {
		r.tags[t.Port] = t
	}
	return r, nil
}

// Set adds or replaces the tag for the given port and persists the registry.
func (r *Registry) Set(t Tag) error {
	r.tags[t.Port] = t
	return r.save()
}

// Remove deletes the tag for the given port and persists the registry.
func (r *Registry) Remove(port int) error {
	delete(r.tags, port)
	return r.save()
}

// Label returns the label for port, or a formatted fallback if none is set.
func (r *Registry) Label(port int) string {
	if t, ok := r.tags[port]; ok {
		return t.Label
	}
	return fmt.Sprintf("%d", port)
}

// List returns all tags sorted by port number.
func (r *Registry) List() []Tag {
	out := make([]Tag, 0, len(r.tags))
	for _, t := range r.tags {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Port < out[j].Port })
	return out
}

func (r *Registry) save() error {
	list := r.List()
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("tags: marshal: %w", err)
	}
	if err := os.WriteFile(r.path, data, 0o644); err != nil {
		return fmt.Errorf("tags: write %s: %w", r.path, err)
	}
	return nil
}

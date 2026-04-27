package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Profile represents a named set of expected open ports for a given context
// (e.g. "dev", "prod", "ci").
type Profile struct {
	Name      string    `json:"name"`
	Ports     []int     `json:"ports"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Registry manages named port profiles stored on disk.
type Registry struct {
	dir string
}

// New returns a Registry that persists profiles under dir.
func New(dir string) (*Registry, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("profile: mkdir %s: %w", dir, err)
	}
	return &Registry{dir: dir}, nil
}

func (r *Registry) path(name string) string {
	return filepath.Join(r.dir, name+".json")
}

// Save creates or overwrites the named profile.
func (r *Registry) Save(name string, ports []int) error {
	if name == "" {
		return errors.New("profile: name must not be empty")
	}
	now := time.Now().UTC()
	p := Profile{Name: name, Ports: ports, CreatedAt: now, UpdatedAt: now}
	// Preserve original CreatedAt if profile already exists.
	if existing, err := r.Load(name); err == nil {
		p.CreatedAt = existing.CreatedAt
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("profile: marshal: %w", err)
	}
	return os.WriteFile(r.path(name), data, 0o644)
}

// Load retrieves the named profile from disk.
func (r *Registry) Load(name string) (*Profile, error) {
	data, err := os.ReadFile(r.path(name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("profile %q not found", name)
		}
		return nil, fmt.Errorf("profile: read: %w", err)
	}
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("profile: unmarshal: %w", err)
	}
	return &p, nil
}

// Delete removes the named profile.
func (r *Registry) Delete(name string) error {
	err := os.Remove(r.path(name))
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("profile %q not found", name)
	}
	return err
}

// List returns all profile names stored in the registry.
func (r *Registry) List() ([]string, error) {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, fmt.Errorf("profile: readdir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			names = append(names, e.Name()[:len(e.Name())-5])
		}
	}
	return names, nil
}

// Diff returns ports present in the profile but not in current, and ports
// present in current but not in the profile.
func Diff(profile []int, current []int) (missing []int, extra []int) {
	pSet := make(map[int]struct{}, len(profile))
	for _, p := range profile {
		pSet[p] = struct{}{}
	}
	cSet := make(map[int]struct{}, len(current))
	for _, p := range current {
		cSet[p] = struct{}{}
	}
	for p := range pSet {
		if _, ok := cSet[p]; !ok {
			missing = append(missing, p)
		}
	}
	for p := range cSet {
		if _, ok := pSet[p]; !ok {
			extra = append(extra, p)
		}
	}
	return
}

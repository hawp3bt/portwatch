package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot represents a named baseline of expected open ports.
type Snapshot struct {
	Name      string    `json:"name"`
	Ports     []int     `json:"ports"`
	CreatedAt time.Time `json:"created_at"`
}

// Store manages named baselines persisted to a JSON file.
type Store struct {
	path string
}

// New returns a Store backed by the given file path.
func New(path string) *Store {
	return &Store{path: path}
}

// Save writes a named baseline of ports to disk, overwriting any existing
// baseline with the same name.
func (s *Store) Save(name string, ports []int) error {
	baselines, err := s.loadAll()
	if err != nil {
		return err
	}
	baselines[name] = Snapshot{
		Name:      name,
		Ports:     ports,
		CreatedAt: time.Now().UTC(),
	}
	return s.writeAll(baselines)
}

// Load returns the baseline snapshot for the given name.
// Returns an error if the name does not exist.
func (s *Store) Load(name string) (Snapshot, error) {
	baselines, err := s.loadAll()
	if err != nil {
		return Snapshot{}, err
	}
	snap, ok := baselines[name]
	if !ok {
		return Snapshot{}, fmt.Errorf("baseline %q not found", name)
	}
	return snap, nil
}

// List returns all stored baseline snapshots.
func (s *Store) List() ([]Snapshot, error) {
	baselines, err := s.loadAll()
	if err != nil {
		return nil, err
	}
	out := make([]Snapshot, 0, len(baselines))
	for _, v := range baselines {
		out = append(out, v)
	}
	return out, nil
}

func (s *Store) loadAll() (map[string]Snapshot, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return make(map[string]Snapshot), nil
	}
	if err != nil {
		return nil, fmt.Errorf("baseline: read %s: %w", s.path, err)
	}
	var m map[string]Snapshot
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("baseline: parse %s: %w", s.path, err)
	}
	return m, nil
}

func (s *Store) writeAll(m map[string]Snapshot) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("baseline: marshal: %w", err)
	}
	return os.WriteFile(s.path, data, 0o644)
}

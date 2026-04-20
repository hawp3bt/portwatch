package state

import (
	"encoding/json"
	"os"
	"time"
)

// Snapshot represents a persisted port state at a point in time.
type Snapshot struct {
	Timestamp time.Time `json:"timestamp"`
	Ports     []int     `json:"ports"`
}

// Store handles reading and writing port state snapshots to disk.
type Store struct {
	path string
}

// New creates a new Store backed by the given file path.
func New(path string) *Store {
	return &Store{path: path}
}

// Save writes the current set of open ports to disk.
func (s *Store) Save(ports []int) error {
	snap := Snapshot{
		Timestamp: time.Now().UTC(),
		Ports:     ports,
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// Load reads the last persisted snapshot from disk.
// Returns an empty Snapshot (no error) if the file does not exist yet.
func (s *Store) Load() (Snapshot, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return Snapshot{}, nil
	}
	if err != nil {
		return Snapshot{}, err
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return Snapshot{}, err
	}
	return snap, nil
}

// Exists reports whether a persisted snapshot file is present on disk.
func (s *Store) Exists() bool {
	_, err := os.Stat(s.path)
	return err == nil
}

package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Entry represents a single port snapshot with metadata.
type Entry struct {
	Ports     []int     `json:"ports"`
	CapturedAt time.Time `json:"captured_at"`
	Label     string    `json:"label,omitempty"`
}

// Manager handles saving and loading named port snapshots.
type Manager struct {
	dir string
}

// New returns a Manager that stores snapshots under dir.
func New(dir string) *Manager {
	return &Manager{dir: dir}
}

// Save writes a named snapshot to disk.
func (m *Manager) Save(name string, ports []int, label string) error {
	if err := os.MkdirAll(m.dir, 0o755); err != nil {
		return fmt.Errorf("snapshot: mkdir: %w", err)
	}
	e := Entry{
		Ports:      ports,
		CapturedAt: time.Now().UTC(),
		Label:      label,
	}
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("snapshot: marshal: %w", err)
	}
	path := m.path(name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("snapshot: write %s: %w", path, err)
	}
	return nil
}

// Load reads a named snapshot from disk.
func (m *Manager) Load(name string) (*Entry, error) {
	data, err := os.ReadFile(m.path(name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("snapshot %q not found", name)
		}
		return nil, fmt.Errorf("snapshot: read: %w", err)
	}
	var e Entry
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("snapshot: unmarshal: %w", err)
	}
	return &e, nil
}

// List returns the names of all saved snapshots.
func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("snapshot: readdir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func (m *Manager) path(name string) string {
	return fmt.Sprintf("%s/%s.json", m.dir, name)
}

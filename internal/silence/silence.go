// Package silence provides a time-window based silencing mechanism that
// prevents alerts from firing during scheduled maintenance periods.
package silence

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Window represents a named maintenance window during which alerts are silenced.
type Window struct {
	Name      string    `json:"name"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	CreatedAt time.Time `json:"created_at"`
}

// Active returns true if the window is currently active.
func (w Window) Active(now time.Time) bool {
	return now.After(w.Start) && now.Before(w.End)
}

// Registry manages silence windows persisted to disk.
type Registry struct {
	mu      sync.RWMutex
	path    string
	windows map[string]Window
}

// New loads an existing registry from path, or returns an empty one if missing.
func New(path string) (*Registry, error) {
	r := &Registry{path: path, windows: make(map[string]Window)}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return r, nil
	}
	if err != nil {
		return nil, fmt.Errorf("silence: read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &r.windows); err != nil {
		return nil, fmt.Errorf("silence: parse %s: %w", path, err)
	}
	return r, nil
}

// Add creates or replaces a silence window.
func (r *Registry) Add(name string, start, end time.Time) error {
	if name == "" {
		return fmt.Errorf("silence: name must not be empty")
	}
	if !end.After(start) {
		return fmt.Errorf("silence: end must be after start")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.windows[name] = Window{Name: name, Start: start, End: end, CreatedAt: time.Now()}
	return r.save()
}

// Remove deletes a silence window by name.
func (r *Registry) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.windows[name]; !ok {
		return fmt.Errorf("silence: window %q not found", name)
	}
	delete(r.windows, name)
	return r.save()
}

// IsSilenced returns true if any active window covers the current time.
func (r *Registry) IsSilenced(now time.Time) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, w := range r.windows {
		if w.Active(now) {
			return true
		}
	}
	return false
}

// List returns all registered windows.
func (r *Registry) List() []Window {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Window, 0, len(r.windows))
	for _, w := range r.windows {
		out = append(out, w)
	}
	return out
}

func (r *Registry) save() error {
	data, err := json.MarshalIndent(r.windows, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, data, 0o644)
}

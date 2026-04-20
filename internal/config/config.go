package config

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

// Config holds the runtime configuration for portwatch.
type Config struct {
	// Interval between successive port scans.
	Interval time.Duration `json:"-"`
	// IntervalSeconds is the on-disk representation (seconds).
	IntervalSeconds int `json:"interval_seconds"`
	// Ports is the explicit list of ports to watch. Empty means scan all.
	Ports []int `json:"ports"`
	// StateFile is the path used to persist port snapshots between runs.
	StateFile string `json:"state_file"`
	// AlertOutput is the file path for alert output; empty means stdout.
	AlertOutput string `json:"alert_output,omitempty"`
}

// Default returns a Config with sensible defaults.
func Default() Config {
	return Config{
		IntervalSeconds: 30,
		Interval:        30 * time.Second,
		Ports:           []int{},
		StateFile:       "/tmp/portwatch_state.json",
	}
}

// Load reads a JSON config file from path and returns a validated Config.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	if cfg.IntervalSeconds <= 0 {
		return Config{}, errors.New("interval_seconds must be greater than zero")
	}
	cfg.Interval = time.Duration(cfg.IntervalSeconds) * time.Second

	if cfg.StateFile == "" {
		cfg.StateFile = Default().StateFile
	}

	return cfg, nil
}

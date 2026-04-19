package config

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

// Config holds portwatch runtime configuration.
type Config struct {
	// Ports to monitor; empty means monitor all detected ports.
	Ports []int `json:"ports"`
	// Interval between scans.
	Interval duration `json:"interval"`
	// Baseline is the set of ports considered "expected".
	Baseline []int `json:"baseline"`
}

// duration is a time.Duration that marshals/unmarshals as a string (e.g. "5s").
type duration struct{ time.Duration }

func (d *duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = v
	return nil
}

func (d duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// Load reads a JSON config file from path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Default returns a sensible default configuration.
func Default() *Config {
	return &Config{
		Interval: duration{30 * time.Second},
	}
}

func (c *Config) validate() error {
	if c.Interval.Duration <= 0 {
		return errors.New("config: interval must be positive")
	}
	return nil
}

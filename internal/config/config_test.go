package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "portwatch-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `{"ports":[80,443],"interval":"10s","baseline":[80]}`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval.Duration != 10*time.Second {
		t.Errorf("expected 10s interval, got %v", cfg.Interval.Duration)
	}
	if len(cfg.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(cfg.Ports))
	}
}

func TestLoad_InvalidInterval(t *testing.T) {
	path := writeTempConfig(t, `{"interval":"-1s"}`)
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for negative interval")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestDefault_ReturnsPositiveInterval(t *testing.T) {
	cfg := Default()
	if cfg.Interval.Duration <= 0 {
		t.Errorf("default interval should be positive, got %v", cfg.Interval.Duration)
	}
}

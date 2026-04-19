package alert

import (
	"bytes"
	"strings"
	"testing"
)

func TestPortOpened_OutputContainsAlert(t *testing.T) {
	var buf bytes.Buffer
	n := New(&buf)
	n.PortOpened(8080)

	out := buf.String()
	if !strings.Contains(out, "ALERT") {
		t.Errorf("expected ALERT in output, got: %s", out)
	}
	if !strings.Contains(out, "8080") {
		t.Errorf("expected port 8080 in output, got: %s", out)
	}
	if !strings.Contains(out, "opened unexpectedly") {
		t.Errorf("expected 'opened unexpectedly' in output, got: %s", out)
	}
}

func TestPortClosed_OutputContainsWarn(t *testing.T) {
	var buf bytes.Buffer
	n := New(&buf)
	n.PortClosed(443)

	out := buf.String()
	if !strings.Contains(out, "WARN") {
		t.Errorf("expected WARN in output, got: %s", out)
	}
	if !strings.Contains(out, "443") {
		t.Errorf("expected port 443 in output, got: %s", out)
	}
	if !strings.Contains(out, "closed unexpectedly") {
		t.Errorf("expected 'closed unexpectedly' in output, got: %s", out)
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	n := New(nil)
	if n.out == nil {
		t.Error("expected non-nil writer when nil passed to New")
	}
}

func TestMultipleEvents_EachOnOwnLine(t *testing.T) {
	var buf bytes.Buffer
	n := New(&buf)
	n.PortOpened(1234)
	n.PortClosed(5678)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %s", len(lines), buf.String())
	}
}

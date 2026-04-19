package scanner

import (
	"net"
	"strconv"
	"testing"
	"time"
)

// startTestServer opens a TCP listener on a random port and returns the port and a cleanup func.
func startTestServer(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	port, _ := strconv.Atoi(ln.Addr().(*net.TCPAddr).Port.String())
	// Re-extract properly
	port = ln.Addr().(*net.TCPAddr).Port
	return port, func() { ln.Close() }
}

func TestScan_OpenPort(t *testing.T) {
	port, cleanup := startTestServer(t)
	defer cleanup()

	states, err := Scan("127.0.0.1", []int{port}, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 result, got %d", len(states))
	}
	if !states[0].Open {
		t.Errorf("expected port %d to be open", port)
	}
}

func TestScan_ClosedPort(t *testing.T) {
	// Port 1 is almost certainly closed in test environments.
	states, err := Scan("127.0.0.1", []int{1}, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if states[0].Open {
		t.Errorf("expected port 1 to be closed")
	}
}

func TestScan_MultiplePortsOpenAndClosed(t *testing.T) {
	port, cleanup := startTestServer(t)
	defer cleanup()

	states, err := Scan("127.0.0.1", []int{port, 1}, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 2 {
		t.Fatalf("expected 2 results, got %d", len(states))
	}
	open := OpenPorts(states)
	if len(open) != 1 {
		t.Errorf("expected 1 open port, got %d", len(open))
	}
	if open[0].Port != port {
		t.Errorf("expected open port %d, got %d", port, open[0].Port)
	}
}

func TestOpenPorts_Filter(t *testing.T) {
	input := []PortState{
		{Port: 80, Open: true},
		{Port: 81, Open: false},
		{Port: 443, Open: true},
	}
	result := OpenPorts(input)
	if len(result) != 2 {
		t.Errorf("expected 2 open ports, got %d", len(result))
	}
}

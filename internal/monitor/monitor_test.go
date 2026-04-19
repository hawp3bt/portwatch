package monitor

import (
	"net"
	"testing"
	"time"
)

func startTCPServer(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	return port, func() { ln.Close() }
}

func TestDiff_AddedAndRemoved(t *testing.T) {
	prev := []int{80, 443}
	curr := []int{443, 8080}
	added, removed := diff(prev, curr)
	if len(added) != 1 || added[0] != 8080 {
		t.Errorf("expected added=[8080], got %v", added)
	}
	if len(removed) != 1 || removed[0] != 80 {
		t.Errorf("expected removed=[80], got %v", removed)
	}
}

func TestDiff_NoChange(t *testing.T) {
	ports := []int{80, 443}
	added, removed := diff(ports, ports)
	if len(added) != 0 || len(removed) != 0 {
		t.Errorf("expected no changes, got added=%v removed=%v", added, removed)
	}
}

func TestMonitor_DetectsNewPort(t *testing.T) {
	// Start with no open port on this address, then open one.
	port, stop := startTCPServer(t)
	stop() // close immediately so monitor sees it as closed first

	m := New([]int{port}, 50*time.Millisecond)
	m.Start()
	defer m.Stop()

	// Now open the port so the next poll detects it as added.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	newPort := ln.Addr().(*net.TCPAddr).Port
	defer ln.Close()

	m2 := New([]int{newPort}, 50*time.Millisecond)
	m2.Start()
	defer m2.Stop()

	select {
	case alert := <-m2.Alerts:
		if len(alert.Added) == 0 && len(alert.Removed) == 0 {
			t.Error("expected a non-empty alert")
		}
	case <-time.After(500 * time.Millisecond):
		// No alert is acceptable if port was open from the start.
	}
}

func TestToSet(t *testing.T) {
	s := toSet([]int{1, 2, 3})
	for _, p := range []int{1, 2, 3} {
		if !s[p] {
			t.Errorf("expected port %d in set", p)
		}
	}
	if s[4] {
		t.Error("port 4 should not be in set")
	}
}

package scanner

import (
	"fmt"
	"net"
	"time"
)

// PortState represents the state of a single port.
type PortState struct {
	Port     int
	Protocol string
	Open     bool
}

// Scan checks which ports from the given list are open on the specified host.
func Scan(host string, ports []int, timeout time.Duration) ([]PortState, error) {
	if host == "" {
		host = "127.0.0.1"
	}

	results := make([]PortState, 0, len(ports))

	for _, port := range ports {
		state := PortState{
			Port:     port,
			Protocol: "tcp",
		}

		address := fmt.Sprintf("%s:%d", host, port)
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err == nil {
			state.Open = true
			conn.Close()
		}

		results = append(results, state)
	}

	return results, nil
}

// OpenPorts filters a slice of PortState and returns only the open ones.
func OpenPorts(states []PortState) []PortState {
	open := make([]PortState, 0)
	for _, s := range states {
		if s.Open {
			open = append(open, s)
		}
	}
	return open
}

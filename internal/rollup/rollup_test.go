package rollup_test

import (
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/user/portwatch/internal/rollup"
)

func collectSummaries(t *testing.T, window time.Duration, events []rollup.Event, flush bool) []rollup.Summary {
	t.Helper()
	var mu sync.Mutex
	var got []rollup.Summary

	r := rollup.New(window, func(s rollup.Summary) {
		mu.Lock()
		got = append(got, s)
		mu.Unlock()
	})

	for _, e := range events {
		r.Push(e)
	}
	if flush {
		r.Flush()
	}
	time.Sleep(20 * time.Millisecond)
	return got
}

func TestFlush_EmitsSummaryWithOpenedAndClosed(t *testing.T) {
	events := []rollup.Event{
		{Port: 80, Opened: true},
		{Port: 443, Opened: true},
		{Port: 8080, Opened: false},
	}
	got := collectSummaries(t, time.Second, events, true)

	if len(got) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(got))
	}
	s := got[0]
	if len(s.Opened) != 2 {
		t.Errorf("expected 2 opened, got %d", len(s.Opened))
	}
	if len(s.Closed) != 1 {
		t.Errorf("expected 1 closed, got %d", len(s.Closed))
	}
}

func TestFlush_EmptyProducesNoSummary(t *testing.T) {
	r := rollup.New(time.Second, func(s rollup.Summary) {
		t.Error("emit called unexpectedly")
	})
	r.Flush()
	time.Sleep(20 * time.Millisecond)
}

func TestPush_OpenThenCloseDeduplicates(t *testing.T) {
	events := []rollup.Event{
		{Port: 9090, Opened: true},
		{Port: 9090, Opened: false}, // close wins
	}
	got := collectSummaries(t, time.Second, events, true)

	if len(got) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(got))
	}
	if len(got[0].Opened) != 0 {
		t.Errorf("port should not appear in opened list")
	}
	if len(got[0].Closed) != 1 || got[0].Closed[0] != 9090 {
		t.Errorf("expected port 9090 in closed list")
	}
}

func TestSummary_String(t *testing.T) {
	s := rollup.Summary{Opened: []int{80, 443}, Closed: []int{8080}}
	str := s.String()
	if str != "2 opened, 1 closed" {
		t.Errorf("unexpected summary string: %q", str)
	}
}

func TestSummary_String_Empty(t *testing.T) {
	s := rollup.Summary{}
	if s.String() != "no changes" {
		t.Errorf("expected 'no changes', got %q", s.String())
	}
}

func TestWindow_AutoFlushAfterDuration(t *testing.T) {
	var mu sync.Mutex
	var got []rollup.Summary

	r := rollup.New(50*time.Millisecond, func(s rollup.Summary) {
		mu.Lock()
		got = append(got, s)
		mu.Unlock()
	})
	r.Push(rollup.Event{Port: 22, Opened: true})

	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("expected auto-flush to produce 1 summary, got %d", len(got))
	}
	sort.Ints(got[0].Opened)
	if len(got[0].Opened) != 1 || got[0].Opened[0] != 22 {
		t.Errorf("unexpected opened ports: %v", got[0].Opened)
	}
}

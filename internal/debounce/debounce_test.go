package debounce_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/debounce"
)

const shortDelay = 50 * time.Millisecond

func TestPush_EventArrivesAfterDelay(t *testing.T) {
	d, ch := debounce.New(shortDelay)
	d.Push(8080, true)

	select {
	case ev := <-ch:
		if ev.Port != 8080 || !ev.Opened {
			t.Fatalf("unexpected event: %+v", ev)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("event never arrived")
	}
}

func TestPush_RapidDuplicatesCoalesced(t *testing.T) {
	d, ch := debounce.New(shortDelay)

	// Push the same port multiple times quickly.
	for i := 0; i < 5; i++ {
		d.Push(9090, true)
		time.Sleep(10 * time.Millisecond)
	}

	// Exactly one event should arrive.
	var count int
	timer := time.NewTimer(200 * time.Millisecond)
	defer timer.Stop()
drain:
	for {
		select {
		case <-ch:
			count++
		case <-timer.C:
			break drain
		}
	}

	if count != 1 {
		t.Fatalf("expected 1 coalesced event, got %d", count)
	}
}

func TestPush_DifferentPortsAreIndependent(t *testing.T) {
	d, ch := debounce.New(shortDelay)
	d.Push(1111, true)
	d.Push(2222, false)

	received := map[int]bool{}
	timer := time.NewTimer(300 * time.Millisecond)
	defer timer.Stop()
collect:
	for len(received) < 2 {
		select {
		case ev := <-ch:
			received[ev.Port] = true
		case <-timer.C:
			break collect
		}
	}

	if !received[1111] || !received[2222] {
		t.Fatalf("did not receive events for both ports: %v", received)
	}
}

func TestFlush_EmitsPendingImmediately(t *testing.T) {
	d, ch := debounce.New(10 * time.Second) // very long delay
	d.Push(3333, true)
	d.Push(4444, false)

	if d.PendingCount() != 2 {
		t.Fatalf("expected 2 pending, got %d", d.PendingCount())
	}

	d.Flush()

	received := map[int]bool{}
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()
flushDrain:
	for len(received) < 2 {
		select {
		case ev := <-ch:
			received[ev.Port] = true
		case <-timer.C:
			break flushDrain
		}
	}

	if !received[3333] || !received[4444] {
		t.Fatalf("flush did not emit all pending events: %v", received)
	}

	if d.PendingCount() != 0 {
		t.Fatalf("expected 0 pending after flush, got %d", d.PendingCount())
	}
}

func TestPendingCount_DecrementsAfterFire(t *testing.T) {
	d, ch := debounce.New(shortDelay)
	d.Push(5555, true)

	if d.PendingCount() != 1 {
		t.Fatalf("expected 1 pending immediately after push")
	}

	<-ch // wait for event to fire
	time.Sleep(10 * time.Millisecond)

	if d.PendingCount() != 0 {
		t.Fatalf("expected 0 pending after event fired")
	}
}

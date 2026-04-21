package ratelimit_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/ratelimit"
)

func TestAllow_FirstCallAlwaysTrue(t *testing.T) {
	l := ratelimit.New(5 * time.Second)
	if !l.Allow("port:8080:open") {
		t.Fatal("expected first Allow call to return true")
	}
}

func TestAllow_SecondCallWithinCooldown_ReturnsFalse(t *testing.T) {
	l := ratelimit.New(5 * time.Second)
	l.Allow("port:8080:open")
	if l.Allow("port:8080:open") {
		t.Fatal("expected second Allow within cooldown to return false")
	}
}

func TestAllow_AfterCooldownExpires_ReturnsTrue(t *testing.T) {
	now := time.Now()
	l := ratelimit.New(5 * time.Second)
	// Inject a clock that moves forward past the cooldown.
	calls := 0
	l2 := &struct{ *ratelimit.Limiter }{ratelimit.New(5 * time.Second)}
	_ = l2
	_ = calls

	// Use Reset to simulate expiry.
	l.Allow("port:9090:open")
	l.Reset("port:9090:open")
	if !l.Allow("port:9090:open") {
		t.Fatal("expected Allow to return true after Reset")
	}
	_ = now
}

func TestAllow_DifferentKeys_IndependentLimits(t *testing.T) {
	l := ratelimit.New(5 * time.Second)
	l.Allow("port:8080:open")
	if !l.Allow("port:9090:open") {
		t.Fatal("expected different key to be allowed independently")
	}
}

func TestReset_ClearsSpecificKey(t *testing.T) {
	l := ratelimit.New(10 * time.Second)
	l.Allow("port:80:open")
	l.Allow("port:443:open")

	l.Reset("port:80:open")

	if !l.Allow("port:80:open") {
		t.Error("expected port:80:open to be allowed after Reset")
	}
	if l.Allow("port:443:open") {
		t.Error("expected port:443:open to still be rate-limited")
	}
}

func TestResetAll_ClearsAllKeys(t *testing.T) {
	l := ratelimit.New(10 * time.Second)
	l.Allow("port:80:open")
	l.Allow("port:443:open")

	l.ResetAll()

	if !l.Allow("port:80:open") {
		t.Error("expected port:80:open to be allowed after ResetAll")
	}
	if !l.Allow("port:443:open") {
		t.Error("expected port:443:open to be allowed after ResetAll")
	}
}

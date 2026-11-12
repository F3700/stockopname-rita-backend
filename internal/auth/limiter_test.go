package auth

import (
	"testing"
	"time"
)

// fixedClock hands the limiter a hand-cranked "now" so the window can be
// crossed without sleeping.
func fixedClock(limiter *Limiter) *time.Time {
	now := time.Date(2026, 11, 12, 8, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }
	return &now
}

func TestLimiterBlocksAfterLimit(t *testing.T) {
	limiter := NewLimiter(2, time.Minute)
	fixedClock(limiter)

	if !limiter.Allowed("10.0.0.1") {
		t.Fatal("the first attempt must be allowed")
	}

	limiter.Fail("10.0.0.1")
	limiter.Fail("10.0.0.1")

	if limiter.Allowed("10.0.0.1") {
		t.Error("expected the client to be blocked after two failures")
	}
}

func TestLimiterForgetsAfterWindow(t *testing.T) {
	limiter := NewLimiter(1, time.Minute)
	now := fixedClock(limiter)

	limiter.Fail("10.0.0.1")
	if limiter.Allowed("10.0.0.1") {
		t.Fatal("expected the client to be blocked inside the window")
	}

	*now = now.Add(61 * time.Second)
	if !limiter.Allowed("10.0.0.1") {
		t.Error("expected the client to be allowed again after the window")
	}
}

func TestLimiterResetOnSuccess(t *testing.T) {
	limiter := NewLimiter(1, time.Minute)
	fixedClock(limiter)

	limiter.Fail("10.0.0.1")
	limiter.Reset("10.0.0.1")

	if !limiter.Allowed("10.0.0.1") {
		t.Error("a successful login must clear the failures")
	}
}

func TestLimiterKeepsClientsSeparate(t *testing.T) {
	limiter := NewLimiter(1, time.Minute)
	fixedClock(limiter)

	limiter.Fail("10.0.0.1")

	if limiter.Allowed("10.0.0.1") {
		t.Fatal("expected the first client to be blocked")
	}
	if !limiter.Allowed("10.0.0.2") {
		t.Error("one client must not block another")
	}
}

func TestLimiterCanBeDisabled(t *testing.T) {
	limiter := NewLimiter(0, time.Minute)
	fixedClock(limiter)

	for range 100 {
		limiter.Fail("10.0.0.1")
	}

	if !limiter.Allowed("10.0.0.1") {
		t.Error("a zero limit must disable the guard")
	}
}

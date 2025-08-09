package limiter

import (
	"testing"
	"time"
)

func TestSlidingWindowLimiter_AllowsWithinLimit(t *testing.T) {
	limiter := NewSlidingWindowLimiter(3, 1*time.Second)

	for i := 0; i < 3; i++ {
		if !limiter.Allow() {
			t.Errorf("expected Allow() to return true on request #%d", i+1)
		}
	}
}

func TestSlidingWindowLimiter_DeniesOverLimit(t *testing.T) {
	limiter := NewSlidingWindowLimiter(2, 1*time.Second)

	limiter.Allow() // 1
	limiter.Allow() // 2
	if limiter.Allow() {
		t.Error("expected Allow() to return false when over the limit")
	}
}

func TestSlidingWindowLimiter_AllowsAfterWindowPasses(t *testing.T) {
	limiter := NewSlidingWindowLimiter(2, 100*time.Millisecond)

	limiter.Allow()
	limiter.Allow()
	if limiter.Allow() {
		t.Error("expected Allow() to return false when over the limit")
	}

	time.Sleep(110 * time.Millisecond)

	if !limiter.Allow() {
		t.Error("expected Allow() to return true after window passed")
	}
}

func TestSlidingWindowLimiter_CleansOldTimestamps(t *testing.T) {
	limiter := NewSlidingWindowLimiter(1, 50*time.Millisecond)

	if !limiter.Allow() {
		t.Fatal("expected initial Allow() to succeed")
	}
	time.Sleep(60 * time.Millisecond)

	// Should allow again because old timestamp is out of window
	if !limiter.Allow() {
		t.Error("expected Allow() to return true after old request expired")
	}
}

package limiter

import (
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	rl := NewLimiter(TokenBucket, 2, time.Duration(1*time.Second))
	out := rl.Allow()
	if !out {
		t.Errorf("should allow first")
	}
	out = rl.Allow()
	if !out {
		t.Errorf("should allow second")
	}
	out = rl.Allow()
	if out {
		t.Errorf("should disallow third")
	}
	time.Sleep(1 * time.Second)
	out = rl.Allow()
	if !out {
		t.Errorf("should allow after reset")
	}
}

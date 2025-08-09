package limiter

import (
	"testing"
	"time"
)

func TestPassFailThenReset(t *testing.T) {
	l := NewFixedWindowLimiter(5, 100*time.Millisecond)
	for i := range 6 {
		out := l.Allow()
		if i == 0 && !out {
			t.Errorf("should pass at beginning")
		}
		if i == 5 && out {
			t.Errorf("should fail after 5")
		}
	}
	time.Sleep(100 * time.Millisecond)
	if !l.Allow() {
		t.Errorf("should pass after reset filter")
	}
}

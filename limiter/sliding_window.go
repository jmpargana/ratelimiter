package limiter

import (
	"sync"
	"time"
)

var _ Limiter = &SlidingWindowLimiter{}

type SlidingWindowLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	requests []time.Time
}

func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	l := &SlidingWindowLimiter{
		limit:    limit,
		window:   window,
		requests: make([]time.Time, 0, limit),
	}
	return l
}

func (l *SlidingWindowLimiter) Allow() bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Drop old timestamps outside the window
	cutoff := now.Add(-l.window)
	i := 0
	for ; i < len(l.requests); i++ {
		if l.requests[i].After(cutoff) {
			break
		}
	}
	l.requests = l.requests[i:]

	if len(l.requests) < l.limit {
		l.requests = append(l.requests, now)
		return true
	}

	return false
}

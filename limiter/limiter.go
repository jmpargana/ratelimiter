package limiter

import (
	"time"

	"golang.org/x/time/rate"
)

var _ Limiter = &rate.Limiter{}

type RateLimiterKind int

const (
	FixedWindow = iota
	SlidingWindow
	TokenBucket
)

type Limiter interface {
	Allow() bool
}

func NewLimiter(
	kind RateLimiterKind,
	limit int,
	window time.Duration,
) Limiter {
	switch kind {
	case FixedWindow:
		return NewFixedWindowLimiter(limit, window)
	case SlidingWindow:
		return NewSlidingWindowLimiter(limit, window)
	case TokenBucket:
		return rate.NewLimiter(rate.Limit(limit/int(window.Seconds())), limit)
	}
	return rate.NewLimiter(rate.Limit(limit/int(window.Seconds())), limit)
}

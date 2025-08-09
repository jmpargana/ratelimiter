package limiter

import (
	"context"
	"sync/atomic"
	"time"
)

var _ Limiter = &FixedWindowLimiter{}

type FixedWindowLimiter struct {
	count    atomic.Int32
	limit    int
	window   time.Duration
	cancelFn context.CancelFunc
}

func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	l := &FixedWindowLimiter{
		limit:    limit,
		window:   window,
		cancelFn: cancel,
	}
	go l.start(ctx)
	return l
}

func (l *FixedWindowLimiter) start(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(l.window))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.count.Store(0)
		case <-ctx.Done():
			return
		}
	}
}

func (l *FixedWindowLimiter) Stop() {
	l.cancelFn()
}

func (l *FixedWindowLimiter) Allow() bool {
	newVal := l.count.Add(1)
	return newVal <= int32(l.limit)
}

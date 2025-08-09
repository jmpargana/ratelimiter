package store

import (
	"context"
	"time"
)

type CounterStore interface {
	Incr(ctx context.Context, key string, ttl time.Duration) (count int, err error)
}

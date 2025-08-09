package store

import (
	"context"
	"sync"
	"time"
)

var _ CounterStore = &InMemoryStore{}

type InMemoryStore struct {
	sync.Map
}

// Incr implements CounterStore.
func (s *InMemoryStore) Incr(ctx context.Context, key string, ttl time.Duration) (count int, err error) {
	// TODO: implement logic from one of the `limiter.Limiter`s
	val, _ := s.LoadOrStore(key, 0)
	c, _ := s.Swap(key, val.(int)+1)
	count = c.(int)
	return
}

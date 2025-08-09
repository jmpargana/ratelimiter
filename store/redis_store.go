package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ CounterStore = &RedisStore{}

// RedisStore implements the CounterStore interface using Redis as backend.
//
// It uses Redis commands to increment counters with an expiration (TTL).
type RedisStore struct {
	rdb *redis.Client
}

// Incr increments the counter associated with the given key within the specified window duration.
//
// This method uses a Redis transaction pipeline to atomically increment the counter and
// set its expiry to the provided window.
//
// Returns the new counter value after incrementing, or an error if the operation fails.
func (r *RedisStore) Incr(ctx context.Context, key string, window time.Duration) (count int, err error) {
	pipe := r.rdb.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return int(incr.Val()), nil
}

// NewRedisStore creates a new RedisStore instance using the provided Redis client.
//
// The returned RedisStore can be used with rate limiter implementations that require a CounterStore.
func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{
		rdb: rdb,
	}
}

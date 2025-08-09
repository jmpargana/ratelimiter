package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ CounterStore = &RedisStore{}

type RedisStore struct {
	rdb *redis.Client
}

// Incr implements CounterStore.
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

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{
		rdb: rdb,
	}
}

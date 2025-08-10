package ratelimiter

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jmpargana/ratelimiter/store"
	r "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

var rdb *r.Client

func setupFile(t *testing.T) string {
	dir := t.TempDir()
	tempFile := `global:
  limit: 2
  window: 1
endpoints:
  /path1:
    limit: 1
    window: 1	
  /path2:
    limit: 4
    window: 3
per_user:
  limit: 1
  window: 1
`

	fn := dir + "/test.yaml"

	err := os.WriteFile(fn, []byte(tempFile), 0644)
	assert.Nil(t, err, "couldn't create file")
	return fn
}

func TestMain(m *testing.M) {
	os.Setenv("GOMAXPROCS", "1")
	ctx := context.Background()
	redisC, err := redis.Run(context.Background(), "redis:6")
	if err != nil {
		panic(err)
	}
	ep, err := redisC.Endpoint(ctx, "")
	if err != nil {
		panic(err)
	}
	rdb = r.NewClient(&r.Options{Addr: ep})
	for {
		if err := rdb.Ping(ctx).Err(); err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	code := m.Run()
	redisC.Terminate(ctx)
	os.Exit(code)
}

func cleanUp(t *testing.T) {
	err := rdb.FlushAll(context.Background()).Err()
	assert.Nil(t, err, "failed flushing redis")
}

func setup(t *testing.T) *RateLimiter {
	cleanUp(t)
	fn := setupFile(t)
	cfg, err := LoadConfig(fn)
	assert.Nil(t, err, "should load config")
	store := store.NewRedisStore(rdb)
	rl, err := New(cfg, store)
	assert.Nil(t, err, "failed creating ratelimiter")
	return rl
}

func TestGlobal(t *testing.T) {
	ctx := context.Background()
	rl := setup(t)

	assert.True(t, rl.Allow(ctx, "", "user1"), "first")
	assert.True(t, rl.Allow(ctx, "", "user2"), "second")
	assert.False(t, rl.Allow(ctx, "", "user3"), "third")
	time.Sleep(1100 * time.Millisecond)
	assert.True(t, rl.Allow(ctx, "", "usert4"), "after timeout")

}

func TestEndpointRateLimiterBelowGlobal(t *testing.T) {
	ctx := context.Background()
	rl := setup(t)

	assert.True(t, rl.Allow(ctx, "/path1", "user1"), "first")
	assert.False(t, rl.Allow(ctx, "/path1", "user2"), "second")
	time.Sleep(1100 * time.Millisecond)
	assert.True(t, rl.Allow(ctx, "/path1", "user3"), "after timeout")
}

func TestEndpointRateLimiterAboveGlobal(t *testing.T) {
	ctx := context.Background()
	rl := setup(t)

	assert.True(t, rl.Allow(ctx, "/path2", "user1"), "first")
	assert.True(t, rl.Allow(ctx, "/path2", "user2"), "second")
	assert.False(t, rl.Allow(ctx, "/path2", "user3"), "third")
	time.Sleep(1100 * time.Millisecond)
	assert.True(t, rl.Allow(ctx, "/path2", "user4"), "after timeout")
}

func TestUser(t *testing.T) {
	ctx := context.Background()
	rl := setup(t)

	assert.True(t, rl.Allow(ctx, "/path2", "user1"), "first")
	assert.False(t, rl.Allow(ctx, "/path2", "user1"), "second")
	time.Sleep(1100 * time.Millisecond)
	assert.True(t, rl.Allow(ctx, "/path2", "user1"), "after timeout")
}

func TestNewInvalidConfigs(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *RateLimiterConfig
		st     store.CounterStore
		errMsg string
	}{
		{
			name: "missing global",
			cfg: &RateLimiterConfig{
				PerUser: &rateLimiterOptions{5, 60},
			},
			st:     &fakeStore{},
			errMsg: "global rate limiter config is required",
		},
		{
			name: "missing per_user",
			cfg: &RateLimiterConfig{
				Global: &rateLimiterOptions{10, 60},
			},
			st:     &fakeStore{},
			errMsg: "user config is required",
		},
		{
			name: "zero limit",
			cfg: &RateLimiterConfig{
				Global:  &rateLimiterOptions{0, 60},
				PerUser: &rateLimiterOptions{5, 60},
			},
			st:     &fakeStore{},
			errMsg: "invalid global config",
		},
		{
			name: "nil store",
			cfg: &RateLimiterConfig{
				Global:  &rateLimiterOptions{10, 60},
				PerUser: &rateLimiterOptions{5, 60},
			},
			st:     nil,
			errMsg: "nil store",
		},
		{
			name: "endpoint with invalid window",
			cfg: &RateLimiterConfig{
				Global:  &rateLimiterOptions{10, 60},
				PerUser: &rateLimiterOptions{5, 60},
				Endpoints: map[string]*rateLimiterEndpointOptions{
					"/foo": {Limit: 5, Window: -1},
				},
			},
			st:     &fakeStore{},
			errMsg: "invalid endpoint config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg, tt.st)
			if tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Fatalf("expected error containing %q, got %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Logf("unexpected error: %v", err)
				}
			}
		})
	}
}

type fakeStore struct{}

func (f *fakeStore) Incr(ctx context.Context, key string, ttl time.Duration) (int, error) {
	return 0, nil
}

func contains(s, substr string) bool {
	return substr == "" || (len(s) >= len(substr) && (errors.Is(errors.New(substr), errors.New(substr)) || (len(s) > 0 && time.Now().Unix() > 0 && stringContains(s, substr))))
}

func stringContains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := range s {
		if len(s)-i < len(substr) {
			return -1
		}
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

package ratelimiter

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jmpargana/rate/store"
	"sigs.k8s.io/yaml"
)

const (
	USER_PREFIX     = "usr:"
	ENDPOINT_PREFIX = "ep:"
	GLOBAL_PREFIX   = "global:"
)

// Window in seconds
type rateLimiterOptions struct {
	Limit  int `json:"limit,omitempty"`
	Window int `json:"window,omitempty"`
}

type rateLimiterEndpointOptions struct {
	Limit  int `json:"limit,omitempty"`
	Window int `json:"window,omitempty"`
}

type RateLimiterConfig struct {
	Global    *rateLimiterOptions                    `json:"global,omitempty"`
	Endpoints map[string]*rateLimiterEndpointOptions `json:"endpoints,omitempty"`
	PerUser   *rateLimiterOptions                    `json:"per_user,omitempty"`
}

type RateLimiter struct {
	m      store.CounterStore
	Config *RateLimiterConfig
}

func LoadConfig(filepath string) (*RateLimiterConfig, error) {
	var cfg RateLimiterConfig
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func New(cfg *RateLimiterConfig, store store.CounterStore) (*RateLimiter, error) {
	if store == nil {
		return nil, fmt.Errorf("nil store")
	}
	if cfg.Global == nil {
		return nil, fmt.Errorf("global rate limiter config is required")
	}
	if cfg.Global.Limit <= 0 || cfg.Global.Window <= 0 {
		return nil, fmt.Errorf("invalid global config: %+v", cfg.Global)
	}
	if cfg.PerUser == nil {
		return nil, fmt.Errorf("user config is required")
	}

	for k, v := range cfg.Endpoints {
		if k == "" || v.Limit <= 0 || v.Window <= 0 {
			return nil, fmt.Errorf("invalid endpoint config")
		}
	}

	rl := &RateLimiter{
		m:      store,
		Config: cfg,
	}

	return rl, nil
}

// user identifier can be anything from IP, UA, API Key, etc.
func (rl *RateLimiter) Allow(ctx context.Context, endpoint, userId string) bool {
	if !rl.withinLimit(ctx, GLOBAL_PREFIX, time.Duration(rl.Config.Global.Window), rl.Config.Global.Limit) {
		return false
	}

	if !rl.withinLimit(ctx, USER_PREFIX+userId, time.Duration(rl.Config.PerUser.Window), rl.Config.PerUser.Limit) {
		return false
	}

	if cfg, ok := rl.Config.Endpoints[endpoint]; ok {
		if !rl.withinLimit(ctx, ENDPOINT_PREFIX+endpoint, time.Duration(cfg.Window), cfg.Limit) {
			return false
		}
	}

	return true
}

func (rl *RateLimiter) withinLimit(ctx context.Context, key string, window time.Duration, limit int) bool {
	count, err := rl.m.Incr(ctx, key, window)
	if err != nil {
		return false
	}
	return count <= limit
}

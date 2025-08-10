package ratelimiter

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jmpargana/ratelimiter/store"
	"sigs.k8s.io/yaml"
)

const (
	USER_PREFIX     = "usr:"
	ENDPOINT_PREFIX = "ep:"
	GLOBAL_PREFIX   = "global:"
)

// rateLimiterOptions defines the rate limit settings for global and per-user rate limiting.
// Limit is the maximum allowed actions per window, and Window is the time window in seconds.
type rateLimiterOptions struct {
	Limit  int `json:"limit,omitempty"`
	Window int `json:"window,omitempty"`
}

// rateLimiterEndpointOptions defines rate limit settings specific to an endpoint.
// Limit is the maximum allowed actions per window, and Window is the time window in seconds.
type rateLimiterEndpointOptions struct {
	Limit  int `json:"limit,omitempty"`
	Window int `json:"window,omitempty"`
}

// RateLimiterConfig holds the configuration for the rate limiter including global,
// per-user, and per-endpoint limits.
type RateLimiterConfig struct {
	Global    *rateLimiterOptions                    `json:"global,omitempty"`
	Endpoints map[string]*rateLimiterEndpointOptions `json:"endpoints,omitempty"`
	PerUser   *rateLimiterOptions                    `json:"per_user,omitempty"`
}

// RateLimiter provides methods to enforce rate limiting based on the provided
// configuration and a counter store implementation.
type RateLimiter struct {
	m   store.CounterStore
	cfg *RateLimiterConfig
}

// LoadConfig reads a YAML configuration file from the given filepath
// and unmarshals it into a RateLimiterConfig struct.
//
// Returns an error if reading the file or unmarshalling fails.
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

// New creates a new RateLimiter instance using the provided configuration and store.
//
// Validates that the store is not nil, the global and per-user configurations
// are present and valid, and all endpoint configs are valid.
//
// Returns an error if validation fails.
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
		m:   store,
		cfg: cfg,
	}

	return rl, nil
}

// Allow checks whether a request from a given user to a specified endpoint
// is allowed based on the global, per-user, and per-endpoint rate limits.
//
// The userId parameter can be any identifier, such as IP, user agent, or API key.
//
// Returns true if the request is within all applicable limits; false otherwise.
func (rl *RateLimiter) Allow(ctx context.Context, endpoint, userId string) bool {
	if !rl.withinLimit(ctx, GLOBAL_PREFIX, time.Duration(rl.cfg.Global.Window)*time.Second, rl.cfg.Global.Limit) {
		return false
	}

	if !rl.withinLimit(ctx, USER_PREFIX+userId, time.Duration(rl.cfg.PerUser.Window)*time.Second, rl.cfg.PerUser.Limit) {
		return false
	}

	if cfg, ok := rl.cfg.Endpoints[endpoint]; ok {
		if !rl.withinLimit(ctx, ENDPOINT_PREFIX+endpoint, time.Duration(cfg.Window)*time.Second, cfg.Limit) {
			return false
		}
	}

	return true
}

// withinLimit increments the counter for the given key within the specified window duration
// and checks if the count is within the allowed limit.
//
// Returns true if the count is less than or equal to the limit; false otherwise.
func (rl *RateLimiter) withinLimit(ctx context.Context, key string, window time.Duration, limit int) bool {
	count, err := rl.m.Incr(ctx, key, window)
	if err != nil {
		return false
	}
	return count <= limit
}

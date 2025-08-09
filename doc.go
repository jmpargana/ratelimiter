// Package ratelimiter provides a configurable rate limiting mechanism.
//
// It supports limiting the number of actions (e.g., API requests) globally,
// per user, and per endpoint, using configurable limits and time windows.
//
// The package relies on an external counter store to track usage counts
// and enforces limits over sliding windows.
//
// Usage:
//
//	// Load configuration from YAML file
//	cfg, err := ratelimiter.LoadConfig("config.yaml")
//	if err != nil {
//	    log.Fatalf("failed to load config: %v", err)
//	}
//
//	// Create a store implementing store.CounterStore interface
//	store := store.NewMemoryStore() // example
//
//	// Create the rate limiter instance
//	rl, err := ratelimiter.New(cfg, store)
//	if err != nil {
//	    log.Fatalf("failed to create rate limiter: %v", err)
//	}
//
//	// Check if request is allowed
//	allowed := rl.Allow(context.Background(), "/api/data", "user123")
//	if !allowed {
//	    fmt.Println("Rate limit exceeded")
//	} else {
//	    fmt.Println("Request allowed")
//	}
//
// Configuration:
//
// The rate limiter configuration supports:
//
//   - Global limits applied to all requests.
//   - Per-user limits identified by any user identifier (IP, API key, etc.).
//   - Endpoint-specific limits overriding global/per-user limits.
//
// The configuration is loaded from a YAML file with the following structure:
//
//	global:
//	  limit: 1000
//	  window: 60
//	per_user:
//	  limit: 100
//	  window: 60
//	endpoints:
//	  /api/data:
//	    limit: 50
//	    window: 60
//
// The limit values represent the maximum allowed requests per window (seconds).
package ratelimiter

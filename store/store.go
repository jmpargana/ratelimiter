package store

import (
	"context"
	"time"
)

// CounterStore defines an interface for a concurrency-safe counter storage.
//
// The Incr method increments the counter associated with the given key,
// sets the expiration time to ttl (time-to-live), and returns the updated count.
//
// If the key does not exist, it is created with initial count 1 and TTL set.
//
// Typical implementations use this interface to track request counts for rate limiting.
type CounterStore interface {
	// Incr increments the counter for the given key with a TTL (expiration time).
	//
	// ctx: context for cancellation and deadlines.
	// key: identifier for the counter to increment.
	// ttl: duration after which the counter expires.
	//
	// Returns the incremented count and any error encountered.
	Incr(ctx context.Context, key string, ttl time.Duration) (count int, err error)
}

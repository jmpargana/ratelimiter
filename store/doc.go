// Package store defines interfaces for counter storage used in rate limiting.
//
// The CounterStore interface provides a method to increment a counter
// associated with a key and automatically expire it after a specified TTL.
//
// Implementations of this interface are responsible for concurrency-safe
// storage of counters, typically used by rate limiters to track usage counts.
//
// Example usage:
//
//	type MemoryStore struct {
//	    // implementation details...
//	}
//
//	func (m *MemoryStore) Incr(ctx context.Context, key string, ttl time.Duration) (int, error) {
//	    // increment the counter for key and set expiry to ttl
//	}
//
//	// Then use this store with a rate limiter or other components that require counting.
package store

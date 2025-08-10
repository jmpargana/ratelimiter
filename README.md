# ratelimiter

`ratelimiter` is a configurable Go package for rate limiting requests.  
It supports **global**, **per-user**, and **per-endpoint** limits, with configurable limits and sliding time windows.

The package works with an external counter store to track usage counts, making it flexible for different storage backends.

---

## Features

- **Global limits** — apply rules to all requests.
- **Per-user limits** — enforce limits based on a user identifier (e.g., IP, API key).
- **Endpoint-specific limits** — override global/per-user limits for specific paths.
- **Sliding window** enforcement.
- **Pluggable storage backend** — works with any store implementing the `store.CounterStore` interface.

---

## Installation

```bash
go get github.com/yourusername/ratelimiter
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/yourusername/ratelimiter"
    "github.com/yourusername/ratelimiter/store"
    "github.com/go-redis/redis/v9"
)

func main() {
    // Connect to a store
    rdb, err := redis.NewClient(redis.Opts{Addr: ""})
    if err != nil {
      log.Fatalf("failed connecting to redis: %v", err)
    }

    // Load configuration from YAML file
    cfg, err := ratelimiter.LoadConfig("config.yaml")
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    // Create a store implementing store.CounterStore interface
    s := store.NewRedisStore(rdb)

    // Create the rate limiter instance
    rl, err := ratelimiter.New(cfg, s)
    if err != nil {
        log.Fatalf("failed to create rate limiter: %v", err)
    }

    // Check if request is allowed
    allowed := rl.Allow(context.Background(), "/api/data", "user123")
    if !allowed {
        fmt.Println("Rate limit exceeded")
    } else {
        fmt.Println("Request allowed")
    }
}
```

## Configuration

The configuration is loaded from a YAML file and supports:

- Global limits for all requests.
- Per-user limits identified by a user ID.
- Endpoint-specific limits that override other limits.

Example `config.yaml`:

```yaml
global:
  limit: 1000
  window: 60
per_user:
  limit: 100
  window: 60
endpoints:
  /api/data:
    limit: 50
    window: 60
```
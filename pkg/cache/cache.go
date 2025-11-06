package cache

import (
	"context"
	"time"
)

// Cache is the interface for cache operations
type Cache interface {
	// Set sets a key-value pair with optional expiry
	Set(ctx context.Context, key string, value interface{}, expiry time.Duration) error

	// Get gets a value by key
	Get(ctx context.Context, key string) (string, error)

	// GetBytes gets a value as bytes
	GetBytes(ctx context.Context, key string) ([]byte, error)

	// Delete deletes a key
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists
	Exists(ctx context.Context, key string) (bool, error)

	// Increment increments a key's value
	Increment(ctx context.Context, key string) (int64, error)

	// Decrement decrements a key's value
	Decrement(ctx context.Context, key string) (int64, error)

	// Expire sets expiry on an existing key
	Expire(ctx context.Context, key string, expiry time.Duration) error

	// Keys gets all keys matching pattern
	Keys(ctx context.Context, pattern string) ([]string, error)

	// FlushAll flushes all keys in current database
	FlushAll(ctx context.Context) error

	// Close closes the cache connection
	Close() error

	// Ping checks if cache is alive
	Ping(ctx context.Context) error
}

// CacheKey is a helper to build cache keys with prefix
type CacheKey struct {
	prefix string
}

// NewCacheKey creates a new cache key builder
func NewCacheKey(prefix string) *CacheKey {
	return &CacheKey{prefix: prefix}
}

// Build builds a cache key with prefix
func (k *CacheKey) Build(parts ...string) string {
	key := k.prefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

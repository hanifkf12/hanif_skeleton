package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryCache implements Cache interface using in-memory map
// For development/testing purposes only - use Redis in production
type MemoryCache struct {
	data   map[string]*cacheItem
	mu     sync.RWMutex
	stopCh chan struct{}
}

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache instance
func NewMemoryCache() Cache {
	mc := &MemoryCache{
		data:   make(map[string]*cacheItem),
		stopCh: make(chan struct{}),
	}

	// Start cleanup goroutine
	go mc.cleanupExpired()

	return mc
}

// Set sets a key-value pair with optional expiry
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := &cacheItem{
		value: value,
	}

	if expiry > 0 {
		item.expiresAt = time.Now().Add(expiry)
	}

	c.data[key] = item
	return nil
}

// Get gets a value by key
func (c *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}

	// Check if expired
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return "", fmt.Errorf("key expired: %s", key)
	}

	return fmt.Sprintf("%v", item.value), nil
}

// GetBytes gets a value as bytes
func (c *MemoryCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

// Delete deletes a key
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

// Exists checks if a key exists
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return false, nil
	}

	// Check if expired
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return false, nil
	}

	return true, nil
}

// Increment increments a key's value
func (c *MemoryCache) Increment(ctx context.Context, key string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.data[key]
	if !exists {
		c.data[key] = &cacheItem{value: int64(1)}
		return 1, nil
	}

	val, ok := item.value.(int64)
	if !ok {
		return 0, fmt.Errorf("value is not an integer")
	}

	val++
	item.value = val
	return val, nil
}

// Decrement decrements a key's value
func (c *MemoryCache) Decrement(ctx context.Context, key string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.data[key]
	if !exists {
		c.data[key] = &cacheItem{value: int64(-1)}
		return -1, nil
	}

	val, ok := item.value.(int64)
	if !ok {
		return 0, fmt.Errorf("value is not an integer")
	}

	val--
	item.value = val
	return val, nil
}

// Expire sets expiry on an existing key
func (c *MemoryCache) Expire(ctx context.Context, key string, expiry time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.data[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	item.expiresAt = time.Now().Add(expiry)
	return nil
}

// Keys gets all keys matching pattern (simple prefix match)
func (c *MemoryCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var keys []string
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys, nil
}

// FlushAll flushes all keys
func (c *MemoryCache) FlushAll(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*cacheItem)
	return nil
}

// Close closes the cache
func (c *MemoryCache) Close() error {
	close(c.stopCh)
	return nil
}

// Ping checks if cache is alive
func (c *MemoryCache) Ping(ctx context.Context) error {
	return nil
}

// cleanupExpired removes expired items periodically
func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for k, item := range c.data {
				if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
					delete(c.data, k)
				}
			}
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}

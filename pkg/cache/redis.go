package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client) Cache {
	return &RedisCache{
		client: client,
	}
}

// Set sets a key-value pair with optional expiry
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	// Convert value to JSON if not string or []byte
	var data interface{}
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = v
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		data = jsonData
	}

	return c.client.Set(ctx, key, data, expiry).Err()
}

// Get gets a value by key
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", err
	}
	return val, nil
}

// GetBytes gets a value as bytes
func (c *RedisCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("key not found: %s", key)
		}
		return nil, err
	}
	return val, nil
}

// Delete deletes a key
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists checks if a key exists
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	val, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// Increment increments a key's value
func (c *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// Decrement decrements a key's value
func (c *RedisCache) Decrement(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// Expire sets expiry on an existing key
func (c *RedisCache) Expire(ctx context.Context, key string, expiry time.Duration) error {
	return c.client.Expire(ctx, key, expiry).Err()
}

// Keys gets all keys matching pattern
func (c *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// FlushAll flushes all keys in current database
func (c *RedisCache) FlushAll(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Ping checks if Redis is alive
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// GetJSON gets a value and unmarshals it from JSON
func (c *RedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.GetBytes(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal(val, dest)
}

// SetJSON marshals value to JSON and sets it
func (c *RedisCache) SetJSON(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.Set(ctx, key, jsonData, expiry)
}

// SetNX sets a key only if it doesn't exist (atomic)
func (c *RedisCache) SetNX(ctx context.Context, key string, value interface{}, expiry time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiry).Result()
}

// GetDel gets a value and deletes it atomically
func (c *RedisCache) GetDel(ctx context.Context, key string) (string, error) {
	val, err := c.client.GetDel(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", err
	}
	return val, nil
}

// MGet gets multiple keys at once
func (c *RedisCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return c.client.MGet(ctx, keys...).Result()
}

// MSet sets multiple key-value pairs at once
func (c *RedisCache) MSet(ctx context.Context, pairs map[string]interface{}) error {
	// Convert map to slice of interface{}
	args := make([]interface{}, 0, len(pairs)*2)
	for k, v := range pairs {
		args = append(args, k, v)
	}
	return c.client.MSet(ctx, args...).Err()
}

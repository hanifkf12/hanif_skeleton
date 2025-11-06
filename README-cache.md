# Cache Package Documentation

## Overview

Cache package menyediakan abstraksi unified untuk caching dengan **Redis** sebagai backend utama dan **in-memory** untuk development/testing. Mengikuti **Clean Architecture** dengan 1 interface dan 2 implementasi.

## Architecture

```
┌─────────────────────────────────────┐
│         UseCase Layer               │
├─────────────────────────────────────┤
│    Cache Interface (Contract)       │  ← Abstraction
├─────────────────────────────────────┤
│    Implementation Layer             │
│    ├─ Redis Cache (Production)     │
│    └─ Memory Cache (Development)   │
└─────────────────────────────────────┘
```

## Cache Interface

```go
type Cache interface {
    Set(ctx context.Context, key string, value interface{}, expiry time.Duration) error
    Get(ctx context.Context, key string) (string, error)
    GetBytes(ctx context.Context, key string) ([]byte, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    Increment(ctx context.Context, key string) (int64, error)
    Decrement(ctx context.Context, key string) (int64, error)
    Expire(ctx context.Context, key string, expiry time.Duration) error
    Keys(ctx context.Context, pattern string) ([]string, error)
    FlushAll(ctx context.Context) error
    Close() error
    Ping(ctx context.Context) error
}
```

## Implementations

### 1. Redis Cache (Production)
- **Driver**: `redis`
- **Use Case**: Production, distributed caching, session storage
- **File**: `pkg/cache/redis.go`
- **Features**: Full Redis support, atomic operations, persistence

### 2. Memory Cache (Development)
- **Driver**: `memory`
- **Use Case**: Development, testing, single instance apps
- **File**: `pkg/cache/memory.go`
- **Features**: Fast in-memory, auto cleanup, no external dependencies

## Configuration

### Environment Variables

Add to `.env`:

```bash
# Cache Configuration
CACHE_DRIVER=redis          # redis or memory
CACHE_HOST=localhost        # Redis host
CACHE_PORT=6379            # Redis port
CACHE_PASSWORD=            # Redis password (optional)
CACHE_DB=0                 # Redis database number (0-15)
```

### Config Struct

File: `pkg/config/cache.go`

```go
type Cache struct {
    Driver   string // redis, memory
    Host     string
    Port     int
    Password string
    DB       int
}
```

## Bootstrap Registry

File: `internal/bootstrap/cache.go`

```go
// Initialize cache based on configuration
cache := bootstrap.RegistryCache(cfg)
defer cache.Close()
```

**Registry automatically:**
- ✅ Selects implementation based on `CACHE_DRIVER`
- ✅ Tests Redis connection (fatal if fails)
- ✅ Provides sensible defaults
- ✅ Logs initialization

---

## Usage Examples

### 1. Basic Set & Get

```go
package usecase

import (
    "context"
    "time"
    "github.com/hanifkf12/hanif_skeleton/pkg/cache"
)

func cacheExample(ctx context.Context, cache cache.Cache) {
    // Set with 5 minute expiry
    err := cache.Set(ctx, "user:1", "John Doe", 5*time.Minute)
    
    // Get value
    value, err := cache.Get(ctx, "user:1")
    // value = "John Doe"
    
    // Set without expiry (never expires)
    cache.Set(ctx, "config:version", "1.0.0", 0)
}
```

### 2. Cache with JSON

```go
import "encoding/json"

func cacheJSON(ctx context.Context, cache cache.Cache) {
    // Struct to cache
    user := User{
        ID:   1,
        Name: "John Doe",
        Email: "john@example.com",
    }
    
    // Marshal to JSON
    jsonData, _ := json.Marshal(user)
    
    // Set in cache
    cache.Set(ctx, "user:1:profile", jsonData, 10*time.Minute)
    
    // Get from cache
    data, _ := cache.GetBytes(ctx, "user:1:profile")
    
    // Unmarshal
    var cachedUser User
    json.Unmarshal(data, &cachedUser)
}
```

### 3. Cache Key Builder

```go
// Build structured cache keys
keyBuilder := cache.NewCacheKey("myapp")

// Build keys with prefix
userKey := keyBuilder.Build("user", "123")
// Result: "myapp:user:123"

sessionKey := keyBuilder.Build("session", "abc-def-ghi")
// Result: "myapp:session:abc-def-ghi"

listKey := keyBuilder.Build("users", "list", "page:1")
// Result: "myapp:users:list:page:1"
```

### 4. Check Existence

```go
exists, err := cache.Exists(ctx, "user:1")
if exists {
    // Key exists, get from cache
    value, _ := cache.Get(ctx, "user:1")
} else {
    // Key doesn't exist, query database
    user := db.GetUser(1)
    cache.Set(ctx, "user:1", user, 5*time.Minute)
}
```

### 5. Delete Cache

```go
// Delete single key
err := cache.Delete(ctx, "user:1")

// Delete multiple keys by pattern
keys, _ := cache.Keys(ctx, "user:*")
for _, key := range keys {
    cache.Delete(ctx, key)
}

// Flush all cache (use with caution!)
cache.FlushAll(ctx)
```

### 6. Counter Operations

```go
// Increment counter
count, err := cache.Increment(ctx, "page:views")
// count = 1

cache.Increment(ctx, "page:views")
// count = 2

// Decrement counter
cache.Decrement(ctx, "inventory:item:123")
```

### 7. Set Expiry on Existing Key

```go
// Set value without expiry
cache.Set(ctx, "temp:data", "value", 0)

// Later, set expiry
cache.Expire(ctx, "temp:data", 1*time.Hour)
```

---

## Integration with UseCase

### Cache-Aside Pattern

```go
package usecase

type userProfile struct {
    userRepo repository.UserRepository
    cache    cache.Cache
}

func (u *userProfile) Serve(data appctx.Data) appctx.Response {
    ctx := data.FiberCtx.UserContext()
    userID := data.FiberCtx.Params("id")
    
    // Build cache key
    cacheKey := cache.NewCacheKey("user").Build(userID, "profile")
    
    // Try cache first
    cachedData, err := u.cache.GetBytes(ctx, cacheKey)
    if err == nil {
        // Cache hit
        var user User
        json.Unmarshal(cachedData, &user)
        return *appctx.NewResponse().WithData(user)
    }
    
    // Cache miss - get from database
    user, err := u.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return *appctx.NewResponse().
            WithCode(fiber.StatusNotFound).
            WithErrors("User not found")
    }
    
    // Store in cache for next time
    jsonData, _ := json.Marshal(user)
    u.cache.Set(ctx, cacheKey, jsonData, 5*time.Minute)
    
    return *appctx.NewResponse().WithData(user)
}
```

### Write-Through Pattern

```go
func (u *updateUser) Serve(data appctx.Data) appctx.Response {
    ctx := data.FiberCtx.UserContext()
    
    // Parse request
    var req UpdateUserRequest
    data.FiberCtx.BodyParser(&req)
    
    // Update database
    err := u.userRepo.UpdateUser(ctx, req)
    if err != nil {
        return *appctx.NewResponse().
            WithCode(fiber.StatusInternalServerError).
            WithErrors(err.Error())
    }
    
    // Update cache immediately (write-through)
    cacheKey := cache.NewCacheKey("user").Build(req.ID, "profile")
    jsonData, _ := json.Marshal(req)
    u.cache.Set(ctx, cacheKey, jsonData, 5*time.Minute)
    
    return *appctx.NewResponse().WithData(req)
}
```

### Cache Invalidation

```go
func (u *deleteUser) Serve(data appctx.Data) appctx.Response {
    ctx := data.FiberCtx.UserContext()
    userID := data.FiberCtx.Params("id")
    
    // Delete from database
    err := u.userRepo.DeleteUser(ctx, userID)
    if err != nil {
        return *appctx.NewResponse().
            WithCode(fiber.StatusInternalServerError).
            WithErrors(err.Error())
    }
    
    // Invalidate cache
    cacheKey := cache.NewCacheKey("user").Build(userID, "profile")
    u.cache.Delete(ctx, cacheKey)
    
    // Also invalidate list cache
    listKey := cache.NewCacheKey("users").Build("list")
    u.cache.Delete(ctx, listKey)
    
    return *appctx.NewResponse().
        WithData(map[string]string{"message": "User deleted"})
}
```

---

## Setup Instructions

### Development (Memory Cache)

1. **Update .env:**
```bash
CACHE_DRIVER=memory
```

2. **No additional setup needed!**

### Production (Redis)

1. **Install Redis:**

```bash
# macOS
brew install redis
brew services start redis

# Ubuntu/Debian
sudo apt-get install redis-server
sudo systemctl start redis

# Docker
docker run -d -p 6379:6379 --name redis redis:alpine
```

2. **Update .env:**
```bash
CACHE_DRIVER=redis
CACHE_HOST=localhost
CACHE_PORT=6379
CACHE_PASSWORD=
CACHE_DB=0
```

3. **Test connection:**
```bash
redis-cli ping
# Expected: PONG
```

### Production (Redis with Password)

```bash
# Set Redis password
redis-cli
> CONFIG SET requirepass "your-strong-password"
> AUTH your-strong-password
> CONFIG REWRITE
```

**Update .env:**
```bash
CACHE_PASSWORD=your-strong-password
```

---

## Cache Patterns & Best Practices

### 1. Cache-Aside (Lazy Loading)

```go
// Read from cache, if miss read from DB and cache it
data, err := cache.Get(ctx, key)
if err != nil {
    // Cache miss
    data = db.Query()
    cache.Set(ctx, key, data, expiry)
}
return data
```

**Pros:** Cache only what's needed
**Cons:** First request is slow (cache miss)

### 2. Write-Through

```go
// Write to DB and cache simultaneously
db.Update(data)
cache.Set(ctx, key, data, expiry)
```

**Pros:** Cache always up-to-date
**Cons:** Extra write latency

### 3. Write-Behind (Write-Back)

```go
// Write to cache first, async write to DB
cache.Set(ctx, key, data, expiry)
go func() {
    db.Update(data)
}()
```

**Pros:** Very fast writes
**Cons:** Risk of data loss if cache fails

### 4. Cache Invalidation

```go
// On update/delete, remove from cache
db.Delete(id)
cache.Delete(ctx, key)

// Or use TTL (Time To Live)
cache.Set(ctx, key, data, 5*time.Minute)
```

**Golden Rule:** "There are only two hard things in Computer Science: cache invalidation and naming things." - Phil Karlton

---

## Performance Tips

### 1. Use Appropriate TTL

```go
// Frequently changing data - short TTL
cache.Set(ctx, "stock:price", price, 10*time.Second)

// Rarely changing data - long TTL
cache.Set(ctx, "config:settings", settings, 24*time.Hour)

// Static data - no expiry
cache.Set(ctx, "product:categories", categories, 0)
```

### 2. Batch Operations

```go
// Instead of multiple Set calls
cache.Set(ctx, "user:1", data1, expiry)
cache.Set(ctx, "user:2", data2, expiry)
cache.Set(ctx, "user:3", data3, expiry)

// Use MSet (Redis only)
pairs := map[string]interface{}{
    "user:1": data1,
    "user:2": data2,
    "user:3": data3,
}
redisCache.MSet(ctx, pairs)
```

### 3. Use Structured Keys

```go
// ✅ GOOD - Structured, easy to query
cache.NewCacheKey("app").Build("user", "123", "profile")
// Result: app:user:123:profile

// ❌ BAD - Unstructured
cache.Set(ctx, "user123profile", data, expiry)
```

### 4. Monitor Cache Hit Rate

```go
type CacheMetrics struct {
    Hits   int64
    Misses int64
}

func (m *CacheMetrics) HitRate() float64 {
    total := m.Hits + m.Misses
    if total == 0 {
        return 0
    }
    return float64(m.Hits) / float64(total) * 100
}

// Target: >80% hit rate for good cache performance
```

---

## Redis-Specific Features

### 1. SetNX (Set if Not Exists)

```go
// Atomic set only if key doesn't exist
success, err := redisCache.SetNX(ctx, "lock:resource", "locked", 10*time.Second)
if success {
    // Lock acquired
    defer cache.Delete(ctx, "lock:resource")
    // Do critical section work
}
```

**Use Case:** Distributed locks, preventing duplicate processing

### 2. GetDel (Get and Delete)

```go
// Atomically get value and delete it
value, err := redisCache.GetDel(ctx, "one-time-token")
// Token is now deleted, can only be used once
```

**Use Case:** One-time tokens, message queues

### 3. Multiple Get/Set

```go
// Get multiple keys at once
values, err := redisCache.MGet(ctx, "key1", "key2", "key3")

// Set multiple keys at once
pairs := map[string]interface{}{
    "key1": "value1",
    "key2": "value2",
}
redisCache.MSet(ctx, pairs)
```

---

## Monitoring & Debugging

### Cache Stats Endpoint

See: `internal/usecase/cache_example.go`

```bash
# Get cache statistics
curl http://localhost:9000/cache/stats

# Response:
{
  "code": 200,
  "data": {
    "status": "ok",
    "keys": ["user:1", "user:2", "session:abc"],
    "count": 3
  }
}
```

### Clear Cache Endpoint

```bash
# Clear all cache
curl -X DELETE http://localhost:9000/cache/clear

# Clear specific key
curl -X DELETE http://localhost:9000/cache/clear?key=user:1
```

### Redis CLI Commands

```bash
# Connect to Redis
redis-cli

# List all keys
KEYS *

# Get value
GET user:1

# Delete key
DEL user:1

# Check if key exists
EXISTS user:1

# Get TTL
TTL user:1

# Flush all
FLUSHDB
```

---

## Error Handling

```go
func handleCacheError(err error) {
    if err != nil {
        // Log error but don't fail request
        logger.Error("Cache error", logger.NewFields("Cache").
            Append(logger.Any("error", err.Error())))
        
        // Fall back to database
        return queryDatabase()
    }
}

// Cache should never break your application
// Always have a fallback to database
```

---

## Testing

### Unit Test with Memory Cache

```go
func TestCachePattern(t *testing.T) {
    // Use memory cache for testing
    cache := cache.NewMemoryCache()
    defer cache.Close()
    
    ctx := context.Background()
    
    // Test set
    err := cache.Set(ctx, "test:key", "value", 1*time.Minute)
    assert.NoError(t, err)
    
    // Test get
    value, err := cache.Get(ctx, "test:key")
    assert.NoError(t, err)
    assert.Equal(t, "value", value)
    
    // Test delete
    cache.Delete(ctx, "test:key")
    exists, _ := cache.Exists(ctx, "test:key")
    assert.False(t, exists)
}
```

### Mock Cache Interface

```go
type MockCache struct {
    mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
    args := m.Called(ctx, key)
    return args.String(0), args.Error(1)
}

// Use in tests
mockCache := new(MockCache)
mockCache.On("Get", mock.Anything, "user:1").Return("John", nil)
```

---

## Security Considerations

### 1. Don't Cache Sensitive Data

```go
// ❌ BAD - Caching passwords
cache.Set(ctx, "user:password", password, expiry)

// ✅ GOOD - Cache non-sensitive data only
cache.Set(ctx, "user:profile", publicProfile, expiry)
```

### 2. Use Separate Redis Databases

```go
// DB 0: User sessions
// DB 1: Application cache
// DB 2: Rate limiting

// Configure per environment
CACHE_DB=0  // Production
CACHE_DB=1  // Staging
```

### 3. Network Security

```bash
# Bind to localhost only (if Redis on same server)
bind 127.0.0.1

# Or use password
requirepass your-strong-password

# Use TLS in production
tls-port 6380
tls-cert-file /path/to/cert.pem
tls-key-file /path/to/key.pem
```

---

## Troubleshooting

### Issue: Connection refused

**Solution:** Check if Redis is running

```bash
# Check Redis status
redis-cli ping

# Start Redis
# macOS
brew services start redis

# Linux
sudo systemctl start redis
```

### Issue: High memory usage

**Solution:** Set max memory and eviction policy

```bash
redis-cli
> CONFIG SET maxmemory 256mb
> CONFIG SET maxmemory-policy allkeys-lru
> CONFIG REWRITE
```

### Issue: Cache always misses

**Solution:** Check key naming and TTL

```go
// Debug cache keys
keys, _ := cache.Keys(ctx, "*")
logger.Info("Cache keys", logger.Any("keys", keys))
```

---

## Summary

Cache package provides:
- ✅ **Unified interface** across Redis and Memory
- ✅ **Production ready** Redis implementation
- ✅ **Development friendly** Memory cache
- ✅ **Bootstrap integration** for easy setup
- ✅ **Cache key builder** for structured keys
- ✅ **Complete operations** (Set, Get, Delete, Increment, etc.)
- ✅ **JSON support** for complex data
- ✅ **Redis-specific features** (SetNX, GetDel, MGet/MSet)
- ✅ **Example usecases** included

**Choose your driver:**
- **memory**: Development, testing, single instance
- **redis**: Production, distributed, high performance

---

**Files:**
- Interface: `pkg/cache/cache.go`
- Redis: `pkg/cache/redis.go`
- Memory: `pkg/cache/memory.go`
- Config: `pkg/config/cache.go`
- Bootstrap: `internal/bootstrap/cache.go`
- Examples: `internal/usecase/cache_example.go`

**Setup Redis:**
```bash
# Install
brew install redis  # macOS
sudo apt install redis-server  # Ubuntu

# Start
brew services start redis  # macOS
sudo systemctl start redis  # Ubuntu

# Test
redis-cli ping  # Should return PONG
```

**Your app now has production-ready caching!** 🚀⚡


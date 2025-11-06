# Custom Middleware System Documentation

## Overview

Custom middleware system yang terintegrasi dengan **Clean Architecture** routing. Middleware berjalan sebelum handler, mengembalikan `appctx.Response`, dan jika code **200** berarti sukses, selain itu eksekusi dihentikan dan response middleware dikembalikan.

## Architecture

```
┌─────────────────────────────────────┐
│     HTTP Request                    │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Middleware 1 (Auth)                │
│  Returns 200? → Continue            │
│  Returns 401? → Stop, return error  │
└────────┬────────────────────────────┘
         │ (if 200)
         ▼
┌─────────────────────────────────────┐
│  Middleware 2 (HMAC)                │
│  Returns 200? → Continue            │
│  Returns 401? → Stop, return error  │
└────────┬────────────────────────────┘
         │ (if 200)
         ▼
┌─────────────────────────────────────┐
│  Handler (UseCase)                  │
│  Business Logic                     │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│     HTTP Response                   │
└─────────────────────────────────────┘
```

## Middleware Interface

```go
// Middleware is a function that processes request before reaching the handler
// Returns appctx.Response with code 200 if middleware passes, otherwise returns error response
type Middleware func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response
```

**Contract:**
- Returns `appctx.Response`
- Code **200** = middleware passed, continue to next
- Code **!= 200** = middleware failed, stop and return response

## Built-in Middlewares

### 1. **Bearer Auth** - `middleware.BearerAuth()`

Validates Bearer token from `Authorization` header.

**Usage:**
```go
middleware.BearerAuth([]string{"valid-token-123", "admin-token-456"})
```

**Expected Header:**
```
Authorization: Bearer valid-token-123
```

**Returns:**
- `200` - Token valid
- `401` - Missing/invalid token

**Example:**
```go
rtr.fiber.Get("/protected", rtr.handleWithMiddleware(
    handler.HttpRequest,
    protectedUseCase,
    middleware.BearerAuth([]string{"token-123"}),
))
```

**Test:**
```bash
# Success
curl -H "Authorization: Bearer valid-token-123" \
  http://localhost:9000/users

# Fail
curl http://localhost:9000/users
# Response: {"code": 401, "errors": "Missing authorization header"}
```

---

### 2. **HMAC Auth** - `middleware.HMACAuth()`

Validates HMAC signature for request integrity.

**Usage:**
```go
middleware.HMACAuth("your-hmac-secret-key")
```

**Expected Headers:**
```
X-Signature: <hmac-sha256-hex-signature>
X-Timestamp: <unix-timestamp>
```

**Signature Calculation:**
```
message = METHOD + PATH + TIMESTAMP + BODY
signature = HMAC-SHA256(secret, message)
```

**Returns:**
- `200` - Signature valid
- `401` - Missing/invalid signature

**Example:**
```go
rtr.fiber.Post("/secure-endpoint", rtr.handleWithMiddleware(
    handler.HttpRequest,
    secureUseCase,
    middleware.HMACAuth("secret-key-123"),
))
```

**Generate Signature (Example in Go):**
```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func generateHMAC(method, path, timestamp, body, secret string) string {
    message := method + path + timestamp + body
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(message))
    return hex.EncodeToString(h.Sum(nil))
}

// Usage
timestamp := "1699200000"
signature := generateHMAC("POST", "/campaigns", timestamp, `{"name":"test"}`, "secret-key-123")
```

**Test:**
```bash
curl -X POST http://localhost:9000/campaigns \
  -H "Content-Type: application/json" \
  -H "X-Timestamp: 1699200000" \
  -H "X-Signature: <calculated-signature>" \
  -d '{"name":"test"}'
```

---

### 3. **API Key Auth** - `middleware.APIKeyAuth()`

Validates API key from custom header.

**Usage:**
```go
middleware.APIKeyAuth("X-API-Key", []string{"key-123", "key-456"})
```

**Expected Header:**
```
X-API-Key: key-123
```

**Returns:**
- `200` - API key valid
- `401` - Missing/invalid API key

**Example:**
```go
rtr.fiber.Get("/api/data", rtr.handleWithMiddleware(
    handler.HttpRequest,
    dataUseCase,
    middleware.APIKeyAuth("X-API-Key", []string{"api-key-123"}),
))
```

**Test:**
```bash
curl -H "X-API-Key: api-key-123" \
  http://localhost:9000/campaigns
```

---

### 4. **Rate Limit** - `middleware.RateLimit()`

Limits requests per IP address.

**Usage:**
```go
middleware.RateLimit(middleware.RateLimitConfig{
    MaxRequests: 10,
    WindowSize:  60, // seconds
})
```

**Returns:**
- `200` - Within limit
- `429` - Rate limit exceeded

**Example:**
```go
rtr.fiber.Post("/public/contact", rtr.handleWithMiddleware(
    handler.HttpRequest,
    contactUseCase,
    middleware.RateLimit(middleware.RateLimitConfig{
        MaxRequests: 5,
        WindowSize:  60,
    }),
))
```

**⚠️ Note:** Current implementation uses in-memory map (for demo). Use **Redis** for production.

---

### 5. **Content Type Validator** - `middleware.ContentTypeValidator()`

Validates `Content-Type` header.

**Usage:**
```go
middleware.ContentTypeValidator([]string{"application/json"})
```

**Returns:**
- `200` - Content type valid
- `415` - Unsupported media type

**Example:**
```go
rtr.fiber.Post("/users", rtr.handleWithMiddleware(
    handler.HttpRequest,
    createUserUseCase,
    middleware.ContentTypeValidator([]string{"application/json"}),
))
```

**Test:**
```bash
# Success
curl -X POST http://localhost:9000/users \
  -H "Content-Type: application/json" \
  -d '{"username":"test"}'

# Fail
curl -X POST http://localhost:9000/users \
  -H "Content-Type: text/plain" \
  -d 'test'
# Response: {"code": 415, "errors": "Unsupported content type"}
```

---

### 6. **IP Whitelist** - `middleware.IPWhitelist()`

Restricts access to whitelisted IP addresses.

**Usage:**
```go
middleware.IPWhitelist([]string{"127.0.0.1", "10.0.0.1"})
```

**Returns:**
- `200` - IP in whitelist
- `403` - Access denied

**Example:**
```go
rtr.fiber.Get("/admin/stats", rtr.handleWithMiddleware(
    handler.HttpRequest,
    statsUseCase,
    middleware.IPWhitelist([]string{"127.0.0.1"}),
))
```

---

## Router Integration

### Updated Router Methods

**1. `handle()` - Without middleware (backward compatible):**
```go
rtr.fiber.Get("/health", rtr.handle(
    handler.HttpRequest,
    healthUseCase,
))
```

**2. `handleWithMiddleware()` - With middleware:**
```go
rtr.fiber.Get("/protected", rtr.handleWithMiddleware(
    handler.HttpRequest,
    protectedUseCase,
    middleware.BearerAuth([]string{"token"}),
    middleware.ContentTypeValidator([]string{"application/json"}),
))
```

### Middleware Execution Order

Middlewares are executed **in order** from left to right:

```go
rtr.fiber.Post("/secure", rtr.handleWithMiddleware(
    handler.HttpRequest,
    secureUseCase,
    middleware.IPWhitelist([]string{"127.0.0.1"}),  // 1st - Check IP
    middleware.BearerAuth([]string{"token"}),       // 2nd - Check Auth
    middleware.HMACAuth("secret"),                   // 3rd - Check HMAC
    middleware.ContentTypeValidator([]string{"application/json"}), // 4th
))
```

**Flow:**
1. IPWhitelist checks IP → 200 → continue
2. BearerAuth checks token → 200 → continue
3. HMACAuth checks signature → 200 → continue
4. ContentTypeValidator checks Content-Type → 200 → continue
5. Handler executes

**If any middleware returns non-200, execution stops immediately.**

---

## Usage Examples

### Example 1: Public Endpoint (No Auth)

```go
healthUseCase := usecase.NewHealth(homeRepo)
rtr.fiber.Get("/health", rtr.handle(
    handler.HttpRequest,
    healthUseCase,
))
```

### Example 2: Protected with Bearer Token

```go
userUseCase := usecase.NewUser(userRepository)
rtr.fiber.Get("/users", rtr.handleWithMiddleware(
    handler.HttpRequest,
    userUseCase,
    middleware.BearerAuth([]string{"valid-token-123"}),
))
```

### Example 3: Protected with API Key

```go
campaignUseCase := usecase.NewCampaign(campaignRepository)
rtr.fiber.Get("/campaigns", rtr.handleWithMiddleware(
    handler.HttpRequest,
    campaignUseCase,
    middleware.APIKeyAuth("X-API-Key", []string{"api-key-123"}),
))
```

### Example 4: Secure Endpoint with HMAC + Content Type

```go
createCampaignUseCase := usecase.NewCreateCampaign(campaignRepository)
rtr.fiber.Post("/campaigns", rtr.handleWithMiddleware(
    handler.HttpRequest,
    createCampaignUseCase,
    middleware.HMACAuth("your-hmac-secret-key"),
    middleware.ContentTypeValidator([]string{"application/json"}),
))
```

### Example 5: Admin Only with Multiple Checks

```go
deleteUserUseCase := usecase.NewDeleteUser(userRepository)
rtr.fiber.Delete("/users/:id", rtr.handleWithMiddleware(
    handler.HttpRequest,
    deleteUserUseCase,
    middleware.BearerAuth([]string{"admin-token-456"}),
    middleware.IPWhitelist([]string{"10.0.0.1", "10.0.0.2"}),
))
```

### Example 6: Rate Limited Public Endpoint

```go
contactUseCase := usecase.NewContact()
rtr.fiber.Post("/public/contact", rtr.handleWithMiddleware(
    handler.HttpRequest,
    contactUseCase,
    middleware.RateLimit(middleware.RateLimitConfig{
        MaxRequests: 5,
        WindowSize:  60, // 5 requests per minute
    }),
    middleware.ContentTypeValidator([]string{"application/json"}),
))
```

---

## Creating Custom Middleware

### Template

```go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "github.com/hanifkf12/hanif_skeleton/internal/appctx"
    "github.com/hanifkf12/hanif_skeleton/pkg/config"
    "github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

// CustomMiddleware does something
func CustomMiddleware(param string) Middleware {
    return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
        lf := logger.NewFields("Middleware.Custom")

        // Your validation logic here
        if !isValid(ctx, param) {
            lf.Append(logger.Any("error", "validation failed"))
            logger.Error("Custom middleware failed", lf)
            return *appctx.NewResponse().
                WithCode(fiber.StatusForbidden).
                WithErrors("Validation failed")
        }

        logger.Info("Custom middleware passed", lf)
        return *appctx.NewResponse().WithCode(fiber.StatusOK)
    }
}
```

### Example: JWT Middleware

```go
package middleware

import (
    "strings"
    "github.com/gofiber/fiber/v2"
    "github.com/golang-jwt/jwt/v5"
    "github.com/hanifkf12/hanif_skeleton/internal/appctx"
    "github.com/hanifkf12/hanif_skeleton/pkg/config"
    "github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

func JWTAuth(secret string) Middleware {
    return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
        lf := logger.NewFields("Middleware.JWTAuth")

        // Get token from header
        authHeader := ctx.Get("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            return *appctx.NewResponse().
                WithCode(fiber.StatusUnauthorized).
                WithErrors("Missing or invalid authorization header")
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // Parse and validate token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(secret), nil
        })

        if err != nil || !token.Valid {
            lf.Append(logger.Any("error", err.Error()))
            logger.Error("JWT validation failed", lf)
            return *appctx.NewResponse().
                WithCode(fiber.StatusUnauthorized).
                WithErrors("Invalid token")
        }

        // Store claims in context
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            ctx.Locals("user_id", claims["user_id"])
            ctx.Locals("role", claims["role"])
        }

        logger.Info("JWT validation successful", lf)
        return *appctx.NewResponse().WithCode(fiber.StatusOK)
    }
}
```

### Example: Role-Based Access Control

```go
func RequireRole(allowedRoles []string) Middleware {
    return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
        lf := logger.NewFields("Middleware.RequireRole")

        // Get role from context (set by previous auth middleware)
        role, ok := ctx.Locals("role").(string)
        if !ok {
            return *appctx.NewResponse().
                WithCode(fiber.StatusForbidden).
                WithErrors("Role not found")
        }

        // Check if role is allowed
        for _, allowedRole := range allowedRoles {
            if role == allowedRole {
                logger.Info("Role check passed", lf)
                return *appctx.NewResponse().WithCode(fiber.StatusOK)
            }
        }

        lf.Append(logger.Any("role", role))
        lf.Append(logger.Any("allowed_roles", allowedRoles))
        logger.Error("Role check failed", lf)
        return *appctx.NewResponse().
            WithCode(fiber.StatusForbidden).
            WithErrors("Insufficient permissions")
    }
}

// Usage
rtr.fiber.Delete("/users/:id", rtr.handleWithMiddleware(
    handler.HttpRequest,
    deleteUserUseCase,
    middleware.JWTAuth("jwt-secret"),
    middleware.RequireRole([]string{"admin", "superadmin"}),
))
```

---

## Testing Middleware

### Unit Test

```go
func TestBearerAuth(t *testing.T) {
    app := fiber.New()
    cfg := &config.Config{}

    // Create middleware
    mw := middleware.BearerAuth([]string{"valid-token"})

    // Test valid token
    ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
    ctx.Request().Header.Set("Authorization", "Bearer valid-token")
    
    resp := mw(ctx, cfg)
    assert.Equal(t, 200, resp.Code)

    // Test invalid token
    ctx.Request().Header.Set("Authorization", "Bearer invalid-token")
    resp = mw(ctx, cfg)
    assert.Equal(t, 401, resp.Code)
}
```

### Integration Test

```bash
# Test protected endpoint
curl -X GET http://localhost:9000/users \
  -H "Authorization: Bearer valid-token-123"

# Expected: {"code": 200, "data": [...]}

# Test without auth
curl -X GET http://localhost:9000/users

# Expected: {"code": 401, "errors": "Missing authorization header"}
```

---

## Best Practices

### 1. Order Matters

```go
// ✅ GOOD - Check cheap operations first
middleware.IPWhitelist(...),        // Fast
middleware.RateLimit(...),          // Fast
middleware.BearerAuth(...),         // Medium
middleware.HMACAuth(...),           // Expensive

// ❌ BAD - Expensive operations first
middleware.HMACAuth(...),           // Runs even for blocked IPs
middleware.IPWhitelist(...),
```

### 2. Store Data in Context

```go
// In middleware
ctx.Locals("user_id", userID)
ctx.Locals("role", "admin")

// In usecase/handler
userID := data.FiberCtx.Locals("user_id").(int)
role := data.FiberCtx.Locals("role").(string)
```

### 3. Use Configuration

```go
// ❌ BAD - Hardcoded
middleware.BearerAuth([]string{"token-123"})

// ✅ GOOD - From config/env
validTokens := cfg.Auth.ValidTokens
middleware.BearerAuth(validTokens)
```

### 4. Logging

```go
// Always log middleware execution
lf := logger.NewFields("Middleware.Name")
lf.Append(logger.Any("path", ctx.Path()))
lf.Append(logger.Any("method", ctx.Method()))
logger.Info("Middleware executed", lf)
```

### 5. Error Messages

```go
// ❌ BAD - Exposes internal details
WithErrors("Token 'abc123' not found in database")

// ✅ GOOD - Generic message
WithErrors("Invalid token")
```

---

## Production Considerations

### 1. Rate Limiting

Use **Redis** instead of in-memory map:

```go
import "github.com/go-redis/redis/v8"

func RateLimitRedis(redis *redis.Client, maxRequests int) Middleware {
    return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
        key := "rate:" + ctx.IP()
        count, _ := redis.Incr(ctx.Context(), key).Result()
        
        if count == 1 {
            redis.Expire(ctx.Context(), key, 60*time.Second)
        }
        
        if count > int64(maxRequests) {
            return *appctx.NewResponse().
                WithCode(fiber.StatusTooManyRequests).
                WithErrors("Rate limit exceeded")
        }
        
        return *appctx.NewResponse().WithCode(fiber.StatusOK)
    }
}
```

### 2. Token Validation

Use proper JWT library or validate against database:

```go
func DatabaseTokenAuth(db Database) Middleware {
    return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
        token := extractToken(ctx)
        
        // Validate in database
        valid, err := db.ValidateToken(ctx.Context(), token)
        if err != nil || !valid {
            return *appctx.NewResponse().
                WithCode(fiber.StatusUnauthorized).
                WithErrors("Invalid token")
        }
        
        return *appctx.NewResponse().WithCode(fiber.StatusOK)
    }
}
```

### 3. HMAC Time Window

Add timestamp validation to prevent replay attacks:

```go
func HMACAuthWithTimeWindow(secret string, windowSeconds int64) Middleware {
    return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
        timestamp := ctx.Get("X-Timestamp")
        reqTime, _ := strconv.ParseInt(timestamp, 10, 64)
        now := time.Now().Unix()
        
        // Check if request is within time window
        if abs(now - reqTime) > windowSeconds {
            return *appctx.NewResponse().
                WithCode(fiber.StatusUnauthorized).
                WithErrors("Request expired")
        }
        
        // ... rest of HMAC validation
    }
}
```

---

## Summary

Custom middleware system provides:
- ✅ **Clean Architecture** integration
- ✅ **Simple contract** (return 200 or error)
- ✅ **Chainable** middlewares
- ✅ **Built-in middlewares** (Auth, HMAC, Rate Limit, etc.)
- ✅ **Easy to extend** (create custom middleware)
- ✅ **Logging** integrated
- ✅ **Context support** (store data for handler)
- ✅ **Production ready** patterns

**Files Created:**
- `internal/middleware/middleware.go` - Interface contract
- `internal/middleware/auth.go` - Bearer & API Key auth
- `internal/middleware/hmac.go` - HMAC signature validation
- `internal/middleware/validators.go` - Rate limit, Content-Type, IP whitelist
- Updated `internal/router/router.go` - Middleware support

**No Fiber default middleware used - 100% custom implementation!** 🎉


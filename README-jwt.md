# JWT Package Documentation

## Overview

JWT (JSON Web Token) package menyediakan operasi lengkap untuk **generate**, **parse**, **validate**, dan **refresh** JWT tokens. Terintegrasi dengan **Clean Architecture** dan middleware system untuk authentication & authorization.

## Architecture

```
┌─────────────────────────────────────┐
│    Login Request                    │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Generate JWT Token                 │
│  (pkg/jwt)                          │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Return Token to Client             │
└─────────────────────────────────────┘


┌─────────────────────────────────────┐
│    Protected Request + Token        │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  JWTAuth Middleware                 │
│  Parse & Validate Token             │
└────────┬────────────────────────────┘
         │ (if valid)
         ▼
┌─────────────────────────────────────┐
│  Extract Claims to Context          │
│  (user_id, role, etc.)              │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Handler (UseCase)                  │
│  Access user data from context      │
└─────────────────────────────────────┘
```

## JWT Interface

```go
type JWT interface {
    // Generate generates a new JWT token with claims
    Generate(claims Claims) (string, error)

    // Parse parses and validates a JWT token
    Parse(tokenString string) (*Claims, error)

    // Refresh refreshes an existing token with new expiry
    Refresh(tokenString string) (string, error)

    // Validate validates a token without parsing claims
    Validate(tokenString string) error
}
```

## Claims Structure

```go
type Claims struct {
    UserID   int64             `json:"user_id"`
    Username string            `json:"username"`
    Email    string            `json:"email"`
    Role     string            `json:"role"`
    Extra    map[string]string `json:"extra,omitempty"`
    
    // Standard JWT claims
    Issuer    string
    IssuedAt  time.Time
    ExpiresAt time.Time
    NotBefore time.Time
}
```

## Configuration

### Environment Variables

Add to `.env`:

```bash
# JWT Configuration
JWT_SECRET_KEY=your-jwt-secret-key-here    # Generate: openssl rand -base64 32
JWT_ISSUER=hanif-skeleton                  # Token issuer name
JWT_EXPIRY=24h                             # Token expiry (24h, 1h, 30m, etc.)
```

### Generate Secret Key

```bash
# Generate secure 32-byte key
openssl rand -base64 32

# Example output:
# k8/JzQ7+FjKxN1mL9pW3vR5tY2nU6hG0iS4eA8bC7dE=
```

### Config Struct

File: `pkg/config/jwt.go`

```go
type JWT struct {
    SecretKey string        // Secret key for signing
    Issuer    string        // Token issuer
    Expiry    time.Duration // Token expiry duration
}
```

## Bootstrap Registry

File: `internal/bootstrap/jwt.go`

```go
// Initialize JWT
jwtInstance := bootstrap.RegistryJWT(cfg)
```

**Registry automatically:**
- ✅ Validates secret key (fatal if missing)
- ✅ Sets default issuer ("hanif-skeleton")
- ✅ Sets default expiry (24 hours)
- ✅ Logs initialization

---

## Usage Examples

### 1. Generate Token (Login)

```go
package usecase

import (
    "github.com/hanifkf12/hanif_skeleton/pkg/jwt"
)

func login(jwtInstance jwt.JWT, userID int64, username, email, role string) (string, error) {
    // Create claims
    claims := jwt.Claims{
        UserID:   userID,
        Username: username,
        Email:    email,
        Role:     role,
    }

    // Generate token
    token, err := jwtInstance.Generate(claims)
    if err != nil {
        return "", err
    }

    return token, nil
    // Returns: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImpvaG4i..."
}
```

### 2. Parse & Validate Token

```go
func validateToken(jwtInstance jwt.JWT, tokenString string) (*jwt.Claims, error) {
    // Parse token
    claims, err := jwtInstance.Parse(tokenString)
    if err != nil {
        if err == jwt.ErrTokenExpired {
            return nil, errors.New("token expired")
        }
        return nil, errors.New("invalid token")
    }

    // Access claims
    userID := claims.UserID
    username := claims.Username
    role := claims.Role

    return claims, nil
}
```

### 3. Refresh Token

```go
func refreshToken(jwtInstance jwt.JWT, oldToken string) (string, error) {
    // Refresh token (can refresh even if expired)
    newToken, err := jwtInstance.Refresh(oldToken)
    if err != nil {
        return "", err
    }

    return newToken, nil
}
```

### 4. Extract Claims Methods

```go
claims, _ := jwtInstance.Parse(token)

// Helper methods
userID := claims.GetUserID()       // int64
username := claims.GetUsername()   // string
email := claims.GetEmail()         // string
role := claims.GetRole()           // string

// Check role
isAdmin := claims.IsAdmin()        // true if role is "admin" or "superadmin"
hasRole := claims.HasRole("admin") // true if role matches
```

---

## HTTP Endpoints

### Login Endpoint

**File:** `internal/usecase/auth.go`

```go
func NewLogin(userRepo repository.UserRepository, hasher *crypto.BcryptHasher, jwtInstance jwt.JWT) contract.UseCase
```

**Request:**
```bash
curl -X POST http://localhost:9000/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "password123"
  }'
```

**Response:**
```json
{
  "code": 200,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user_id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "role": "user",
    "expires_in": "24h"
  }
}
```

### Refresh Token Endpoint

```bash
curl -X POST http://localhost:9000/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Response:**
```json
{
  "code": 200,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": "24h"
  }
}
```

---

## Middleware Integration

### JWTAuth Middleware

**File:** `internal/middleware/auth.go`

```go
middleware.JWTAuth(jwtInstance)
```

**Features:**
- ✅ Extracts token from `Authorization: Bearer <token>`
- ✅ Parses and validates JWT
- ✅ Checks expiry
- ✅ Stores claims in context
- ✅ Returns 401 if invalid/expired

**Usage in Router:**
```go
rtr.fiber.Get("/users", rtr.handleWithMiddleware(
    handler.HttpRequest,
    userUseCase,
    middleware.JWTAuth(jwtInstance),
))
```

**Test:**
```bash
# Get token first
TOKEN=$(curl -X POST http://localhost:9000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"pass123"}' | jq -r '.data.token')

# Use token
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:9000/users
```

### RequireRole Middleware

**Usage:**
```go
rtr.fiber.Post("/users", rtr.handleWithMiddleware(
    handler.HttpRequest,
    createUserUseCase,
    middleware.JWTAuth(jwtInstance),
    middleware.RequireRole([]string{"admin"}), // Only admin can create
))
```

**Flow:**
1. JWTAuth extracts & validates token
2. JWTAuth stores role in context
3. RequireRole checks if role matches
4. If not → 403 Forbidden
5. If yes → Continue to handler

---

## Access Claims in Handler

### From Context

```go
func (u *yourUseCase) Serve(data appctx.Data) appctx.Response {
    // Get user ID
    userID, ok := data.FiberCtx.Locals("user_id").(int64)
    if !ok {
        return *appctx.NewResponse().
            WithCode(fiber.StatusUnauthorized).
            WithErrors("User ID not found")
    }

    // Get username
    username := data.FiberCtx.Locals("username").(string)

    // Get role
    role := data.FiberCtx.Locals("role").(string)

    // Get full claims
    claims, ok := data.FiberCtx.Locals("claims").(*jwt.Claims)
    if ok {
        email := claims.Email
        // ... use claims
    }

    // Your business logic using user data
    // ...
}
```

---

## Token Lifecycle

### 1. Login Flow

```
User Login
    ↓
Validate Credentials
    ↓
Generate JWT Token
    ↓
Return Token to Client
    ↓
Client stores token (localStorage, cookie, etc.)
```

### 2. Protected Request Flow

```
Client sends request with token
    ↓
JWTAuth Middleware
    ↓
Parse & Validate Token
    ↓
Extract Claims → Context
    ↓
Handler processes request
    ↓
Return response
```

### 3. Token Expiry Flow

```
Token expires after 24h (configurable)
    ↓
Client sends expired token
    ↓
JWTAuth returns 401 "Token expired"
    ↓
Client calls /auth/refresh
    ↓
Get new token
    ↓
Continue using new token
```

---

## Error Handling

### JWT Errors

```go
var (
    ErrInvalidToken      = errors.New("invalid token")
    ErrTokenExpired      = errors.New("token expired")
    ErrInvalidSignMethod = errors.New("invalid signing method")
    ErrMissingClaims     = errors.New("missing claims")
)
```

### Handling in Middleware

```go
claims, err := jwtInstance.Parse(token)
if err != nil {
    if err == jwt.ErrTokenExpired {
        return "Token expired - please refresh"
    }
    return "Invalid token"
}
```

---

## Security Best Practices

### 1. Secret Key Management

```bash
# ❌ BAD - Weak key
JWT_SECRET_KEY=mysecret123

# ✅ GOOD - Strong random key
JWT_SECRET_KEY=k8/JzQ7+FjKxN1mL9pW3vR5tY2nU6hG0iS4eA8bC7dE=
```

### 2. Token Expiry

```bash
# Short-lived tokens for sensitive operations
JWT_EXPIRY=1h

# Longer for regular apps
JWT_EXPIRY=24h

# Use refresh tokens for extended sessions
```

### 3. HTTPS Only

```go
// Always use HTTPS in production
// Tokens in plain HTTP can be intercepted
```

### 4. Secure Storage (Client)

```javascript
// ✅ GOOD - HttpOnly cookie
document.cookie = "token=" + token + "; HttpOnly; Secure; SameSite=Strict";

// ⚠️ OK - localStorage (vulnerable to XSS)
localStorage.setItem('token', token);

// ❌ BAD - Plain cookie
document.cookie = "token=" + token;
```

### 5. Token Refresh Strategy

```go
// Option 1: Refresh before expiry (recommended)
if timeUntilExpiry < 5*time.Minute {
    newToken := refreshToken(oldToken)
}

// Option 2: Refresh on 401
if response.status == 401 && response.error == "Token expired" {
    newToken := refreshToken(oldToken)
    retryRequest(newToken)
}
```

### 6. Role-Based Access

```go
// Always check permissions
middleware.JWTAuth(jwtInstance),
middleware.RequireRole([]string{"admin"}),
```

### 7. Logout

```go
// Server-side: Blacklist token (use Redis)
func logout(token string) {
    redis.Set("blacklist:"+token, "1", 24*time.Hour)
}

// In middleware: Check blacklist
if redis.Exists("blacklist:"+token) {
    return "Token has been revoked"
}
```

---

## Production Considerations

### 1. Token Blacklist (Redis)

```go
import "github.com/go-redis/redis/v8"

func JWTAuthWithBlacklist(jwtInstance jwt.JWT, redis *redis.Client) Middleware {
    return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
        token := extractToken(ctx)
        
        // Check if token is blacklisted
        val, _ := redis.Get(ctx.Context(), "blacklist:"+token).Result()
        if val != "" {
            return *appctx.NewResponse().
                WithCode(fiber.StatusUnauthorized).
                WithErrors("Token has been revoked")
        }
        
        // Normal JWT validation
        claims, err := jwtInstance.Parse(token)
        // ...
    }
}
```

### 2. Refresh Token Strategy

```go
// Use separate refresh token (longer expiry)
type RefreshToken struct {
    Token     string
    ExpiresAt time.Time
    UserID    int64
}

// Store in database with access token
// Refresh token: 30 days
// Access token: 1 hour
```

### 3. Multiple Device Support

```go
// Store device info in claims
claims := jwt.Claims{
    UserID: userID,
    Extra: map[string]string{
        "device_id": deviceID,
        "ip": clientIP,
    },
}
```

### 4. Token Versioning

```go
// Add version to claims
claims := jwt.Claims{
    Extra: map[string]string{
        "version": "v1",
    },
}

// Invalidate old versions
if claims.Extra["version"] != "v1" {
    return errors.New("token version mismatch")
}
```

---

## Testing

### Unit Test JWT

```go
func TestJWT_Generate(t *testing.T) {
    jwtInstance, _ := jwt.NewJWT(jwt.Config{
        SecretKey: "test-secret",
        Issuer:    "test",
        Expiry:    1 * time.Hour,
    })

    claims := jwt.Claims{
        UserID:   1,
        Username: "test",
        Role:     "user",
    }

    token, err := jwtInstance.Generate(claims)
    assert.NoError(t, err)
    assert.NotEmpty(t, token)

    // Parse back
    parsed, err := jwtInstance.Parse(token)
    assert.NoError(t, err)
    assert.Equal(t, int64(1), parsed.UserID)
    assert.Equal(t, "test", parsed.Username)
}
```

### Integration Test

```bash
# Test login
TOKEN=$(curl -X POST http://localhost:9000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"pass"}' | jq -r '.data.token')

# Test protected endpoint
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:9000/users

# Test invalid token
curl -H "Authorization: Bearer invalid-token" \
  http://localhost:9000/users
# Expected: {"code": 401, "errors": "Invalid token"}
```

---

## Troubleshooting

### Issue: "Token expired" immediately

**Solution:** Check server time sync

```bash
# Check server time
date

# Sync with NTP
sudo ntpdate -s time.nist.gov
```

### Issue: "Invalid signature"

**Solution:** Secret key mismatch

```bash
# Verify JWT_SECRET_KEY is same across:
# - Environment variables
# - Config file
# - All servers (if load balanced)
```

### Issue: Claims not in context

**Solution:** Ensure JWTAuth middleware runs first

```go
// ✅ GOOD
middleware.JWTAuth(jwtInstance),
middleware.RequireRole(...),

// ❌ BAD
middleware.RequireRole(...),
middleware.JWTAuth(jwtInstance),
```

---

## Summary

JWT package provides:
- ✅ **Complete JWT operations** (generate, parse, refresh, validate)
- ✅ **HS256 signing** (HMAC with SHA-256)
- ✅ **Standard claims** + custom fields
- ✅ **Bootstrap integration** for easy setup
- ✅ **Middleware integration** (JWTAuth, RequireRole)
- ✅ **Login/Refresh endpoints** included
- ✅ **Context support** (store claims for handlers)
- ✅ **Error handling** (expired, invalid, etc.)
- ✅ **Production ready** patterns

**Files Created:**
- `pkg/jwt/jwt.go` - JWT interface & implementation
- `pkg/config/jwt.go` - JWT config struct
- `internal/bootstrap/jwt.go` - Registry initialization
- `internal/middleware/auth.go` - Updated with JWTAuth & RequireRole
- `internal/usecase/auth.go` - Login & RefreshToken usecases
- Updated `.env` - JWT configuration

**Generate secret key:**
```bash
openssl rand -base64 32
```

**Usage:**
```go
// Initialize
jwtInstance := bootstrap.RegistryJWT(cfg)

// Generate token
token, _ := jwtInstance.Generate(claims)

// Protect route
rtr.fiber.Get("/protected", rtr.handleWithMiddleware(
    handler.HttpRequest,
    useCase,
    middleware.JWTAuth(jwtInstance),
))
```

**Your API now has production-ready JWT authentication!** 🔐🎯


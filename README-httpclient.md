## HTTP Client Package Documentation

## Overview

HTTP Client package menyediakan abstraksi unified untuk melakukan HTTP requests ke 3rd party services/APIs. Dilengkapi dengan retry mechanism, timeout handling, dan logging terintegrasi, mengikuti **Clean Architecture** pattern.

## Architecture

```
┌─────────────────────────────────────┐
│         UseCase Layer               │
│    (Call 3rd Party APIs)            │
├─────────────────────────────────────┤
│   HTTPClient Interface (Contract)   │  ← Abstraction
├─────────────────────────────────────┤
│    Implementation Layer             │
│    ├─ Standard HTTP Client         │
│    └─ Mock Client (Testing)        │
└─────────────────────────────────────┘
```

## HTTPClient Interface

```go
type HTTPClient interface {
    Get(ctx context.Context, url string, headers map[string]string) (*Response, error)
    Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error)
    Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error)
    Patch(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error)
    Delete(ctx context.Context, url string, headers map[string]string) (*Response, error)
    Do(ctx context.Context, req *Request) (*Response, error)
}
```

## Features

### ✅ HTTP Methods Support
- GET, POST, PUT, PATCH, DELETE
- Custom method via `Do()`

### ✅ Automatic Retry
- Configurable retry attempts
- Exponential backoff
- Skip retry on 4xx errors

### ✅ Timeout Handling
- Per-request timeout
- Context cancellation support

### ✅ Logging Integration
- Request/response logging
- Duration tracking
- Error logging

### ✅ Response Helpers
- JSON unmarshal
- Success/Error checking
- String conversion

### ✅ Testing Support
- Mock client included
- Easy to test

## Configuration

### Environment Variables

Add to `.env`:

```bash
# HTTP Client Configuration
HTTP_CLIENT_TIMEOUT=30s                    # Request timeout
HTTP_CLIENT_MAX_RETRIES=3                  # Max retry attempts (0 = no retry)
HTTP_CLIENT_RETRY_WAIT_TIME=1s            # Wait between retries
HTTP_CLIENT_FOLLOW_REDIRECT=true          # Follow redirects
```

### Config Struct

File: `pkg/config/httpclient.go`

```go
type HTTPClient struct {
    Timeout        time.Duration
    MaxRetries     int
    RetryWaitTime  time.Duration
    FollowRedirect bool
}
```

### Client Config

```go
type Config struct {
    Timeout         time.Duration
    MaxRetries      int
    RetryWaitTime   time.Duration
    DefaultHeaders  map[string]string
    FollowRedirects bool
    BaseURL         string  // Base URL for relative paths
}
```

## Bootstrap Registry

File: `internal/bootstrap/httpclient.go`

```go
// Initialize HTTP client
httpClient := bootstrap.RegistryHTTPClient(cfg)
```

**Registry automatically:**
- ✅ Sets default timeout (30s)
- ✅ Sets default retry (3 attempts)
- ✅ Adds User-Agent header
- ✅ Logs initialization

---

## Usage Examples

### 1. Basic GET Request

```go
package usecase

import (
    "context"
    "github.com/hanifkf12/hanif_skeleton/pkg/httpclient"
)

func getUsers(ctx context.Context, client httpclient.HTTPClient) error {
    // Simple GET request
    resp, err := client.Get(ctx, "https://api.example.com/users", nil)
    if err != nil {
        return err
    }

    // Check success
    if !resp.IsSuccess() {
        return fmt.Errorf("API returned status %d", resp.StatusCode)
    }

    // Parse JSON response
    var users []User
    if err := resp.JSON(&users); err != nil {
        return err
    }

    return nil
}
```

### 2. POST Request with Body

```go
func createUser(ctx context.Context, client httpclient.HTTPClient) error {
    // Request body
    payload := map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   30,
    }

    // Headers
    headers := map[string]string{
        "Authorization": "Bearer token-abc-123",
        "Content-Type":  "application/json",
    }

    // POST request
    resp, err := client.Post(ctx, "https://api.example.com/users", payload, headers)
    if err != nil {
        return err
    }

    if !resp.IsSuccess() {
        return fmt.Errorf("failed to create user: %s", resp.String())
    }

    return nil
}
```

### 3. With Custom Headers

```go
func apiWithAuth(ctx context.Context, client httpclient.HTTPClient) error {
    headers := map[string]string{
        "Authorization": "Bearer " + apiToken,
        "X-API-Key":     apiKey,
        "Accept":        "application/json",
    }

    resp, err := client.Get(ctx, "https://api.example.com/data", headers)
    if err != nil {
        return err
    }

    return nil
}
```

### 4. With Timeout

```go
func apiWithTimeout(client httpclient.HTTPClient) error {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := client.Get(ctx, "https://slow-api.example.com/data", nil)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return errors.New("request timeout")
        }
        return err
    }

    return nil
}
```

### 5. Custom Request with Do()

```go
func customRequest(ctx context.Context, client httpclient.HTTPClient) error {
    req := &httpclient.Request{
        Method: "POST",
        URL:    "https://api.example.com/webhook",
        Headers: map[string]string{
            "X-Webhook-Signature": signature,
        },
        Body: webhookData,
    }

    resp, err := client.Do(ctx, req)
    if err != nil {
        return err
    }

    return nil
}
```

---

## Integration with UseCase

### Example: Weather Service

```go
package usecase

type weatherService struct {
    httpClient httpclient.HTTPClient
}

func NewWeatherService(httpClient httpclient.HTTPClient) contract.UseCase {
    return &weatherService{httpClient: httpClient}
}

func (u *weatherService) Serve(data appctx.Data) appctx.Response {
    ctx := data.FiberCtx.UserContext()

    // Parse request
    var req WeatherRequest
    data.FiberCtx.BodyParser(&req)

    // Call weather API
    url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", 
        apiKey, req.City)

    resp, err := u.httpClient.Get(ctx, url, map[string]string{
        "Accept": "application/json",
    })

    if err != nil {
        return *appctx.NewResponse().
            WithCode(fiber.StatusServiceUnavailable).
            WithErrors("Weather service unavailable")
    }

    if !resp.IsSuccess() {
        return *appctx.NewResponse().
            WithCode(fiber.StatusServiceUnavailable).
            WithErrors("Failed to fetch weather data")
    }

    // Parse response
    var weather WeatherData
    resp.JSON(&weather)

    return *appctx.NewResponse().WithData(weather)
}
```

### Example: Payment Gateway

```go
type paymentGateway struct {
    httpClient httpclient.HTTPClient
    apiKey     string
    baseURL    string
}

func (u *paymentGateway) Serve(data appctx.Data) appctx.Response {
    ctx := data.FiberCtx.UserContext()

    var req PaymentRequest
    data.FiberCtx.BodyParser(&req)

    // Prepare payment request
    url := u.baseURL + "/transactions"
    headers := map[string]string{
        "Authorization": "Bearer " + u.apiKey,
    }

    // Call payment gateway
    resp, err := u.httpClient.Post(ctx, url, req, headers)
    if err != nil {
        return *appctx.NewResponse().
            WithCode(fiber.StatusServiceUnavailable).
            WithErrors("Payment gateway unavailable")
    }

    if !resp.IsSuccess() {
        return *appctx.NewResponse().
            WithCode(fiber.StatusPaymentRequired).
            WithErrors("Payment failed")
    }

    var payment PaymentResponse
    resp.JSON(&payment)

    return *appctx.NewResponse().WithData(payment)
}
```

---

## Response Helpers

### Check Response Status

```go
resp, err := client.Get(ctx, url, nil)

// Check if success (2xx)
if resp.IsSuccess() {
    // Handle success
}

// Check if error (4xx or 5xx)
if resp.IsError() {
    // Handle error
}

// Check specific status code
if resp.StatusCode == 404 {
    // Handle not found
}
```

### Parse JSON Response

```go
// Method 1: Using JSON helper
var data MyStruct
err := resp.JSON(&data)

// Method 2: Manual unmarshal
var data MyStruct
json.Unmarshal(resp.Body, &data)
```

### Get Response as String

```go
resp, _ := client.Get(ctx, url, nil)

// Get body as string
body := resp.String()
fmt.Println(body)
```

### Access Headers

```go
resp, _ := client.Get(ctx, url, nil)

// Get specific header
contentType := resp.Headers.Get("Content-Type")
rateLimit := resp.Headers.Get("X-RateLimit-Remaining")
```

---

## Retry Mechanism

### How It Works

```
Request → Try #1 → Failed → Wait 1s → Try #2 → Failed → Wait 1s → Try #3 → Success
```

**Retry Logic:**
- ✅ Retries on network errors
- ✅ Retries on 5xx errors
- ❌ No retry on 4xx errors (client errors)
- ✅ Configurable max attempts
- ✅ Configurable wait time

### Configure Retry

```bash
# .env
HTTP_CLIENT_MAX_RETRIES=3         # 3 retry attempts
HTTP_CLIENT_RETRY_WAIT_TIME=1s    # Wait 1 second between retries
```

### Disable Retry

```bash
HTTP_CLIENT_MAX_RETRIES=0  # No retry
```

---

## Testing with Mock Client

### Create Mock Client

```go
package usecase_test

import (
    "testing"
    "github.com/hanifkf12/hanif_skeleton/pkg/httpclient"
)

func TestWeatherService(t *testing.T) {
    // Create mock client
    mockClient := httpclient.NewMockClient()

    // Setup mock response
    mockResponse := &httpclient.Response{
        StatusCode: 200,
        Body: []byte(`{
            "city": "Jakarta",
            "temperature": 28.5,
            "description": "Sunny"
        }`),
    }
    mockClient.OnGet("https://api.weatherapi.com/v1/current.json", mockResponse, nil)

    // Use in test
    service := NewWeatherService(mockClient)
    result := service.Serve(data)

    // Assertions
    assert.Equal(t, 200, result.Code)
}
```

### Mock Error Response

```go
func TestAPIError(t *testing.T) {
    mockClient := httpclient.NewMockClient()

    // Mock error
    mockClient.OnGet(
        "https://api.example.com/users",
        nil,
        errors.New("connection timeout"),
    )

    // Test should handle error
    resp, err := mockClient.Get(ctx, "https://api.example.com/users", nil)
    assert.Error(t, err)
}
```

### Mock Different Status Codes

```go
func TestDifferentStatusCodes(t *testing.T) {
    mockClient := httpclient.NewMockClient()

    // Mock 404
    mockClient.OnGet("https://api.example.com/not-found", &httpclient.Response{
        StatusCode: 404,
        Body: []byte(`{"error": "not found"}`),
    }, nil)

    // Mock 500
    mockClient.OnPost("https://api.example.com/error", &httpclient.Response{
        StatusCode: 500,
        Body: []byte(`{"error": "internal server error"}`),
    }, nil)
}
```

---

## Advanced Usage

### 1. With Base URL

```go
config := httpclient.Config{
    BaseURL: "https://api.example.com",
    Timeout: 30 * time.Second,
}

client := httpclient.NewHTTPClient(config)

// Use relative paths
resp, _ := client.Get(ctx, "/users", nil)        // https://api.example.com/users
resp, _ := client.Get(ctx, "/products", nil)      // https://api.example.com/products
```

### 2. With Default Headers

```go
config := httpclient.Config{
    DefaultHeaders: map[string]string{
        "Authorization": "Bearer " + token,
        "X-API-Key":     apiKey,
        "User-Agent":    "MyApp/1.0",
    },
}

client := httpclient.NewHTTPClient(config)

// All requests will have these headers
resp, _ := client.Get(ctx, url, nil)
```

### 3. Multiple Clients

```go
// Client for service A
configA := httpclient.Config{
    BaseURL: "https://service-a.com",
    Timeout: 10 * time.Second,
}
clientA := httpclient.NewHTTPClient(configA)

// Client for service B
configB := httpclient.Config{
    BaseURL: "https://service-b.com",
    Timeout: 30 * time.Second,
}
clientB := httpclient.NewHTTPClient(configB)
```

### 4. Parallel Requests

```go
func fetchMultiple(ctx context.Context, client httpclient.HTTPClient) error {
    var wg sync.WaitGroup
    errChan := make(chan error, 2)

    // Request 1
    wg.Add(1)
    go func() {
        defer wg.Done()
        resp, err := client.Get(ctx, "https://api1.example.com/data", nil)
        if err != nil {
            errChan <- err
        }
    }()

    // Request 2
    wg.Add(1)
    go func() {
        defer wg.Done()
        resp, err := client.Get(ctx, "https://api2.example.com/data", nil)
        if err != nil {
            errChan <- err
        }
    }()

    wg.Wait()
    close(errChan)

    // Check errors
    for err := range errChan {
        return err
    }

    return nil
}
```

---

## Common Use Cases

### 1. REST API Integration

```go
// GET list
resp, _ := client.Get(ctx, baseURL+"/users", headers)

// GET single
resp, _ := client.Get(ctx, baseURL+"/users/123", headers)

// CREATE
resp, _ := client.Post(ctx, baseURL+"/users", userData, headers)

// UPDATE
resp, _ := client.Put(ctx, baseURL+"/users/123", userData, headers)

// DELETE
resp, _ := client.Delete(ctx, baseURL+"/users/123", headers)
```

### 2. Webhook Notification

```go
func sendWebhook(ctx context.Context, client httpclient.HTTPClient, event Event) error {
    payload := map[string]interface{}{
        "event": event.Type,
        "data":  event.Data,
        "timestamp": time.Now().Unix(),
    }

    // Calculate signature
    signature := calculateHMAC(payload, secret)

    headers := map[string]string{
        "X-Webhook-Signature": signature,
    }

    resp, err := client.Post(ctx, webhookURL, payload, headers)
    return err
}
```

### 3. OAuth Token Refresh

```go
func refreshAccessToken(ctx context.Context, client httpclient.HTTPClient, refreshToken string) (string, error) {
    payload := map[string]interface{}{
        "grant_type":    "refresh_token",
        "refresh_token": refreshToken,
        "client_id":     clientID,
        "client_secret": clientSecret,
    }

    resp, err := client.Post(ctx, "https://oauth.example.com/token", payload, nil)
    if err != nil {
        return "", err
    }

    var tokenResp struct {
        AccessToken string `json:"access_token"`
    }
    resp.JSON(&tokenResp)

    return tokenResp.AccessToken, nil
}
```

### 4. File Download

```go
func downloadFile(ctx context.Context, client httpclient.HTTPClient, fileURL string) ([]byte, error) {
    resp, err := client.Get(ctx, fileURL, nil)
    if err != nil {
        return nil, err
    }

    if !resp.IsSuccess() {
        return nil, fmt.Errorf("download failed: %d", resp.StatusCode)
    }

    return resp.Body, nil
}
```

---

## Error Handling

### Handle Different Error Types

```go
resp, err := client.Get(ctx, url, nil)
if err != nil {
    // Network error, timeout, etc.
    if ctx.Err() == context.DeadlineExceeded {
        return errors.New("request timeout")
    }
    return fmt.Errorf("request failed: %w", err)
}

// HTTP error (4xx, 5xx)
if !resp.IsSuccess() {
    switch resp.StatusCode {
    case 400:
        return errors.New("bad request")
    case 401:
        return errors.New("unauthorized")
    case 404:
        return errors.New("not found")
    case 500:
        return errors.New("server error")
    default:
        return fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
}
```

---

## Best Practices

### ✅ DO:
- Use context with timeout
- Check response status codes
- Handle both network errors and HTTP errors
- Log requests for debugging
- Use retry for transient failures
- Mock HTTP client in tests

### ❌ DON'T:
- Ignore errors
- Use infinite timeout
- Retry on 4xx errors
- Log sensitive data (tokens, passwords)
- Use HTTP client without timeout

---

## Troubleshooting

### Issue: Request timeout

**Solution:** Increase timeout or optimize API

```bash
HTTP_CLIENT_TIMEOUT=60s  # Increase to 60 seconds
```

### Issue: Too many retries

**Solution:** Reduce retry attempts

```bash
HTTP_CLIENT_MAX_RETRIES=1  # Retry only once
```

### Issue: Connection refused

**Solution:** Check API endpoint and network

```go
resp, err := client.Get(ctx, url, nil)
if err != nil {
    log.Printf("Connection error: %v", err)
    // Check if API is up, firewall, DNS, etc.
}
```

---

## Summary

HTTP Client package provides:
- ✅ **Clean interface** for HTTP requests
- ✅ **Automatic retry** with backoff
- ✅ **Timeout handling** with context
- ✅ **Logging integration** (request/response)
- ✅ **Response helpers** (JSON, String, Status checks)
- ✅ **Mock client** for testing
- ✅ **Bootstrap integration** for easy setup
- ✅ **Production ready** error handling

**Choose for:**
- Calling 3rd party APIs
- REST API integration
- Webhook notifications
- OAuth flows
- File downloads
- Any external HTTP service

---

**Files:**
- Interface: `pkg/httpclient/httpclient.go`
- Implementation: `pkg/httpclient/client.go`
- Mock: `pkg/httpclient/mock.go`
- Config: `pkg/config/httpclient.go`
- Bootstrap: `internal/bootstrap/httpclient.go`
- Examples: `internal/usecase/httpclient_example.go`

**Usage:**
```go
// Initialize
httpClient := bootstrap.RegistryHTTPClient(cfg)

// GET request
resp, err := httpClient.Get(ctx, "https://api.example.com/users", nil)

// POST request
resp, err := httpClient.Post(ctx, url, body, headers)

// Check response
if resp.IsSuccess() {
    var data MyStruct
    resp.JSON(&data)
}
```

**Your app is ready to call any 3rd party API!** 🚀🌐


package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// HTTPClient is the interface for HTTP client operations
type HTTPClient interface {
	// Get makes a GET request
	Get(ctx context.Context, url string, headers map[string]string) (*Response, error)

	// Post makes a POST request
	Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error)

	// Put makes a PUT request
	Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error)

	// Patch makes a PATCH request
	Patch(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error)

	// Delete makes a DELETE request
	Delete(ctx context.Context, url string, headers map[string]string) (*Response, error)

	// Do executes a custom HTTP request
	Do(ctx context.Context, req *Request) (*Response, error)
}

// Request represents an HTTP request
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    interface{}
	Timeout time.Duration
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
	RawRequest *http.Request
}

// JSON unmarshals response body to a struct
func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// String returns response body as string
func (r *Response) String() string {
	return string(r.Body)
}

// IsSuccess checks if status code is 2xx
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsError checks if status code is 4xx or 5xx
func (r *Response) IsError() bool {
	return r.StatusCode >= 400
}

// Config holds HTTP client configuration
type Config struct {
	Timeout         time.Duration     // Request timeout
	MaxRetries      int               // Max retry attempts
	RetryWaitTime   time.Duration     // Wait time between retries
	DefaultHeaders  map[string]string // Default headers for all requests
	FollowRedirects bool              // Follow redirects
	BaseURL         string            // Base URL for relative paths
}

// DefaultConfig returns default HTTP client configuration
func DefaultConfig() Config {
	return Config{
		Timeout:         30 * time.Second,
		MaxRetries:      3,
		RetryWaitTime:   1 * time.Second,
		DefaultHeaders:  make(map[string]string),
		FollowRedirects: true,
	}
}

package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

// standardClient implements HTTPClient interface using standard net/http
type standardClient struct {
	client *http.Client
	config Config
}

// NewHTTPClient creates a new HTTP client instance
func NewHTTPClient(config Config) HTTPClient {
	// Set defaults if not provided
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryWaitTime == 0 {
		config.RetryWaitTime = 1 * time.Second
	}
	if config.DefaultHeaders == nil {
		config.DefaultHeaders = make(map[string]string)
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Disable redirects if configured
	if !config.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return &standardClient{
		client: client,
		config: config,
	}
}

// Get makes a GET request
func (c *standardClient) Get(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  http.MethodGet,
		URL:     c.buildURL(url),
		Headers: headers,
	})
}

// Post makes a POST request
func (c *standardClient) Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  http.MethodPost,
		URL:     c.buildURL(url),
		Body:    body,
		Headers: headers,
	})
}

// Put makes a PUT request
func (c *standardClient) Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  http.MethodPut,
		URL:     c.buildURL(url),
		Body:    body,
		Headers: headers,
	})
}

// Patch makes a PATCH request
func (c *standardClient) Patch(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  http.MethodPatch,
		URL:     c.buildURL(url),
		Body:    body,
		Headers: headers,
	})
}

// Delete makes a DELETE request
func (c *standardClient) Delete(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  http.MethodDelete,
		URL:     c.buildURL(url),
		Headers: headers,
	})
}

// Do executes an HTTP request with retry logic
func (c *standardClient) Do(ctx context.Context, req *Request) (*Response, error) {
	lf := logger.NewFields("HTTPClient.Do")
	lf.Append(logger.Any("method", req.Method))
	lf.Append(logger.Any("url", req.URL))

	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			lf.Append(logger.Any("retry_attempt", attempt))
			logger.Info("Retrying HTTP request", lf)
			time.Sleep(c.config.RetryWaitTime)
		}

		resp, err := c.doRequest(ctx, req)
		if err == nil && resp.IsSuccess() {
			return resp, nil
		}

		lastErr = err
		if err != nil {
			lf.Append(logger.Any("error", err.Error()))
			logger.Error("HTTP request failed", lf)
		} else {
			lf.Append(logger.Any("status_code", resp.StatusCode))
			logger.Error("HTTP request returned error status", lf)
		}

		// Don't retry on client errors (4xx)
		if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return resp, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
		}
	}

	return nil, fmt.Errorf("HTTP request failed after %d attempts: %w", c.config.MaxRetries+1, lastErr)
}

// doRequest executes a single HTTP request
func (c *standardClient) doRequest(ctx context.Context, req *Request) (*Response, error) {
	// Prepare request body
	var bodyReader io.Reader
	if req.Body != nil {
		jsonData, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	for key, value := range c.config.DefaultHeaders {
		httpReq.Header.Set(key, value)
	}

	// Set request headers (override defaults)
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set Content-Type if body is present and not set
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	startTime := time.Now()
	httpResp, err := c.client.Do(httpReq)
	duration := time.Since(startTime)

	lf := logger.NewFields("HTTPClient.doRequest")
	lf.Append(logger.Any("method", req.Method))
	lf.Append(logger.Any("url", req.URL))
	lf.Append(logger.Any("duration_ms", duration.Milliseconds()))

	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("HTTP request execution failed", lf)
		return nil, fmt.Errorf("request execution failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to read response body", lf)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	lf.Append(logger.Any("status_code", httpResp.StatusCode))
	lf.Append(logger.Any("response_size", len(body)))

	response := &Response{
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Headers:    httpResp.Header,
		Body:       body,
		RawRequest: httpReq,
	}

	if response.IsSuccess() {
		logger.Info("HTTP request successful", lf)
	} else {
		logger.Error("HTTP request failed", lf)
	}

	return response, nil
}

// buildURL combines base URL with path if base URL is set
func (c *standardClient) buildURL(path string) string {
	if c.config.BaseURL != "" {
		return c.config.BaseURL + path
	}
	return path
}

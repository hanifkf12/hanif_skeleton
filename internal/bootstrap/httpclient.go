package bootstrap

import (
	"time"

	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/httpclient"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

// RegistryHTTPClient creates and returns an HTTP client instance based on configuration
func RegistryHTTPClient(cfg *config.Config) httpclient.HTTPClient {
	lf := logger.NewFields("RegistryHTTPClient")

	// Set defaults
	timeout := cfg.HTTPClient.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	maxRetries := cfg.HTTPClient.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	retryWaitTime := cfg.HTTPClient.RetryWaitTime
	if retryWaitTime == 0 {
		retryWaitTime = 1 * time.Second
	}

	followRedirect := cfg.HTTPClient.FollowRedirect
	// Default to true if not set
	if timeout == 30*time.Second && maxRetries == 3 {
		followRedirect = true
	}

	lf.Append(logger.Any("timeout", timeout.String()))
	lf.Append(logger.Any("max_retries", maxRetries))
	lf.Append(logger.Any("retry_wait_time", retryWaitTime.String()))
	lf.Append(logger.Any("follow_redirect", followRedirect))

	clientConfig := httpclient.Config{
		Timeout:         timeout,
		MaxRetries:      maxRetries,
		RetryWaitTime:   retryWaitTime,
		FollowRedirects: followRedirect,
		DefaultHeaders: map[string]string{
			"User-Agent": "hanif-skeleton-http-client/1.0",
		},
	}

	client := httpclient.NewHTTPClient(clientConfig)

	logger.Info("HTTP client initialized successfully", lf)
	return client
}

package config

import "time"

// HTTPClient holds HTTP client configuration
type HTTPClient struct {
	Timeout        time.Duration `mapstructure:"HTTP_CLIENT_TIMEOUT"`         // Request timeout
	MaxRetries     int           `mapstructure:"HTTP_CLIENT_MAX_RETRIES"`     // Max retry attempts
	RetryWaitTime  time.Duration `mapstructure:"HTTP_CLIENT_RETRY_WAIT_TIME"` // Wait time between retries
	FollowRedirect bool          `mapstructure:"HTTP_CLIENT_FOLLOW_REDIRECT"` // Follow redirects
}

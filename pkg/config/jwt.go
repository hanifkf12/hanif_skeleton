package config

import "time"

// JWT holds JWT configuration
type JWT struct {
	SecretKey string        `mapstructure:"JWT_SECRET_KEY"` // Secret key for signing JWT
	Issuer    string        `mapstructure:"JWT_ISSUER"`     // Token issuer
	Expiry    time.Duration `mapstructure:"JWT_EXPIRY"`     // Token expiry in seconds (will be converted to duration)
}

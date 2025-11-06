package config

// Cache holds cache configuration
type Cache struct {
	Driver   string `mapstructure:"CACHE_DRIVER"`   // redis, memory
	Host     string `mapstructure:"CACHE_HOST"`     // Redis host
	Port     int    `mapstructure:"CACHE_PORT"`     // Redis port
	Password string `mapstructure:"CACHE_PASSWORD"` // Redis password
	DB       int    `mapstructure:"CACHE_DB"`       // Redis database number
}

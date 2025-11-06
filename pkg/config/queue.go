package config

// Queue holds queue configuration
type Queue struct {
	Driver   string `mapstructure:"QUEUE_DRIVER"`   // asynq, memory
	Host     string `mapstructure:"QUEUE_HOST"`     // Redis host (for asynq)
	Port     int    `mapstructure:"QUEUE_PORT"`     // Redis port
	Password string `mapstructure:"QUEUE_PASSWORD"` // Redis password
	DB       int    `mapstructure:"QUEUE_DB"`       // Redis database
}

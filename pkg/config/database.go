package config

type Database struct {
	Host     string `mapstructure:"db_host"`
	Port     string `mapstructure:"db_port"`
	Username string `mapstructure:"db_username"`
	Password string `mapstructure:"db_password"`
	Name     string `mapstructure:"db_name"`
}

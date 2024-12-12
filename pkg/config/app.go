package config

type App struct {
	Name string `mapstructure:"name"`
	Port string `mapstructure:"port"`
}

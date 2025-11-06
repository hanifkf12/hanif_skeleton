package config

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	App        `mapstructure:",squash"`
	Database   `mapstructure:",squash"`
	Storage    `mapstructure:",squash"`
	Crypto     `mapstructure:",squash"`
	JWT        `mapstructure:",squash"`
	Cache      `mapstructure:",squash"`
	HTTPClient `mapstructure:",squash"`
}

func LoadAllConfigs() (*Config, error) {
	var cfg Config

	viper.AutomaticEnv()

	viper.WatchConfig()

	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	_ = viper.MergeInConfig()

	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)
	})

	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

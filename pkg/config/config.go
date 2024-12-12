package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	App      `mapstructure:",squash"`
	Database `mapstructure:",squash"`
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

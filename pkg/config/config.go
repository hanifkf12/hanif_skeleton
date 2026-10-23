package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	App        `mapstructure:",squash"`
	Database   `mapstructure:",squash"`
	Crypto     `mapstructure:",squash"`
	JWT        `mapstructure:",squash"`
	Cache      `mapstructure:",squash"`
	HTTPClient `mapstructure:",squash"`
	Queue      `mapstructure:",squash"`
}

func LoadAllConfigs() (*Config, error) {
	var cfg Config

	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	v.AutomaticEnv()

	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.SetConfigName(".env")
	_ = v.MergeInConfig()

	err := v.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

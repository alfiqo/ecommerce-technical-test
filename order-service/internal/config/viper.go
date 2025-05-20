package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// AppConfig wraps Viper to provide application configuration
type AppConfig struct {
	Viper *viper.Viper
}

// NewAppConfig creates a new AppConfig instance
func NewAppConfig(v *viper.Viper) *AppConfig {
	return &AppConfig{
		Viper: v,
	}
}

// NewViper is a function to load config from config.json
// You can change the implementation, for example load from env file, consul, etcd, etc
func NewViper() *viper.Viper {
	config := viper.New()

	config.SetConfigName("config")
	config.SetConfigType("json")
	config.AddConfigPath("./../")
	config.AddConfigPath("./")
	err := config.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	return config
}
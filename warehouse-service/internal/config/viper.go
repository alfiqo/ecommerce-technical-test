package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// NewViper is a function to load config from config.json and environment variables
// Environment variables take precedence over config file values
func NewViper() *viper.Viper {
	config := viper.New()

	// Load from config file
	config.SetConfigName("config")
	config.SetConfigType("json")
	config.AddConfigPath("./../")
	config.AddConfigPath("./")
	err := config.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}
	
	// Also load from environment variables
	config.AutomaticEnv()
	
	// Map specific environment variables to config keys
	if port := config.GetString("PORT"); port != "" {
		config.Set("web.port", port)
	}
	
	return config
}

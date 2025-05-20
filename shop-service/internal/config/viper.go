package config

import (
	"fmt"
	"os"
	"github.com/spf13/viper"
)

// NewViper is a function to load config from config.json
// You can change the implementation, for example load from env file, consul, etcd, etc
func NewViper() *viper.Viper {
	config := viper.New()

	// Use custom config file if specified
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config"
	} else {
		// If a full filename is provided, extract the name without extension
		ext := ".json" // Assume JSON for now
		if len(configFile) > len(ext) && configFile[len(configFile)-len(ext):] == ext {
			configFile = configFile[:len(configFile)-len(ext)]
		}
	}

	config.SetConfigName(configFile)
	config.SetConfigType("json")
	config.AddConfigPath("./../")
	config.AddConfigPath("./")
	err := config.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	return config
}

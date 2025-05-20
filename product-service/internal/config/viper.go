package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewViper() *viper.Viper {
	config := viper.New()
	
	// Load OS environment variables
	config.AutomaticEnv()
	
	// Check for environment-specific config
	env := config.GetString("ENV")
	configFile := "config.json"
	
	if env == "e2e" {
		configFile = "config.e2e.json"
		logrus.Infof("Using e2e configuration: %s", configFile)
	}
	
	config.SetConfigFile(configFile)
	err := config.ReadInConfig()
	if err != nil {
		logrus.Fatalf("Failed to read config file: %v", err)
	}
	return config
}
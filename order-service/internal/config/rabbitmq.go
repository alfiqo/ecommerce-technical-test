package config

import (
	"fmt"
	"order-service/internal/messaging"
)

// RabbitMQConfig holds configuration for RabbitMQ
type RabbitMQConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Exchange string `mapstructure:"exchange"`
	Queue    string `mapstructure:"queue"`
}

// GetRabbitMQConfig returns the RabbitMQ configuration
func (c *AppConfig) GetRabbitMQConfig() *RabbitMQConfig {
	return &RabbitMQConfig{
		Host:     c.Viper.GetString("rabbitmq.host"),
		Port:     c.Viper.GetInt("rabbitmq.port"),
		Username: c.Viper.GetString("rabbitmq.username"),
		Password: c.Viper.GetString("rabbitmq.password"),
		Exchange: c.Viper.GetString("rabbitmq.exchange"),
		Queue:    c.Viper.GetString("rabbitmq.queue"),
	}
}

// GetRabbitMQMessagingConfig converts the RabbitMQ config to messaging.RabbitMQConfig
func (c *RabbitMQConfig) GetRabbitMQMessagingConfig() messaging.RabbitMQConfig {
	return messaging.RabbitMQConfig{
		URL:       fmt.Sprintf("%s:%d", c.Host, c.Port),
		Exchange:  c.Exchange,
		QueueName: c.Queue,
		Username:  c.Username,
		Password:  c.Password,
	}
}
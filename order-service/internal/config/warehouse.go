package config

import (
	"time"
)

// WarehouseConfig holds configuration for the warehouse service integration
type WarehouseConfig struct {
	BaseURL     string        `mapstructure:"base_url"`
	Timeout     time.Duration `mapstructure:"timeout"`
	APIKey      string        `mapstructure:"api_key"`
	MaxRetries  int           `mapstructure:"max_retries"`
	RetryDelay  time.Duration `mapstructure:"retry_delay"`
	AsyncMode   bool          `mapstructure:"async_mode"`
	QueueName   string        `mapstructure:"queue_name"`
	MQAddress   string        `mapstructure:"mq_address"`
	MQUsername  string        `mapstructure:"mq_username"`
	MQPassword  string        `mapstructure:"mq_password"`
}

// GetWarehouseConfig returns the warehouse service configuration
func (c *AppConfig) GetWarehouseConfig() *WarehouseConfig {
	return &WarehouseConfig{
		BaseURL:     c.Viper.GetString("warehouse.base_url"),
		Timeout:     c.Viper.GetDuration("warehouse.timeout"),
		APIKey:      c.Viper.GetString("warehouse.api_key"),
		MaxRetries:  c.Viper.GetInt("warehouse.max_retries"),
		RetryDelay:  c.Viper.GetDuration("warehouse.retry_delay"),
		AsyncMode:   c.Viper.GetBool("warehouse.async_mode"),
		QueueName:   c.Viper.GetString("warehouse.queue_name"),
		MQAddress:   c.Viper.GetString("warehouse.mq_address"),
		MQUsername:  c.Viper.GetString("warehouse.mq_username"),
		MQPassword:  c.Viper.GetString("warehouse.mq_password"),
	}
}
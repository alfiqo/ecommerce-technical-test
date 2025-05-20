package services

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// ServiceConfig holds the configuration for external service connections
type ServiceConfig struct {
	URL     string        // Base URL of the service
	Timeout time.Duration // Timeout for requests in milliseconds
}

// ServicesConfig holds the configuration for all external services
type ServicesConfig struct {
	Warehouse ServiceConfig
}

// NewServicesConfig creates a new configuration for external services
func NewServicesConfig(config *viper.Viper) *ServicesConfig {
	return &ServicesConfig{
		Warehouse: ServiceConfig{
			URL:     config.GetString("services.warehouse.url"),
			Timeout: time.Duration(config.GetInt("services.warehouse.timeout")) * time.Millisecond,
		},
	}
}

// GetEndpointURL returns the full URL for a specific service endpoint
func (s *ServiceConfig) GetEndpointURL(endpoint string) string {
	// If endpoint already starts with '/', don't add another one
	if len(endpoint) > 0 && endpoint[0] == '/' {
		return fmt.Sprintf("%s%s", s.URL, endpoint)
	}
	
	return fmt.Sprintf("%s/%s", s.URL, endpoint)
}
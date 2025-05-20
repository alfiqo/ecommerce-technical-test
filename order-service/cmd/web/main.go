package main

import (
	"flag"
	"fmt"
	"order-service/internal/bootstrap"
	"order-service/internal/config"

	_ "order-service/docs" // Import swagger docs
)

// @title Order Service API
// @version 1.0
// @description API documentation for Order Service
// @contact.name API Support
// @contact.email support@example.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:3000
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	// Parse command line flags
	port := flag.Int("port", 0, "Port to listen on (overrides config file)")
	flag.Parse()

	// Initialize configuration and all required components
	viperConfig := config.NewViper()
	appConfig := config.NewAppConfig(viperConfig)
	log := config.NewLogger(viperConfig)
	db := config.NewDatabase(viperConfig, log)
	validate := config.NewValidator(viperConfig)
	app := config.NewFiber(viperConfig)

	// Bootstrap the application
	bootstrap.Bootstrap(&bootstrap.BootstrapConfig{
		DB:       db,
		App:      app,
		Log:      log,
		Validate: validate,
		Config:   appConfig,
	})

	// Get the web port from config or command line flag
	webPort := viperConfig.GetInt("web.port")
	if *port > 0 {
		webPort = *port
		log.Infof("Overriding config port with command line port: %d", webPort)
	}

	// Start the server
	log.Infof("Starting server on port %d", webPort)
	err := app.Listen(fmt.Sprintf("0.0.0.0:%d", webPort))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
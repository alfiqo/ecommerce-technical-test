package main

import (
	"fmt"
	"os"
	"shop-service/internal/config"
	"shop-service/internal/config/services"
	"strconv"

	_ "shop-service/docs" // Import the generated swagger docs
)

// @title Shop Service API
// @version 1.0
// @description API for managing shops and their warehouse associations
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@example.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:3000
// @BasePath /api/v1
func main() {
	viperConfig := config.NewViper()
	log := config.NewLogger(viperConfig)
	db := config.NewDatabase(viperConfig, log)
	validate := config.NewValidator(viperConfig)
	app := config.NewFiber(viperConfig)

	// Initialize services config
	servicesConfig := services.NewServicesConfig(viperConfig)

	config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		App:      app,
		Log:      log,
		Validate: validate,
		Config:   viperConfig,
		Services: servicesConfig,
	})

	// Get port from environment variable or config
	port := os.Getenv("PORT")
	webPort := viperConfig.GetInt("web.port")

	// If PORT env var is set, use it
	if port != "" {
		if portNum, err := strconv.Atoi(port); err == nil {
			webPort = portNum
		}
	}

	err := app.Listen(fmt.Sprintf("0.0.0.0:%d", webPort))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

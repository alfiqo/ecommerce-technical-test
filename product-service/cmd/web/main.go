package main

import (
	"fmt"
	"product-service/internal/config"
	
	// Import Swagger docs
	_ "product-service/docs"
)

// @title Product Service API
// @version 1.0
// @description This is a product service API in Go using Fiber framework.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// 
// @x-extension-openapi {"example": "Accessible at http://localhost:8080/swagger/index.html or http://localhost:8080/api/v1/docs/index.html"}
func main() {
	viperConfig := config.NewViper()
	log := config.NewLogger(viperConfig)
	db := config.NewDatabase(viperConfig, log)
	validate := config.NewValidator(viperConfig)
	app := config.NewFiber(viperConfig)

	config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		App:      app,
		Log:      log,
		Validate: validate,
		Config:   viperConfig,
	})

	webPort := viperConfig.GetInt("web.port")
	err := app.Listen(fmt.Sprintf("0.0.0.0:%d", webPort))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
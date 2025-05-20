package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	swagger "github.com/swaggo/fiber-swagger"
)

// NewSwagger initializes Swagger configuration
func NewSwagger(app *fiber.App, log *logrus.Logger) {
	log.Info("Initializing Swagger documentation")
	
	// Set up Swagger routes - we add it here for documentation but the actual route is in route.go
	app.Get("/swagger/*", swagger.FiberWrapHandler())
	
	log.Info("Swagger documentation initialized and available at /swagger/index.html")
}
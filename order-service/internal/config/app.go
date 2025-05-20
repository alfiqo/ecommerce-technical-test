package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// BootstrapConfig represents the configuration for the application bootstrap
// Note: This is only used as a reference type for other packages
// The actual bootstrap functionality has been moved to the bootstrap package
type BootstrapConfig struct {
	DB       *gorm.DB
	App      *fiber.App
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *AppConfig
}
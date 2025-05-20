package config

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spf13/viper"
)

func NewFiber(config *viper.Viper) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Second * time.Duration(config.GetInt("web.read_timeout")),
		WriteTimeout: time.Second * time.Duration(config.GetInt("web.write_timeout")),
		IdleTimeout:  time.Second * time.Duration(config.GetInt("web.idle_timeout")),
	})
	
	// Middleware
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())
	
	return app
}
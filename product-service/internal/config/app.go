package config

import (
	"product-service/internal/delivery/http/middleware"
	"product-service/internal/delivery/http/route"
	"product-service/internal/entity"
	"product-service/internal/handler"
	"product-service/internal/repository"
	"product-service/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	DB       *gorm.DB
	App      *fiber.App
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
}

func Bootstrap(config *BootstrapConfig) {
	// Configure the Fiber app with our custom error handler
	appConfig := fiber.Config{
		ErrorHandler: middleware.ErrorHandler(config.Log),
	}
	app := fiber.New(appConfig)
	*config.App = *app

	// Initialize Swagger documentation
	NewSwagger(config.App, config.Log)

	// Auto-migrate database if needed
	if config.Config.GetBool("database.auto_migrate") {
		config.Log.Info("Auto-migrating database tables...")
		// Auto-migrate products table
		err := config.DB.AutoMigrate(&entity.Product{})
		if err != nil {
			config.Log.Fatalf("Failed to migrate database: %v", err)
		}
		config.Log.Info("Database migration completed")
	}

	// Setup repositories
	productRepository := repository.NewProductRepository(config.Log, config.DB)

	// Setup use cases
	productUseCase := usecase.NewProductUseCase(config.DB, config.Log, config.Validate, productRepository)

	// Setup handlers
	productHandler := handler.NewProductHandler(productUseCase, config.Log)

	// Setup routes
	routeConfig := route.RouteConfig{
		App:            config.App,
		ProductHandler: productHandler,
		DB:             config.DB,
		ProductRepo:    productRepository,
		Logger:         config.Log,
	}
	routeConfig.Setup()
}
package config

import (
	"shop-service/internal/config/services"
	"shop-service/internal/delivery/http/route"
	"shop-service/internal/entity"
	"shop-service/internal/gateway"
	"shop-service/internal/handler"
	"shop-service/internal/repository"
	"shop-service/internal/usecase"

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
	Services *services.ServicesConfig
}

func Bootstrap(config *BootstrapConfig) {
	// Configure logger for more contextual information
	config.Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
	})

	// Set log level from configuration
	logLevel := config.Config.GetInt("log.level")
	config.Log.SetLevel(logrus.Level(logLevel))
	
	// Initialize services config if not provided
	if config.Services == nil {
		config.Services = services.NewServicesConfig(config.Config)
	}
	
	config.Log.Info("Bootstrapping application...")

	// Auto-migrate database if needed
	if config.Config.GetBool("database.auto_migrate") {
		config.Log.Info("Auto-migrating database tables...")
		err := config.DB.AutoMigrate(
			&entity.Shop{},
			&entity.ShopWarehouse{},
		)
		if err != nil {
			config.Log.WithField("error", err.Error()).Fatal("Failed to migrate database")
		}
		config.Log.Info("Database migration completed")
	}

	// Setup repositories
	shopRepository := repository.NewShopRepository(config.Log)
	shopWarehouseRepository := repository.NewShopWarehouseRepository(config.Log)
	
	// Setup gateways
	warehouseGateway := gateway.NewWarehouseGateway(config.Log, config.Services)
	
	// Setup usecases
	shopUsecase := usecase.NewShopUsecase(
		config.DB,
		config.Log,
		config.Validate,
		shopRepository,
		shopWarehouseRepository,
		warehouseGateway,
	)
	
	// Setup handlers
	shopHandler := handler.NewShopHandler(shopUsecase, config.Log)
	
	// Configure routes
	routeConfig := route.RouteConfig{
		App:         config.App,
		DB:          config.DB,
		Log:         config.Log,
		ShopHandler: shopHandler,
	}
	
	// Setup routes
	routeConfig.Setup()
	
	config.Log.Info("Application bootstrap completed")
}
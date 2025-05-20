package config

import (
	"warehouse-service/internal/delivery/http/middleware"
	"warehouse-service/internal/delivery/http/route"
	"warehouse-service/internal/entity"
	"warehouse-service/internal/gateway/product"
	"warehouse-service/internal/handler"
	"warehouse-service/internal/repository"
	"warehouse-service/internal/usecase"

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
	// Configure logger for more contextual information
	config.Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
	})

	// Set log level from configuration
	logLevel := config.Config.GetInt("log.level")
	config.Log.SetLevel(logrus.Level(logLevel))
	
	config.Log.Info("Bootstrapping application...")

	// Auto-migrate database if needed
	if config.Config.GetBool("database.auto_migrate") {
		config.Log.Info("Auto-migrating database tables...")
		// Auto-migrate tables
		err := config.DB.AutoMigrate(
			&entity.Warehouse{},
			&entity.WarehouseStock{},
			&entity.StockTransfer{},
			&entity.StockMovement{},
			&entity.ReservationLog{},
		)
		if err != nil {
			config.Log.WithField("error", err.Error()).Fatal("Failed to migrate database")
		}
		config.Log.Info("Database migration completed")
	}

	// setup repositories
	warehouseRepository := repository.NewWarehouseRepository(config.Log, config.DB)
	reservationRepository := repository.NewReservationRepository(config.Log, config.DB)
	stockRepository := repository.NewStockRepository(config.Log, config.DB)
	
	// setup product client
	productClient := product.NewProductClient(config.Log)

	// setup use cases
	warehouseUseCase := usecase.NewWarehouseUseCase(config.DB, config.Log, config.Validate, warehouseRepository)
	reservationUseCase := usecase.NewReservationUseCase(config.DB, config.Log, config.Validate, reservationRepository, warehouseRepository)
	stockUseCase := usecase.NewStockUseCase(config.DB, config.Log, config.Validate, stockRepository, warehouseRepository, productClient)

	// setup handlers
	warehouseHandler := handler.NewWarehouseHandler(warehouseUseCase, config.Log)
	reservationHandler := handler.NewReservationHandler(reservationUseCase, config.Log)
	stockHandler := handler.NewStockHandler(stockUseCase, config.Log)

	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(config.DB)
	authMiddleware.SetLogger(config.Log)

	// Configure routes
	routeConfig := route.RouteConfig{
		App:                config.App,
		WarehouseHandler:   warehouseHandler,
		ReservationHandler: reservationHandler,
		StockHandler:       stockHandler,
		DB:                 config.DB,
		WarehouseRepo:      warehouseRepository,
		Log:                config.Log,
	}
	
	// Setup routes
	routeConfig.Setup()
	
	config.Log.Info("Application bootstrap completed")
}
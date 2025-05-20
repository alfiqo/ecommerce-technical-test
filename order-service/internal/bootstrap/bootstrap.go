package bootstrap

import (
	"order-service/internal/config"
	"order-service/internal/delivery/http/middleware"
	"order-service/internal/delivery/http/route"
	"order-service/internal/entity"
	"order-service/internal/factory"
	"order-service/internal/handler"
	"order-service/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	DB       *gorm.DB
	App      *fiber.App
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *config.AppConfig
}

func Bootstrap(config *BootstrapConfig) {
	// Configure logger for more contextual information
	config.Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
	})

	// Set log level from configuration
	logLevel := config.Config.Viper.GetInt("log.level")
	config.Log.SetLevel(logrus.Level(logLevel))
	
	config.Log.Info("Bootstrapping application...")

	// Auto-migrate database if needed
	if config.Config.Viper.GetBool("database.auto_migrate") {
		config.Log.Info("Auto-migrating database tables...")
		// Auto-migrate tables - removed Inventory entity as it's now handled by warehouse service
		err := config.DB.AutoMigrate(&entity.Order{}, &entity.OrderItem{}, &entity.Reservation{})
		if err != nil {
			config.Log.WithField("error", err.Error()).Fatal("Failed to migrate database")
		}
		config.Log.Info("Database migration completed")
	}

	// Create factory for dependency injection
	appFactory := factory.NewFactory(
		config.DB,
		config.Config,
		config.Log,
		config.Validate,
	)

	// Create repositories through factory
	orderRepository := appFactory.CreateOrderRepository()
	reservationRepository := appFactory.CreateReservationRepository()

	// Setup other use cases
	orderUseCase := appFactory.CreateOrderUseCase()
	reservationUseCase := usecase.NewReservationUseCase(
		config.DB,
		config.Log,
		config.Validate,
		reservationRepository,
		orderRepository,
	)

	// Setup handlers
	orderHandler := handler.NewOrderHandler(orderUseCase, config.Log)
	reservationHandler := handler.NewReservationHandler(reservationUseCase, config.Log)
	warehouseHandler := handler.NewWarehouseHandler(config.Log, appFactory.CreateWarehouseGateway())

	// Create simple auth middleware
	authMiddleware := middleware.NewSimpleAuthMiddleware(config.Log)

	// Configure routes
	routeConfig := route.RouteConfig{
		App:                config.App,
		OrderHandler:       orderHandler,
		ReservationHandler: reservationHandler,
		WarehouseHandler:   warehouseHandler,
		Log:                config.Log,
		AuthMiddleware:     authMiddleware,
	}
	
	// Setup routes
	routeConfig.Setup()
	
	config.Log.Info("Application bootstrap completed")
}
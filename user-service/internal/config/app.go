package config

import (
	"user-service/internal/delivery/http/middleware"
	"user-service/internal/delivery/http/route"
	"user-service/internal/entity"
	"user-service/internal/handler"
	"user-service/internal/repository"
	"user-service/internal/usecase"

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
		// Auto-migrate users table
		err := config.DB.AutoMigrate(&entity.User{})
		if err != nil {
			config.Log.WithField("error", err.Error()).Fatal("Failed to migrate database")
		}
		config.Log.Info("Database migration completed")
	}

	// setup repositories
	userRepository := repository.NewUserRepository(config.Log, config.DB)

	// setup use cases
	userUseCase := usecase.NewUserUseCase(config.DB, config.Log, config.Validate, userRepository)

	// setup handler
	userHandler := handler.NewUserHandler(userUseCase, config.Log)

	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(config.DB, userRepository)
	authMiddleware.SetLogger(config.Log)

	// Configure routes
	routeConfig := route.RouteConfig{
		App:         config.App,
		UserHandler: userHandler,
		DB:          config.DB,
		UserRepo:    userRepository,
		Log:         config.Log,
	}
	
	// Setup routes
	routeConfig.Setup()
	
	config.Log.Info("Application bootstrap completed")
}
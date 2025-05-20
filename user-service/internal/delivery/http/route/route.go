package route

import (
	"github.com/google/uuid"
	"user-service/internal/delivery/http/middleware"
	"user-service/internal/delivery/http/response"
	"user-service/internal/errors"
	"user-service/internal/handler"
	"user-service/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RouteConfig struct {
	App         *fiber.App
	UserHandler *handler.UserHandler
	DB          *gorm.DB
	UserRepo    repository.UserRepositoryInterface
	Log         *logrus.Logger
}

func (c *RouteConfig) Setup() {
	// Add request ID middleware (must be first)
	c.App.Use(func(ctx *fiber.Ctx) error {
		requestID := ctx.Get("X-Request-ID")
		if requestID == "" {
			ctx.Set("X-Request-ID", uuid.New().String())
		}
		return ctx.Next()
	})

	// Apply logger middleware
	c.App.Use(middleware.Logger(c.Log))

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(c.DB, c.UserRepo)
	authMiddleware.SetLogger(c.Log)

	// Swagger documentation
	c.App.Get("/swagger/*", swagger.HandlerDefault)

	// Set up API routes
	api := c.App.Group("/api")
	v1 := api.Group("/v1")

	// Health check endpoint
	v1.Get("/health", func(ctx *fiber.Ctx) error {
		return response.JSONSuccess(ctx, map[string]string{"status": "ok"})
	})

	// Public user endpoints
	v1.Post("/users", c.UserHandler.Register)
	v1.Post("/users/login", c.UserHandler.Login)

	// Protected user endpoints - require authentication
	v1.Get("/users/:id", authMiddleware.RequireAuth(), c.UserHandler.GetUser)

	// 404 Handler
	c.App.Use(func(ctx *fiber.Ctx) error {
		return response.JSONError(ctx, errors.ErrResourceNotFound, c.Log)
	})
}
package route

import (
	"shop-service/internal/delivery/http/middleware"
	"shop-service/internal/delivery/http/response"
	"shop-service/internal/errors"
	"shop-service/internal/handler"

	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"gorm.io/gorm"
)

type RouteConfig struct {
	App         *fiber.App
	DB          *gorm.DB
	Log         *logrus.Logger
	ShopHandler *handler.ShopHandler
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
	
	// Set up Swagger documentation endpoint
	c.App.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Set up API routes
	api := c.App.Group("/api")
	v1 := api.Group("/v1")

	// Health check endpoint
	v1.Get("/health", func(ctx *fiber.Ctx) error {
		return response.JSONSuccess(ctx, map[string]string{"status": "ok"})
	})

	// Shop endpoints
	v1.Get("/shops", c.ShopHandler.ListShops)
	v1.Post("/shops", c.ShopHandler.CreateShop)
	v1.Get("/shops/:id", c.ShopHandler.GetShopByID)
	v1.Get("/shops/:id/warehouses", c.ShopHandler.GetShopWarehouses)

	// 404 Handler
	c.App.Use(func(ctx *fiber.Ctx) error {
		return response.JSONError(ctx, errors.ErrResourceNotFound, c.Log)
	})
}

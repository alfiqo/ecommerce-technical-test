package route

import (
	"product-service/internal/delivery/http/middleware"
	"product-service/internal/delivery/http/response"
	"product-service/internal/errors"
	"product-service/internal/handler"
	"product-service/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	swagger "github.com/swaggo/fiber-swagger"
	"gorm.io/gorm"
)

type RouteConfig struct {
	App            *fiber.App
	ProductHandler *handler.ProductHandler
	DB             *gorm.DB
	ProductRepo    repository.ProductRepositoryInterface
	Logger         *logrus.Logger
}

func (c *RouteConfig) Setup() {
	// Add the request ID middleware to all routes
	c.App.Use(middleware.RequestID())
	
	// Add the logger middleware to all routes
	c.App.Use(middleware.Logger(c.Logger))
	
	// Set up API routes
	api := c.App.Group("/api")
	v1 := api.Group("/v1")

	// Health check endpoint with standardized response
	v1.Get("/health", func(ctx *fiber.Ctx) error {
		return response.JSONSuccess(ctx, fiber.Map{"status": "ok"})
	})
	
	// Swagger documentation endpoint
	v1.Get("/docs/*", swagger.FiberWrapHandler())

	// Product endpoints
	products := v1.Group("/products")
	products.Get("/", c.ProductHandler.GetProducts)
	products.Post("/", c.ProductHandler.CreateProduct)
	
	// IMPORTANT: Order matters in Fiber routing! 
	// Specific paths must come before parameter paths to avoid conflicts
	// For example, "/search" must be defined before "/:id", otherwise "/search" will be matched as an ID
	products.Get("/search", c.ProductHandler.SearchProducts)
	products.Get("/category/:category", c.ProductHandler.GetProductsByCategory)
	
	// Generic parameter routes come after specific routes
	products.Get("/:id", c.ProductHandler.GetProductByID)
	products.Put("/:id", c.ProductHandler.UpdateProduct)
	products.Delete("/:id", c.ProductHandler.DeleteProduct)
	
	// 404 handler for undefined routes
	c.App.Use(func(ctx *fiber.Ctx) error {
		return response.JSONError(ctx, errors.ErrResourceNotFound, c.Logger)
	})
}
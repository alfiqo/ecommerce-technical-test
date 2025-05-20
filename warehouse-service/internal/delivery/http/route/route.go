package route

import (
	"github.com/google/uuid"
	"warehouse-service/internal/delivery/http/middleware"
	"warehouse-service/internal/delivery/http/response"
	"warehouse-service/internal/errors"
	"warehouse-service/internal/handler"
	"warehouse-service/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RouteConfig struct {
	App                *fiber.App
	WarehouseHandler   *handler.WarehouseHandler
	ReservationHandler *handler.ReservationHandler
	StockHandler       *handler.StockHandler
	DB                 *gorm.DB
	WarehouseRepo      repository.WarehouseRepositoryInterface
	Log                *logrus.Logger
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
	authMiddleware := middleware.NewAuthMiddleware(c.DB)
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

	// Admin-only warehouse routes - all endpoints require authentication
	warehouses := v1.Group("/warehouses")
	
	// Apply auth middleware to all warehouse routes
	warehouses.Use(authMiddleware.RequireAuth())
	
	// Warehouse endpoints
	warehouses.Get("/", c.WarehouseHandler.ListWarehouses)
	warehouses.Post("/", c.WarehouseHandler.CreateWarehouse)
	warehouses.Get("/:id", c.WarehouseHandler.GetWarehouse)
	warehouses.Put("/:id", c.WarehouseHandler.UpdateWarehouse)
	warehouses.Delete("/:id", c.WarehouseHandler.DeleteWarehouse)
	
	// Stock management endpoints for warehouses
	warehouses.Get("/:warehouseId/stock", c.StockHandler.GetWarehouseStock)
	warehouses.Post("/:warehouseId/stock", c.StockHandler.AddStock)

	// Inventory routes
	inventory := v1.Group("/inventory")
	
	// Apply auth middleware to all inventory routes
	inventory.Use(authMiddleware.RequireAuth())
	
	// Reservation endpoints
	inventory.Post("/reserve", c.ReservationHandler.ReserveStock)
	inventory.Post("/reserve/cancel", c.ReservationHandler.CancelReservation)
	inventory.Post("/reserve/commit", c.ReservationHandler.CommitReservation)
	
	// Reservation history endpoint 
	inventory.Get("/warehouses/:warehouse_id/products/:product_id/reservations", 
		c.ReservationHandler.GetReservationHistory)
	
	// Stock transfer endpoint (requires authentication)
	stockGroup := v1.Group("/stock") 
	stockGroup.Use(authMiddleware.RequireAuth())
	stockGroup.Post("/transfer", c.StockHandler.TransferStock)
		
	// 404 Handler
	c.App.Use(func(ctx *fiber.Ctx) error {
		return response.JSONError(ctx, errors.ErrResourceNotFound, c.Log)
	})
}
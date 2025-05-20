package route

import (
	"github.com/google/uuid"
	"order-service/internal/delivery/http/middleware"
	"order-service/internal/delivery/http/response"
	"order-service/internal/errors"
	"order-service/internal/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/sirupsen/logrus"
)

type RouteConfig struct {
	App                *fiber.App
	OrderHandler       *handler.OrderHandler
	ReservationHandler *handler.ReservationHandler
	WarehouseHandler   *handler.WarehouseHandler
	Log                *logrus.Logger
	AuthMiddleware     *middleware.SimpleAuthMiddleware
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

	// Swagger documentation
	c.App.Get("/swagger/*", swagger.HandlerDefault)

	// Set up API routes
	api := c.App.Group("/api")
	v1 := api.Group("/v1")

	// Health check endpoint
	v1.Get("/health", func(ctx *fiber.Ctx) error {
		return response.JSONSuccess(ctx, map[string]string{"status": "ok"})
	})

	// Order endpoints
	orders := v1.Group("/orders")
	orders.Post("/", c.AuthMiddleware.RequireAuth(), c.OrderHandler.CreateOrder)
	orders.Get("/", c.AuthMiddleware.RequireAuth(), c.OrderHandler.GetUserOrders)
	orders.Get("/:id", c.AuthMiddleware.RequireAuth(), c.OrderHandler.GetOrder)
	orders.Patch("/:id/status", c.AuthMiddleware.RequireAuth(), c.OrderHandler.UpdateOrderStatus)
	orders.Post("/:id/payment", c.AuthMiddleware.RequireAuth(), c.OrderHandler.ProcessPayment)

	// Order reservation endpoints
	orders.Get("/:order_id/reservations", c.AuthMiddleware.RequireAuth(), c.ReservationHandler.GetOrderReservations)

	// Reservation endpoints
	reservations := v1.Group("/reservations")
	reservations.Post("/", c.AuthMiddleware.RequireAuth(), c.ReservationHandler.CreateReservation)
	reservations.Post("/:id/deactivate", c.AuthMiddleware.RequireAuth(), c.ReservationHandler.DeactivateReservation)
	reservations.Post("/cleanup", c.AuthMiddleware.RequireAuth(), c.ReservationHandler.CleanupExpiredReservations)

	// Inventory endpoints
	inventory := v1.Group("/inventory")
	inventory.Get("/:product_id/:warehouse_id", c.AuthMiddleware.RequireAuth(), c.WarehouseHandler.GetInventory)
	inventory.Post("/batch", c.AuthMiddleware.RequireAuth(), c.WarehouseHandler.GetInventoryBatch)
	inventory.Post("/reserve", c.AuthMiddleware.RequireAuth(), c.WarehouseHandler.ReserveStock)
	inventory.Post("/confirm", c.AuthMiddleware.RequireAuth(), c.WarehouseHandler.ConfirmStockDeduction)
	inventory.Post("/release", c.AuthMiddleware.RequireAuth(), c.WarehouseHandler.ReleaseReservation)

	// 404 Handler
	c.App.Use(func(ctx *fiber.Ctx) error {
		return response.JSONError(ctx, errors.ErrResourceNotFound, c.Log)
	})
}
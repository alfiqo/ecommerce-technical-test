package usecase

import (
	"context"
	"errors"
	"order-service/internal/entity"
	"order-service/internal/model"
	"order-service/internal/model/converter"
	"order-service/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type OrderUseCaseInterface interface {
	CreateOrder(ctx context.Context, request *model.CreateOrderRequest) (*model.OrderResponse, error)
	GetOrderByID(ctx context.Context, orderID uint) (*model.OrderResponse, error)
	GetOrdersByUserID(ctx context.Context, userID string, page, limit int) ([]model.OrderResponse, int64, error)
	UpdateOrderStatus(ctx context.Context, orderID uint, status string) error
	ProcessPayment(ctx context.Context, orderID uint) error
	CancelExpiredOrders(ctx context.Context) error
}

type OrderUseCase struct {
	DB                    *gorm.DB
	Log                   *logrus.Logger
	Validate              *validator.Validate
	OrderRepository       repository.OrderRepositoryInterface
	ReservationRepository repository.ReservationRepositoryInterface
	InventoryUseCase      InventoryUseCaseInterface
}

func NewOrderUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	orderRepository repository.OrderRepositoryInterface,
	reservationRepository repository.ReservationRepositoryInterface,
	inventoryUseCase InventoryUseCaseInterface,
) OrderUseCaseInterface {
	return &OrderUseCase{
		DB:                    db,
		Log:                   logger,
		Validate:              validate,
		OrderRepository:       orderRepository,
		ReservationRepository: reservationRepository,
		InventoryUseCase:      inventoryUseCase,
	}
}

func (c *OrderUseCase) CreateOrder(ctx context.Context, request *model.CreateOrderRequest) (*model.OrderResponse, error) {
	// Validate request first before attempting any resource reservation
	err := c.Validate.Struct(request)
	if err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	// Check if items exist
	if len(request.Items) == 0 {
		c.Log.Warn("Order must contain at least one item")
		return nil, fiber.ErrBadRequest
	}

	// Create a separate context for inventory operations
	inventoryCtx, inventoryCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer inventoryCancel()

	// Check and lock stock before starting the transaction
	// This is a critical step to prevent overselling
	if err := c.InventoryUseCase.CheckAndReserveStock(inventoryCtx, request.Items); err != nil {
		c.Log.Warnf("Failed to reserve stock: %+v", err)

		// Check if it's a stock insufficiency error
		if errors.Is(err, entity.ErrInsufficientStock) {
			return nil, errors.New("insufficient stock available for one or more items")
		}

		return nil, fiber.ErrInternalServerError
	}

	// Create a new context with a longer timeout for database operations
	// This ensures that the database transaction can complete even if the original context deadline is close
	dbCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start a transaction for the order creation with the new context
	tx := c.DB.WithContext(dbCtx).Begin()
	defer tx.Rollback()

	// Calculate total amount
	var totalAmount float64
	orderItems := make([]entity.OrderItem, len(request.Items))

	for i, item := range request.Items {
		totalPrice := float64(item.Quantity) * item.UnitPrice
		totalAmount += totalPrice

		orderItems[i] = entity.OrderItem{
			ProductID:   item.ProductID,
			WarehouseID: item.WarehouseID,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  totalPrice,
		}
	}

	// Set payment deadline to 24 hours from now
	paymentDeadline := time.Now().Add(24 * time.Hour)

	// Create order
	order := &entity.Order{
		UserID:          request.UserID,
		Status:          entity.OrderStatusPending,
		TotalAmount:     totalAmount,
		ShippingAddress: request.ShippingAddress,
		PaymentMethod:   request.PaymentMethod,
		PaymentDeadline: paymentDeadline,
	}

	if err := c.OrderRepository.CreateOrder(tx, order); err != nil {
		c.Log.Warnf("Failed to create order: %+v", err)

		// Release the reserved stock since we're aborting the order
		c.releaseStockForItems(ctx, request.Items)

		return nil, fiber.ErrInternalServerError
	}

	// Set order ID for each item
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}

	// Create order items
	if err := c.OrderRepository.CreateOrderItems(tx, orderItems); err != nil {
		c.Log.Warnf("Failed to create order items: %+v", err)

		// Release the reserved stock since we're aborting the order
		c.releaseStockForItems(ctx, request.Items)

		return nil, fiber.ErrInternalServerError
	}

	// Create stock reservations in the reservation tracking table
	reservations := make([]entity.Reservation, len(request.Items))
	for i, item := range request.Items {
		reservations[i] = entity.Reservation{
			OrderID:     order.ID,
			ProductID:   item.ProductID,
			WarehouseID: item.WarehouseID,
			Quantity:    item.Quantity,
			ExpiresAt:   paymentDeadline,
			IsActive:    true,
		}
	}

	if err := c.ReservationRepository.CreateReservationBatch(tx, reservations); err != nil {
		c.Log.Warnf("Failed to create stock reservations: %+v", err)

		// Release the reserved stock since we're aborting the order
		c.releaseStockForItems(ctx, request.Items)

		return nil, fiber.ErrInternalServerError
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed to commit transaction: %+v", err)

		// Release the reserved stock since we're aborting the order
		c.releaseStockForItems(ctx, request.Items)

		return nil, fiber.ErrInternalServerError
	}

	// Create a new context for loading the created order
	loadCtx, loadCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer loadCancel()

	// Load the created order with its items
	createdOrder, err := c.OrderRepository.FindOrderByID(c.DB.WithContext(loadCtx), order.ID)
	if err != nil {
		c.Log.Warnf("Failed to load created order: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.OrderToResponse(createdOrder), nil
}

// Helper method to release stock for items when an order fails
func (c *OrderUseCase) releaseStockForItems(ctx context.Context, items []model.OrderItemRequest) {
	// Create a new context for inventory operations
	inventoryCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Convert to order items for the inventory usecase
	orderItems := make([]entity.OrderItem, len(items))
	for i, item := range items {
		orderItems[i] = entity.OrderItem{
			ProductID:   item.ProductID,
			WarehouseID: item.WarehouseID,
			Quantity:    item.Quantity,
		}
	}

	// Attempt to release the stock, but just log errors rather than returning them
	if err := c.InventoryUseCase.ReleaseReservation(inventoryCtx, orderItems); err != nil {
		c.Log.Warnf("Failed to release stock reservations: %+v", err)
	}
}

func (c *OrderUseCase) GetOrderByID(ctx context.Context, orderID uint) (*model.OrderResponse, error) {
	// Create a new context with a timeout for database operations
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	order, err := c.OrderRepository.FindOrderByID(c.DB.WithContext(dbCtx), orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Log.Warnf("Order not found: %d", orderID)
			return nil, fiber.ErrNotFound
		}
		c.Log.Warnf("Failed to find order: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.OrderToResponse(order), nil
}

func (c *OrderUseCase) GetOrdersByUserID(ctx context.Context, userID string, page, limit int) ([]model.OrderResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Create a new context with a timeout for database operations
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orders, total, err := c.OrderRepository.FindOrdersByUserID(c.DB.WithContext(dbCtx), userID, page, limit)
	if err != nil {
		c.Log.Warnf("Failed to find orders by user ID: %+v", err)
		return nil, 0, fiber.ErrInternalServerError
	}

	return converter.OrdersToResponse(orders), total, nil
}

func (c *OrderUseCase) UpdateOrderStatus(ctx context.Context, orderID uint, status string) error {
	// Create a new context with a timeout for database operations
	dbCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx := c.DB.WithContext(dbCtx).Begin()
	defer tx.Rollback()

	// Validate status
	orderStatus := entity.OrderStatus(status)
	validStatuses := map[entity.OrderStatus]bool{
		entity.OrderStatusPending:   true,
		entity.OrderStatusPaid:      true,
		entity.OrderStatusCancelled: true,
		entity.OrderStatusCompleted: true,
	}

	if !validStatuses[orderStatus] {
		c.Log.Warnf("Invalid order status: %s", status)
		return fiber.ErrBadRequest
	}

	// Get current order to check current status
	order, err := c.OrderRepository.FindOrderByID(tx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Log.Warnf("Order not found: %d", orderID)
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed to find order: %+v", err)
		return fiber.ErrInternalServerError
	}

	// Handle inventory and reservation updates based on status change
	if orderStatus == entity.OrderStatusCancelled {
		// For cancellations, release the inventory reservation
		if order.Status == entity.OrderStatusPending {
			// First deactivate reservations in the reservation tracking table
			if err := c.ReservationRepository.DeactivateReservationsByOrderID(tx, orderID); err != nil {
				c.Log.Warnf("Failed to deactivate reservations: %+v", err)
				return fiber.ErrInternalServerError
			}

			// Update order status before external service calls
			if err := c.OrderRepository.UpdateOrderStatus(tx, orderID, orderStatus); err != nil {
				c.Log.Warnf("Failed to update order status: %+v", err)
				return fiber.ErrInternalServerError
			}

			// Commit transaction before making external service call
			if err := tx.Commit().Error; err != nil {
				c.Log.Warnf("Failed to commit transaction: %+v", err)
				return fiber.ErrInternalServerError
			}

			// Create a separate context for inventory operations
			inventoryCtx, inventoryCancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer inventoryCancel()

			// Release stock in inventory system - this is now outside the transaction
			if err := c.InventoryUseCase.ReleaseReservation(inventoryCtx, order.OrderItems); err != nil {
				c.Log.Warnf("Failed to release inventory reservation: %+v", err)
				// The order is already marked as cancelled, so this is just a warning
				// We don't want to fail the operation if just the inventory release fails
				return nil
			}

			// Early return since we've already committed the transaction
			return nil
		}
	} else if orderStatus == entity.OrderStatusPaid && order.Status == entity.OrderStatusPending {
		// For payment confirmation, update the database first
		if err := c.OrderRepository.UpdateOrderStatus(tx, orderID, orderStatus); err != nil {
			c.Log.Warnf("Failed to update order status: %+v", err)
			return fiber.ErrInternalServerError
		}

		// Commit transaction before making external service call
		if err := tx.Commit().Error; err != nil {
			c.Log.Warnf("Failed to commit transaction: %+v", err)
			return fiber.ErrInternalServerError
		}

		// Create a separate context for inventory operations
		inventoryCtx, inventoryCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer inventoryCancel()

		// Deduct stock permanently - this is now outside the transaction
		if err := c.InventoryUseCase.ConfirmStockDeduction(inventoryCtx, order.OrderItems); err != nil {
			c.Log.Warnf("Failed to confirm stock deduction: %+v", err)
			// The order is already marked as paid, so this is just a warning
			// We'll need an operational process to reconcile these edge cases
			return nil
		}

		// Early return since we've already committed the transaction
		return nil
	}

	// Update order status
	if err := c.OrderRepository.UpdateOrderStatus(tx, orderID, orderStatus); err != nil {
		c.Log.Warnf("Failed to update order status: %+v", err)
		return fiber.ErrInternalServerError
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed to commit transaction: %+v", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func (c *OrderUseCase) ProcessPayment(ctx context.Context, orderID uint) error {
	// This is a simplified implementation
	// In a real system, this would integrate with a payment gateway

	// Create a new context with a timeout for database operations
	dbCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx := c.DB.WithContext(dbCtx).Begin()
	defer tx.Rollback()

	// Get order with its items
	order, err := c.OrderRepository.FindOrderByID(tx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Log.Warnf("Order not found: %d", orderID)
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed to find order: %+v", err)
		return fiber.ErrInternalServerError
	}

	// Check if order is in pending status
	if order.Status != entity.OrderStatusPending {
		c.Log.Warnf("Cannot process payment for non-pending order: %d", orderID)
		return fiber.ErrBadRequest
	}

	// First update local database
	// Update order status to paid
	if err := c.OrderRepository.UpdateOrderStatus(tx, orderID, entity.OrderStatusPaid); err != nil {
		c.Log.Warnf("Failed to update order status: %+v", err)
		return fiber.ErrInternalServerError
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed to commit transaction: %+v", err)
		return fiber.ErrInternalServerError
	}

	// Create a separate context for inventory operations
	inventoryCtx, inventoryCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer inventoryCancel()

	// Now that the database transaction is committed, make the external service call
	// Permanently deduct stock from inventory (converting reservation to actual sale)
	if err := c.InventoryUseCase.ConfirmStockDeduction(inventoryCtx, order.OrderItems); err != nil {
		c.Log.Warnf("Failed to confirm stock deduction: %+v", err)
		// The order is already marked as paid, so log but don't fail the operation
		// This would typically trigger an alert for manual reconciliation
	}

	return nil
}

func (c *OrderUseCase) CancelExpiredOrders(ctx context.Context) error {
	// Create a new context with a timeout for database operations
	dbCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx := c.DB.WithContext(dbCtx).Begin()
	defer tx.Rollback()

	currentTime := time.Now()

	// Find expired pending orders with their items
	expiredOrders, err := c.OrderRepository.FindExpiredOrders(tx, currentTime)
	if err != nil {
		c.Log.Warnf("Failed to find expired orders: %+v", err)
		return fiber.ErrInternalServerError
	}

	for _, order := range expiredOrders {
		// First update the database to mark the order as cancelled and deactivate reservations
		// This ensures we don't leave the database in an inconsistent state if inventory release fails

		// Update order status to cancelled
		if err := c.OrderRepository.UpdateOrderStatus(tx, order.ID, entity.OrderStatusCancelled); err != nil {
			c.Log.Warnf("Failed to update order status: %+v", err)
			return fiber.ErrInternalServerError
		}

		// Deactivate reservations in tracking table
		if err := c.ReservationRepository.DeactivateReservationsByOrderID(tx, order.ID); err != nil {
			c.Log.Warnf("Failed to deactivate reservations: %+v", err)
			return fiber.ErrInternalServerError
		}

		// Create a separate context for inventory operations
		inventoryCtx, inventoryCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer inventoryCancel()

		// Release stock in inventory system (after database updates)
		// If this fails, at least the database is consistent and we can retry inventory releases
		if err := c.InventoryUseCase.ReleaseReservation(inventoryCtx, order.OrderItems); err != nil {
			c.Log.Warnf("Failed to release inventory for expired order %d: %+v", order.ID, err)
			// Continue with other orders rather than failing the entire process
			// The warehouse service should have its own cleanup process for handling orphaned reservations
		}
	}

	// Find expired reservations that are still active
	expiredReservations, err := c.ReservationRepository.FindExpiredReservations(tx, currentTime)
	if err != nil {
		c.Log.Warnf("Failed to find expired reservations: %+v", err)
		return fiber.ErrInternalServerError
	}

	// Process expired reservations in groups by order
	reservationsByOrder := make(map[uint][]entity.Reservation)
	for _, reservation := range expiredReservations {
		reservationsByOrder[reservation.OrderID] = append(reservationsByOrder[reservation.OrderID], reservation)
	}

	// For each order's reservations, update tracking table first, then release inventory
	for orderID, reservations := range reservationsByOrder {
		// First update reservation status in tracking table
		for _, res := range reservations {
			if err := c.ReservationRepository.UpdateReservationStatus(tx, res.ID, false); err != nil {
				c.Log.Warnf("Failed to deactivate reservation: %+v", err)
				return fiber.ErrInternalServerError
			}
		}

		// Convert reservations to order items for inventory release
		var orderItems []entity.OrderItem
		for _, res := range reservations {
			orderItems = append(orderItems, entity.OrderItem{
				ProductID:   res.ProductID,
				WarehouseID: res.WarehouseID,
				Quantity:    res.Quantity,
			})
		}

		// Release inventory after database is updated
		// Create a separate context for inventory operations
		invCtx, invCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer invCancel()

		if err := c.InventoryUseCase.ReleaseReservation(invCtx, orderItems); err != nil {
			c.Log.Warnf("Failed to release inventory for expired reservations of order %d: %+v", orderID, err)
			// Continue with other reservations rather than failing the entire process
			// The warehouse service should have its own cleanup process
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed to commit transaction: %+v", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

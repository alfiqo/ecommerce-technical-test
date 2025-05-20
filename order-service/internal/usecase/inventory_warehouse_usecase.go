package usecase

import (
	"context"
	"fmt"
	"order-service/internal/entity"
	"order-service/internal/gateway/warehouse"
	"order-service/internal/model"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InventoryWarehouseUseCase implements inventory business logic using the warehouse service
type InventoryWarehouseUseCase struct {
	DB               *gorm.DB
	Log              *logrus.Logger
	WarehouseGateway warehouse.WarehouseGatewayInterface
}

// NewInventoryWarehouseUseCase creates a new InventoryWarehouseUseCase instance
func NewInventoryWarehouseUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	_ interface{}, // Kept for compatibility but not used
	warehouseGateway warehouse.WarehouseGatewayInterface,
) *InventoryWarehouseUseCase {
	return &InventoryWarehouseUseCase{
		DB:               db,
		Log:              log,
		WarehouseGateway: warehouseGateway,
	}
}

// CheckAndReserveStock checks and reserves stock for multiple items
func (uc *InventoryWarehouseUseCase) CheckAndReserveStock(ctx context.Context, items []model.OrderItemRequest) error {
	// Create a new context with a longer timeout for database operations
	// This ensures that the database transaction can complete even if the original context deadline is close
	dbCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Start a transaction to store our local reservation data
	tx := uc.DB.WithContext(dbCtx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate a temporary order ID if not provided (will be updated later)
	// In real implementation, we would use the actual order ID
	var orderID uint = 0
	for _, item := range items {
		if item.OrderID > 0 {
			orderID = item.OrderID
			break
		}
	}

	// Create a context with a timeout for the warehouse service call
	// Using a separate context with a longer timeout that's not tied to the original request
	warehouseCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Log the warehouse service call for debugging
	uc.Log.WithFields(logrus.Fields{
		"orderID": orderID,
		"items":   items,
	}).Info("Calling warehouse service to reserve stock")
	
	// Call warehouse service to reserve stock
	reservationResp, err := uc.WarehouseGateway.CheckAndReserveStock(warehouseCtx, orderID, items, "")
	if err != nil {
		tx.Rollback()
		if err == warehouse.ErrInsufficientStock {
			return entity.ErrInsufficientStock
		}
		// Improve error logging to help diagnose the issue
		uc.Log.WithError(err).WithFields(logrus.Fields{
			"orderID": orderID,
		}).Error("Failed to reserve stock in warehouse")
		return fmt.Errorf("failed to reserve stock in warehouse: %w", err)
	}

	// Store successful reservation information in our local database for tracking
	for _, item := range reservationResp.Items {
		if !item.Available {
			tx.Rollback()
			return entity.ErrInsufficientStock
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// ConfirmStockDeduction commits reserved stock as sold (after payment)
func (uc *InventoryWarehouseUseCase) ConfirmStockDeduction(ctx context.Context, orderItems []entity.OrderItem) error {
	// Get order ID from the first item
	if len(orderItems) == 0 {
		return fmt.Errorf("no order items provided")
	}

	orderID := orderItems[0].OrderID

	// Get reservation ID for the order
	// In a real implementation, we would store the reservation ID from the warehouse service
	// For now, we'll just use the order ID as a placeholder
	reservationID := fmt.Sprintf("res_%d", orderID)

	// Create a context with a timeout for the warehouse service call
	warehouseCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Log the warehouse service call for debugging
	uc.Log.WithFields(logrus.Fields{
		"orderID":       orderID,
		"reservationID": reservationID,
	}).Info("Calling warehouse service to confirm stock deduction")
	
	// Call warehouse service to confirm stock deduction
	_, err := uc.WarehouseGateway.ConfirmStockDeduction(warehouseCtx, orderID, reservationID)
	if err != nil {
		uc.Log.WithError(err).WithFields(logrus.Fields{
			"orderID":       orderID,
			"reservationID": reservationID,
		}).Error("Failed to confirm stock deduction in warehouse")
		return fmt.Errorf("failed to confirm stock deduction in warehouse: %w", err)
	}

	return nil
}

// ReleaseReservation releases stock back to available inventory (e.g., cancelled order)
func (uc *InventoryWarehouseUseCase) ReleaseReservation(ctx context.Context, orderItems []entity.OrderItem) error {
	// Get order ID from the first item
	if len(orderItems) == 0 {
		return fmt.Errorf("no order items provided")
	}

	orderID := orderItems[0].OrderID

	// Get reservation ID for the order
	// In a real implementation, we would store the reservation ID from the warehouse service
	// For now, we'll just use the order ID as a placeholder
	reservationID := fmt.Sprintf("res_%d", orderID)

	// Create a context with a timeout for the warehouse service call
	warehouseCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Log the warehouse service call for debugging
	uc.Log.WithFields(logrus.Fields{
		"orderID":       orderID,
		"reservationID": reservationID,
	}).Info("Calling warehouse service to release reservation")
	
	// Create the request with order item details
	releaseRequest := warehouse.ReservationReleaseRequest{
		WarehouseID: orderItems[0].WarehouseID,
		ProductID:   orderItems[0].ProductID,
		Quantity:    orderItems[0].Quantity,
		Reference:   reservationID,
	}
	
	// Call warehouse service to release reservation
	_, err := uc.WarehouseGateway.ReleaseReservation(warehouseCtx, orderID, releaseRequest)
	if err != nil {
		if err == warehouse.ErrReservationNotFound {
			// If the reservation doesn't exist, consider it already released
			uc.Log.Warnf("Reservation not found for order %d, considering it already released", orderID)
			return nil
		}
		uc.Log.WithError(err).WithFields(logrus.Fields{
			"orderID":     orderID,
			"warehouseID": releaseRequest.WarehouseID,
			"productID":   releaseRequest.ProductID,
			"quantity":    releaseRequest.Quantity,
			"reference":   releaseRequest.Reference,
		}).Error("Failed to release reservation in warehouse")
		return fmt.Errorf("failed to release reservation in warehouse: %w", err)
	}

	return nil
}

// GetInventory gets current inventory level for a product
func (uc *InventoryWarehouseUseCase) GetInventory(ctx context.Context, productID, warehouseID uint) (*entity.Inventory, error) {
	// Create a context with a timeout for the warehouse service call
	warehouseCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Log the warehouse service call for debugging
	uc.Log.WithFields(logrus.Fields{
		"productID":   productID,
		"warehouseID": warehouseID,
	}).Info("Calling warehouse service to get inventory")
	
	// Call warehouse service to get inventory
	inventoryResp, err := uc.WarehouseGateway.GetInventory(warehouseCtx, productID, warehouseID)
	if err != nil {
		uc.Log.WithError(err).WithFields(logrus.Fields{
			"productID":   productID,
			"warehouseID": warehouseID,
		}).Error("Failed to get inventory from warehouse")
		return nil, fmt.Errorf("failed to get inventory from warehouse: %w", err)
	}

	// Convert response to entity
	inventory := &entity.Inventory{
		ProductID:       inventoryResp.ProductID,
		WarehouseID:     inventoryResp.WarehouseID,
		Quantity:        inventoryResp.Quantity,
		ReservedQuantity: inventoryResp.ReservedQuantity,
		CreatedAt:       inventoryResp.CreatedAt,
		UpdatedAt:       inventoryResp.UpdatedAt,
	}

	return inventory, nil
}

// UpdateInventory updates inventory quantity
func (uc *InventoryWarehouseUseCase) UpdateInventory(ctx context.Context, inventory *entity.Inventory) error {
	// Create a context with a timeout for the warehouse service call
	warehouseCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Log the warehouse service call for debugging
	uc.Log.WithFields(logrus.Fields{
		"productID":   inventory.ProductID,
		"warehouseID": inventory.WarehouseID,
		"quantity":    inventory.Quantity,
	}).Info("Calling warehouse service to update inventory")
	
	// Call warehouse service to update inventory
	_, err := uc.WarehouseGateway.UpdateInventory(warehouseCtx, inventory)
	if err != nil {
		uc.Log.WithError(err).WithFields(logrus.Fields{
			"productID":   inventory.ProductID,
			"warehouseID": inventory.WarehouseID,
		}).Error("Failed to update inventory in warehouse")
		return fmt.Errorf("failed to update inventory in warehouse: %w", err)
	}

	return nil
}
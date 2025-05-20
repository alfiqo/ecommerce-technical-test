package usecase

import (
	"context"
	"order-service/internal/entity"
	"order-service/internal/model"
)

// InventoryUseCaseInterface defines the contract for inventory business logic
type InventoryUseCaseInterface interface {
	// CheckAndReserveStock checks and reserves stock for multiple items in a transaction
	CheckAndReserveStock(ctx context.Context, items []model.OrderItemRequest) error

	// ConfirmStockDeduction commits reserved stock as sold (after payment)
	ConfirmStockDeduction(ctx context.Context, orderItems []entity.OrderItem) error

	// ReleaseReservation releases stock back to available inventory (e.g., cancelled order)
	ReleaseReservation(ctx context.Context, orderItems []entity.OrderItem) error

	// GetInventory gets current inventory level for a product
	GetInventory(ctx context.Context, productID, warehouseID uint) (*entity.Inventory, error)

	// UpdateInventory updates inventory quantity
	UpdateInventory(ctx context.Context, inventory *entity.Inventory) error
}
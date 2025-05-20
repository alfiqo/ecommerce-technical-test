package warehouse

import (
	"context"
	"order-service/internal/entity"
	"order-service/internal/model"
)

// InventoryQuery represents a query for a specific product/warehouse inventory
type InventoryQuery struct {
	ProductID   uint
	WarehouseID uint
}

// WarehouseGatewayInterface defines the contract for interacting with the warehouse service
type WarehouseGatewayInterface interface {
	// CheckAndReserveStock checks and reserves stock for multiple items
	CheckAndReserveStock(ctx context.Context, orderID uint, items []model.OrderItemRequest, reserveUntil string) (*ReservationResponse, error)

	// ConfirmStockDeduction commits reserved stock as sold (after payment)
	ConfirmStockDeduction(ctx context.Context, orderID uint, reservationID string) (*StockOperationResponse, error)

	// ReleaseReservation releases stock back to available inventory (e.g., cancelled order)
	ReleaseReservation(ctx context.Context, orderID uint, reservation ReservationReleaseRequest) (*StockOperationResponse, error)

	// GetInventory gets current inventory level for a product
	GetInventory(ctx context.Context, productID, warehouseID uint) (*InventoryResponse, error)

	// GetInventoryBatch gets current inventory levels for multiple products
	GetInventoryBatch(ctx context.Context, items []InventoryQuery) (map[string]*InventoryResponse, error)

	// UpdateInventory updates inventory quantity (admin operation)
	UpdateInventory(ctx context.Context, inventory *entity.Inventory) (*StockOperationResponse, error)
}
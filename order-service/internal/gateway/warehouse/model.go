package warehouse

import (
	"time"
)

// InventoryResponse represents the inventory data returned from the warehouse service
type InventoryResponse struct {
	ID               uint      `json:"id"`
	ProductID        uint      `json:"product_id"`
	WarehouseID      uint      `json:"warehouse_id"`
	Quantity         int       `json:"quantity"`
	ReservedQuantity int       `json:"reserved_quantity"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// AvailableQuantity returns the quantity available for reservation
func (i *InventoryResponse) AvailableQuantity() int {
	return i.Quantity - i.ReservedQuantity
}

// ReservationRequest represents a request to reserve inventory
type ReservationRequest struct {
	OrderID      uint                   `json:"order_id"`
	OrderItems   []ReservationOrderItem `json:"order_items"`
	ReserveUntil time.Time              `json:"reserve_until"`
}

type ReserveStockRequest struct {
	WarehouseID uint `json:"warehouse_id" validate:"required"`
	ProductID   uint `json:"product_id" validate:"required"`
	Quantity    int  `json:"quantity" validate:"required,gt=0"`
}

// ReservationOrderItem represents an item to be reserved in warehouse inventory
type ReservationOrderItem struct {
	ProductID   uint    `json:"product_id"`
	WarehouseID uint    `json:"warehouse_id"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price,omitempty"`
}

// ReservationResponse represents a response from the warehouse service for a reservation
type ReservationResponse struct {
	ReservationID string                    `json:"reservation_id"`
	OrderID       uint                      `json:"order_id"`
	Items         []ReservationResponseItem `json:"items"`
	Success       bool                      `json:"success"`
	Message       string                    `json:"message,omitempty"`
	ReservedUntil time.Time                 `json:"reserved_until"`
}

// ReservationResponseItem represents an item in a reservation response
type ReservationResponseItem struct {
	ProductID   uint   `json:"product_id"`
	WarehouseID uint   `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
	Available   bool   `json:"available"`
	Message     string `json:"message,omitempty"`
}

// ReservationReleaseRequest represents a request to release a reservation
type ReservationReleaseRequest struct {
	WarehouseID uint   `json:"warehouse_id" validate:"required"`
	ProductID   uint   `json:"product_id" validate:"required"`
	Quantity    int    `json:"quantity" validate:"required,gt=0"`
	Reference   string `json:"reference" validate:"required"`
}

// ReservationConfirmRequest represents a request to confirm (deduct) a reservation
type ReservationConfirmRequest struct {
	ReservationID string `json:"reservation_id,omitempty"`
	OrderID       uint   `json:"order_id"`
}

// InventoryQueryRequest represents a request to query inventory
type InventoryQueryRequest struct {
	ProductID   uint `json:"product_id"`
	WarehouseID uint `json:"warehouse_id"`
}

// InventoryQuery is already defined in interface.go

// InventoryBatchQueryRequest represents a request to query multiple inventory items
type InventoryBatchQueryRequest struct {
	Items []InventoryQueryRequest `json:"items"`
}

// StockOperationResponse represents a generic response from the warehouse service
type StockOperationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

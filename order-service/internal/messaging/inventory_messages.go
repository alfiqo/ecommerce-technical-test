package messaging

import (
	"order-service/internal/model"
	"time"
)

// Message types for inventory operations
const (
	MessageTypeReserveStock    = "inventory.reserve"
	MessageTypeConfirmStock    = "inventory.confirm"
	MessageTypeReleaseStock    = "inventory.release"
	MessageTypeGetInventory    = "inventory.get"
	MessageTypeUpdateInventory = "inventory.update"
)

// InventoryMessage is the base message structure for all inventory messages
type InventoryMessage struct {
	Type          string    `json:"type"`
	OrderID       uint      `json:"order_id"`
	ReservationID string    `json:"reservation_id,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	CorrelationID string    `json:"correlation_id,omitempty"`
}

// ReserveStockMessage is used to request stock reservation
type ReserveStockMessage struct {
	InventoryMessage
	Items        []ReserveStockItem `json:"items"`
	ReserveUntil time.Time          `json:"reserve_until,omitempty"`
}

// ReserveStockItem represents an item to reserve stock for
type ReserveStockItem struct {
	ProductID   uint    `json:"product_id"`
	WarehouseID uint    `json:"warehouse_id"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price,omitempty"`
}

// NewReserveStockMessage creates a new ReserveStockMessage
func NewReserveStockMessage(orderID uint, items []model.OrderItemRequest, reserveUntil time.Time) *ReserveStockMessage {
	stockItems := make([]ReserveStockItem, len(items))
	for i, item := range items {
		stockItems[i] = ReserveStockItem{
			ProductID:   item.ProductID,
			WarehouseID: item.WarehouseID,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		}
	}

	return &ReserveStockMessage{
		InventoryMessage: InventoryMessage{
			Type:      MessageTypeReserveStock,
			OrderID:   orderID,
			Timestamp: time.Now(),
		},
		Items:        stockItems,
		ReserveUntil: reserveUntil,
	}
}

// ConfirmStockMessage is used to confirm stock deduction after payment
type ConfirmStockMessage struct {
	InventoryMessage
}

// NewConfirmStockMessage creates a new ConfirmStockMessage
func NewConfirmStockMessage(orderID uint, reservationID string) *ConfirmStockMessage {
	return &ConfirmStockMessage{
		InventoryMessage: InventoryMessage{
			Type:          MessageTypeConfirmStock,
			OrderID:       orderID,
			ReservationID: reservationID,
			Timestamp:     time.Now(),
		},
	}
}

// ReleaseStockMessage is used to release previously reserved stock
type ReleaseStockMessage struct {
	InventoryMessage
	WarehouseID uint   `json:"warehouse_id"`
	ProductID   uint   `json:"product_id"`
	Quantity    int    `json:"quantity"`
	Reference   string `json:"reference"`
}

// NewReleaseStockMessage creates a new ReleaseStockMessage
func NewReleaseStockMessage(orderID uint, reservationID string) *ReleaseStockMessage {
	return &ReleaseStockMessage{
		InventoryMessage: InventoryMessage{
			Type:          MessageTypeReleaseStock,
			OrderID:       orderID,
			ReservationID: reservationID,
			Timestamp:     time.Now(),
		},
		// Use default values which should be replaced by the caller
		WarehouseID: 1,
		ProductID:   1,
		Quantity:    1,
		Reference:   reservationID,
	}
}

// InventoryResponseMessage is the base response message for inventory operations
type InventoryResponseMessage struct {
	Type          string    `json:"type"`
	OrderID       uint      `json:"order_id,omitempty"`
	ReservationID string    `json:"reservation_id,omitempty"`
	Success       bool      `json:"success"`
	Message       string    `json:"message,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	CorrelationID string    `json:"correlation_id,omitempty"`
}

// ReserveStockResponseMessage is the response for a stock reservation request
type ReserveStockResponseMessage struct {
	InventoryResponseMessage
	Items         []ReserveStockResponseItem `json:"items,omitempty"`
	ReservationID string                     `json:"reservation_id,omitempty"`
}

// ReserveStockResponseItem represents the result of a stock reservation for an item
type ReserveStockResponseItem struct {
	ProductID   uint   `json:"product_id"`
	WarehouseID uint   `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
	Available   bool   `json:"available"`
	Message     string `json:"message,omitempty"`
}
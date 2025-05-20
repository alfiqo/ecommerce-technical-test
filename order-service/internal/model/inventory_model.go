package model

// InventoryResponse represents the inventory data for a product at a warehouse
type InventoryResponse struct {
	ProductID        uint `json:"product_id"`
	WarehouseID      uint `json:"warehouse_id"`
	Quantity         int  `json:"quantity"`
	ReservedQuantity int  `json:"reserved_quantity"`
	AvailableQuantity int `json:"available_quantity"`
}

// InventoryQuery represents a query for a specific product's inventory
type InventoryQuery struct {
	ProductID   uint `json:"product_id" validate:"required"`
	WarehouseID uint `json:"warehouse_id" validate:"required"`
}

// InventoryBatchRequest represents a request to get inventory for multiple products
type InventoryBatchRequest struct {
	Items []InventoryQuery `json:"items" validate:"required,dive"`
}

// InventoryBatchResponse represents a response with inventory for multiple products
type InventoryBatchResponse struct {
	Items map[string]InventoryResponse `json:"items"`
}

// StockReservationRequest represents a request to reserve stock for an order
type StockReservationRequest struct {
	OrderID      uint              `json:"order_id" validate:"required"`
	Items        []OrderItemRequest `json:"items" validate:"required,dive"`
	ReserveUntil string            `json:"reserve_until,omitempty"`
}

// StockReservationItem represents an item in a stock reservation response
type StockReservationItem struct {
	ProductID   uint   `json:"product_id"`
	WarehouseID uint   `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
	Available   bool   `json:"available"`
	Message     string `json:"message,omitempty"`
}

// StockReservationResponse represents a response to a stock reservation request
type StockReservationResponse struct {
	ReservationID string                `json:"reservation_id"`
	OrderID       uint                  `json:"order_id"`
	Items         []StockReservationItem `json:"items"`
	Success       bool                  `json:"success"`
	Message       string                `json:"message,omitempty"`
	ReservedUntil string                `json:"reserved_until"`
}

// StockOperationRequest represents a request for a stock operation like confirm or release
type StockOperationRequest struct {
	OrderID       uint   `json:"order_id" validate:"required"`
	ReservationID string `json:"reservation_id,omitempty"`
	WarehouseID   uint   `json:"warehouse_id,omitempty"`
	ProductID     uint   `json:"product_id,omitempty"`
	Quantity      int    `json:"quantity,omitempty"`
}

// StockOperationResponse represents a response to a stock operation
type StockOperationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
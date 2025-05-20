package model

import (
	"time"
)

// CreateOrderRequest is a struct for order creation requests
type CreateOrderRequest struct {
	UserID          string               `json:"user_id" validate:"required"`
	ShippingAddress string               `json:"shipping_address" validate:"required"`
	PaymentMethod   string               `json:"payment_method" validate:"required,max=50"`
	Items           []OrderItemRequest   `json:"items" validate:"required,dive"`
}

// OrderItemRequest represents an item in the order creation request
type OrderItemRequest struct {
	OrderID     uint    `json:"order_id"`      // Added for compatibility with warehouse service
	ProductID   uint    `json:"product_id" validate:"required"`
	WarehouseID uint    `json:"warehouse_id" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,min=0"`
}

// UpdateOrderStatusRequest is used to update an order's status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending paid cancelled completed"`
}

// OrderResponse represents the response structure for an order
type OrderResponse struct {
	ID              uint                  `json:"id"`
	UserID          string                `json:"user_id"`
	Status          string                `json:"status"`
	TotalAmount     float64               `json:"total_amount"`
	ShippingAddress string                `json:"shipping_address"`
	PaymentMethod   string                `json:"payment_method"`
	PaymentDeadline string                `json:"payment_deadline"`
	CreatedAt       string                `json:"created_at"`
	UpdatedAt       string                `json:"updated_at"`
	Items           []OrderItemResponse   `json:"items,omitempty"`
}

// OrderItemResponse represents an item in the order response
type OrderItemResponse struct {
	ID          uint    `json:"id"`
	ProductID   uint    `json:"product_id"`
	WarehouseID uint    `json:"warehouse_id"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

// OrderFilter represents query parameters for filtering orders
type OrderFilter struct {
	UserID    string `query:"user_id"`
	Status    string `query:"status"`
	StartDate string `query:"start_date"`
	EndDate   string `query:"end_date"`
	Page      int    `query:"page"`
	Limit     int    `query:"limit"`
}

// ReservationRequest is used for creating stock reservations
type ReservationRequest struct {
	OrderID     uint      `json:"order_id" validate:"required"`
	ProductID   uint      `json:"product_id" validate:"required"`
	WarehouseID uint      `json:"warehouse_id" validate:"required"`
	Quantity    int       `json:"quantity" validate:"required,min=1"`
	ExpiresAt   time.Time `json:"expires_at" validate:"required"`
}

// ReservationResponse represents the response structure for a reservation
type ReservationResponse struct {
	ID          uint      `json:"id"`
	OrderID     uint      `json:"order_id"`
	ProductID   uint      `json:"product_id"`
	WarehouseID uint      `json:"warehouse_id"`
	Quantity    int       `json:"quantity"`
	ExpiresAt   string    `json:"expires_at"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   string    `json:"created_at"`
}
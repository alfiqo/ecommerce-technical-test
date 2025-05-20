package model

// ReservationStatus represents the status of a reservation
type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "pending"
	ReservationStatusCommitted ReservationStatus = "committed"
	ReservationStatusCancelled ReservationStatus = "cancelled"
)

// ReserveStockRequest represents a request to reserve stock
type ReserveStockRequest struct {
	WarehouseID uint `json:"warehouse_id" validate:"required"`
	ProductID   uint `json:"product_id" validate:"required"`
	Quantity    int  `json:"quantity" validate:"required,gt=0"`
}

// CancelReservationRequest represents a request to cancel a reservation
type CancelReservationRequest struct {
	WarehouseID uint   `json:"warehouse_id" validate:"required"`
	ProductID   uint   `json:"product_id" validate:"required"`
	Quantity    int    `json:"quantity" validate:"required,gt=0"`
	Reference   string `json:"reference" validate:"required"`
}

// CommitReservationRequest represents a request to commit a reservation
type CommitReservationRequest struct {
	WarehouseID uint   `json:"warehouse_id" validate:"required"`
	ProductID   uint   `json:"product_id" validate:"required"`
	Quantity    int    `json:"quantity" validate:"required,gt=0"`
	Reference   string `json:"reference" validate:"required"`
}

// ReservationResponse represents a response to a stock reservation request
type ReservationResponse struct {
	WarehouseID        uint             `json:"warehouse_id"`
	ProductID          uint             `json:"product_id"`
	ReservedQuantity   int              `json:"reserved_quantity"`
	AvailableQuantity  int              `json:"available_quantity"`
	TotalQuantity      int              `json:"total_quantity"`
	Reference          string           `json:"reference"`
	Status             ReservationStatus `json:"status"`
	ReservationTime    string           `json:"reservation_time"`
}

// ReservationLogResponse represents a single reservation log entry in the history
type ReservationLogResponse struct {
	Quantity    int               `json:"quantity"`
	Status      ReservationStatus `json:"status"`
	Reference   string            `json:"reference"`
	CreatedAt   string            `json:"created_at"`
}

// ReservationHistoryResponse represents a response to a reservation history request
type ReservationHistoryResponse struct {
	WarehouseID uint                    `json:"warehouse_id"`
	ProductID   uint                    `json:"product_id"`
	Total       int64                   `json:"total"`
	Page        int                     `json:"page"`
	Limit       int                     `json:"limit"`
	Logs        []ReservationLogResponse `json:"logs"`
}
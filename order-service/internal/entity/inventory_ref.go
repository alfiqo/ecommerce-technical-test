package entity

import (
	"time"
)

// Inventory is a reference type for interacting with the warehouse service
// It is not stored locally but provides a common structure for code using inventory data
type Inventory struct {
	ID              uint      `json:"id"`
	ProductID       uint      `json:"product_id"`
	WarehouseID     uint      `json:"warehouse_id"`
	Quantity        int       `json:"quantity"`
	ReservedQuantity int      `json:"reserved_quantity"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AvailableQuantity returns the quantity available for reservation
func (i *Inventory) AvailableQuantity() int {
	return i.Quantity - i.ReservedQuantity
}
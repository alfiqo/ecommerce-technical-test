package entity

import (
	"time"
)

// WarehouseStock represents a warehouse stock entity
type WarehouseStock struct {
	ID               uint      `gorm:"column:id;primaryKey;autoIncrement"`
	WarehouseID      uint      `gorm:"column:warehouse_id;not null;index:idx_warehouse_product,unique"`
	ProductID        uint      `gorm:"column:product_id;not null;index:idx_warehouse_product,unique"`
	Quantity         int       `gorm:"column:quantity;default:0;not null"`
	ReservedQuantity int       `gorm:"column:reserved_quantity;default:0;not null"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
	
	// Virtual field (not stored in database)
	AvailableQuantity int `gorm:"-"`
	
	// Relationships
	Warehouse         Warehouse `gorm:"foreignKey:WarehouseID"`
}

func (ws *WarehouseStock) TableName() string {
	return "warehouse_stock"
}

// CalculateAvailableQuantity sets the AvailableQuantity field based on Quantity and ReservedQuantity
func (ws *WarehouseStock) CalculateAvailableQuantity() {
	ws.AvailableQuantity = ws.Quantity - ws.ReservedQuantity
	if ws.AvailableQuantity < 0 {
		ws.AvailableQuantity = 0
	}
}
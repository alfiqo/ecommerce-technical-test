package entity

import (
	"time"
)

// MovementType represents the type of stock movement
type MovementType string

const (
	MovementTypeStockIn    MovementType = "stock_in"
	MovementTypeStockOut   MovementType = "stock_out"
	MovementTypeTransferIn MovementType = "transfer_in"
	MovementTypeTransferOut MovementType = "transfer_out"
)

// StockMovement represents a record of stock quantity changes
type StockMovement struct {
	ID            uint         `gorm:"column:id;primaryKey;autoIncrement"`
	WarehouseID   uint         `gorm:"column:warehouse_id;not null;index:idx_warehouse_product"`
	ProductID     uint         `gorm:"column:product_id;not null;index:idx_warehouse_product"` // References external product service
	ProductSKU    string       `gorm:"column:product_sku;type:varchar(100);not null"`         // Store SKU for reference
	MovementType  MovementType `gorm:"column:movement_type;type:enum('stock_in','stock_out','transfer_in','transfer_out');not null"`
	Quantity      int          `gorm:"column:quantity;not null"`
	ReferenceType string       `gorm:"column:reference_type;type:varchar(50)"`
	ReferenceID   string       `gorm:"column:reference_id;type:varchar(100)"`
	Notes         string       `gorm:"column:notes;type:text"`
	CreatedAt     time.Time    `gorm:"column:created_at;autoCreateTime"`
	
	// Relationships
	Warehouse     Warehouse    `gorm:"foreignKey:WarehouseID"`
}

func (sm *StockMovement) TableName() string {
	return "stock_movements"
}
package entity

import (
	"time"
)

// TransferStatus represents the status of a stock transfer
type TransferStatus string

const (
	StatusPending   TransferStatus = "pending"
	StatusCompleted TransferStatus = "completed"
	StatusFailed    TransferStatus = "failed"
)

// StockTransfer represents a stock transfer between warehouses
type StockTransfer struct {
	ID                uint           `gorm:"column:id;primaryKey;autoIncrement"`
	SourceWarehouseID uint           `gorm:"column:source_warehouse_id;not null"`
	TargetWarehouseID uint           `gorm:"column:target_warehouse_id;not null"`
	ProductID         uint           `gorm:"column:product_id;not null"`
	Quantity          int            `gorm:"column:quantity;not null"`
	Status            TransferStatus `gorm:"column:status;type:enum('pending','completed','failed');default:pending;not null"`
	TransferReference string         `gorm:"column:transfer_reference;type:varchar(50);not null;index"`
	CreatedAt         time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	
	// Relationships
	SourceWarehouse   Warehouse      `gorm:"foreignKey:SourceWarehouseID"`
	TargetWarehouse   Warehouse      `gorm:"foreignKey:TargetWarehouseID"`
}

func (st *StockTransfer) TableName() string {
	return "stock_transfers"
}
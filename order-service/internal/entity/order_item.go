package entity

import (
	"time"

	"gorm.io/gorm"
)

// OrderItem represents an item within an order
type OrderItem struct {
	ID          uint      `gorm:"column:id;primaryKey;autoIncrement"`
	OrderID     uint      `gorm:"column:order_id;not null;index:idx_order_id"`
	ProductID   uint      `gorm:"column:product_id;not null;index:idx_product_id"`
	WarehouseID uint      `gorm:"column:warehouse_id;not null;index:idx_warehouse_id"`
	Quantity    int       `gorm:"column:quantity;not null"`
	UnitPrice   float64   `gorm:"column:unit_price;type:decimal(10,2);not null"`
	TotalPrice  float64   `gorm:"column:total_price;type:decimal(10,2);not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	Order       *Order    `gorm:"foreignKey:OrderID"`
}

func (oi *OrderItem) TableName() string {
	return "order_items"
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	oi.CreatedAt = time.Now()
	oi.UpdatedAt = time.Now()
	return
}
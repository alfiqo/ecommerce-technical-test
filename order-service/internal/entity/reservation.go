package entity

import (
	"time"

	"gorm.io/gorm"
)

// Reservation represents a stock reservation for an order
type Reservation struct {
	ID          uint      `gorm:"column:id;primaryKey;autoIncrement"`
	OrderID     uint      `gorm:"column:order_id;not null;index:idx_order_id"`
	ProductID   uint      `gorm:"column:product_id;not null;index:idx_product_id"`
	WarehouseID uint      `gorm:"column:warehouse_id;not null;index:idx_warehouse_id"`
	Quantity    int       `gorm:"column:quantity;not null"`
	ExpiresAt   time.Time `gorm:"column:expires_at;not null;index:idx_expires_at"`
	IsActive    bool      `gorm:"column:is_active;default:true;index:idx_is_active"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	Order       *Order    `gorm:"foreignKey:OrderID"`
}

func (r *Reservation) TableName() string {
	return "stock_reservations"
}

func (r *Reservation) BeforeCreate(tx *gorm.DB) (err error) {
	r.CreatedAt = time.Now()
	return
}
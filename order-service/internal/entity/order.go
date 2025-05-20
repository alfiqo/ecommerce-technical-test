package entity

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatus represents the possible states of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusCompleted OrderStatus = "completed"
)

// Order represents an order entity
type Order struct {
	ID              uint         `gorm:"column:id;primaryKey;autoIncrement"`
	UserID          string       `gorm:"column:user_id;type:char(36);not null;index:idx_user_id"`
	Status          OrderStatus  `gorm:"column:status;type:enum('pending','paid','cancelled','completed');default:pending;index:idx_status"`
	TotalAmount     float64      `gorm:"column:total_amount;type:decimal(10,2);not null"`
	ShippingAddress string       `gorm:"column:shipping_address;type:text;not null"`
	PaymentMethod   string       `gorm:"column:payment_method;type:varchar(50);not null"`
	PaymentDeadline time.Time    `gorm:"column:payment_deadline;not null"`
	CreatedAt       time.Time    `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time    `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	OrderItems      []OrderItem  `gorm:"foreignKey:OrderID"`
	Reservations    []Reservation `gorm:"foreignKey:OrderID"`
}

func (o *Order) TableName() string {
	return "orders"
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	return
}
package entity

import (
	"time"
)

// ReservationStatus represents the status of a reservation
type ReservationStatus string

const (
	// ReservationStatusPending represents a pending reservation
	ReservationStatusPending ReservationStatus = "pending"
	
	// ReservationStatusCommitted represents a committed reservation (stock removed)
	ReservationStatusCommitted ReservationStatus = "committed"
	
	// ReservationStatusCancelled represents a cancelled reservation
	ReservationStatusCancelled ReservationStatus = "cancelled"
)

// ReservationLog represents a log entry for stock reservations
type ReservationLog struct {
	ID          uint             `gorm:"column:id;primaryKey;autoIncrement"`
	WarehouseID uint             `gorm:"column:warehouse_id;not null;index"`
	ProductID   uint             `gorm:"column:product_id;not null;index"`
	Quantity    int              `gorm:"column:quantity;not null"`
	Status      string           `gorm:"column:status;type:enum('pending','committed','cancelled');default:pending;not null"`
	Reference   string           `gorm:"column:reference;type:varchar(100)"`
	CreatedAt   time.Time        `gorm:"column:created_at;autoCreateTime"`
	
	// Relationships
	Warehouse   Warehouse        `gorm:"foreignKey:WarehouseID"`
}

func (r *ReservationLog) TableName() string {
	return "reservation_logs"
}
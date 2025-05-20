package entity

import (
	"time"
)

// Warehouse represents a warehouse entity
type Warehouse struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name;type:varchar(255);not null"`
	Location  string    `gorm:"column:location;type:varchar(255);not null"`
	Address   string    `gorm:"column:address;type:varchar(500);not null"`
	IsActive  bool      `gorm:"column:is_active;default:true;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
}

func (w *Warehouse) TableName() string {
	return "warehouses"
}
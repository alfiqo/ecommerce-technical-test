package entity

import (
	"time"
)

// Shop represents a shop entity in the database
type Shop struct {
	ID           uint           `gorm:"primaryKey;column:id"`
	Name         string         `gorm:"column:name;type:varchar(255);index:,unique;not null"`
	Description  string         `gorm:"column:description;type:text"`
	Address      string         `gorm:"column:address;type:text;not null"`
	ContactEmail string         `gorm:"column:contact_email;type:varchar(255);not null"`
	ContactPhone string         `gorm:"column:contact_phone;type:varchar(50);not null"`
	IsActive     bool           `gorm:"column:is_active;default:true;not null;index"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime;not null"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime;not null"`
	Warehouses   []ShopWarehouse `gorm:"foreignKey:ShopID"`
}

// TableName returns the table name for the Shop entity
func (Shop) TableName() string {
	return "shops"
}

// ShopWarehouse represents a junction table between Shop and Warehouse
type ShopWarehouse struct {
	ID          uint      `gorm:"primaryKey;column:id"`
	ShopID      uint      `gorm:"column:shop_id;not null"`
	WarehouseID uint      `gorm:"column:warehouse_id;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;not null"`
}

// TableName returns the table name for the ShopWarehouse entity
func (ShopWarehouse) TableName() string {
	return "shop_warehouses"
}
package model

import (
	"time"
)

// WarehouseResponse holds the data returned for a warehouse from the warehouse service
// @Description Information about a warehouse
type WarehouseResponse struct {
	ID          uint      `json:"id" example:"42"`
	Name        string    `json:"name" example:"Main Distribution Center"`
	Address     string    `json:"address" example:"789 Warehouse Blvd, Springfield, IL 62701"`
	Capacity    int       `json:"capacity" example:"5000"`
	IsActive    bool      `json:"is_active" example:"true"`
	CreatedAt   time.Time `json:"created_at" example:"2025-05-01T08:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2025-05-02T10:30:00Z"`
}

// ShopWarehousesResponse represents a collection of warehouses for a shop
// @Description Collection of warehouses associated with a shop
type ShopWarehousesResponse struct {
	ShopID     uint                `json:"shop_id" example:"1"`
	Warehouses []WarehouseResponse `json:"warehouses"`
}
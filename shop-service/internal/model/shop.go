package model

import (
	"time"
)

// CreateShopRequest holds the data needed to create a new shop
// @Description Request to create a new shop
type CreateShopRequest struct {
	Name         string `json:"name" validate:"required,min=3,max=255" example:"Downtown Bookstore"`
	Description  string `json:"description" validate:"omitempty" example:"Our flagship bookstore location"`
	Address      string `json:"address" validate:"required" example:"123 Main St, New York, NY 10001"`
	ContactEmail string `json:"contact_email" validate:"required,email" example:"contact@example.com"`
	ContactPhone string `json:"contact_phone" validate:"required" example:"+1-555-123-4567"`
	IsActive     bool   `json:"is_active" example:"true"`
}

// UpdateShopRequest holds the data needed to update a shop
// @Description Request to update an existing shop
type UpdateShopRequest struct {
	Name         string `json:"name" validate:"omitempty,min=3,max=255" example:"Updated Bookstore"`
	Description  string `json:"description" validate:"omitempty" example:"Our rebranded bookstore location"`
	Address      string `json:"address" validate:"omitempty" example:"456 Broadway, New York, NY 10012"`
	ContactEmail string `json:"contact_email" validate:"omitempty,email" example:"new-contact@example.com"`
	ContactPhone string `json:"contact_phone" validate:"omitempty" example:"+1-555-987-6543"`
	IsActive     *bool  `json:"is_active" validate:"omitempty" example:"false"`
}

// ShopResponse holds the data returned for a shop
// @Description Response containing shop information
type ShopResponse struct {
	ID           uint      `json:"id" example:"1"`
	Name         string    `json:"name" example:"Downtown Bookstore"`
	Description  string    `json:"description" example:"Our flagship bookstore location"`
	Address      string    `json:"address" example:"123 Main St, New York, NY 10001"`
	ContactEmail string    `json:"contact_email" example:"contact@example.com"`
	ContactPhone string    `json:"contact_phone" example:"+1-555-123-4567"`
	IsActive     bool      `json:"is_active" example:"true"`
	CreatedAt    time.Time `json:"created_at" example:"2025-05-18T08:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2025-05-18T08:00:00Z"`
}

// WarehouseID holds the ID of a warehouse
// @Description Reference to a warehouse by ID
type WarehouseID struct {
	ID uint `json:"id" example:"42"`
}

// ShopDetailResponse includes warehouses IDs associated with the shop
// @Description Detailed shop information including associated warehouses
type ShopDetailResponse struct {
	ShopResponse
	WarehouseIDs []WarehouseID `json:"warehouse_ids,omitempty"`
}

// AssignWarehouseRequest holds the data needed to assign a warehouse to a shop
// @Description Request to assign a warehouse to a shop
type AssignWarehouseRequest struct {
	WarehouseID uint `json:"warehouse_id" validate:"required" example:"42"`
}

// ShopListResponse holds a list of shops for pagination
// @Description Paginated list of shops
type ShopListResponse struct {
	Shops      []ShopResponse `json:"shops"`
	TotalCount int64          `json:"total_count" example:"42"`
	Page       int            `json:"page" example:"1"`
	PageSize   int            `json:"page_size" example:"10"`
}
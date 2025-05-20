package model

type GetWarehouseRequest struct {
	ID uint `json:"id" validate:"required"`
}

type CreateWarehouseRequest struct {
	Name     string `json:"name" validate:"required,max=255"`
	Location string `json:"location" validate:"required,max=255"`
	Address  string `json:"address" validate:"required,max=500"`
	IsActive bool   `json:"is_active"`
}

type UpdateWarehouseRequest struct {
	ID       uint   `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required,max=255"`
	Location string `json:"location" validate:"required,max=255"`
	Address  string `json:"address" validate:"required,max=500"`
	IsActive bool   `json:"is_active"`
}

type ListWarehouseRequest struct {
	Page  int `json:"page" validate:"min=1"`
	Limit int `json:"limit" validate:"min=1,max=100"`
}

type WarehouseStatsDTO struct {
	TotalProducts int64 `json:"total_products"`
	TotalItems    int64 `json:"total_items"`
}

type WarehouseResponse struct {
	ID        uint             `json:"id,omitempty"`
	Name      string           `json:"name,omitempty"`
	Location  string           `json:"location,omitempty"`
	Address   string           `json:"address,omitempty"`
	IsActive  bool             `json:"is_active,omitempty"`
	Stats     *WarehouseStatsDTO `json:"stats,omitempty"`
	CreatedAt string           `json:"created_at,omitempty"`
	UpdatedAt string           `json:"updated_at,omitempty"`
}

type WarehouseListResponse struct {
	Warehouses []WarehouseResponse `json:"warehouses"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
}
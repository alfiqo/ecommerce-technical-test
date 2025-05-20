package model

type ProductResponse struct {
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price,omitempty"`
	Stock       int     `json:"stock,omitempty"`
	Category    string  `json:"category,omitempty"`
	SKU         string  `json:"sku,omitempty"`
	ImageURL    string  `json:"image_url,omitempty"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
}

type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
	Count    int64            `json:"count"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"min=0"`
	Category    string  `json:"category" validate:"max=100"`
	SKU         string  `json:"sku" validate:"max=50"`
	ImageURL    string  `json:"image_url" validate:"max=255"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name" validate:"max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"gt=0"`
	Stock       int     `json:"stock" validate:"min=0"`
	Category    string  `json:"category" validate:"max=100"`
	SKU         string  `json:"sku" validate:"max=50"`
	ImageURL    string  `json:"image_url" validate:"max=255"`
}
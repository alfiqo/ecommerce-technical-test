package model

// AddStockRequest represents a request to add stock to a warehouse
type AddStockRequest struct {
	WarehouseID uint   `json:"warehouse_id" validate:"required"`
	ProductID   uint   `json:"product_id" validate:"required"`
	ProductSKU  string `json:"product_sku" validate:"required"`
	Quantity    int    `json:"quantity" validate:"required,gt=0"`
	Reference   string `json:"reference" validate:"required"`
	Notes       string `json:"notes"`
}

// StockResponse represents a response to a stock operation
type StockResponse struct {
	WarehouseID       uint   `json:"warehouse_id"`
	ProductID         uint   `json:"product_id"`
	ProductName       string `json:"product_name,omitempty"`
	SKU               string `json:"sku,omitempty"`
	Quantity          int    `json:"quantity"`
	ReservedQuantity  int    `json:"reserved_quantity"`
	AvailableQuantity int    `json:"available_quantity"`
	UpdatedAt         string `json:"updated_at"`
}

// StockItemResponse represents a single stock item in a list
type StockItemResponse struct {
	WarehouseID       uint   `json:"warehouse_id"`
	ProductID         uint   `json:"product_id"`
	ProductName       string `json:"product_name,omitempty"`
	SKU               string `json:"sku,omitempty"`
	Quantity          int    `json:"quantity"`
	ReservedQuantity  int    `json:"reserved_quantity"`
	AvailableQuantity int    `json:"available_quantity"`
	UpdatedAt         string `json:"updated_at"`
}

// WarehouseStockListResponse represents a paginated list of stock items
type WarehouseStockListResponse struct {
	WarehouseID uint               `json:"warehouse_id"`
	Total       int64              `json:"total"`
	Page        int                `json:"page"`
	Limit       int                `json:"limit"`
	TotalPages  int64              `json:"total_pages"`
	Items       []StockItemResponse `json:"items"`
}

// StockTransferRequest represents a request to transfer stock between warehouses
type StockTransferRequest struct {
	SourceWarehouseID uint   `json:"source_warehouse_id" validate:"required"`
	TargetWarehouseID uint   `json:"target_warehouse_id" validate:"required"`
	ProductID         uint   `json:"product_id" validate:"required"`
	ProductSKU        string `json:"product_sku" validate:"required"`
	Quantity          int    `json:"quantity" validate:"required,gt=0"`
	Reference         string `json:"reference"`
	Notes             string `json:"notes"`
}

// StockTransferResponse represents a response to a stock transfer
type StockTransferResponse struct {
	TransferID        uint   `json:"transfer_id"`
	SourceWarehouseID uint   `json:"source_warehouse_id"`
	TargetWarehouseID uint   `json:"target_warehouse_id"`
	ProductID         uint   `json:"product_id"`
	Quantity          int    `json:"quantity"`
	Status            string `json:"status"`
	TransferReference string `json:"transfer_reference"`
	CreatedAt         string `json:"created_at"`
}
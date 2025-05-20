package warehouse

import (
	"context"
	"errors"
	"fmt"
	"order-service/internal/entity"
	"order-service/internal/model"

	"github.com/sirupsen/logrus"
)

var (
	// ErrConnectionFailed is returned when we can't connect to the warehouse service
	ErrConnectionFailed = errors.New("failed to connect to warehouse service")

	// ErrInsufficientStock is returned when the warehouse service reports insufficient stock
	ErrInsufficientStock = errors.New("insufficient stock available")

	// ErrReservationNotFound is returned when a requested reservation is not found
	ErrReservationNotFound = errors.New("reservation not found")
)

// WarehouseGateway implements the WarehouseGatewayInterface
type WarehouseGateway struct {
	Client *Client
	Log    *logrus.Logger
}

// NewWarehouseGateway creates a new warehouse gateway
func NewWarehouseGateway(client *Client, log *logrus.Logger) *WarehouseGateway {
	return &WarehouseGateway{
		Client: client,
		Log:    log,
	}
}

// CheckAndReserveStock checks and reserves stock for multiple items
func (g *WarehouseGateway) CheckAndReserveStock(ctx context.Context, orderID uint, items []model.OrderItemRequest, reserveUntil string) (*ReservationResponse, error) {
	// Parse reserveUntil string into time.Time if provided, otherwise use 24h from now
	// var reserveUntilTime time.Time
	var err error

	// if reserveUntil != "" {
	// 	reserveUntilTime, err = time.Parse(time.RFC3339, reserveUntil)
	// 	if err != nil {
	// 		g.Log.Warnf("Invalid reserveUntil format: %s, defaulting to 24h", reserveUntil)
	// 		reserveUntilTime = time.Now().Add(24 * time.Hour)
	// 	}
	// } else {
	// 	reserveUntilTime = time.Now().Add(24 * time.Hour)
	// }

	// Prepare request
	orderItems := make([]ReservationOrderItem, len(items))
	for i, item := range items {
		orderItems[i] = ReservationOrderItem{
			ProductID:   item.ProductID,
			WarehouseID: item.WarehouseID,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		}
	}

	request := ReserveStockRequest{
		WarehouseID: items[0].WarehouseID,
		ProductID:   items[0].ProductID,
		Quantity:    items[0].Quantity,
	}

	// Make API call
	var response ReservationResponse
	err = g.Client.doRequest(ctx, "POST", "/api/v1/inventory/reserve", request, &response)
	if err != nil {
		g.Log.Errorf("Failed to reserve stock: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	if !response.Success {
		g.Log.Warnf("Stock reservation failed: %s", response.Message)
		return &response, ErrInsufficientStock
	}

	return &response, nil
}

// ConfirmStockDeduction commits reserved stock as sold (after payment)
func (g *WarehouseGateway) ConfirmStockDeduction(ctx context.Context, orderID uint, reservationID string) (*StockOperationResponse, error) {
	request := ReservationConfirmRequest{
		ReservationID: reservationID,
		OrderID:       orderID,
	}

	var response StockOperationResponse
	err := g.Client.doRequest(ctx, "POST", "/api/v1/inventory/commit", request, &response)
	if err != nil {
		g.Log.Errorf("Failed to confirm stock deduction: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	if !response.Success {
		g.Log.Warnf("Stock confirmation failed: %s", response.Message)
		if response.Message == "reservation not found" {
			return &response, ErrReservationNotFound
		}
		return &response, errors.New(response.Message)
	}

	return &response, nil
}

// ReleaseReservation releases stock back to available inventory (e.g., cancelled order)
func (g *WarehouseGateway) ReleaseReservation(ctx context.Context, orderID uint, reservation ReservationReleaseRequest) (*StockOperationResponse, error) {
	// Use the provided reservation data directly
	var response StockOperationResponse
	err := g.Client.doRequest(ctx, "POST", "/api/v1/inventory/reserve/cancel", reservation, &response)
	if err != nil {
		g.Log.Errorf("Failed to release stock reservation: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	if !response.Success {
		g.Log.Warnf("Stock release failed: %s", response.Message)
		if response.Message == "reservation not found" {
			return &response, ErrReservationNotFound
		}
		return &response, errors.New(response.Message)
	}

	return &response, nil
}

// GetInventory gets current inventory level for a product
func (g *WarehouseGateway) GetInventory(ctx context.Context, productID, warehouseID uint) (*InventoryResponse, error) {
	request := InventoryQueryRequest{
		ProductID:   productID,
		WarehouseID: warehouseID,
	}

	var response InventoryResponse
	err := g.Client.doRequest(ctx, "POST", "/api/v1/inventory/get", request, &response)
	if err != nil {
		g.Log.Errorf("Failed to get inventory: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	return &response, nil
}

// GetInventoryBatch gets current inventory levels for multiple products
func (g *WarehouseGateway) GetInventoryBatch(ctx context.Context, items []InventoryQuery) (map[string]*InventoryResponse, error) {
	// Prepare request
	queryItems := make([]InventoryQueryRequest, len(items))
	for i, item := range items {
		queryItems[i] = InventoryQueryRequest{
			ProductID:   item.ProductID,
			WarehouseID: item.WarehouseID,
		}
	}

	request := InventoryBatchQueryRequest{
		Items: queryItems,
	}

	// Make API call
	var response map[string]InventoryResponse
	err := g.Client.doRequest(ctx, "POST", "/api/v1/inventory/batch", request, &response)
	if err != nil {
		g.Log.Errorf("Failed to get inventory batch: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	// Convert response to expected return type
	result := make(map[string]*InventoryResponse)
	for key, value := range response {
		valueCopy := value // Create a copy to avoid referencing loop variable
		result[key] = &valueCopy
	}

	return result, nil
}

// UpdateInventory updates inventory quantity (admin operation)
func (g *WarehouseGateway) UpdateInventory(ctx context.Context, inventory *entity.Inventory) (*StockOperationResponse, error) {
	// Convert entity to appropriate request type
	request := InventoryResponse{
		ProductID:        inventory.ProductID,
		WarehouseID:      inventory.WarehouseID,
		Quantity:         inventory.Quantity,
		ReservedQuantity: inventory.ReservedQuantity,
	}

	var response StockOperationResponse
	err := g.Client.doRequest(ctx, "POST", "/api/v1/inventory/update", request, &response)
	if err != nil {
		g.Log.Errorf("Failed to update inventory: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	if !response.Success {
		g.Log.Warnf("Inventory update failed: %s", response.Message)
		return &response, errors.New(response.Message)
	}

	return &response, nil
}

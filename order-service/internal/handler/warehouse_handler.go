package handler

import (
	"context"
	"order-service/internal/delivery/http/response"
	"order-service/internal/errors"
	"order-service/internal/gateway/warehouse"
	"order-service/internal/model"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// WarehouseHandler handles HTTP requests related to warehouse inventory operations
type WarehouseHandler struct {
	log              *logrus.Logger
	warehouseGateway warehouse.WarehouseGatewayInterface
}

// NewWarehouseHandler creates a new handler for warehouse operations
func NewWarehouseHandler(log *logrus.Logger, warehouseGateway warehouse.WarehouseGatewayInterface) *WarehouseHandler {
	return &WarehouseHandler{
		log:              log,
		warehouseGateway: warehouseGateway,
	}
}

// GetInventory godoc
// @Summary Get inventory for a product
// @Description Returns current inventory level for a specific product at a warehouse
// @Tags Inventory
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Param warehouse_id path int true "Warehouse ID"
// @Success 200 {object} model.InventoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /inventory/{product_id}/{warehouse_id} [get]
func (h *WarehouseHandler) GetInventory(c *fiber.Ctx) error {
	productID, err := strconv.ParseUint(c.Params("product_id"), 10, 64)
	if err != nil {
		h.log.Warnf("Invalid product ID: %v", err)
		return response.JSONError(c, errors.ErrInvalidInput, h.log)
	}

	warehouseID, err := strconv.ParseUint(c.Params("warehouse_id"), 10, 64)
	if err != nil {
		h.log.Warnf("Invalid warehouse ID: %v", err)
		return response.JSONError(c, errors.ErrInvalidInput, h.log)
	}

	ctx := context.Background()
	inventoryResponse, err := h.warehouseGateway.GetInventory(ctx, uint(productID), uint(warehouseID))
	if err != nil {
		h.log.Warnf("Failed to get inventory: %v", err)
		return response.JSONError(c, errors.ErrInternalServer, h.log)
	}

	if inventoryResponse == nil {
		return response.JSONError(c, errors.ErrResourceNotFound, h.log)
	}

	// Map the gateway response to the model response
	result := model.InventoryResponse{
		ProductID:        inventoryResponse.ProductID,
		WarehouseID:      inventoryResponse.WarehouseID,
		Quantity:         inventoryResponse.Quantity,
		ReservedQuantity: inventoryResponse.ReservedQuantity,
		AvailableQuantity: inventoryResponse.Quantity - inventoryResponse.ReservedQuantity,
	}

	return response.JSONSuccess(c, result)
}

// GetInventoryBatch godoc
// @Summary Get inventory for multiple products
// @Description Returns current inventory levels for multiple products across warehouses
// @Tags Inventory
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body model.InventoryBatchRequest true "Inventory query request"
// @Success 200 {object} model.InventoryBatchResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /inventory/batch [post]
func (h *WarehouseHandler) GetInventoryBatch(c *fiber.Ctx) error {
	var request model.InventoryBatchRequest
	if err := c.BodyParser(&request); err != nil {
		h.log.Warnf("Invalid request body: %v", err)
		return response.JSONError(c, errors.ErrInvalidInput, h.log)
	}

	// Convert model request to gateway request
	var items []warehouse.InventoryQuery
	for _, item := range request.Items {
		items = append(items, warehouse.InventoryQuery{
			ProductID:   item.ProductID,
			WarehouseID: item.WarehouseID,
		})
	}

	ctx := context.Background()
	inventoryResponses, err := h.warehouseGateway.GetInventoryBatch(ctx, items)
	if err != nil {
		h.log.Warnf("Failed to get inventory batch: %v", err)
		return response.JSONError(c, errors.ErrInternalServer, h.log)
	}

	// Map the gateway responses to the model responses
	result := model.InventoryBatchResponse{
		Items: make(map[string]model.InventoryResponse),
	}

	for key, item := range inventoryResponses {
		result.Items[key] = model.InventoryResponse{
			ProductID:        item.ProductID,
			WarehouseID:      item.WarehouseID,
			Quantity:         item.Quantity,
			ReservedQuantity: item.ReservedQuantity,
			AvailableQuantity: item.Quantity - item.ReservedQuantity,
		}
	}

	return response.JSONSuccess(c, result)
}

// ReserveStock godoc
// @Summary Reserve stock for an order
// @Description Creates stock reservations for items in an order
// @Tags Inventory
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body model.StockReservationRequest true "Stock reservation request"
// @Success 200 {object} model.StockReservationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /inventory/reserve [post]
func (h *WarehouseHandler) ReserveStock(c *fiber.Ctx) error {
	var request model.StockReservationRequest
	if err := c.BodyParser(&request); err != nil {
		h.log.Warnf("Invalid request body: %v", err)
		return response.JSONError(c, errors.ErrInvalidInput, h.log)
	}

	// Default expiration time to 24 hours from now if not provided
	reserveUntil := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	if request.ReserveUntil != "" {
		reserveUntil = request.ReserveUntil
	}

	ctx := context.Background()
	reservationResp, err := h.warehouseGateway.CheckAndReserveStock(ctx, request.OrderID, request.Items, reserveUntil)
	if err != nil {
		h.log.Warnf("Failed to reserve stock: %v", err)
		return response.JSONError(c, errors.ErrInternalServer, h.log)
	}

	// Map the gateway response to the model response
	var responseItems []model.StockReservationItem
	for _, item := range reservationResp.Items {
		responseItems = append(responseItems, model.StockReservationItem{
			ProductID:   item.ProductID,
			WarehouseID: item.WarehouseID,
			Quantity:    item.Quantity,
			Available:   item.Available,
			Message:     item.Message,
		})
	}

	result := model.StockReservationResponse{
		ReservationID: reservationResp.ReservationID,
		OrderID:       reservationResp.OrderID,
		Items:         responseItems,
		Success:       reservationResp.Success,
		Message:       reservationResp.Message,
		ReservedUntil: reservationResp.ReservedUntil.Format(time.RFC3339),
	}

	return response.JSONSuccess(c, result)
}

// ConfirmStockDeduction godoc
// @Summary Confirm stock deduction for an order
// @Description Confirms the deduction of reserved stock for an order (typically after payment)
// @Tags Inventory
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body model.StockOperationRequest true "Stock operation request"
// @Success 200 {object} model.StockOperationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /inventory/confirm [post]
func (h *WarehouseHandler) ConfirmStockDeduction(c *fiber.Ctx) error {
	var request model.StockOperationRequest
	if err := c.BodyParser(&request); err != nil {
		h.log.Warnf("Invalid request body: %v", err)
		return response.JSONError(c, errors.ErrInvalidInput, h.log)
	}

	ctx := context.Background()
	opResponse, err := h.warehouseGateway.ConfirmStockDeduction(ctx, request.OrderID, request.ReservationID)
	if err != nil {
		h.log.Warnf("Failed to confirm stock deduction: %v", err)
		return response.JSONError(c, errors.ErrInternalServer, h.log)
	}

	result := model.StockOperationResponse{
		Success: opResponse.Success,
		Message: opResponse.Message,
	}

	return response.JSONSuccess(c, result)
}

// ReleaseReservation godoc
// @Summary Release a stock reservation
// @Description Releases a previously created stock reservation, making the stock available again
// @Tags Inventory
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body model.StockOperationRequest true "Stock operation request"
// @Success 200 {object} model.StockOperationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /inventory/release [post]
func (h *WarehouseHandler) ReleaseReservation(c *fiber.Ctx) error {
	var request model.StockOperationRequest
	if err := c.BodyParser(&request); err != nil {
		h.log.Warnf("Invalid request body: %v", err)
		return response.JSONError(c, errors.ErrInvalidInput, h.log)
	}

	ctx := context.Background()
	
	// Create the release request with the data from the API request
	releaseRequest := warehouse.ReservationReleaseRequest{
		WarehouseID: request.WarehouseID,
		ProductID:   request.ProductID,
		Quantity:    request.Quantity,
		Reference:   request.ReservationID,
	}
	
	opResponse, err := h.warehouseGateway.ReleaseReservation(ctx, request.OrderID, releaseRequest)
	if err != nil {
		h.log.Warnf("Failed to release reservation: %v", err)
		return response.JSONError(c, errors.ErrInternalServer, h.log)
	}

	result := model.StockOperationResponse{
		Success: opResponse.Success,
		Message: opResponse.Message,
	}

	return response.JSONSuccess(c, result)
}
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	appErrors "shop-service/internal/errors"
	"shop-service/internal/model"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShopHandler_GetShopWarehouses_Success(t *testing.T) {
	// Setup
	mockShopUsecase, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Get("/api/v1/shops/:id/warehouses", handler.GetShopWarehouses)
	
	// Mock data
	shopID := uint(1)
	now := time.Now()
	
	mockWarehousesResponse := &model.ShopWarehousesResponse{
		ShopID: shopID,
		Warehouses: []model.WarehouseResponse{
			{
				ID:        101,
				Name:      "Warehouse 1",
				Address:   "123 Storage St",
				Capacity:  1000,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:        102,
				Name:      "Warehouse 2",
				Address:   "456 Depot Rd",
				Capacity:  2000,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
	
	// Set mock expectations with any context parameter
	mockShopUsecase.On("GetShopWarehouses", mock.Anything, shopID).
		Return(mockWarehousesResponse, nil)
	
	// Create a test request
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/shops/%d/warehouses", shopID), nil)
	assert.NoError(t, err)
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	var responseBody struct {
		Success bool                        `json:"success"`
		Data    model.ShopWarehousesResponse `json:"data"`
	}
	
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	err = json.Unmarshal(body, &responseBody)
	assert.NoError(t, err)
	
	// Assert response content
	assert.True(t, responseBody.Success)
	assert.Equal(t, shopID, responseBody.Data.ShopID)
	assert.Equal(t, 2, len(responseBody.Data.Warehouses))
	assert.Equal(t, uint(101), responseBody.Data.Warehouses[0].ID)
	assert.Equal(t, uint(102), responseBody.Data.Warehouses[1].ID)
	
	// Verify expectations
	mockShopUsecase.AssertExpectations(t)
}

func TestShopHandler_GetShopWarehouses_InvalidID(t *testing.T) {
	// Setup
	mockShopUsecase, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Get("/api/v1/shops/:id/warehouses", handler.GetShopWarehouses)
	
	// Create a test request with invalid ID
	req, err := http.NewRequest("GET", "/api/v1/shops/invalid-id/warehouses", nil)
	assert.NoError(t, err)
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// Parse response body
	var responseBody struct {
		Success bool         `json:"success"`
		Error   *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	err = json.Unmarshal(body, &responseBody)
	assert.NoError(t, err)
	
	// Assert error response
	assert.False(t, responseBody.Success)
	assert.NotNil(t, responseBody.Error)
	assert.Equal(t, "INVALID_INPUT", responseBody.Error.Code)
	
	// Verify no expectations called on usecase
	mockShopUsecase.AssertNotCalled(t, "GetShopWarehouses")
}

func TestShopHandler_GetShopWarehouses_NotFound(t *testing.T) {
	// Setup
	mockShopUsecase, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Get("/api/v1/shops/:id/warehouses", handler.GetShopWarehouses)
	
	// Mock error
	shopID := uint(999)
	mockError := appErrors.ErrShopNotFound
	
	// Set mock expectations with any context parameter
	mockShopUsecase.On("GetShopWarehouses", mock.Anything, shopID).
		Return(nil, mockError)
	
	// Create a test request
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/shops/%d/warehouses", shopID), nil)
	assert.NoError(t, err)
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	
	// Parse response body
	var responseBody struct {
		Success bool         `json:"success"`
		Error   *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	err = json.Unmarshal(body, &responseBody)
	assert.NoError(t, err)
	
	// Assert error response
	assert.False(t, responseBody.Success)
	assert.NotNil(t, responseBody.Error)
	assert.Equal(t, "SHOP_NOT_FOUND", responseBody.Error.Code)
	
	// Verify expectations
	mockShopUsecase.AssertExpectations(t)
}

func TestShopHandler_GetShopWarehouses_ServiceError(t *testing.T) {
	// Setup
	mockShopUsecase, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Get("/api/v1/shops/:id/warehouses", handler.GetShopWarehouses)
	
	// Mock error
	shopID := uint(1)
	mockError := appErrors.ErrExternalServiceUnavailable
	
	// Set mock expectations with any context parameter
	mockShopUsecase.On("GetShopWarehouses", mock.Anything, shopID).
		Return(nil, mockError)
	
	// Create a test request
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/shops/%d/warehouses", shopID), nil)
	assert.NoError(t, err)
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	
	// Parse response body
	var responseBody struct {
		Success bool         `json:"success"`
		Error   *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	err = json.Unmarshal(body, &responseBody)
	assert.NoError(t, err)
	
	// Assert error response
	assert.False(t, responseBody.Success)
	assert.NotNil(t, responseBody.Error)
	assert.Equal(t, "EXTERNAL_SERVICE_UNAVAILABLE", responseBody.Error.Code)
	
	// Verify expectations
	mockShopUsecase.AssertExpectations(t)
}
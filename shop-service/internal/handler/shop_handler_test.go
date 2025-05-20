package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	appErrors "shop-service/internal/errors"
	"shop-service/internal/entity"
	"shop-service/internal/model"
	"shop-service/internal/model/converter"
	mockUsecase "shop-service/mocks/usecase"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupShopHandlerTest(t *testing.T) (*mockUsecase.ShopUsecaseMock, *ShopHandler, *fiber.App) {
	// Create mock usecase
	mockShopUsecase := new(mockUsecase.ShopUsecaseMock)
	
	// Create logger
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Properly disable logging for tests
	
	// Create handler
	handler := NewShopHandler(mockShopUsecase, logger)
	
	// Create fiber app for testing
	app := fiber.New()
	
	return mockShopUsecase, handler, app
}

func TestShopHandler_ListShops_Success(t *testing.T) {
	// Setup
	mockShopUsecase, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Get("/api/v1/shops", handler.ListShops)
	
	// Mock data
	now := time.Now()
	mockShops := []entity.Shop{
		{
			ID:           1,
			Name:         "Shop 1",
			Description:  "Description 1",
			Address:      "Address 1",
			ContactEmail: "shop1@example.com",
			ContactPhone: "1234567890",
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           2,
			Name:         "Shop 2",
			Description:  "Description 2",
			Address:      "Address 2",
			ContactEmail: "shop2@example.com",
			ContactPhone: "0987654321",
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	mockTotalCount := int64(2)
	
	// Create expected response
	_ = converter.ToShopListResponse(mockShops, mockTotalCount, 1, 10)
	
	// Set mock expectations
	mockShopUsecase.On("ListShops", 1, 10, "", false).
		Return(mockShops, mockTotalCount, nil)
	
	// Create a test request
	req, err := http.NewRequest("GET", "/api/v1/shops", nil)
	assert.NoError(t, err)
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	var responseBody struct {
		Success bool                   `json:"success"`
		Data    model.ShopListResponse `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(t, err)
	
	// Assert response content
	assert.True(t, responseBody.Success)
	assert.Equal(t, 2, len(responseBody.Data.Shops))
	assert.Equal(t, int64(2), responseBody.Data.TotalCount)
	assert.Equal(t, 1, responseBody.Data.Page)
	assert.Equal(t, 10, responseBody.Data.PageSize)
	
	// Verify expectations
	mockShopUsecase.AssertExpectations(t)
}

func TestShopHandler_ListShops_WithQueryParams(t *testing.T) {
	// Setup
	mockShopUsecase, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Get("/api/v1/shops", handler.ListShops)
	
	// Mock data
	now := time.Now()
	mockShops := []entity.Shop{
		{
			ID:           1,
			Name:         "Market",
			Description:  "Supermarket",
			Address:      "Address 1",
			ContactEmail: "market@example.com",
			ContactPhone: "1234567890",
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	mockTotalCount := int64(1)
	
	// Expected parameters from query
	page := 2
	pageSize := 5
	searchTerm := "Market"
	includeInactive := true
	
	// Set mock expectations
	mockShopUsecase.On("ListShops", page, pageSize, searchTerm, includeInactive).
		Return(mockShops, mockTotalCount, nil)
	
	// Create a test request with query parameters
	req, err := http.NewRequest("GET", "/api/v1/shops?page=2&page_size=5&search=Market&include_inactive=true", nil)
	assert.NoError(t, err)
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Verify expectations
	mockShopUsecase.AssertExpectations(t)
}

func TestShopHandler_ListShops_Error(t *testing.T) {
	// Setup
	mockShopUsecase, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Get("/api/v1/shops", handler.ListShops)
	
	// Mock error
	mockError := errors.New("database error")
	
	// Set mock expectations
	mockShopUsecase.On("ListShops", 1, 10, "", false).
		Return([]entity.Shop{}, int64(0), mockError)
	
	// Create a test request
	req, err := http.NewRequest("GET", "/api/v1/shops", nil)
	assert.NoError(t, err)
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code - should be error
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	
	// Verify expectations
	mockShopUsecase.AssertExpectations(t)
}

func TestShopHandler_GetShopByID_Success(t *testing.T) {
	// Setup
	mockShopUsecase, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Get("/api/v1/shops/:id", handler.GetShopByID)
	
	// Mock data
	now := time.Now()
	shopID := uint(1)
	mockShop := &entity.Shop{
		ID:           shopID,
		Name:         "Shop 1",
		Description:  "Description 1",
		Address:      "Address 1",
		ContactEmail: "shop1@example.com",
		ContactPhone: "1234567890",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Warehouses: []entity.ShopWarehouse{
			{
				ID:          1,
				ShopID:      shopID,
				WarehouseID: 101,
				CreatedAt:   now,
			},
			{
				ID:          2,
				ShopID:      shopID,
				WarehouseID: 102,
				CreatedAt:   now,
			},
		},
	}
	
	// Set mock expectations
	mockShopUsecase.On("GetShopWithWarehouses", shopID).Return(mockShop, nil)
	
	// Create a test request
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/shops/%d", shopID), nil)
	assert.NoError(t, err)
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	var responseBody struct {
		Success bool                     `json:"success"`
		Data    model.ShopDetailResponse `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(t, err)
	
	// Assert response content
	assert.True(t, responseBody.Success)
	assert.Equal(t, uint(1), responseBody.Data.ID)
	assert.Equal(t, "Shop 1", responseBody.Data.Name)
	assert.Equal(t, 2, len(responseBody.Data.WarehouseIDs))
	
	// Verify expectations
	mockShopUsecase.AssertExpectations(t)
}

func TestShopHandler_GetShopByID_InvalidID(t *testing.T) {
	// Setup
	_, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Get("/api/v1/shops/:id", handler.GetShopByID)
	
	// Create a test request with invalid ID
	req, err := http.NewRequest("GET", "/api/v1/shops/invalid", nil)
	assert.NoError(t, err)
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code - should be bad request
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// No need to verify expectations as the handler should return early
}

func TestShopHandler_GetShopByID_NotFound(t *testing.T) {
	// Setup
	mockShopUsecase, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Get("/api/v1/shops/:id", handler.GetShopByID)
	
	// Mock error
	shopID := uint(999)
	mockError := appErrors.ErrShopNotFound
	
	// Set mock expectations
	mockShopUsecase.On("GetShopWithWarehouses", shopID).Return(nil, mockError)
	
	// Create a test request
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/shops/%d", shopID), nil)
	assert.NoError(t, err)
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code - should be 404 Not Found
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	
	// Verify expectations
	mockShopUsecase.AssertExpectations(t)
}

func TestShopHandler_CreateShop_Success(t *testing.T) {
	// Setup
	mockShopUsecase, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Post("/api/v1/shops", handler.CreateShop)
	
	// Mock request data
	createRequest := model.CreateShopRequest{
		Name:         "New Shop",
		Description:  "A new shop for testing",
		Address:      "123 Test St",
		ContactEmail: "test@example.com",
		ContactPhone: "1234567890",
		IsActive:     true,
	}
	
	requestBody, err := json.Marshal(createRequest)
	assert.NoError(t, err)
	
	// Mock data - the created shop
	now := time.Now()
	mockShop := &entity.Shop{
		ID:           1,
		Name:         createRequest.Name,
		Description:  createRequest.Description,
		Address:      createRequest.Address,
		ContactEmail: createRequest.ContactEmail,
		ContactPhone: createRequest.ContactPhone,
		IsActive:     createRequest.IsActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	
	// Set mock expectations
	mockShopUsecase.On("CreateShop", &createRequest).Return(mockShop, nil)
	
	// Create a test request
	req, err := http.NewRequest("POST", "/api/v1/shops", bytes.NewBuffer(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code - should be 201 Created
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	
	// Parse response body
	var responseBody struct {
		Success bool               `json:"success"`
		Data    model.ShopResponse `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(t, err)
	
	// Assert response content
	assert.True(t, responseBody.Success)
	assert.Equal(t, uint(1), responseBody.Data.ID)
	assert.Equal(t, "New Shop", responseBody.Data.Name)
	assert.Equal(t, "A new shop for testing", responseBody.Data.Description)
	assert.Equal(t, "123 Test St", responseBody.Data.Address)
	assert.Equal(t, "test@example.com", responseBody.Data.ContactEmail)
	assert.Equal(t, "1234567890", responseBody.Data.ContactPhone)
	assert.True(t, responseBody.Data.IsActive)
	
	// Verify expectations
	mockShopUsecase.AssertExpectations(t)
}

func TestShopHandler_CreateShop_InvalidRequest(t *testing.T) {
	// Setup
	mockShopUsecase, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Post("/api/v1/shops", handler.CreateShop)
	
	// Invalid request - missing required fields
	invalidRequest := `{"name": "Invalid Shop"}`
	
	// Mock validation error
	mockShopUsecase.On("CreateShop", mock.Anything).Return(nil, appErrors.WithError(appErrors.ErrInvalidInput, errors.New("validation error")))
	
	// Create a test request with invalid body
	req, err := http.NewRequest("POST", "/api/v1/shops", bytes.NewBuffer([]byte(invalidRequest)))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code - should be 400 Bad Request
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// Verify expectations
	mockShopUsecase.AssertExpectations(t)
}

func TestShopHandler_CreateShop_UsecaseError(t *testing.T) {
	// Setup
	mockShopUsecase, handler, app := setupShopHandlerTest(t)
	
	// Register the route
	app.Post("/api/v1/shops", handler.CreateShop)
	
	// Mock request data
	createRequest := model.CreateShopRequest{
		Name:         "New Shop",
		Description:  "A new shop for testing",
		Address:      "123 Test St",
		ContactEmail: "test@example.com",
		ContactPhone: "1234567890",
		IsActive:     true,
	}
	
	requestBody, err := json.Marshal(createRequest)
	assert.NoError(t, err)
	
	// Mock error
	mockError := appErrors.WithError(appErrors.ErrInternalServer, errors.New("database error"))
	
	// Set mock expectations
	mockShopUsecase.On("CreateShop", &createRequest).Return(nil, mockError)
	
	// Create a test request
	req, err := http.NewRequest("POST", "/api/v1/shops", bytes.NewBuffer(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	
	// Perform the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Assert status code - should be 500 Internal Server Error
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	
	// Verify expectations
	mockShopUsecase.AssertExpectations(t)
}
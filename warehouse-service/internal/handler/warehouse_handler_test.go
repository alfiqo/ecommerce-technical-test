package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"warehouse-service/internal/entity"
	"warehouse-service/internal/model"
	"warehouse-service/mocks/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupWarehouseHandlerTest(t *testing.T) (*WarehouseHandler, *usecase.MockWarehouseUseCaseInterface, *fiber.App) {
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Suppress log output during tests
	
	ctrl := gomock.NewController(t)
	mockUsecase := usecase.NewMockWarehouseUseCaseInterface(ctrl)
	
	handler := &WarehouseHandler{
		UseCase: mockUsecase,
		Log:     logger,
	}
	
	app := fiber.New()
	
	return handler, mockUsecase, app
}

func TestWarehouseHandler_GetWarehouse(t *testing.T) {
	handler, mockUsecase, app := setupWarehouseHandlerTest(t)
	
	// Setup route
	app.Get("/api/v1/warehouses/:id", handler.GetWarehouse)
	
	// Create test warehouse
	warehouseID := uint(1)
	warehouse := &entity.Warehouse{
		ID:       warehouseID,
		Name:     "Test Warehouse",
		Location: "Test Location",
		Address:  "Test Address",
		IsActive: true,
	}
	
	// Setup statistics
	productCount := int64(10)
	totalItemCount := int64(100)
	
	// Setup mock expectations
	warehouseResponse := &model.WarehouseResponse{
		ID:        warehouseID,
		Name:      warehouse.Name,
		Location:  warehouse.Location,
		Address:   warehouse.Address,
		IsActive:  warehouse.IsActive,
		Stats:     &model.WarehouseStatsDTO{
			TotalProducts: productCount,
			TotalItems:    totalItemCount,
		},
	}
	mockUsecase.EXPECT().GetWarehouse(gomock.Any(), warehouseID).Return(warehouseResponse, nil)
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/warehouses/1", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin_token_here")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Check response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	// Verify response data
	assert.Equal(t, float64(warehouseID), result["data"].(map[string]interface{})["id"])
	assert.Equal(t, warehouse.Name, result["data"].(map[string]interface{})["name"])
	assert.Equal(t, warehouse.Location, result["data"].(map[string]interface{})["location"])
	assert.Equal(t, warehouse.Address, result["data"].(map[string]interface{})["address"])
	assert.Equal(t, warehouse.IsActive, result["data"].(map[string]interface{})["is_active"])
	
	stats := result["data"].(map[string]interface{})["stats"].(map[string]interface{})
	assert.Equal(t, float64(productCount), stats["total_products"])
	assert.Equal(t, float64(totalItemCount), stats["total_items"])
}

func TestWarehouseHandler_GetWarehouse_InvalidID(t *testing.T) {
	handler, _, app := setupWarehouseHandlerTest(t)
	
	// Setup route
	app.Get("/api/v1/warehouses/:id", handler.GetWarehouse)
	
	// Create request with invalid ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/warehouses/invalid", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin_token_here")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	
	// Check response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	// Verify error structure
	assert.Equal(t, false, result["success"])
	assert.NotNil(t, result["error"])
}

func TestWarehouseHandler_GetWarehouse_NotFound(t *testing.T) {
	handler, mockUsecase, app := setupWarehouseHandlerTest(t)
	
	// Setup route
	app.Get("/api/v1/warehouses/:id", handler.GetWarehouse)
	
	warehouseID := uint(999)
	
	// Setup mock expectations for not found case
	mockUsecase.EXPECT().GetWarehouse(gomock.Any(), warehouseID).Return(nil, fiber.ErrNotFound)
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/warehouses/999", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin_token_here")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	
	// Check response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	// Verify error structure 
	assert.Equal(t, false, result["success"])
	assert.NotNil(t, result["error"])
}

func TestWarehouseHandler_GetWarehouse_InternalError(t *testing.T) {
	handler, mockUsecase, app := setupWarehouseHandlerTest(t)
	
	// Setup route
	app.Get("/api/v1/warehouses/:id", handler.GetWarehouse)
	
	warehouseID := uint(1)
	
	// Setup mock expectations for internal error case
	mockUsecase.EXPECT().GetWarehouse(gomock.Any(), warehouseID).Return(nil, errors.New("database error"))
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/warehouses/1", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin_token_here")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	
	// Check response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	// Verify error structure
	assert.Equal(t, false, result["success"])
	assert.NotNil(t, result["error"])
}

func TestWarehouseHandler_CreateWarehouse(t *testing.T) {
	handler, mockUsecase, app := setupWarehouseHandlerTest(t)
	
	// Setup route
	app.Post("/api/v1/warehouses", handler.CreateWarehouse)
	
	// Create test request data
	createRequest := &model.CreateWarehouseRequest{
		Name:     "New Warehouse",
		Location: "New Location",
		Address:  "New Address",
		IsActive: true,
	}
	
	// Expected response
	warehouseID := uint(1)
	warehouseResponse := &model.WarehouseResponse{
		ID:        warehouseID,
		Name:      createRequest.Name,
		Location:  createRequest.Location,
		Address:   createRequest.Address,
		IsActive:  createRequest.IsActive,
		Stats:     &model.WarehouseStatsDTO{
			TotalProducts: 0,
			TotalItems:    0,
		},
	}
	
	// Setup mock expectations
	mockUsecase.EXPECT().CreateWarehouse(gomock.Any(), gomock.Any()).Return(warehouseResponse, nil)
	
	// Create request
	reqBody, _ := json.Marshal(createRequest)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/warehouses", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin_token_here")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Check response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	// Verify response data
	assert.Equal(t, float64(warehouseID), result["data"].(map[string]interface{})["id"])
	assert.Equal(t, createRequest.Name, result["data"].(map[string]interface{})["name"])
	assert.Equal(t, createRequest.Location, result["data"].(map[string]interface{})["location"])
	assert.Equal(t, createRequest.Address, result["data"].(map[string]interface{})["address"])
	assert.Equal(t, createRequest.IsActive, result["data"].(map[string]interface{})["is_active"])
}

func TestWarehouseHandler_CreateWarehouse_InvalidBody(t *testing.T) {
	handler, _, app := setupWarehouseHandlerTest(t)
	
	// Setup route
	app.Post("/api/v1/warehouses", handler.CreateWarehouse)
	
	// Create invalid request body
	reqBody := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/warehouses", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin_token_here")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	
	// Check response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	// Verify error structure
	assert.Equal(t, false, result["success"])
	assert.NotNil(t, result["error"])
}

func TestWarehouseHandler_UpdateWarehouse(t *testing.T) {
	handler, mockUsecase, app := setupWarehouseHandlerTest(t)
	
	// Setup route
	app.Put("/api/v1/warehouses/:id", handler.UpdateWarehouse)
	
	// Create test data
	warehouseID := uint(1)
	updateRequest := &model.UpdateWarehouseRequest{
		ID:       warehouseID,
		Name:     "Updated Warehouse",
		Location: "Updated Location",
		Address:  "Updated Address",
		IsActive: true,
	}
	
	// Expected response
	warehouseResponse := &model.WarehouseResponse{
		ID:        warehouseID,
		Name:      updateRequest.Name,
		Location:  updateRequest.Location,
		Address:   updateRequest.Address,
		IsActive:  updateRequest.IsActive,
		Stats:     &model.WarehouseStatsDTO{
			TotalProducts: 5,
			TotalItems:    100,
		},
	}
	
	// Setup mock expectations
	mockUsecase.EXPECT().UpdateWarehouse(gomock.Any(), gomock.Any()).Return(warehouseResponse, nil)
	
	// Create request
	reqBody, _ := json.Marshal(updateRequest)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/warehouses/1", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin_token_here")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Check response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	// Verify response data
	assert.Equal(t, float64(warehouseID), result["data"].(map[string]interface{})["id"])
	assert.Equal(t, updateRequest.Name, result["data"].(map[string]interface{})["name"])
	assert.Equal(t, updateRequest.Location, result["data"].(map[string]interface{})["location"])
	assert.Equal(t, updateRequest.Address, result["data"].(map[string]interface{})["address"])
	assert.Equal(t, updateRequest.IsActive, result["data"].(map[string]interface{})["is_active"])
}

func TestWarehouseHandler_DeleteWarehouse(t *testing.T) {
	handler, mockUsecase, app := setupWarehouseHandlerTest(t)
	
	// Setup route
	app.Delete("/api/v1/warehouses/:id", handler.DeleteWarehouse)
	
	warehouseID := uint(1)
	
	// Setup mock expectations
	mockUsecase.EXPECT().DeleteWarehouse(gomock.Any(), warehouseID).Return(nil)
	
	// Create request
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/warehouses/1", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin_token_here")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Check response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	// Verify success message
	assert.Equal(t, true, result["success"])
	assert.Equal(t, "Warehouse deleted successfully", result["data"].(map[string]interface{})["message"])
}

func TestWarehouseHandler_ListWarehouses(t *testing.T) {
	handler, mockUsecase, app := setupWarehouseHandlerTest(t)
	
	// Setup route
	app.Get("/api/v1/warehouses", handler.ListWarehouses)
	
	// Mock warehouses data
	warehouses := []model.WarehouseResponse{
		{
			ID:        1,
			Name:      "Warehouse 1",
			Location:  "Location 1",
			Address:   "Address 1",
			IsActive:  true,
			Stats:     &model.WarehouseStatsDTO{TotalProducts: 10, TotalItems: 100},
		},
		{
			ID:        2,
			Name:      "Warehouse 2",
			Location:  "Location 2",
			Address:   "Address 2",
			IsActive:  true,
			Stats:     &model.WarehouseStatsDTO{TotalProducts: 20, TotalItems: 200},
		},
	}
	
	total := int64(2)
	page := 1
	limit := 20
	
	listResponse := &model.WarehouseListResponse{
		Warehouses: warehouses,
		Total:      total,
		Page:       page,
		Limit:      limit,
	}
	
	// Setup mock expectations
	mockUsecase.EXPECT().ListWarehouses(gomock.Any(), gomock.Any()).Return(listResponse, nil)
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/warehouses?page=1&limit=20", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin_token_here")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Check response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	// Verify response data
	data := result["data"].(map[string]interface{})
	assert.Equal(t, float64(total), data["total"])
	assert.Equal(t, float64(page), data["page"])
	assert.Equal(t, float64(limit), data["limit"])
	
	// Verify warehouses array
	warehousesData := data["warehouses"].([]interface{})
	assert.Equal(t, 2, len(warehousesData))
	
	// Check first warehouse data
	w1 := warehousesData[0].(map[string]interface{})
	assert.Equal(t, float64(1), w1["id"])
	assert.Equal(t, "Warehouse 1", w1["name"])
}

func TestWarehouseHandler_ListWarehouses_InvalidParams(t *testing.T) {
	handler, _, app := setupWarehouseHandlerTest(t)
	
	// Setup route
	app.Get("/api/v1/warehouses", handler.ListWarehouses)
	
	// Create request with invalid page
	req := httptest.NewRequest(http.MethodGet, "/api/v1/warehouses?page=invalid&limit=20", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin_token_here")
	
	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	
	// Check response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	// Verify error structure
	assert.Equal(t, false, result["success"])
	assert.NotNil(t, result["error"])
}
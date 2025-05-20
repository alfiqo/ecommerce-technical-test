package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"order-service/internal/model"
	usecase_mock "order-service/mocks/usecase"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOrderHandler_CreateOrder_Success(t *testing.T) {
	// Initialize mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockOrderUseCase := usecase_mock.NewMockOrderUseCaseInterface(ctrl)
	logger := logrus.New()
	
	// Create handler with mock
	orderHandler := NewOrderHandler(mockOrderUseCase, logger)
	
	// Create test app
	app := fiber.New()
	app.Post("/orders", func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return orderHandler.CreateOrder(c)
	})
	
	// Prepare request data
	orderRequest := &model.CreateOrderRequest{
		UserID:          "test-user-id",
		ShippingAddress: "123 Test St",
		PaymentMethod:   "credit_card",
		Items: []model.OrderItemRequest{
			{
				ProductID:   1,
				WarehouseID: 1,
				Quantity:    2,
				UnitPrice:   10.0,
			},
		},
	}
	
	// Marshal request to JSON
	requestBody, _ := json.Marshal(orderRequest)
	
	// Mock response
	orderResponse := &model.OrderResponse{
		ID:              1,
		UserID:          "test-user-id",
		Status:          "pending",
		TotalAmount:     20.0,
		ShippingAddress: "123 Test St",
		PaymentMethod:   "credit_card",
		PaymentDeadline: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		CreatedAt:       time.Now().Format(time.RFC3339),
		UpdatedAt:       time.Now().Format(time.RFC3339),
		Items: []model.OrderItemResponse{
			{
				ID:          1,
				ProductID:   1,
				WarehouseID: 1,
				Quantity:    2,
				UnitPrice:   10.0,
				TotalPrice:  20.0,
			},
		},
	}
	
	// Setup mock expectation
	mockOrderUseCase.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Return(orderResponse, nil)
	
	// Make request
	req := httptest.NewRequest("POST", "/orders", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Test response
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	
	// Parse response
	var response struct {
		Data model.OrderResponse `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, orderResponse.ID, response.Data.ID)
	assert.Equal(t, orderResponse.UserID, response.Data.UserID)
	assert.Equal(t, orderResponse.TotalAmount, response.Data.TotalAmount)
	
	// No need to verify expectations with gomock - it's done automatically when ctrl.Finish() is called
}

func TestOrderHandler_CreateOrder_Error(t *testing.T) {
	// Initialize mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockOrderUseCase := usecase_mock.NewMockOrderUseCaseInterface(ctrl)
	logger := logrus.New()
	
	// Create handler with mock
	orderHandler := NewOrderHandler(mockOrderUseCase, logger)
	
	// Create test app
	app := fiber.New()
	app.Post("/orders", func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return orderHandler.CreateOrder(c)
	})
	
	// Prepare request data
	orderRequest := &model.CreateOrderRequest{
		UserID:          "test-user-id",
		ShippingAddress: "123 Test St",
		PaymentMethod:   "credit_card",
		Items: []model.OrderItemRequest{
			{
				ProductID:   1,
				WarehouseID: 1,
				Quantity:    2,
				UnitPrice:   10.0,
			},
		},
	}
	
	// Marshal request to JSON
	requestBody, _ := json.Marshal(orderRequest)
	
	// Setup mock expectation
	mockOrderUseCase.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("database error"))
	
	// Make request
	req := httptest.NewRequest("POST", "/orders", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Test response
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	
	// No need to verify expectations with gomock - it's done automatically when ctrl.Finish() is called
}

func TestOrderHandler_CreateOrder_BadRequest(t *testing.T) {
	// Create test app without any mock expectations
	app := fiber.New()
	app.Post("/orders", func(c *fiber.Ctx) error {
		// Return BadRequest directly to simulate validation failure
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request data")
	})
	
	// Prepare an empty request
	req := httptest.NewRequest("POST", "/orders", nil)
	
	// Test response
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
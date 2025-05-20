package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"product-service/internal/delivery/http/middleware"
	"product-service/internal/delivery/http/response"
	appErrors "product-service/internal/errors"
	"product-service/internal/model"
	mockUsecase "product-service/mocks/usecase"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProductHandlerTestSuite struct {
	suite.Suite
	app                *fiber.App
	mockProductUseCase *mockUsecase.MockProductUseCase
	productHandler     *ProductHandler
	logger             *logrus.Logger
}

func (suite *ProductHandlerTestSuite) SetupTest() {
	// Setup logger
	suite.logger = logrus.New()
	
	// Setup Fiber app
	suite.app = fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler(suite.logger),
	})
	
	// Setup middleware
	suite.app.Use(middleware.RequestID())
	
	// Setup mock usecase
	suite.mockProductUseCase = new(mockUsecase.MockProductUseCase)
	
	// Setup handler
	suite.productHandler = NewProductHandler(suite.mockProductUseCase, suite.logger)
	
	// Setup routes
	api := suite.app.Group("/api")
	v1 := api.Group("/v1")
	products := v1.Group("/products")
	products.Get("/", suite.productHandler.GetProducts)
	products.Post("/", suite.productHandler.CreateProduct)
	// Special routes that could be matched by the :id parameter need to be defined first
	products.Get("/search", suite.productHandler.SearchProducts)
	products.Get("/category/:category", suite.productHandler.GetProductsByCategory)
	// Generic parameter routes come after specific routes
	products.Get("/:id", suite.productHandler.GetProductByID)
	products.Put("/:id", suite.productHandler.UpdateProduct)
	products.Delete("/:id", suite.productHandler.DeleteProduct)
}

func (suite *ProductHandlerTestSuite) TestGetProducts() {
	t := suite.T()
	
	// Setup mock data
	mockProductResponse := &model.ProductListResponse{
		Products: []model.ProductResponse{
			{
				ID:          "f47ac10b-58cc-4372-a567-0e02b2c3d479",
				Name:        "Test Product 1",
				Description: "Test Description 1",
				Price:       99.99,
				Stock:       10,
				Category:    "Test Category",
				SKU:         "TEST-SKU-001",
			},
			{
				ID:          "f47ac10b-58cc-4372-a567-0e02b2c3d480",
				Name:        "Test Product 2",
				Description: "Test Description 2",
				Price:       149.99,
				Stock:       20,
				Category:    "Another Category",
				SKU:         "TEST-SKU-002",
			},
		},
		Count:  2,
		Limit:  10,
		Offset: 0,
	}
	
	// Setup expectations
	suite.mockProductUseCase.On("GetProducts", mock.Anything, 10, 0).Return(mockProductResponse, nil)
	
	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	resp, err := suite.app.Test(req)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	// Unmarshal response
	var apiResponse response.Response
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)
	
	// Validate response data
	assert.True(t, apiResponse.Success)
	assert.Nil(t, apiResponse.Error)
	
	// Extract and verify the data
	dataJSON, err := json.Marshal(apiResponse.Data)
	assert.NoError(t, err)
	
	var productListResp model.ProductListResponse
	err = json.Unmarshal(dataJSON, &productListResp)
	assert.NoError(t, err)
	
	assert.Equal(t, 2, len(productListResp.Products))
	assert.Equal(t, "Test Product 1", productListResp.Products[0].Name)
	assert.Equal(t, float64(99.99), productListResp.Products[0].Price)
	
	// Verify expectations
	suite.mockProductUseCase.AssertExpectations(t)
}

func (suite *ProductHandlerTestSuite) TestGetProductByID() {
	t := suite.T()
	
	// Setup mock data
	mockProductID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	mockProductResponse := &model.ProductResponse{
		ID:          mockProductID,
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Stock:       10,
		Category:    "Test Category",
		SKU:         "TEST-SKU-001",
	}
	
	// Setup expectations
	suite.mockProductUseCase.On("GetProductByID", mock.Anything, mockProductID).Return(mockProductResponse, nil)
	
	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/"+mockProductID, nil)
	resp, err := suite.app.Test(req)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	// Unmarshal response
	var apiResponse response.Response
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)
	
	// Validate response data
	assert.True(t, apiResponse.Success)
	assert.Nil(t, apiResponse.Error)
	
	// Extract and verify the data
	dataJSON, err := json.Marshal(apiResponse.Data)
	assert.NoError(t, err)
	
	var productResp model.ProductResponse
	err = json.Unmarshal(dataJSON, &productResp)
	assert.NoError(t, err)
	
	assert.Equal(t, mockProductID, productResp.ID)
	assert.Equal(t, "Test Product", productResp.Name)
	assert.Equal(t, float64(99.99), productResp.Price)
	
	// Verify expectations
	suite.mockProductUseCase.AssertExpectations(t)
}

func (suite *ProductHandlerTestSuite) TestGetProductByID_NotFound() {
	t := suite.T()
	
	// Setup mock data
	mockProductID := "non-existent-id"
	
	// Setup expectations
	suite.mockProductUseCase.On("GetProductByID", mock.Anything, mockProductID).Return(nil, appErrors.ErrProductNotFound)
	
	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/"+mockProductID, nil)
	resp, err := suite.app.Test(req)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	
	// Parse response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	// Unmarshal response
	var apiResponse response.Response
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)
	
	// Validate response data
	assert.False(t, apiResponse.Success)
	assert.NotNil(t, apiResponse.Error)
	assert.Equal(t, "PRODUCT_NOT_FOUND", apiResponse.Error.Code)
	
	// Verify expectations
	suite.mockProductUseCase.AssertExpectations(t)
}

func (suite *ProductHandlerTestSuite) TestCreateProduct() {
	t := suite.T()
	
	// Setup mock data
	mockCreateRequest := model.CreateProductRequest{
		Name:        "New Test Product",
		Description: "New Test Description",
		Price:       149.99,
		Stock:       15,
		Category:    "New Category",
		SKU:         "NEW-SKU-001",
	}
	
	mockProductResponse := &model.ProductResponse{
		ID:          "f47ac10b-58cc-4372-a567-0e02b2c3d481",
		Name:        "New Test Product",
		Description: "New Test Description",
		Price:       149.99,
		Stock:       15,
		Category:    "New Category",
		SKU:         "NEW-SKU-001",
	}
	
	// Setup expectations
	suite.mockProductUseCase.On("CreateProduct", mock.Anything, mock.MatchedBy(func(req *model.CreateProductRequest) bool {
		return req.Name == mockCreateRequest.Name && req.SKU == mockCreateRequest.SKU
	})).Return(mockProductResponse, nil)
	
	// Create request body
	reqBody, _ := json.Marshal(mockCreateRequest)
	
	// Create request
	req := httptest.NewRequest("POST", "/api/v1/products", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Test the request
	resp, err := suite.app.Test(req)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	// Unmarshal response
	var apiResponse response.Response
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)
	
	// Validate response data
	assert.True(t, apiResponse.Success)
	assert.Nil(t, apiResponse.Error)
	
	// Extract and verify the data
	dataJSON, err := json.Marshal(apiResponse.Data)
	assert.NoError(t, err)
	
	var productResp model.ProductResponse
	err = json.Unmarshal(dataJSON, &productResp)
	assert.NoError(t, err)
	
	assert.Equal(t, mockProductResponse.ID, productResp.ID)
	assert.Equal(t, mockCreateRequest.Name, productResp.Name)
	assert.Equal(t, mockCreateRequest.Price, productResp.Price)
	
	// Verify expectations
	suite.mockProductUseCase.AssertExpectations(t)
}

func (suite *ProductHandlerTestSuite) TestCreateProduct_InvalidInput() {
	t := suite.T()
	
	// Setup invalid mock data (missing required fields)
	mockInvalidRequest := map[string]interface{}{
		"name":     "", // Name is required
		"price":    -10, // Price should be positive
		"category": "Test Category",
	}
	
	// Mock behavior for invalid request - the handler should validate input before calling usecase
	createReq := &model.CreateProductRequest{
		Name:     "",
		Price:    -10,
		Category: "Test Category",
	}
	suite.mockProductUseCase.On("CreateProduct", mock.Anything, mock.MatchedBy(func(req *model.CreateProductRequest) bool {
		// This should not be called, but add a fallback just in case
		return req.Name == createReq.Name && req.Price == createReq.Price && req.Category == createReq.Category
	})).Return(nil, appErrors.WithMessage(appErrors.ErrInvalidInput, "Invalid input"))
	
	// Create request body
	reqBody, _ := json.Marshal(mockInvalidRequest)
	
	// Create request
	req := httptest.NewRequest("POST", "/api/v1/products", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Test the request
	resp, err := suite.app.Test(req)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// Parse response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	// Unmarshal response
	var apiResponse response.Response
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)
	
	// Validate response data
	assert.False(t, apiResponse.Success)
	assert.NotNil(t, apiResponse.Error)
	assert.Equal(t, "INVALID_INPUT", apiResponse.Error.Code)
}

func (suite *ProductHandlerTestSuite) TestUpdateProduct() {
	t := suite.T()
	
	// Setup mock data
	mockProductID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	mockUpdateRequest := model.UpdateProductRequest{
		Name:        "Updated Product",
		Description: "Updated Description",
		Price:       199.99,
	}
	
	mockProductResponse := &model.ProductResponse{
		ID:          mockProductID,
		Name:        "Updated Product",
		Description: "Updated Description",
		Price:       199.99,
		Stock:       10,
		Category:    "Test Category",
		SKU:         "TEST-SKU-001",
	}
	
	// Setup expectations
	suite.mockProductUseCase.On("UpdateProduct", mock.Anything, mockProductID, mock.MatchedBy(func(req *model.UpdateProductRequest) bool {
		return req.Name == mockUpdateRequest.Name && req.Price == mockUpdateRequest.Price
	})).Return(mockProductResponse, nil)
	
	// Create request body
	reqBody, _ := json.Marshal(mockUpdateRequest)
	
	// Create request
	req := httptest.NewRequest("PUT", "/api/v1/products/"+mockProductID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Test the request
	resp, err := suite.app.Test(req)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	// Unmarshal response
	var apiResponse response.Response
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)
	
	// Validate response data
	assert.True(t, apiResponse.Success)
	assert.Nil(t, apiResponse.Error)
	
	// Extract and verify the data
	dataJSON, err := json.Marshal(apiResponse.Data)
	assert.NoError(t, err)
	
	var productResp model.ProductResponse
	err = json.Unmarshal(dataJSON, &productResp)
	assert.NoError(t, err)
	
	assert.Equal(t, mockProductID, productResp.ID)
	assert.Equal(t, mockUpdateRequest.Name, productResp.Name)
	assert.Equal(t, mockUpdateRequest.Price, productResp.Price)
	
	// Verify expectations
	suite.mockProductUseCase.AssertExpectations(t)
}

func (suite *ProductHandlerTestSuite) TestDeleteProduct() {
	t := suite.T()
	
	// Setup mock data
	mockProductID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	
	// Setup expectations
	suite.mockProductUseCase.On("DeleteProduct", mock.Anything, mockProductID).Return(nil)
	
	// Create request
	req := httptest.NewRequest("DELETE", "/api/v1/products/"+mockProductID, nil)
	resp, err := suite.app.Test(req)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	
	// Verify expectations
	suite.mockProductUseCase.AssertExpectations(t)
}

func (suite *ProductHandlerTestSuite) TestSearchProducts() {
	t := suite.T()
	
	// Setup mock data
	searchQuery := "test"
	mockProductResponse := &model.ProductListResponse{
		Products: []model.ProductResponse{
			{
				ID:          "f47ac10b-58cc-4372-a567-0e02b2c3d479",
				Name:        "Test Product 1",
				Description: "Test Description 1",
				Price:       99.99,
				Stock:       10,
				Category:    "Test Category",
				SKU:         "TEST-SKU-001",
			},
		},
		Count:  1,
		Limit:  10,
		Offset: 0,
	}
	
	// Setup expectations
	suite.mockProductUseCase.On("SearchProducts", mock.Anything, searchQuery, 10, 0).Return(mockProductResponse, nil)
	
	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/search?q="+searchQuery, nil)
	resp, err := suite.app.Test(req)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	// Unmarshal response
	var apiResponse response.Response
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)
	
	// Validate response data
	assert.True(t, apiResponse.Success)
	assert.Nil(t, apiResponse.Error)
	
	// Extract and verify the data
	dataJSON, err := json.Marshal(apiResponse.Data)
	assert.NoError(t, err)
	
	var productListResp model.ProductListResponse
	err = json.Unmarshal(dataJSON, &productListResp)
	assert.NoError(t, err)
	
	assert.Equal(t, 1, len(productListResp.Products))
	assert.Equal(t, "Test Product 1", productListResp.Products[0].Name)
	
	// Verify expectations
	suite.mockProductUseCase.AssertExpectations(t)
}

func (suite *ProductHandlerTestSuite) TestGetProductsByCategory() {
	t := suite.T()
	
	// Setup mock data
	category := "electronics"
	mockProductResponse := &model.ProductListResponse{
		Products: []model.ProductResponse{
			{
				ID:          "f47ac10b-58cc-4372-a567-0e02b2c3d479",
				Name:        "iPhone",
				Description: "Apple Smartphone",
				Price:       999.99,
				Stock:       10,
				Category:    "electronics",
				SKU:         "IPHONE-001",
			},
			{
				ID:          "f47ac10b-58cc-4372-a567-0e02b2c3d480",
				Name:        "Samsung Galaxy",
				Description: "Android Smartphone",
				Price:       899.99,
				Stock:       15,
				Category:    "electronics",
				SKU:         "SAMSUNG-001",
			},
		},
		Count:  2,
		Limit:  10,
		Offset: 0,
	}
	
	// Setup expectations
	suite.mockProductUseCase.On("GetProductsByCategory", mock.Anything, category, 10, 0).Return(mockProductResponse, nil)
	
	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products/category/"+category, nil)
	resp, err := suite.app.Test(req)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	
	// Parse response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	// Unmarshal response
	var apiResponse response.Response
	err = json.Unmarshal(body, &apiResponse)
	assert.NoError(t, err)
	
	// Validate response data
	assert.True(t, apiResponse.Success)
	assert.Nil(t, apiResponse.Error)
	
	// Extract and verify the data
	dataJSON, err := json.Marshal(apiResponse.Data)
	assert.NoError(t, err)
	
	var productListResp model.ProductListResponse
	err = json.Unmarshal(dataJSON, &productListResp)
	assert.NoError(t, err)
	
	assert.Equal(t, 2, len(productListResp.Products))
	assert.Equal(t, "iPhone", productListResp.Products[0].Name)
	assert.Equal(t, "Samsung Galaxy", productListResp.Products[1].Name)
	
	// Verify expectations
	suite.mockProductUseCase.AssertExpectations(t)
}

func TestProductHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProductHandlerTestSuite))
}
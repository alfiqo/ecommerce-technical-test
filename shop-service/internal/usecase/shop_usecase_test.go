package usecase

import (
	"context"
	"errors"
	"io"
	"shop-service/internal/entity"
	"shop-service/internal/model"
	"shop-service/mocks/gateway"
	repoMocks "shop-service/mocks/repository_mocks"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupShopUsecaseTest(t *testing.T) (*gorm.DB, *logrus.Logger, *validator.Validate, *repoMocks.ShopRepositoryMock, *repoMocks.ShopWarehouseRepositoryMock, *gateway.WarehouseGatewayMock, ShopUsecaseInterface) {
	// Create mock DB
	db, err := gorm.Open(nil, &gorm.Config{})
	assert.NoError(t, err)
	
	// Create mock logger
	logger := logrus.New()
	// Better way to disable logger in tests
	logger.SetOutput(io.Discard)
	
	// Create validator
	validate := validator.New()
	
	// Create mock repositories and gateway
	mockShopRepo := new(repoMocks.ShopRepositoryMock)
	mockShopWarehouseRepo := new(repoMocks.ShopWarehouseRepositoryMock)
	mockWarehouseGateway := new(gateway.WarehouseGatewayMock)
	
	// Create usecase
	usecase := NewShopUsecase(
		db, 
		logger, 
		validate, 
		mockShopRepo, 
		mockShopWarehouseRepo, 
		mockWarehouseGateway,
	)
	
	return db, logger, validate, mockShopRepo, mockShopWarehouseRepo, mockWarehouseGateway, usecase
}

func TestShopUsecase_ListShops_Success(t *testing.T) {
	// Setup
	db, _, _, mockShopRepo, _, _, usecase := setupShopUsecaseTest(t)
	
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
	
	// Test parameters
	page := 1
	pageSize := 10
	searchTerm := ""
	includeInactive := false
	
	// Set expectations
	mockShopRepo.On("FindAll", db, page, pageSize, searchTerm, includeInactive).
		Return(mockShops, mockTotalCount, nil)
	
	// Execute
	shops, totalCount, err := usecase.ListShops(page, pageSize, searchTerm, includeInactive)
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, mockTotalCount, totalCount)
	assert.Equal(t, len(mockShops), len(shops))
	assert.Equal(t, mockShops[0].ID, shops[0].ID)
	assert.Equal(t, mockShops[1].ID, shops[1].ID)
	
	// Verify expectations
	mockShopRepo.AssertExpectations(t)
}

func TestShopUsecase_ListShops_DatabaseError(t *testing.T) {
	// Setup
	db, _, _, mockShopRepo, _, _, usecase := setupShopUsecaseTest(t)
	
	// Test parameters
	page := 1
	pageSize := 10
	searchTerm := ""
	includeInactive := false
	
	// Mock error
	mockError := errors.New("database error")
	
	// Set expectations
	mockShopRepo.On("FindAll", db, page, pageSize, searchTerm, includeInactive).
		Return([]entity.Shop{}, int64(0), mockError)
	
	// Execute
	shops, totalCount, err := usecase.ListShops(page, pageSize, searchTerm, includeInactive)
	
	// Assertions
	assert.Error(t, err)
	assert.Equal(t, mockError, err)
	assert.Equal(t, int64(0), totalCount)
	assert.Empty(t, shops)
	
	// Verify expectations
	mockShopRepo.AssertExpectations(t)
}

func TestShopUsecase_ListShops_InvalidPage(t *testing.T) {
	// Setup
	db, _, _, mockShopRepo, _, _, usecase := setupShopUsecaseTest(t)
	
	// Mock data
	mockShops := []entity.Shop{}
	mockTotalCount := int64(0)
	
	// Test parameters with invalid page
	page := 0  // Invalid
	normalizedPage := 1  // Should be normalized to 1
	pageSize := 10
	searchTerm := ""
	includeInactive := false
	
	// Set expectations - should call with normalized page
	mockShopRepo.On("FindAll", db, normalizedPage, pageSize, searchTerm, includeInactive).
		Return(mockShops, mockTotalCount, nil)
	
	// Execute
	shops, totalCount, err := usecase.ListShops(page, pageSize, searchTerm, includeInactive)
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, mockTotalCount, totalCount)
	assert.Empty(t, shops)
	
	// Verify expectations
	mockShopRepo.AssertExpectations(t)
}

func TestShopUsecase_ListShops_InvalidPageSize(t *testing.T) {
	// Setup
	db, _, _, mockShopRepo, _, _, usecase := setupShopUsecaseTest(t)
	
	// Mock data
	mockShops := []entity.Shop{}
	mockTotalCount := int64(0)
	
	// Test parameters with invalid page size
	page := 1
	pageSize := 0  // Invalid
	normalizedPageSize := 10  // Should be normalized to 10
	searchTerm := ""
	includeInactive := false
	
	// Set expectations - should call with normalized page size
	mockShopRepo.On("FindAll", db, page, normalizedPageSize, searchTerm, includeInactive).
		Return(mockShops, mockTotalCount, nil)
	
	// Execute
	shops, totalCount, err := usecase.ListShops(page, pageSize, searchTerm, includeInactive)
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, mockTotalCount, totalCount)
	assert.Empty(t, shops)
	
	// Verify expectations
	mockShopRepo.AssertExpectations(t)
}

func TestShopUsecase_GetShopByID_Success(t *testing.T) {
	// Setup
	db, _, _, mockShopRepo, _, _, usecase := setupShopUsecaseTest(t)
	
	// Mock data
	now := time.Now()
	mockShop := &entity.Shop{
		ID:           1,
		Name:         "Shop 1",
		Description:  "Description 1",
		Address:      "Address 1",
		ContactEmail: "shop1@example.com",
		ContactPhone: "1234567890",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	
	// Test parameters
	shopID := uint(1)
	
	// Set expectations
	mockShopRepo.On("FindByID", db, shopID).
		Return(mockShop, nil)
	
	// Execute
	shop, err := usecase.GetShopByID(shopID)
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, mockShop.ID, shop.ID)
	assert.Equal(t, mockShop.Name, shop.Name)
	
	// Verify expectations
	mockShopRepo.AssertExpectations(t)
}

func TestShopUsecase_GetShopByID_NotFound(t *testing.T) {
	// Setup
	db, _, _, mockShopRepo, _, _, usecase := setupShopUsecaseTest(t)
	
	// Test parameters
	shopID := uint(999)
	
	// Mock error
	mockError := gorm.ErrRecordNotFound
	
	// Set expectations
	mockShopRepo.On("FindByID", db, shopID).
		Return(nil, mockError)
	
	// Execute
	shop, err := usecase.GetShopByID(shopID)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, shop)
	assert.Equal(t, "Shop not found", err.Error())
	
	// Verify expectations
	mockShopRepo.AssertExpectations(t)
}

func TestShopUsecase_GetShopByID_InvalidID(t *testing.T) {
	// Setup
	_, _, _, _, _, _, usecase := setupShopUsecaseTest(t)
	
	// Test with invalid ID
	shopID := uint(0)
	
	// Execute
	shop, err := usecase.GetShopByID(shopID)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, shop)
	assert.Equal(t, "Invalid input data", err.Error())
}

func TestShopUsecase_GetShopWarehouses_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, _, _, mockShopRepo, mockShopWarehouseRepo, mockWarehouseGateway, usecase := setupShopUsecaseTest(t)
	
	// Test parameters
	shopID := uint(1)
	
	// Mock data
	now := time.Now()
	mockShop := &entity.Shop{
		ID:        shopID,
		Name:      "Test Shop",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	mockWarehouseIDs := []uint{101, 102}
	
	mockWarehouseResponse1 := &model.WarehouseResponse{
		ID:        101,
		Name:      "Warehouse 1",
		Address:   "123 Storage St",
		Capacity:  1000,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	mockWarehouseResponse2 := &model.WarehouseResponse{
		ID:        102,
		Name:      "Warehouse 2",
		Address:   "456 Depot Rd",
		Capacity:  2000,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// Set expectations
	mockShopRepo.On("FindByID", db, shopID).Return(mockShop, nil)
	mockShopWarehouseRepo.On("FindWarehouseIDsByShopID", db, shopID).Return(mockWarehouseIDs, nil)
	mockWarehouseGateway.On("GetWarehouseByID", mock.Anything, uint(101)).Return(mockWarehouseResponse1, nil)
	mockWarehouseGateway.On("GetWarehouseByID", mock.Anything, uint(102)).Return(mockWarehouseResponse2, nil)
	
	// Execute
	response, err := usecase.GetShopWarehouses(ctx, shopID)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, shopID, response.ShopID)
	assert.Equal(t, 2, len(response.Warehouses))
	assert.Equal(t, uint(101), response.Warehouses[0].ID)
	assert.Equal(t, "Warehouse 1", response.Warehouses[0].Name)
	assert.Equal(t, uint(102), response.Warehouses[1].ID)
	assert.Equal(t, "Warehouse 2", response.Warehouses[1].Name)
	
	// Verify expectations
	mockShopRepo.AssertExpectations(t)
	mockShopWarehouseRepo.AssertExpectations(t)
	mockWarehouseGateway.AssertExpectations(t)
}

func TestShopUsecase_GetShopWarehouses_ShopNotFound(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, _, _, mockShopRepo, _, _, usecase := setupShopUsecaseTest(t)
	
	// Test parameters
	shopID := uint(999)
	
	// Mock error
	mockError := gorm.ErrRecordNotFound
	
	// Set expectations
	mockShopRepo.On("FindByID", db, shopID).Return(nil, mockError)
	
	// Execute
	response, err := usecase.GetShopWarehouses(ctx, shopID)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "Shop not found", err.Error())
	
	// Verify expectations
	mockShopRepo.AssertExpectations(t)
}

func TestShopUsecase_GetShopWarehouses_NoWarehouses(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, _, _, mockShopRepo, mockShopWarehouseRepo, _, usecase := setupShopUsecaseTest(t)
	
	// Test parameters
	shopID := uint(1)
	
	// Mock data
	now := time.Now()
	mockShop := &entity.Shop{
		ID:        shopID,
		Name:      "Test Shop",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// Empty warehouse IDs
	var mockWarehouseIDs []uint
	
	// Set expectations
	mockShopRepo.On("FindByID", db, shopID).Return(mockShop, nil)
	mockShopWarehouseRepo.On("FindWarehouseIDsByShopID", db, shopID).Return(mockWarehouseIDs, nil)
	
	// Execute
	response, err := usecase.GetShopWarehouses(ctx, shopID)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, shopID, response.ShopID)
	assert.Empty(t, response.Warehouses)
	
	// Verify expectations
	mockShopRepo.AssertExpectations(t)
	mockShopWarehouseRepo.AssertExpectations(t)
}

func TestShopUsecase_GetShopWarehouses_InvalidShopID(t *testing.T) {
	// Setup
	ctx := context.Background()
	_, _, _, _, _, _, usecase := setupShopUsecaseTest(t)
	
	// Test with invalid ID
	shopID := uint(0)
	
	// Execute
	response, err := usecase.GetShopWarehouses(ctx, shopID)
	
	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "Invalid input data", err.Error())
}
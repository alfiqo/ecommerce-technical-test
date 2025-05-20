package repository

import (
	"shop-service/internal/entity"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupShopRepositoryTest(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, ShopRepositoryInterface) {
	// Create a new SQL mock
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	// Create GORM DB with the SQL mock
	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)
	
	// Create the repository with a mock logger
	logger := logrus.New()
	logger.Out = nil // Disable logging for tests
	
	repo := NewShopRepository(logger)

	return db, mock, repo
}

func TestShopRepository_FindAll(t *testing.T) {
	// Setup
	db, mock, repo := setupShopRepositoryTest(t)
	
	// Mock data
	now := time.Now()
	expectedShops := []entity.Shop{
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
	
	totalCount := int64(2)
	page := 1
	pageSize := 10
	searchTerm := ""
	includeInactive := false
	
	// Mock the count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
	mock.ExpectQuery("^SELECT count.*FROM `shops`").
		WithArgs().
		WillReturnRows(countRows)
	
	// Mock the find query
	rows := sqlmock.NewRows([]string{"id", "name", "description", "address", "contact_email", "contact_phone", "is_active", "created_at", "updated_at"})
	for _, shop := range expectedShops {
		rows.AddRow(shop.ID, shop.Name, shop.Description, shop.Address, shop.ContactEmail, shop.ContactPhone, shop.IsActive, shop.CreatedAt, shop.UpdatedAt)
	}
	
	mock.ExpectQuery("^SELECT.*FROM `shops`").
		WithArgs().
		WillReturnRows(rows)
	
	// Execute the method
	actualShops, actualCount, err := repo.FindAll(db, page, pageSize, searchTerm, includeInactive)
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, totalCount, actualCount)
	assert.Len(t, actualShops, len(expectedShops))
	
	for i, shop := range actualShops {
		assert.Equal(t, expectedShops[i].ID, shop.ID)
		assert.Equal(t, expectedShops[i].Name, shop.Name)
		assert.Equal(t, expectedShops[i].Description, shop.Description)
		assert.Equal(t, expectedShops[i].Address, shop.Address)
		assert.Equal(t, expectedShops[i].ContactEmail, shop.ContactEmail)
		assert.Equal(t, expectedShops[i].ContactPhone, shop.ContactPhone)
		assert.Equal(t, expectedShops[i].IsActive, shop.IsActive)
	}
}

func TestShopRepository_FindAll_WithSearch(t *testing.T) {
	// Setup
	db, mock, repo := setupShopRepositoryTest(t)
	
	// Mock data
	now := time.Now()
	expectedShops := []entity.Shop{
		{
			ID:           1,
			Name:         "Market",
			Description:  "Supermarket with groceries",
			Address:      "Address 1",
			ContactEmail: "market@example.com",
			ContactPhone: "1234567890",
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	
	totalCount := int64(1)
	page := 1
	pageSize := 10
	searchTerm := "Market"
	includeInactive := false
	
	// Mock the count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
	mock.ExpectQuery("^SELECT count.*FROM `shops`").
		WithArgs("%"+searchTerm+"%", "%"+searchTerm+"%", true).
		WillReturnRows(countRows)
	
	// Mock the find query
	rows := sqlmock.NewRows([]string{"id", "name", "description", "address", "contact_email", "contact_phone", "is_active", "created_at", "updated_at"})
	for _, shop := range expectedShops {
		rows.AddRow(shop.ID, shop.Name, shop.Description, shop.Address, shop.ContactEmail, shop.ContactPhone, shop.IsActive, shop.CreatedAt, shop.UpdatedAt)
	}
	
	mock.ExpectQuery("^SELECT.*FROM `shops`").
		WithArgs("%"+searchTerm+"%", "%"+searchTerm+"%", true, pageSize).
		WillReturnRows(rows)
	
	// Execute the method
	actualShops, actualCount, err := repo.FindAll(db, page, pageSize, searchTerm, includeInactive)
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, totalCount, actualCount)
	assert.Len(t, actualShops, len(expectedShops))
	
	for i, shop := range actualShops {
		assert.Equal(t, expectedShops[i].ID, shop.ID)
		assert.Equal(t, expectedShops[i].Name, shop.Name)
		assert.Equal(t, expectedShops[i].Description, shop.Description)
	}
}

func TestShopRepository_FindAll_WithInactive(t *testing.T) {
	// Setup
	db, mock, repo := setupShopRepositoryTest(t)
	
	// Mock data
	now := time.Now()
	expectedShops := []entity.Shop{
		{
			ID:           1,
			Name:         "Active Shop",
			Description:  "This shop is active",
			Address:      "Address 1",
			ContactEmail: "active@example.com",
			ContactPhone: "1234567890",
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           2,
			Name:         "Inactive Shop",
			Description:  "This shop is inactive",
			Address:      "Address 2",
			ContactEmail: "inactive@example.com",
			ContactPhone: "0987654321",
			IsActive:     false,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	
	totalCount := int64(2)
	page := 1
	pageSize := 10
	searchTerm := ""
	includeInactive := true
	
	// Mock the count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
	mock.ExpectQuery("^SELECT count.*FROM `shops`").
		WithArgs().
		WillReturnRows(countRows)
	
	// Mock the find query
	rows := sqlmock.NewRows([]string{"id", "name", "description", "address", "contact_email", "contact_phone", "is_active", "created_at", "updated_at"})
	for _, shop := range expectedShops {
		rows.AddRow(shop.ID, shop.Name, shop.Description, shop.Address, shop.ContactEmail, shop.ContactPhone, shop.IsActive, shop.CreatedAt, shop.UpdatedAt)
	}
	
	mock.ExpectQuery("^SELECT.*FROM `shops`").
		WithArgs().
		WillReturnRows(rows)
	
	// Execute the method
	actualShops, actualCount, err := repo.FindAll(db, page, pageSize, searchTerm, includeInactive)
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, totalCount, actualCount)
	assert.Len(t, actualShops, len(expectedShops))
	
	// Check that we have both active and inactive shops
	hasActive := false
	hasInactive := false
	for _, shop := range actualShops {
		if shop.IsActive {
			hasActive = true
		} else {
			hasInactive = true
		}
	}
	
	assert.True(t, hasActive)
	assert.True(t, hasInactive)
}
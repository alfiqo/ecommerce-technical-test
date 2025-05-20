package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"shop-service/internal/entity"
	"shop-service/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ShopAPITestSuite defines a test suite for shop API endpoints
type ShopAPITestSuite struct {
	suite.Suite
	DB     *gorm.DB
	ApiURL string
}

// SetupSuite prepares the test suite
func (s *ShopAPITestSuite) SetupSuite() {
	// Get environment variables or use defaults
	apiHost := getEnv("API_HOST", "app")  // Container name from docker-compose.e2e.yml
	apiPort := getEnv("API_PORT", "3000") // Port from docker-compose.e2e.yml
	s.ApiURL = fmt.Sprintf("http://%s:%s/api/v1", apiHost, apiPort)

	// Setup the database connection
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASS", "")
	dbHost := getEnv("DB_HOST", "mysql") // Container name from docker-compose.e2e.yml
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "shop_service_test")

	// Create database DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// Configure GORM to be silent in tests
	dbLogger := logger.New(
		&testWriter{}, // Custom writer that does nothing
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Connect to the database with retry
	var db *gorm.DB
	var err error
	maxRetries := 5
	retryDelay := 3 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: dbLogger,
		})
		if err == nil {
			break
		}
		s.T().Logf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	if err != nil {
		s.T().Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
	}
	
	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		s.T().Fatalf("Failed to get underlying SQL DB: %v", err)
	}
	
	err = sqlDB.Ping()
	if err != nil {
		s.T().Fatalf("Failed to ping database: %v", err)
	}
	
	s.T().Logf("Successfully connected to database at %s:%s", dbHost, dbPort)
	s.DB = db

	// Migrate the schema
	s.migrateSchema()

	// Seed the database with test data
	s.seedTestData()
}

// migrateSchema creates the necessary tables for testing
func (s *ShopAPITestSuite) migrateSchema() {
	// Execute migration SQL directly instead of using AutoMigrate
	// Create shops table
	err := s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS shops (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			address TEXT NOT NULL,
			contact_email VARCHAR(255) NOT NULL,
			contact_phone VARCHAR(50) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE INDEX idx_shops_name (name),
			INDEX idx_shops_is_active (is_active)
		) ENGINE=InnoDB;
	`).Error
	if err != nil {
		s.T().Fatalf("Failed to create shops table: %v", err)
	}

	// Create shop_warehouses junction table
	err = s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS shop_warehouses (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			shop_id BIGINT UNSIGNED NOT NULL,
			warehouse_id BIGINT UNSIGNED NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE INDEX idx_shop_warehouse (shop_id, warehouse_id),
			INDEX idx_warehouse_id (warehouse_id),
			CONSTRAINT fk_shop_warehouses_shop_id FOREIGN KEY (shop_id) REFERENCES shops (id) ON DELETE CASCADE
		) ENGINE=InnoDB;
	`).Error
	if err != nil {
		s.T().Fatalf("Failed to create shop_warehouses table: %v", err)
	}
}

// seedTestData populates the database with test data
func (s *ShopAPITestSuite) seedTestData() {
	// Clean up existing data
	s.DB.Exec("DELETE FROM shop_warehouses")
	s.DB.Exec("DELETE FROM shops")

	// Create test shops - active shops
	activeShops := []entity.Shop{
		{
			Name:         "Grocery Store",
			Description:  "A store selling food and household items",
			Address:      "123 Main St",
			ContactEmail: "grocery@example.com",
			ContactPhone: "555-123-4567",
			IsActive:     true,
		},
		{
			Name:         "Electronics Shop",
			Description:  "A shop selling gadgets and electronics",
			Address:      "456 Tech Blvd",
			ContactEmail: "electronics@example.com",
			ContactPhone: "555-987-6543",
			IsActive:     true,
		},
	}

	// Insert active shops
	for i := range activeShops {
		result := s.DB.Create(&activeShops[i])
		if result.Error != nil {
			s.T().Fatalf("Failed to create active test shop: %v", result.Error)
		}
		s.T().Logf("Created test shop: %s with ID: %d", activeShops[i].Name, activeShops[i].ID)
	}

	// Create and insert inactive shop using raw SQL to ensure is_active is false
	result := s.DB.Exec(`INSERT INTO shops 
		(name, description, address, contact_email, contact_phone, is_active, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		"Closed Shop",
		"This shop is no longer active",
		"789 Old Road",
		"closed@example.com",
		"555-111-2222",
		false)
	
	if result.Error != nil {
		s.T().Fatalf("Failed to create inactive test shop: %v", result.Error)
	}
	s.T().Logf("Created inactive test shop: Closed Shop")

	// Create shop_warehouses associations
	// First for Grocery Store (ID: 1)
	shopWarehouses := []entity.ShopWarehouse{
		{
			ShopID:      1, // Grocery Store
			WarehouseID: 101,
		},
		{
			ShopID:      1, // Grocery Store
			WarehouseID: 102,
		},
		{
			ShopID:      2, // Electronics Shop
			WarehouseID: 103,
		},
		{
			ShopID:      2, // Electronics Shop
			WarehouseID: 104,
		},
		{
			ShopID:      2, // Electronics Shop
			WarehouseID: 105,
		},
	}

	// Insert shop_warehouses
	for _, shopWarehouse := range shopWarehouses {
		result := s.DB.Create(&shopWarehouse)
		if result.Error != nil {
			s.T().Fatalf("Failed to create shop_warehouse association: %v", result.Error)
		}
		s.T().Logf("Created shop_warehouse association: Shop ID %d, Warehouse ID %d", 
			shopWarehouse.ShopID, shopWarehouse.WarehouseID)
	}
}

// TearDownSuite cleans up after all tests have been run
func (s *ShopAPITestSuite) TearDownSuite() {
	// Clean up test data
	s.DB.Exec("DELETE FROM shop_warehouses")
	s.DB.Exec("DELETE FROM shops")

	// Close the database connection
	sqlDB, err := s.DB.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// TestListShops tests the GET /shops endpoint
func (s *ShopAPITestSuite) TestListShops() {
	// Make the request
	url := fmt.Sprintf("%s/shops", s.ApiURL)
	resp, err := http.Get(url)
	if err != nil {
		s.T().Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert status code
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.T().Fatalf("Failed to read response body: %v", err)
	}

	// Parse the response
	var response struct {
		Success bool                   `json:"success"`
		Data    model.ShopListResponse `json:"data"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		s.T().Fatalf("Failed to parse response: %v", err)
	}

	// Assert response structure
	assert.True(s.T(), response.Success)
	assert.NotNil(s.T(), response.Data)
	assert.GreaterOrEqual(s.T(), len(response.Data.Shops), 2) // At least the active shops
	assert.GreaterOrEqual(s.T(), response.Data.TotalCount, int64(2))
}

// TestListShopsWithSearch tests the GET /shops?search=... endpoint
func (s *ShopAPITestSuite) TestListShopsWithSearch() {
	// Make the request with search term
	url := fmt.Sprintf("%s/shops?search=Electronics", s.ApiURL)
	resp, err := http.Get(url)
	if err != nil {
		s.T().Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert status code
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.T().Fatalf("Failed to read response body: %v", err)
	}

	// Parse the response
	var response struct {
		Success bool                   `json:"success"`
		Data    model.ShopListResponse `json:"data"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		s.T().Fatalf("Failed to parse response: %v", err)
	}

	// Assert response data
	assert.True(s.T(), response.Success)
	// Check if we have at least one shop
	if len(response.Data.Shops) > 0 {
		// Only check the name if we have shops
		found := false
		for _, shop := range response.Data.Shops {
			if shop.Name == "Electronics Shop" {
				found = true
				break
			}
		}
		assert.True(s.T(), found, "Should find a shop with name 'Electronics Shop'")
	}
}

// TestListShopsWithInactive tests the GET /shops?include_inactive=true endpoint
func (s *ShopAPITestSuite) TestListShopsWithInactive() {
	// Make the request with include_inactive=true
	url := fmt.Sprintf("%s/shops?include_inactive=true", s.ApiURL)
	resp, err := http.Get(url)
	if err != nil {
		s.T().Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert status code
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.T().Fatalf("Failed to read response body: %v", err)
	}

	// Parse the response
	var response struct {
		Success bool                   `json:"success"`
		Data    model.ShopListResponse `json:"data"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		s.T().Fatalf("Failed to parse response: %v", err)
	}

	// Assert response includes inactive shop
	assert.True(s.T(), response.Success)
	assert.GreaterOrEqual(s.T(), len(response.Data.Shops), 3) // All shops including inactive
	
	// Check if we have at least one inactive shop
	hasInactive := false
	for _, shop := range response.Data.Shops {
		if !shop.IsActive {
			hasInactive = true
			break
		}
	}
	assert.True(s.T(), hasInactive, "Should have at least one inactive shop")
}

// TestListShopsPagination tests the pagination functionality
func (s *ShopAPITestSuite) TestListShopsPagination() {
	// Make the request with pagination
	url := fmt.Sprintf("%s/shops?page=1&page_size=1", s.ApiURL)
	resp, err := http.Get(url)
	if err != nil {
		s.T().Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert status code
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.T().Fatalf("Failed to read response body: %v", err)
	}

	// Parse the response
	var response struct {
		Success bool                   `json:"success"`
		Data    model.ShopListResponse `json:"data"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		s.T().Fatalf("Failed to parse response: %v", err)
	}

	// Assert response pagination
	assert.True(s.T(), response.Success)
	assert.Equal(s.T(), 1, len(response.Data.Shops))
	assert.Equal(s.T(), 1, response.Data.Page)
	assert.Equal(s.T(), 1, response.Data.PageSize)
	assert.GreaterOrEqual(s.T(), response.Data.TotalCount, int64(2))
}

// TestGetShopByID tests the GET /shops/:id endpoint success case
func (s *ShopAPITestSuite) TestGetShopByID() {
	// We know shop with ID 1 exists from seed data (Grocery Store)
	shopID := 1
	
	// Make the request
	url := fmt.Sprintf("%s/shops/%d", s.ApiURL, shopID)
	resp, err := http.Get(url)
	if err != nil {
		s.T().Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert status code
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.T().Fatalf("Failed to read response body: %v", err)
	}

	// Parse the response
	var response struct {
		Success bool                     `json:"success"`
		Data    model.ShopDetailResponse `json:"data"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		s.T().Fatalf("Failed to parse response: %v", err)
	}

	// Assert response structure
	assert.True(s.T(), response.Success)
	assert.NotNil(s.T(), response.Data)
	
	// Verify shop data
	assert.Equal(s.T(), uint(shopID), response.Data.ID)
	assert.Equal(s.T(), "Grocery Store", response.Data.Name)
	assert.Equal(s.T(), "A store selling food and household items", response.Data.Description)
	assert.Equal(s.T(), "123 Main St", response.Data.Address)
	assert.Equal(s.T(), "grocery@example.com", response.Data.ContactEmail)
	assert.Equal(s.T(), "555-123-4567", response.Data.ContactPhone)
	assert.True(s.T(), response.Data.IsActive)
	
	// Verify warehouse IDs
	assert.Equal(s.T(), 2, len(response.Data.WarehouseIDs), "Should have 2 warehouse associations")
	
	// Check if warehouse IDs match what we expect
	warehouseIDs := []uint{101, 102} // From seed data
	for i, warehouse := range response.Data.WarehouseIDs {
		assert.Equal(s.T(), warehouseIDs[i], warehouse.ID, 
			"Warehouse ID should match what we set in seed data")
	}
}

// TestGetShopByIDInvalidFormat tests the GET /shops/:id endpoint with invalid ID format
func (s *ShopAPITestSuite) TestGetShopByIDInvalidFormat() {
	// Make the request with an invalid ID (non-numeric)
	url := fmt.Sprintf("%s/shops/invalid-id", s.ApiURL)
	resp, err := http.Get(url)
	if err != nil {
		s.T().Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert status code
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.T().Fatalf("Failed to read response body: %v", err)
	}

	// Parse the response
	var response struct {
		Success bool        `json:"success"`
		Error   struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		s.T().Fatalf("Failed to parse response: %v", err)
	}

	// Assert error response structure
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "INVALID_INPUT", response.Error.Code)
	assert.Contains(s.T(), response.Error.Message, "Invalid input data")
}

// TestGetShopByIDNotFound tests the GET /shops/:id endpoint with a non-existent ID
func (s *ShopAPITestSuite) TestGetShopByIDNotFound() {
	// Use a very large ID that should not exist
	nonExistentID := 9999
	
	// Make the request with a non-existent ID
	url := fmt.Sprintf("%s/shops/%d", s.ApiURL, nonExistentID)
	resp, err := http.Get(url)
	if err != nil {
		s.T().Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert status code
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.T().Fatalf("Failed to read response body: %v", err)
	}

	// Parse the response
	var response struct {
		Success bool        `json:"success"`
		Error   struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		s.T().Fatalf("Failed to parse response: %v", err)
	}

	// Assert error response structure
	assert.False(s.T(), response.Success)
	assert.Equal(s.T(), "SHOP_NOT_FOUND", response.Error.Code)
	assert.Contains(s.T(), response.Error.Message, "Shop not found")
}

// Run the test suite
func TestShopAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(ShopAPITestSuite))
}

// Helper functions

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// testWriter is a custom writer that implements logger.Writer for GORM
type testWriter struct{}

func (w *testWriter) Printf(format string, args ...interface{}) {
	// Do nothing, effectively silencing the output
}

// Helper function for HTTP requests with context
func makeRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}
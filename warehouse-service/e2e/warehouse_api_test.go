package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	baseURL         = "http://app:3000"
	warehousesURL   = baseURL + "/api/v1/warehouses"
	warehouseIDURL  = baseURL + "/api/v1/warehouses/" // Will be appended with warehouse ID
	adminToken      = "admin_token_here"
)

// Response structures
type WebResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *ErrorInfo      `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type WarehouseResponse struct {
	ID        uint             `json:"id,omitempty"`
	Name      string           `json:"name,omitempty"`
	Location  string           `json:"location,omitempty"`
	Address   string           `json:"address,omitempty"`
	IsActive  bool             `json:"is_active,omitempty"`
	Stats     *WarehouseStatsDTO `json:"stats,omitempty"`
	CreatedAt string           `json:"created_at,omitempty"`
	UpdatedAt string           `json:"updated_at,omitempty"`
}

type WarehouseStatsDTO struct {
	TotalProducts int64 `json:"total_products"`
	TotalItems    int64 `json:"total_items"`
}

// Test warehouse data
var (
	testWarehouse = struct {
		Name     string
		Location string
		Address  string
		IsActive bool
	}{
		Name:     "E2E Test Warehouse",
		Location: "Test Location",
		Address:  fmt.Sprintf("123 Test Street, Suite %d", time.Now().Unix()%1000),
		IsActive: true,
	}
)

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Setup: wait for the service to be ready
	waitForService()

	// Run the tests
	exitCode := m.Run()

	// Cleanup could go here if needed
	
	os.Exit(exitCode)
}

// waitForService waits for the service to be available
func waitForService() {
	healthEndpoint := baseURL + "/api/v1/health"
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(healthEndpoint)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		fmt.Printf("Waiting for service to be ready... (%d/%d)\n", i+1, maxRetries)
		time.Sleep(time.Second)
	}
	fmt.Println("Service did not become ready in time")
}

// TestWarehouseFlow tests the complete warehouse flow
func TestWarehouseFlow(t *testing.T) {
	// 1. Create a new warehouse
	warehouseID := testCreateWarehouse(t)
	require.NotZero(t, warehouseID, "Warehouse ID should not be empty")

	// 2. Get the warehouse details
	warehouse := testGetWarehouse(t, warehouseID)
	require.Equal(t, testWarehouse.Name, warehouse.Name, "Warehouse name should match")

	// 3. Update the warehouse
	updatedWarehouse := testUpdateWarehouse(t, warehouseID)
	require.Equal(t, "Updated "+testWarehouse.Name, updatedWarehouse.Name, "Updated warehouse name should match")

	// 4. Delete the warehouse
	testDeleteWarehouse(t, warehouseID)
}

// testCreateWarehouse tests creating a new warehouse
func testCreateWarehouse(t *testing.T) uint {
	t.Log("Testing warehouse creation...")

	// Create request payload
	payload := map[string]interface{}{
		"name":      testWarehouse.Name,
		"location":  testWarehouse.Location,
		"address":   testWarehouse.Address,
		"is_active": testWarehouse.IsActive,
	}
	jsonPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	// Create a client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create the request
	req, err := http.NewRequest("POST", warehousesURL, bytes.NewBuffer(jsonPayload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	// Send the request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read and log response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	t.Logf("Create Warehouse Response (status %d): %s", resp.StatusCode, string(body))

	// Check status code
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Parse the response
	var response WebResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "Failed to parse JSON response")
	require.True(t, response.Success, "Expected success to be true")

	var warehouseResponse WarehouseResponse
	err = json.Unmarshal(response.Data, &warehouseResponse)
	require.NoError(t, err, "Failed to parse warehouse response")
	require.NotZero(t, warehouseResponse.ID, "Warehouse ID should not be empty")

	return warehouseResponse.ID
}

// testGetWarehouse tests retrieving a warehouse by ID
func testGetWarehouse(t *testing.T, warehouseID uint) *WarehouseResponse {
	t.Log("Testing get warehouse...")

	// Create a client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create the request
	url := fmt.Sprintf("%s%d", warehouseIDURL, warehouseID)
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	// Send the request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read and log response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	t.Logf("Get Warehouse Response (status %d): %s", resp.StatusCode, string(body))

	// Check status code
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Parse the response
	var response WebResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "Failed to parse JSON response")
	require.True(t, response.Success, "Expected success to be true")

	var warehouseResponse WarehouseResponse
	err = json.Unmarshal(response.Data, &warehouseResponse)
	require.NoError(t, err, "Failed to parse warehouse response")
	require.Equal(t, warehouseID, warehouseResponse.ID, "Warehouse ID should match")

	return &warehouseResponse
}

// testUpdateWarehouse tests updating a warehouse
func testUpdateWarehouse(t *testing.T, warehouseID uint) *WarehouseResponse {
	t.Log("Testing update warehouse...")

	// Create request payload with updated name
	payload := map[string]interface{}{
		"name":      "Updated " + testWarehouse.Name,
		"location":  testWarehouse.Location,
		"address":   testWarehouse.Address,
		"is_active": testWarehouse.IsActive,
	}
	jsonPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	// Create a client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create the request
	url := fmt.Sprintf("%s%d", warehouseIDURL, warehouseID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	// Send the request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read and log response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	t.Logf("Update Warehouse Response (status %d): %s", resp.StatusCode, string(body))

	// Check status code
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Parse the response
	var response WebResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "Failed to parse JSON response")
	require.True(t, response.Success, "Expected success to be true")

	var warehouseResponse WarehouseResponse
	err = json.Unmarshal(response.Data, &warehouseResponse)
	require.NoError(t, err, "Failed to parse warehouse response")
	require.Equal(t, warehouseID, warehouseResponse.ID, "Warehouse ID should match")
	require.Equal(t, "Updated "+testWarehouse.Name, warehouseResponse.Name, "Updated warehouse name should match")

	return &warehouseResponse
}

// testDeleteWarehouse tests deleting a warehouse
func testDeleteWarehouse(t *testing.T, warehouseID uint) {
	t.Log("Testing delete warehouse...")

	// Create a client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create the request
	url := fmt.Sprintf("%s%d", warehouseIDURL, warehouseID)
	req, err := http.NewRequest("DELETE", url, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	// Send the request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read and log response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	t.Logf("Delete Warehouse Response (status %d): %s", resp.StatusCode, string(body))

	// Check status code
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Parse the response
	var response WebResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "Failed to parse JSON response")
	require.True(t, response.Success, "Expected success to be true")
}

// TestUnauthorizedAccess tests accessing endpoints without authentication
func TestUnauthorizedAccess(t *testing.T) {
	t.Log("Testing unauthorized access...")

	// Create a client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create the request without authentication token
	req, err := http.NewRequest("GET", warehousesURL, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read and log response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	t.Logf("Unauthorized Response (status %d): %s", resp.StatusCode, string(body))

	// Check status code
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Expected status code 401")

	// Parse the response
	var response WebResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "Failed to parse JSON response")
	require.False(t, response.Success, "Expected success to be false")
	require.NotNil(t, response.Error, "Expected error to be present")
}
package converter

import (
	"testing"
	"time"
	"warehouse-service/internal/entity"
	"warehouse-service/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestWarehouseToResponse(t *testing.T) {
	// Create test data
	createdAt := time.Now().UTC()
	updatedAt := time.Now().UTC()
	
	warehouse := &entity.Warehouse{
		ID:        1,
		Name:      "Test Warehouse",
		Location:  "Test Location",
		Address:   "Test Address",
		IsActive:  true,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	
	stats := &model.WarehouseStatsDTO{
		TotalProducts: 10,
		TotalItems:    100,
	}
	
	// Call the function
	response := WarehouseToResponse(warehouse, stats)
	
	// Assert the results
	assert.NotNil(t, response)
	assert.Equal(t, warehouse.ID, response.ID)
	assert.Equal(t, warehouse.Name, response.Name)
	assert.Equal(t, warehouse.Location, response.Location)
	assert.Equal(t, warehouse.Address, response.Address)
	assert.Equal(t, warehouse.IsActive, response.IsActive)
	assert.Equal(t, stats, response.Stats)
	assert.Equal(t, createdAt.Format(time.RFC3339), response.CreatedAt)
	assert.Equal(t, updatedAt.Format(time.RFC3339), response.UpdatedAt)
}

func TestWarehouseRequestToEntity(t *testing.T) {
	// Create test data
	request := &model.CreateWarehouseRequest{
		Name:     "New Warehouse",
		Location: "New Location",
		Address:  "New Address",
		IsActive: true,
	}
	
	// Call the function
	entity := WarehouseRequestToEntity(request)
	
	// Assert the results
	assert.NotNil(t, entity)
	assert.Equal(t, request.Name, entity.Name)
	assert.Equal(t, request.Location, entity.Location)
	assert.Equal(t, request.Address, entity.Address)
	assert.Equal(t, request.IsActive, entity.IsActive)
}

func TestUpdateWarehouseFromRequest(t *testing.T) {
	// Create test data
	warehouse := &entity.Warehouse{
		ID:       1,
		Name:     "Old Warehouse",
		Location: "Old Location",
		Address:  "Old Address",
		IsActive: false,
	}
	
	request := &model.UpdateWarehouseRequest{
		ID:       1,
		Name:     "Updated Warehouse",
		Location: "Updated Location",
		Address:  "Updated Address",
		IsActive: true,
	}
	
	// Call the function
	UpdateWarehouseFromRequest(warehouse, request)
	
	// Assert the results
	assert.Equal(t, request.ID, warehouse.ID)
	assert.Equal(t, request.Name, warehouse.Name)
	assert.Equal(t, request.Location, warehouse.Location)
	assert.Equal(t, request.Address, warehouse.Address)
	assert.Equal(t, request.IsActive, warehouse.IsActive)
}
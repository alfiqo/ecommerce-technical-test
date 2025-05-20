package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWarehouseStock_CalculateAvailableQuantity(t *testing.T) {
	// Test case 1: Normal calculation
	stock1 := &WarehouseStock{
		WarehouseID:      1,
		ProductID:        1,
		Quantity:         100,
		ReservedQuantity: 30,
	}
	stock1.CalculateAvailableQuantity()
	assert.Equal(t, 70, stock1.AvailableQuantity)
	
	// Test case 2: Reserved equals quantity
	stock2 := &WarehouseStock{
		WarehouseID:      1,
		ProductID:        2,
		Quantity:         50,
		ReservedQuantity: 50,
	}
	stock2.CalculateAvailableQuantity()
	assert.Equal(t, 0, stock2.AvailableQuantity)
	
	// Test case 3: Reserved greater than quantity (edge case)
	stock3 := &WarehouseStock{
		WarehouseID:      1,
		ProductID:        3,
		Quantity:         20,
		ReservedQuantity: 30,
	}
	stock3.CalculateAvailableQuantity()
	assert.Equal(t, 0, stock3.AvailableQuantity)
	
	// Test case 4: Zero quantity
	stock4 := &WarehouseStock{
		WarehouseID:      1,
		ProductID:        4,
		Quantity:         0,
		ReservedQuantity: 0,
	}
	stock4.CalculateAvailableQuantity()
	assert.Equal(t, 0, stock4.AvailableQuantity)
}

func TestTransferStatus_Values(t *testing.T) {
	// Test transfer status values
	assert.Equal(t, TransferStatus("pending"), StatusPending)
	assert.Equal(t, TransferStatus("completed"), StatusCompleted)
	assert.Equal(t, TransferStatus("failed"), StatusFailed)
}
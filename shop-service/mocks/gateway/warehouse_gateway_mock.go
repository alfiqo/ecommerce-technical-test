package gateway

import (
	"context"
	"shop-service/internal/model"

	"github.com/stretchr/testify/mock"
)

// WarehouseGatewayMock is a mock for WarehouseGatewayInterface
type WarehouseGatewayMock struct {
	mock.Mock
}

// GetWarehouseByID mocks the GetWarehouseByID method
func (m *WarehouseGatewayMock) GetWarehouseByID(ctx context.Context, warehouseID uint) (*model.WarehouseResponse, error) {
	args := m.Called(ctx, warehouseID)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).(*model.WarehouseResponse), args.Error(1)
}
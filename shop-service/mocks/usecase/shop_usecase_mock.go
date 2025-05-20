package mocks

import (
	"context"
	"shop-service/internal/entity"
	"shop-service/internal/model"

	"github.com/stretchr/testify/mock"
)

// ShopUsecaseMock is a mock for ShopUsecaseInterface
type ShopUsecaseMock struct {
	mock.Mock
}

// ListShops provides a mock function
func (_m *ShopUsecaseMock) ListShops(page int, pageSize int, searchTerm string, includeInactive bool) ([]entity.Shop, int64, error) {
	ret := _m.Called(page, pageSize, searchTerm, includeInactive)

	var r0 []entity.Shop
	var r1 int64
	var r2 error

	if rf, ok := ret.Get(0).(func(int, int, string, bool) []entity.Shop); ok {
		r0 = rf(page, pageSize, searchTerm, includeInactive)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entity.Shop)
		}
	}

	if rf, ok := ret.Get(1).(func(int, int, string, bool) int64); ok {
		r1 = rf(page, pageSize, searchTerm, includeInactive)
	} else {
		r1 = ret.Get(1).(int64)
	}

	if rf, ok := ret.Get(2).(func(int, int, string, bool) error); ok {
		r2 = rf(page, pageSize, searchTerm, includeInactive)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetShopByID provides a mock function
func (_m *ShopUsecaseMock) GetShopByID(id uint) (*entity.Shop, error) {
	ret := _m.Called(id)

	var r0 *entity.Shop
	var r1 error

	if rf, ok := ret.Get(0).(func(uint) *entity.Shop); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.Shop)
		}
	}

	if rf, ok := ret.Get(1).(func(uint) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetShopWithWarehouses provides a mock function
func (_m *ShopUsecaseMock) GetShopWithWarehouses(id uint) (*entity.Shop, error) {
	ret := _m.Called(id)

	var r0 *entity.Shop
	var r1 error

	if rf, ok := ret.Get(0).(func(uint) *entity.Shop); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.Shop)
		}
	}

	if rf, ok := ret.Get(1).(func(uint) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetShopWarehouses provides a mock function
func (_m *ShopUsecaseMock) GetShopWarehouses(ctx context.Context, shopID uint) (*model.ShopWarehousesResponse, error) {
	ret := _m.Called(ctx, shopID)

	var r0 *model.ShopWarehousesResponse
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uint) *model.ShopWarehousesResponse); ok {
		r0 = rf(ctx, shopID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ShopWarehousesResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uint) error); ok {
		r1 = rf(ctx, shopID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateShop provides a mock function
func (_m *ShopUsecaseMock) CreateShop(req *model.CreateShopRequest) (*entity.Shop, error) {
	ret := _m.Called(req)

	var r0 *entity.Shop
	var r1 error

	if rf, ok := ret.Get(0).(func(*model.CreateShopRequest) *entity.Shop); ok {
		r0 = rf(req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.Shop)
		}
	}

	if rf, ok := ret.Get(1).(func(*model.CreateShopRequest) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
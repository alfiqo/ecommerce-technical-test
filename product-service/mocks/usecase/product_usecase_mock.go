package usecase

import (
	"context"
	"product-service/internal/model"

	"github.com/stretchr/testify/mock"
)

type MockProductUseCase struct {
	mock.Mock
}

func (m *MockProductUseCase) GetProducts(ctx context.Context, limit, offset int) (*model.ProductListResponse, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProductListResponse), args.Error(1)
}

func (m *MockProductUseCase) GetProductByID(ctx context.Context, id string) (*model.ProductResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProductResponse), args.Error(1)
}

func (m *MockProductUseCase) CreateProduct(ctx context.Context, request *model.CreateProductRequest) (*model.ProductResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProductResponse), args.Error(1)
}

func (m *MockProductUseCase) UpdateProduct(ctx context.Context, id string, request *model.UpdateProductRequest) (*model.ProductResponse, error) {
	args := m.Called(ctx, id, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProductResponse), args.Error(1)
}

func (m *MockProductUseCase) DeleteProduct(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductUseCase) SearchProducts(ctx context.Context, query string, limit, offset int) (*model.ProductListResponse, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProductListResponse), args.Error(1)
}

func (m *MockProductUseCase) GetProductsByCategory(ctx context.Context, category string, limit, offset int) (*model.ProductListResponse, error) {
	args := m.Called(ctx, category, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProductListResponse), args.Error(1)
}
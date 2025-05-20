package usecase_mock

import (
	"context"
	"order-service/internal/model"

	"github.com/stretchr/testify/mock"
)

// ReservationUseCaseMock is a mock implementation of the ReservationUseCaseInterface
type ReservationUseCaseMock struct {
	mock.Mock
}

// CreateReservation mocks the CreateReservation method
func (m *ReservationUseCaseMock) CreateReservation(ctx context.Context, request *model.ReservationRequest) (*model.ReservationResponse, error) {
	args := m.Called(ctx, request)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).(*model.ReservationResponse), args.Error(1)
}

// GetReservationsByOrderID mocks the GetReservationsByOrderID method
func (m *ReservationUseCaseMock) GetReservationsByOrderID(ctx context.Context, orderID uint) ([]model.ReservationResponse, error) {
	args := m.Called(ctx, orderID)
	
	return args.Get(0).([]model.ReservationResponse), args.Error(1)
}

// DeactivateReservation mocks the DeactivateReservation method
func (m *ReservationUseCaseMock) DeactivateReservation(ctx context.Context, reservationID uint) error {
	args := m.Called(ctx, reservationID)
	return args.Error(0)
}

// CleanupExpiredReservations mocks the CleanupExpiredReservations method
func (m *ReservationUseCaseMock) CleanupExpiredReservations(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
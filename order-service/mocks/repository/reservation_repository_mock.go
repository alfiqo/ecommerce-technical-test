package repository_mock

import (
	"order-service/internal/entity"
	"time"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ReservationRepositoryMock is a mock implementation of the ReservationRepositoryInterface
type ReservationRepositoryMock struct {
	mock.Mock
}

// CreateReservation mocks the CreateReservation method
func (m *ReservationRepositoryMock) CreateReservation(tx *gorm.DB, reservation *entity.Reservation) error {
	args := m.Called(tx, reservation)
	return args.Error(0)
}

// CreateReservationBatch mocks the CreateReservationBatch method
func (m *ReservationRepositoryMock) CreateReservationBatch(tx *gorm.DB, reservations []entity.Reservation) error {
	args := m.Called(tx, reservations)
	return args.Error(0)
}

// FindReservationsByOrderID mocks the FindReservationsByOrderID method
func (m *ReservationRepositoryMock) FindReservationsByOrderID(tx *gorm.DB, orderID uint) ([]entity.Reservation, error) {
	args := m.Called(tx, orderID)
	
	return args.Get(0).([]entity.Reservation), args.Error(1)
}

// UpdateReservationStatus mocks the UpdateReservationStatus method
func (m *ReservationRepositoryMock) UpdateReservationStatus(tx *gorm.DB, reservationID uint, isActive bool) error {
	args := m.Called(tx, reservationID, isActive)
	return args.Error(0)
}

// DeactivateReservationsByOrderID mocks the DeactivateReservationsByOrderID method
func (m *ReservationRepositoryMock) DeactivateReservationsByOrderID(tx *gorm.DB, orderID uint) error {
	args := m.Called(tx, orderID)
	return args.Error(0)
}

// FindExpiredReservations mocks the FindExpiredReservations method
func (m *ReservationRepositoryMock) FindExpiredReservations(tx *gorm.DB, currentTime time.Time) ([]entity.Reservation, error) {
	args := m.Called(tx, currentTime)
	
	return args.Get(0).([]entity.Reservation), args.Error(1)
}
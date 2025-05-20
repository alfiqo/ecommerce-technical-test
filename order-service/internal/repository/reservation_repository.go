package repository

import (
	"order-service/internal/entity"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReservationRepositoryInterface interface {
	CreateReservation(tx *gorm.DB, reservation *entity.Reservation) error
	CreateReservationBatch(tx *gorm.DB, reservations []entity.Reservation) error
	FindReservationsByOrderID(tx *gorm.DB, orderID uint) ([]entity.Reservation, error)
	UpdateReservationStatus(tx *gorm.DB, reservationID uint, isActive bool) error
	DeactivateReservationsByOrderID(tx *gorm.DB, orderID uint) error
	FindExpiredReservations(tx *gorm.DB, currentTime time.Time) ([]entity.Reservation, error)
}

type ReservationRepository struct {
	DB  *gorm.DB
	Log *logrus.Logger
}

func NewReservationRepository(log *logrus.Logger, db *gorm.DB) ReservationRepositoryInterface {
	return &ReservationRepository{
		DB:  db,
		Log: log,
	}
}

func (r *ReservationRepository) CreateReservation(tx *gorm.DB, reservation *entity.Reservation) error {
	return tx.Create(reservation).Error
}

func (r *ReservationRepository) CreateReservationBatch(tx *gorm.DB, reservations []entity.Reservation) error {
	return tx.Create(&reservations).Error
}

func (r *ReservationRepository) FindReservationsByOrderID(tx *gorm.DB, orderID uint) ([]entity.Reservation, error) {
	var reservations []entity.Reservation
	
	err := tx.Where("order_id = ?", orderID).Find(&reservations).Error
	if err != nil {
		return nil, err
	}
	
	return reservations, nil
}

func (r *ReservationRepository) UpdateReservationStatus(tx *gorm.DB, reservationID uint, isActive bool) error {
	return tx.Model(&entity.Reservation{}).Where("id = ?", reservationID).Update("is_active", isActive).Error
}

func (r *ReservationRepository) DeactivateReservationsByOrderID(tx *gorm.DB, orderID uint) error {
	return tx.Model(&entity.Reservation{}).Where("order_id = ?", orderID).Update("is_active", false).Error
}

func (r *ReservationRepository) FindExpiredReservations(tx *gorm.DB, currentTime time.Time) ([]entity.Reservation, error) {
	var reservations []entity.Reservation
	
	err := tx.Where("expires_at < ? AND is_active = true", currentTime).Find(&reservations).Error
	if err != nil {
		return nil, err
	}
	
	return reservations, nil
}
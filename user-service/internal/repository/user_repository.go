package repository

import (
	"user-service/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	Create(db *gorm.DB, user *entity.User) error
	FindByEmail(db *gorm.DB, user *entity.User, email string) error
	FindByToken(db *gorm.DB, token string) (*entity.User, error)
	Update(db *gorm.DB, user *entity.User) error
}

type UserRepository struct {
	DB  *gorm.DB
	Log *logrus.Logger
}

func NewUserRepository(log *logrus.Logger, db *gorm.DB) UserRepositoryInterface {
	return &UserRepository{
		DB:  db,
		Log: log,
	}
}

func (r *UserRepository) Create(db *gorm.DB, user *entity.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepository) FindByEmail(db *gorm.DB, user *entity.User, email string) error {
	return r.DB.Where("email = ?", email).Take(user).Error
}

func (r *UserRepository) Update(db *gorm.DB, user *entity.User) error {
	return r.DB.Save(user).Error
}

func (r *UserRepository) FindByToken(db *gorm.DB, token string) (*entity.User, error) {
	user := new(entity.User)
	if err := db.Where("token = ?", token).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

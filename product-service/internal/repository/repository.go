package repository

import "gorm.io/gorm"

// Repository is an interface for repository operations
type Repository interface {
	GetDB() *gorm.DB
}
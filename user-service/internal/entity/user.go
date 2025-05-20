package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User is a struct that represents a user entity
type User struct {
	ID        uuid.UUID `gorm:"column:uuid;primaryKey"`
	Name      string    `gorm:"column:name;type:varchar(255);not null"`
	Email     string    `gorm:"column:email;type:varchar(255);uniqueIndex;not null"`
	Phone     string    `gorm:"column:phone;type:varchar(50);uniqueIndex"`
	Password  string    `gorm:"column:password;type:varchar(100);not null"`
	Token     string    `gorm:"column:token;type:varchar(255)"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"` // Menggunakan time.Time untuk timestamp
	UpdatedAt time.Time `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return
}

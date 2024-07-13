package models

import (
	"time"

	"gorm.io/gorm"
)

// User model
type User struct {
	ID          string         `gorm:"primary_key;type:varchar(36)" json:"user_id"`
	FirstName   string         `gorm:"type:varchar(255)" json:"first_name"`
	LastName    string         `gorm:"type:varchar(255)" json:"last_name"`
	Email       string         `gorm:"type:varchar(255);unique" json:"email"`
	PhoneNumber string         `gorm:"type:varchar(255)" json:"phone_number"`
	Password    string         `gorm:"type:varchar(255)" json:"password"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at"`
	Addresses   []AddressUser  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"addresses"`
}

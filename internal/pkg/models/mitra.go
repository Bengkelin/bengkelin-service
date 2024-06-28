package models

import (
	"time"

	"gorm.io/gorm"
)

type Mitra struct {
	ID         string         `gorm:"primary_key;type:varchar(36)" json:"mitra_id"`
	FirstName  string         `gorm:"type:varchar(255)" json:"first_name"`
	LastName   string         `gorm:"type:varchar(255)" json:"last_name"`
	Email      string         `gorm:"type:varchar(255);unique" json:"email"`
	Password   string         `gorm:"type:varchar(255)" json:"password"`
	NoTelp     string         `gorm:"type:varchar(255)" json:"no_telp"`
	BankName   string         `gorm:"type:varchar(255)" json:"bank_name"`
	BankNumber string         `gorm:"type:varchar(255)" json:"bank_number"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at"`
}

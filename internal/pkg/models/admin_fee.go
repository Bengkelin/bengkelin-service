package models

import "time"

type AdminFee struct {
	ID        string    `gorm:"primaryKey;varchar(36)" json:"id"`
	AdminFee  float64   `json:"admin_fee"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

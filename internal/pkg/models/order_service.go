package models

import "time"

type OrderService struct {
	ID        uint      `gorm:"primary_key;autoIncrement" json:"id"`
	OrderID   string    `gorm:"type:varchar(36);index" json:"order_id"`
	Title     string    `gorm:"type:varchar(255)" json:"title"`
	Detail    string    `gorm:"type:varchar(255)" json:"detail"`
	Price     float64   `gorm:"type:float" json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

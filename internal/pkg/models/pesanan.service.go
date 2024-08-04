package models

import "time"

type PesananService struct {
	ID          uint      `gorm:"primary_key;autoIncrement" json:"id"`
	PesananID   string    `gorm:"primary_key;reference;type:varchar(36)" json:"pesanan_id"`
	ServiceName string    `gorm:"type:varchar(255)" json:"service_name"`
	Note        string    `gorm:"type:varchar(255)" json:"note"`
	Price       float64   `gorm:"type:float" json:"price"`
	TotalPrice  float64   `gorm:"type:float" json:"total_price"`
	Pesanan     Pesanan   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

package models

import "time"

type Pesanan struct {
	ID             string           `gorm:"primary_key;type:varchar(36)" json:"id"`
	UserID         string           `gorm:"type:varchar(36)" json:"user_id"`
	BengkelID      string           `gorm:"type:varchar(36)" json:"bengkel_id"`
	VehicleID      string           `gorm:"type:varchar(36)" json:"vehicle_id"`
	Status         uint             `gorm:"type:int" json:"status"`
	IsHomeService  bool             `gorm:"type:bool" json:"is_home_service"`
	TotalPrice     float64          `gorm:"type:float" json:"total_price"`
	HandlingFee    float64          `gorm:"type:float" json:"handling_fee"`
	ServiceFee     float64          `gorm:"type:float" json:"service_fee"`
	HomeServiceFee float64          `gorm:"type:float" json:"home_service_fee"`
	Note           string           `gorm:"type:text" json:"note"`
	User           User             `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
	Bengkel        Bengkel          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"bengkel"`
	Vehicle        Vehicle          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"vehicle"`
	PesananService []PesananService `gorm:"foreignKey:PesananID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"pesanan_service"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

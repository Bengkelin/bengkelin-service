package models

import "time"

type Order struct {
	ID                  string           `gorm:"primary_key;type:varchar(36)" json:"id"`
	UserID              string           `gorm:"type:varchar(36)" json:"user_id"`
	BengkelID           string           `gorm:"type:varchar(36)" json:"bengkel_id"`
	VehicleID           uint             `gorm:"type:int" json:"vehicle_id"`
	Status              uint             `gorm:"type:int" json:"status"`
	IsHomeService       *bool            `gorm:"type:bool" json:"is_home_service"`
	TotalPrice          float64          `gorm:"type:float" json:"total_price"`
	AdminFee            float64          `gorm:"type:float" json:"admin_fee"`
	HomeServiceFee      float64          `gorm:"type:float" json:"home_service_fee"`
	HomeServiceSchedule string         `gorm:"type:varchar(50)" json:"home_service_schedule"`
	PaymentMethod       string         `gorm:"type:varchar(50)" json:"payment_method"`
	Note                string         `gorm:"type:text" json:"note"`
	User                User           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
	Bengkel             Bengkel        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"bengkel"`
	Vehicle             Vehicle        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"vehicle"`
	OrderServices       []OrderService `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"order_services"`
	ConfirmedAt         *time.Time     `json:"confirmed_at"`
	PaidAt              *time.Time     `json:"paid_at"`
	FinishedAt          *time.Time     `json:"finished_at"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

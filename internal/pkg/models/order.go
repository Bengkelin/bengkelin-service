package models

import "time"

type Order struct {
	ID                  string             `gorm:"primary_key;type:varchar(36)" json:"id"`
	UserID              string             `gorm:"type:varchar(36);index:idx_orders_user_status_created,priority:1" json:"user_id"`
	BengkelID           string             `gorm:"type:varchar(36);index:idx_orders_bengkel_status_created,priority:1" json:"bengkel_id"`
	VehicleID           uint               `gorm:"type:int" json:"vehicle_id"`
	Status              OrderStatus        `gorm:"type:int;index:idx_orders_bengkel_status_created,priority:2;index:idx_orders_user_status_created,priority:2" json:"status"`
	IsHomeService       *bool              `gorm:"type:bool" json:"is_home_service"`
	TotalPrice          float64            `gorm:"type:float" json:"total_price"`
	AdminFee            float64            `gorm:"type:float" json:"admin_fee"`
	HomeServiceFee      float64            `gorm:"type:float" json:"home_service_fee"`
	HomeServiceSchedule string             `gorm:"type:varchar(50)" json:"home_service_schedule"`
	PaymentMethod       string             `gorm:"type:varchar(50)" json:"payment_method"`
	Note                string             `gorm:"type:text" json:"note"`
	CancelledBy         string             `gorm:"type:varchar(36)" json:"cancelled_by,omitempty"` // User ID or Mitra ID who cancelled
	CancelledReason     CancellationReason `gorm:"type:varchar(50)" json:"cancelled_reason,omitempty"`
	User                User               `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
	Bengkel             Bengkel            `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"bengkel"`
	Vehicle             Vehicle            `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"vehicle"`
	OrderServices       []OrderService     `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"order_services"`
	ConfirmedAt         *time.Time         `json:"confirmed_at"`
	PaidAt              *time.Time         `json:"paid_at"`
	FinishedAt          *time.Time         `json:"finished_at"`
	CancelledAt         *time.Time         `json:"cancelled_at"`
	CreatedAt           time.Time          `gorm:"index:idx_orders_bengkel_status_created,priority:3,sort:desc;index:idx_orders_user_status_created,priority:3,sort:desc;index:idx_orders_created_at,sort:desc" json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
}

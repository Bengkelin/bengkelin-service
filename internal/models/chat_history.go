package models

import "time"

type ChatHistory struct {
	ID             string    `gorm:"primary_key;type:varchar(36)" json:"id"`
	MessageText    string    `gorm:"type:text" json:"message_text"`
	Type           uint      `gorm:"type:int" json:"type"`
	ImageUrl       string    `gorm:"type:varchar(500)" json:"image_url"`
	SenderUserId   string    `gorm:"type:varchar(36);not null" json:"sender_user_id"`
	ReceiverUserId string    `gorm:"type:varchar(36);not null" json:"receiver_user_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

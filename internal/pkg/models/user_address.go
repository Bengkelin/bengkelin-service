package models

type UserAddress struct {
	ID           uint    `gorm:"primary_key;auto_increment" json:"address_id"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	AddressLabel string  `json:"address_label"`
	FullAddress  string  `json:"full_address"`
	Note         string  `json:"note"`
	UserID       string  `gorm:"type:varchar(36);index" json:"user_id"`
	User         User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

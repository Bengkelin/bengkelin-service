package models

type BengkelAddress struct {
	ID           uint    `gorm:"primary_key;auto_increment" json:"id"`
	BengkelID    string  `gorm:"type:varchar(36);index" json:"-"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	AddressLabel string  `json:"address_label"`
	FullAddress  string  `json:"full_address"`
	Note         string  `json:"note"`
	Bengkel      Bengkel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

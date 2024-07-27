package models

type BengkelAddress struct {
	BengkelID    string  `gorm:"primary_key;reference;type:varchar(36)" json:"-"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	AddressLabel string  `json:"address_label"`
	FullAddress  string  `json:"full_address"`
	Note         string  `json:"note"`
	Bengkel      Bengkel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

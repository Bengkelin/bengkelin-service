package models

type AddressMitra struct {
	ID           uint    `gorm:"primary_key;auto_increment" json:"address_id"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	AddressLabel string  `json:"address_label"`
	FullAddress  string  `json:"full_address"`
	Note         string  `json:"note"`
	MitraID      string  `gorm:"primary_key;reference" json:"mitra_id"`
	Mitra        Mitra   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
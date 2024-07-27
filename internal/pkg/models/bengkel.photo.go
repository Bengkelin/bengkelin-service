package models

type BengkelPhoto struct {
	BengkelID string  `gorm:"primary_key;reference;type:varchar(36)" json:"bengkel_id"`
	PhotoURL  string  `gorm:"type:varchar(255)" json:"photo_url"`
	Bengkel   Bengkel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

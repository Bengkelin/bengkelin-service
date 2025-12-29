package models

type BengkelPhoto struct {
	ID        uint    `gorm:"primary_key;auto_increment" json:"photo_id"`
	BengkelID string  `gorm:"type:varchar(36);index" json:"-"`
	PhotoURL  string  `gorm:"type:varchar(255)" json:"photo_url"`
	Bengkel   Bengkel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

package models

type BengkelService struct {
	BengkelID   string  `gorm:"primary_key;reference;type:varchar(36)" json:"bengkel_id"`
	NamaService string  `gorm:"type:varchar(100)" json:"nama_service"`
	Bengkel     Bengkel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

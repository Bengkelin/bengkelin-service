package models

type BengkelService struct {
	ID          uint    `gorm:"primary_key;auto_increment" json:"id"`
	BengkelID   string  `gorm:"type:varchar(36);index" json:"-"`
	NamaService string  `gorm:"type:varchar(100)" json:"nama_service"`
	Bengkel     Bengkel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

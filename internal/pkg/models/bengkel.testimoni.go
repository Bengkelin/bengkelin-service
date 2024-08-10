package models

type BengkelTestimoni struct {
	ID        uint    `gorm:"primary_key;autoIncrement" json:"id"`
	BengkelID string  `gorm:"primary_key;type:varchar(36)" json:"-"`
	PesananID string  `gorm:"type:varchar(36)" json:"pesanan_id"`
	UserID    string  `gorm:"reference;type:varchar(36)" json:"-"`
	Testimoni string  `gorm:"type:text" json:"testimoni"`
	Rating    int     `gorm:"type:int" json:"rating"`
	User      User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
	Bengkel   Bengkel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Pesanan   Pesanan `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

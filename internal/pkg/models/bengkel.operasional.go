package models

type BengkelOperasional struct {
	BengkelID string  `gorm:"primary_key;reference;type:varchar(36)" json:"-"`
	Hari      string  `gorm:"primary_key;type:varchar(10)" json:"hari"`
	JamBuka   string  `gorm:"type:varchar(20)" json:"jam_buka"`
	Buka      *bool   `gorm:"type:bool" json:"-"`
	Bengkel   Bengkel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

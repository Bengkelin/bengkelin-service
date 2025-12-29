package models

type BengkelOperational struct {
	ID        uint    `gorm:"primary_key;auto_increment" json:"id"`
	BengkelID string  `gorm:"type:varchar(36);index" json:"-"`
	Hari      string  `gorm:"type:varchar(10)" json:"hari"`
	JamBuka   string  `gorm:"type:varchar(20)" json:"jam_buka"`
	Bengkel   Bengkel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

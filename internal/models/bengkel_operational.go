package models

type BengkelOperational struct {
	ID        uint    `gorm:"primary_key;auto_increment" json:"id"`
	BengkelID string  `gorm:"type:varchar(36);index:idx_bengkel_operational_bengkel_id" json:"-"`
	Hari      string  `gorm:"type:varchar(10)" json:"hari"`
	JamBuka   string  `gorm:"type:varchar(20)" json:"jam_buka"`
	JamTutup  string  `gorm:"type:varchar(20)" json:"jam_tutup"`
	IsActive  *bool   `gorm:"type:boolean;default:true" json:"is_active"`
	Bengkel   Bengkel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

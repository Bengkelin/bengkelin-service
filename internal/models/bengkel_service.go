package models

type BengkelService struct {
	ID          uint    `gorm:"primary_key;auto_increment" json:"id"`
	BengkelID   string  `gorm:"type:varchar(36);index:idx_bengkel_services_bengkel_id" json:"-"`
	NamaService string  `gorm:"type:varchar(100)" json:"nama_service"`
	Description string  `gorm:"type:text" json:"description"`
	Price       float64 `gorm:"type:decimal(10,2)" json:"price"`
	IsAvailable *bool   `gorm:"type:boolean;default:true" json:"is_available"`
	Bengkel     Bengkel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

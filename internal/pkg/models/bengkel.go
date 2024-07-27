package models

type Bengkel struct {
	ID           string               `gorm:"primary_key;type:varchar(36)" json:"bengkel_id"`
	MitraID      string               `gorm:"reference" json:"-"`
	BengkelName  string               `gorm:"type:varchar(255)" json:"bengkel_name"`
	BengkelPhone string               `gorm:"type:varchar(255)" json:"bengkel_phone"`
	JumlahMontir uint                 `gorm:"type:int" json:"jumlah_montir"`
	HomeService  *bool                `gorm:"type:bool" json:"home_service"`
	StoreService *bool                `gorm:"type:bool" json:"store_service"`
	Operasionals []BengkelOperasional `gorm:"foreignKey:BengkelID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"operasionals"`
	Photos       []BengkelPhoto       `gorm:"foreignKey:BengkelID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"photos"`
	Services     []BengkelService     `gorm:"foreignKey:BengkelID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"services"`
	Addresses    []BengkelAddress     `gorm:"foreignKey:BengkelID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"addresses"`
}

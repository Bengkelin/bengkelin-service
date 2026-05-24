package models

type VehiclePhoto struct {
	ID        uint    `gorm:"primary_key;auto_increment" json:"photo_id"`
	VehicleID uint    `gorm:"type:uint;index:idx_vehicle_photos_vehicle_id" json:"vehicle_id"`
	Vehicle   Vehicle `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	PhotoURL  string  `json:"photo_url"`
}

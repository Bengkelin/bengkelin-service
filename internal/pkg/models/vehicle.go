package models

type Vehicle struct {
	ID          uint           `gorm:"primary_key;auto_increment" json:"vehicle_id"`
	VehicleType string         `json:"vehicle_type"`
	PlateNumber string         `json:"plate_number"`
	UserID      string         `gorm:"primary_key;reference" json:"user_id"`
	User        User           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Photos      []VehiclePhoto `gorm:"foreignKey:VehicleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"photos"`
}

package models

type Vehicle struct {
	ID            uint           `gorm:"primary_key;auto_increment" json:"vehicle_id"`
	VehicleType   string         `form:"vehicle_type" json:"vehicle_type"`
	VehicleNumber string         `form:"vehicle_number" json:"vehicle_number"`
	VehicleColor  string         `form:"vehicle_color" json:"vehicle_color"`
	UserID        string         `gorm:"type:varchar(36);index:idx_vehicles_user_id" json:"user_id"`
	User          User           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Photos        []VehiclePhoto `gorm:"foreignKey:VehicleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"photos"`
}

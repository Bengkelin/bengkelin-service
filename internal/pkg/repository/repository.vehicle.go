package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	vehicleRepository *VehicleRepository
)

type VehicleRepositoryInterface interface {
	CreateVehicle(vehicle models.Vehicle) (models.Vehicle, error)
	UpdateVehicleById(vehicleId uint, userId string, vehicle *models.Vehicle) error
	GetVehicleById(userId string) (*models.Vehicle, error)
}

type VehicleRepository struct{}

func GetVehicleRepository() VehicleRepositoryInterface {
	if vehicleRepository == nil {
		vehicleRepository = &VehicleRepository{}
	}
	return vehicleRepository
}

// CreateVehicle implements VehicleRepositoryInterface.
func (repo *VehicleRepository) CreateVehicle(vehicle models.Vehicle) (models.Vehicle, error) {
	err := Create(&vehicle)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.Vehicle{}, err
	}
	return vehicle, nil
}

// UpdateVehicleById implements VehicleRepositoryInterface.
func (*VehicleRepository) UpdateVehicleById(vehicleId uint, userId string, vehicle *models.Vehicle) error {
	err := db.GetDB().Model(&models.Vehicle{}).Where("id = ? AND user_id = ?", vehicleId, userId).Updates(vehicle).Error

	if err != nil {
		return err
	}
	return nil
}

// GetVehicleById implements VehicleRepositoryInterface.
func (*VehicleRepository) GetVehicleById(userId string) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	where := models.Vehicle{}
	where.UserID = userId
	_, err := First(where, &vehicle, []string{"Photos"})
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

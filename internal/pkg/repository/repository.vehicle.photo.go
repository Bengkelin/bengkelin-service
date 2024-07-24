package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	vehiclePhotoRepository *VehiclePhotoRepository
)

type VehiclePhotoRepositoryInterface interface {
	CreateVehiclePhoto(vehiclePicture models.VehiclePhoto) (models.VehiclePhoto, error)
	UpdateVehiclePhotoById(vehiclePhotoId uint, vehiclePhoto *models.VehiclePhoto) error
	GetVehiclePhotoById(vehicleId uint) (*models.VehiclePhoto, error)
}

type VehiclePhotoRepository struct{}

func GetVehiclePhotoRepository() VehiclePhotoRepositoryInterface {
	if vehiclePhotoRepository == nil {
		vehiclePhotoRepository = &VehiclePhotoRepository{}
	}
	return vehiclePhotoRepository
}

// CreateVehiclePhoto implements VehiclePhotoRepositoryInterface.
func (repo *VehiclePhotoRepository) CreateVehiclePhoto(vehiclePicture models.VehiclePhoto) (models.VehiclePhoto, error) {
	err := Create(&vehiclePicture)
	if err != nil {
		return models.VehiclePhoto{}, err
	}
	return vehiclePicture, nil
}

// UpdateVehiclePhotoById implements VehiclePhotoRepositoryInterface.
func (*VehiclePhotoRepository) UpdateVehiclePhotoById(vehiclePhotoId uint, vehiclePhoto *models.VehiclePhoto) error {
	err := db.GetDB().Model(&models.VehiclePhoto{}).Where("id = ?", vehiclePhotoId).Updates(vehiclePhoto).Error

	if err != nil {
		return err
	}
	return nil
}

// GetVehiclePhotoById implements VehiclePhotoRepositoryInterface.
func (*VehiclePhotoRepository) GetVehiclePhotoById(vehicleId uint) (*models.VehiclePhoto, error) {
	var vehiclePhoto models.VehiclePhoto
	where := models.VehiclePhoto{}
	where.VehicleID = vehicleId
	_, err := First(where, &vehiclePhoto, nil)
	if err != nil {
		return nil, err
	}
	return &vehiclePhoto, nil
}

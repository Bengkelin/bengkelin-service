package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/db"
	"github.com/Bengkelin/bengkelin-service/internal/models"
)

var (
	vehiclePhotoRepository *VehiclePhotoRepository
	vehiclePhotoOnce       sync.Once
)

type VehiclePhotoRepositoryInterface interface {
	CreateVehiclePhoto(ctx context.Context, vehiclePicture models.VehiclePhoto) (models.VehiclePhoto, error)
	UpdateVehiclePhotoById(ctx context.Context, vehiclePhotoId uint, vehiclePhoto *models.VehiclePhoto) error
	GetVehiclePhotoById(ctx context.Context, vehicleId uint) (*models.VehiclePhoto, error)
}

type VehiclePhotoRepository struct{}

func GetVehiclePhotoRepository() VehiclePhotoRepositoryInterface {
	vehiclePhotoOnce.Do(func() {
		vehiclePhotoRepository = &VehiclePhotoRepository{}
	})
	return vehiclePhotoRepository
}

// CreateVehiclePhoto implements VehiclePhotoRepositoryInterface.
func (repo *VehiclePhotoRepository) CreateVehiclePhoto(ctx context.Context, vehiclePicture models.VehiclePhoto) (models.VehiclePhoto, error) {
	err := Create(ctx, &vehiclePicture)
	if err != nil {
		return models.VehiclePhoto{}, err
	}
	return vehiclePicture, nil
}

// UpdateVehiclePhotoById implements VehiclePhotoRepositoryInterface.
func (*VehiclePhotoRepository) UpdateVehiclePhotoById(ctx context.Context, vehiclePhotoId uint, vehiclePhoto *models.VehiclePhoto) error {
	err := db.GetDB().WithContext(ctx).Model(&models.VehiclePhoto{}).Where("id = ?", vehiclePhotoId).Updates(vehiclePhoto).Error

	if err != nil {
		return err
	}
	return nil
}

// GetVehiclePhotoById implements VehiclePhotoRepositoryInterface.
func (*VehiclePhotoRepository) GetVehiclePhotoById(ctx context.Context, vehicleId uint) (*models.VehiclePhoto, error) {
	var vehiclePhoto models.VehiclePhoto
	where := models.VehiclePhoto{}
	where.VehicleID = vehicleId
	_, err := First(ctx, where, &vehiclePhoto, nil)
	if err != nil {
		return nil, err
	}
	return &vehiclePhoto, nil
}

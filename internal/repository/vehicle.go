package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/db"
	"github.com/Bengkelin/bengkelin-service/internal/models"
)

var (
	vehicleRepository *VehicleRepository
	vehicleOnce       sync.Once
)

type VehicleRepositoryInterface interface {
	CreateVehicle(ctx context.Context, vehicle models.Vehicle) (models.Vehicle, error)
	UpdateVehicleById(ctx context.Context, vehicleId uint, userId string, vehicle *models.Vehicle) error
	GetVehicleById(ctx context.Context, userId string, vehicleId uint) (*models.Vehicle, error)
	GetAllVehiclesByUserId(ctx context.Context, userId string) ([]models.Vehicle, error)
	DeleteVehicleById(ctx context.Context, vehicleId uint, userId string) error
}

type VehicleRepository struct{}

func GetVehicleRepository() VehicleRepositoryInterface {
	vehicleOnce.Do(func() {
		vehicleRepository = &VehicleRepository{}
	})
	return vehicleRepository
}

// CreateVehicle implements VehicleRepositoryInterface.
func (repo *VehicleRepository) CreateVehicle(ctx context.Context, vehicle models.Vehicle) (models.Vehicle, error) {
	err := Create(ctx, &vehicle)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.Vehicle{}, err
	}
	return vehicle, nil
}

// UpdateVehicleById implements VehicleRepositoryInterface.
func (*VehicleRepository) UpdateVehicleById(ctx context.Context, vehicleId uint, userId string, vehicle *models.Vehicle) error {
	err := db.GetDB().WithContext(ctx).Model(&models.Vehicle{}).Where("id = ? AND user_id = ?", vehicleId, userId).Updates(vehicle).Error

	if err != nil {
		return err
	}
	return nil
}

// GetVehicleById implements VehicleRepositoryInterface.
func (*VehicleRepository) GetVehicleById(ctx context.Context, userId string, vehicleId uint) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	where := models.Vehicle{}
	where.UserID = userId
	where.ID = vehicleId
	_, err := First(ctx, where, &vehicle, []string{"Photos"})
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

// GetAllVehiclesByUserId implements VehicleRepositoryInterface.
func (*VehicleRepository) GetAllVehiclesByUserId(ctx context.Context, userId string) ([]models.Vehicle, error) {
	var vehicles []models.Vehicle
	err := db.GetDB().WithContext(ctx).Where("user_id = ?", userId).Preload("Photos").Find(&vehicles).Error
	if err != nil {
		return nil, err
	}
	return vehicles, nil
}

// DeleteVehicleById implements VehicleRepositoryInterface.
func (*VehicleRepository) DeleteVehicleById(ctx context.Context, vehicleId uint, userId string) error {
	err := db.GetDB().WithContext(ctx).Where("id = ? AND user_id = ?", vehicleId, userId).Delete(&models.Vehicle{}).Error

	if err != nil {
		return err
	}
	return nil
}

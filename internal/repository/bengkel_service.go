package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/db"
	"github.com/Bengkelin/bengkelin-service/internal/models"
)

var (
	bengkelServiceRepository *BengkelServiceRepository
	bengkelServiceOnce       sync.Once
)

type BengkelServiceRepositoryInterface interface {
	CreateBengkelService(ctx context.Context, bengkelService models.BengkelService) (models.BengkelService, error)
	UpdateBengkelServiceById(ctx context.Context, bengkelServiceId string, bengkelService *models.BengkelService) error
	UpdateBengkelService(ctx context.Context, serviceId uint, bengkelService *models.BengkelService) error
	GetBengkelServiceById(ctx context.Context, bengkelServiceId string) (*models.BengkelService, error)
	GetBengkelServiceByServiceId(ctx context.Context, serviceId uint) (*models.BengkelService, error)
}

type BengkelServiceRepository struct{}

func GetBengkelServiceRepository() BengkelServiceRepositoryInterface {
	bengkelServiceOnce.Do(func() {
		bengkelServiceRepository = &BengkelServiceRepository{}
	})
	return bengkelServiceRepository
}

// CreateBengkelService implements BengkelServiceRepositoryInterface.
func (repo *BengkelServiceRepository) CreateBengkelService(ctx context.Context, bengkelService models.BengkelService) (models.BengkelService, error) {
	err := Create(ctx, &bengkelService)
	if err != nil {
		return models.BengkelService{}, err
	}
	return bengkelService, nil
}

// UpdateBengkelService implements BengkelServiceRepositoryInterface.
func (*BengkelServiceRepository) UpdateBengkelService(ctx context.Context, serviceId uint, bengkelService *models.BengkelService) error {
	err := db.GetDB().WithContext(ctx).Model(&models.BengkelService{}).Where("id = ?", serviceId).Updates(bengkelService).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelServiceByServiceId implements BengkelServiceRepositoryInterface.
func (*BengkelServiceRepository) GetBengkelServiceByServiceId(ctx context.Context, serviceId uint) (*models.BengkelService, error) {
	var bengkelService models.BengkelService
	err := db.GetDB().WithContext(ctx).Where("id = ?", serviceId).First(&bengkelService).Error
	if err != nil {
		return nil, err
	}
	return &bengkelService, nil
}

// UpdateBengkelServiceById implements BengkelServiceRepositoryInterface.
func (*BengkelServiceRepository) UpdateBengkelServiceById(ctx context.Context, bengkelServiceId string, bengkelService *models.BengkelService) error {
	err := db.GetDB().WithContext(ctx).Model(&models.BengkelService{}).Where("bengkel_id = ?", bengkelServiceId).Updates(bengkelService).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelServiceById implements BengkelServiceRepositoryInterface.
func (*BengkelServiceRepository) GetBengkelServiceById(ctx context.Context, bengkelServiceId string) (*models.BengkelService, error) {
	var bengkelService models.BengkelService
	where := models.BengkelService{}
	where.BengkelID = bengkelServiceId
	_, err := First(ctx, where, &bengkelService, nil)
	if err != nil {
		return nil, err
	}
	return &bengkelService, nil
}

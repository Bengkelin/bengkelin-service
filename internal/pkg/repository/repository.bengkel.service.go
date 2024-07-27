package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	bengkelServiceRepository *BengkelServiceRepository
)

type BengkelServiceRepositoryInterface interface {
	CreateBengkelService(bengkelService models.BengkelService) (models.BengkelService, error)
	UpdateBengkelServiceById(bengkelServiceId string, bengkelService *models.BengkelService) error
	GetBengkelServiceById(bengkelServiceId string) (*models.BengkelService, error)
}

type BengkelServiceRepository struct{}

func GetBengkelServiceRepository() BengkelServiceRepositoryInterface {
	if bengkelServiceRepository == nil {
		bengkelServiceRepository = &BengkelServiceRepository{}
	}
	return bengkelServiceRepository
}

// CreateBengkelService implements BengkelServiceRepositoryInterface.
func (repo *BengkelServiceRepository) CreateBengkelService(bengkelService models.BengkelService) (models.BengkelService, error) {
	err := Create(&bengkelService)
	if err != nil {
		return models.BengkelService{}, err
	}
	return bengkelService, nil
}

// UpdateBengkelServiceById implements BengkelServiceRepositoryInterface.
func (*BengkelServiceRepository) UpdateBengkelServiceById(bengkelServiceId string, bengkelService *models.BengkelService) error {
	err := db.GetDB().Model(&models.BengkelService{}).Where("bengkel_id = ?", bengkelServiceId).Updates(bengkelService).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelServiceById implements BengkelServiceRepositoryInterface.
func (*BengkelServiceRepository) GetBengkelServiceById(bengkelServiceId string) (*models.BengkelService, error) {
	var bengkelService models.BengkelService
	where := models.BengkelService{}
	where.BengkelID = bengkelServiceId
	_, err := First(where, &bengkelService, nil)
	if err != nil {
		return nil, err
	}
	return &bengkelService, nil
}

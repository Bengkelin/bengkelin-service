package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	pesananServiceRepository *PesananServiceRepository
)

type PesananServiceRepositoryInterface interface {
	CreatePesananService(pesananService models.PesananService) (models.PesananService, error)
	UpdatePesananServiceById(pesananServiceId string, pesananService *models.PesananService) error
	GetAllPesananService() ([]models.PesananService, error)
	GetAllPesananServicePaginate(page int, limit int, userId string) ([]models.PesananService, int, error)
}

type PesananServiceRepository struct{}

func GetPesananServiceRepository() PesananServiceRepositoryInterface {
	if pesananServiceRepository == nil {
		pesananServiceRepository = &PesananServiceRepository{}
	}
	return pesananServiceRepository
}

// CreatePesananService implements PesananServiceRepositoryInterface.
func (repo *PesananServiceRepository) CreatePesananService(pesananService models.PesananService) (models.PesananService, error) {
	err := db.GetDB().Create(&pesananService).Error
	if err != nil {
		return models.PesananService{}, err
	}
	return pesananService, nil
}

// UpdatePesananServiceById implements PesananServiceRepositoryInterface.
func (*PesananServiceRepository) UpdatePesananServiceById(pesananServiceId string, pesananService *models.PesananService) error {
	err := db.GetDB().Model(&models.PesananService{}).Where("id = ?", pesananServiceId).Updates(pesananService).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAllPesananService implements PesananServiceRepositoryInterface.
func (*PesananServiceRepository) GetAllPesananService() ([]models.PesananService, error) {
	var pesananService []models.PesananService
	err := db.GetDB().Find(&pesananService).Error
	if err != nil {
		return nil, err
	}
	return pesananService, nil
}

// GetAllPesananServicePaginate implements PesananServiceRepositoryInterface.
func (*PesananServiceRepository) GetAllPesananServicePaginate(page int, limit int, userId string) ([]models.PesananService, int, error) {
	var pesananService []models.PesananService
	var total int64
	err := db.GetDB().Model(&models.PesananService{}).Where("user_id = ?", userId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.GetDB().Where("user_id = ?", userId).Offset((page - 1) * limit).Limit(limit).Find(&pesananService).Error
	if err != nil {
		return nil, 0, err
	}

	return pesananService, int(total), nil
}
	
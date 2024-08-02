package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	bengkelTestimoniRepository *BengkelTestimoniRepository
)

type BengkelTestimoniRepositoryInterface interface {
	CreateBengkelTestimoni(bengkelTestimoni models.BengkelTestimoni) (models.BengkelTestimoni, error)
	UpdateBengkelTestimoniById(bengkelTestimoniId string, bengkelTestimoni *models.BengkelTestimoni) error
	GetBengkelTestimoniById(bengkelTestimoniId string) (*models.BengkelTestimoni, error)
	GetAllBengkelTestimoniPaginate(page int, limit int) ([]models.BengkelTestimoni, int, error)
}

type BengkelTestimoniRepository struct{}

func GetBengkelTestimoniRepository() BengkelTestimoniRepositoryInterface {
	if bengkelTestimoniRepository == nil {
		bengkelTestimoniRepository = &BengkelTestimoniRepository{}
	}
	return bengkelTestimoniRepository
}

// CreateBengkelTestimoni implements BengkelTestimoniRepositoryInterface.
func (repo *BengkelTestimoniRepository) CreateBengkelTestimoni(bengkelTestimoni models.BengkelTestimoni) (models.BengkelTestimoni, error) {
	err := db.GetDB().Create(&bengkelTestimoni).Error
	if err != nil {
		return models.BengkelTestimoni{}, err
	}
	return bengkelTestimoni, nil
}

// UpdateBengkelTestimoniById implements BengkelTestimoniRepositoryInterface.
func (*BengkelTestimoniRepository) UpdateBengkelTestimoniById(bengkelTestimoniId string, bengkelTestimoni *models.BengkelTestimoni) error {
	err := db.GetDB().Model(&models.BengkelTestimoni{}).Where("id = ?", bengkelTestimoniId).Updates(bengkelTestimoni).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelTestimoniById implements BengkelTestimoniRepositoryInterface.
func (*BengkelTestimoniRepository) GetBengkelTestimoniById(bengkelTestimoniId string) (*models.BengkelTestimoni, error) {
	var bengkelTestimoni models.BengkelTestimoni
	where := models.BengkelTestimoni{}
	where.UserID = bengkelTestimoniId
	err := db.GetDB().First(where, &bengkelTestimoni, nil).Error
	if err != nil {
		return nil, err
	}
	return &bengkelTestimoni, nil
}

// GetAllBengkelTestimoniPaginate implements BengkelTestimoniRepositoryInterface.
func (*BengkelTestimoniRepository) GetAllBengkelTestimoniPaginate(page int, limit int) ([]models.BengkelTestimoni, int, error) {
	var bengkelTestimoni []models.BengkelTestimoni
	var total int64
	err := db.GetDB().Model(&models.BengkelTestimoni{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.GetDB().Offset((page - 1) * limit).Limit(limit).Find(&bengkelTestimoni).Error
	if err != nil {
		return nil, 0, err
	}
	return bengkelTestimoni, int(total), nil
}

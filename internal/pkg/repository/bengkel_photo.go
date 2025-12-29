package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	bengkelPhotoRepository *BengkelPhotoRepository
)

type BengkelPhotoRepositoryInterface interface {
	CreateBengkelPhoto(bengkelPhoto models.BengkelPhoto) (models.BengkelPhoto, error)

	UpdateBengkelPhotoById(bengkelPhotoId string, bengkelPhoto *models.BengkelPhoto) error

	GetBengkelPhotoById(bengkelId string) (*models.BengkelPhoto, error)
}

type BengkelPhotoRepository struct{}

func GetBengkelPhotoRepository() BengkelPhotoRepositoryInterface {
	if bengkelPhotoRepository == nil {
		bengkelPhotoRepository = &BengkelPhotoRepository{}
	}
	return bengkelPhotoRepository
}

// CreateBengkelPhoto implements BengkelPhotoRepositoryInterface.
func (repo *BengkelPhotoRepository) CreateBengkelPhoto(bengkelPhoto models.BengkelPhoto) (models.BengkelPhoto, error) {
	err := Create(&bengkelPhoto)
	if err != nil {
		return models.BengkelPhoto{}, err
	}
	return bengkelPhoto, nil
}

// UpdateBengkelPhotoById implements BengkelPhotoRepositoryInterface.
func (*BengkelPhotoRepository) UpdateBengkelPhotoById(bengkelPhotoId string, bengkelPhoto *models.BengkelPhoto) error {
	err := db.GetDB().Model(&models.BengkelPhoto{}).Where("bengkel_id = ?", bengkelPhotoId).Updates(bengkelPhoto).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelPhotoById implements BengkelPhotoRepositoryInterface.
func (*BengkelPhotoRepository) GetBengkelPhotoById(bengkelId string) (*models.BengkelPhoto, error) {
	var bengkelPhoto models.BengkelPhoto
	where := models.BengkelPhoto{}
	where.BengkelID = bengkelId
	_, err := First(where, &bengkelPhoto, nil)
	if err != nil {
		return nil, err
	}
	return &bengkelPhoto, nil
}
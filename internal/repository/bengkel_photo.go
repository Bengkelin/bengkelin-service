package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/db"
	"github.com/Bengkelin/bengkelin-service/internal/models"
)

var (
	bengkelPhotoRepository *BengkelPhotoRepository
	bengkelPhotoOnce       sync.Once
)

type BengkelPhotoRepositoryInterface interface {
	CreateBengkelPhoto(ctx context.Context, bengkelPhoto models.BengkelPhoto) (models.BengkelPhoto, error)

	UpdateBengkelPhotoById(ctx context.Context, bengkelPhotoId string, bengkelPhoto *models.BengkelPhoto) error

	GetBengkelPhotoById(ctx context.Context, bengkelId string) (*models.BengkelPhoto, error)

	GetBengkelPhotoByPK(ctx context.Context, photoId string) (*models.BengkelPhoto, error)

	DeleteBengkelPhotoById(ctx context.Context, photoId string) error
}

type BengkelPhotoRepository struct{}

func GetBengkelPhotoRepository() BengkelPhotoRepositoryInterface {
	bengkelPhotoOnce.Do(func() {
		bengkelPhotoRepository = &BengkelPhotoRepository{}
	})
	return bengkelPhotoRepository
}

// CreateBengkelPhoto implements BengkelPhotoRepositoryInterface.
func (repo *BengkelPhotoRepository) CreateBengkelPhoto(ctx context.Context, bengkelPhoto models.BengkelPhoto) (models.BengkelPhoto, error) {
	err := Create(ctx, &bengkelPhoto)
	if err != nil {
		return models.BengkelPhoto{}, err
	}
	return bengkelPhoto, nil
}

// UpdateBengkelPhotoById implements BengkelPhotoRepositoryInterface.
func (*BengkelPhotoRepository) UpdateBengkelPhotoById(ctx context.Context, bengkelPhotoId string, bengkelPhoto *models.BengkelPhoto) error {
	err := db.GetDB().WithContext(ctx).Model(&models.BengkelPhoto{}).Where("bengkel_id = ?", bengkelPhotoId).Updates(bengkelPhoto).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelPhotoById implements BengkelPhotoRepositoryInterface.
func (*BengkelPhotoRepository) GetBengkelPhotoById(ctx context.Context, bengkelId string) (*models.BengkelPhoto, error) {
	var bengkelPhoto models.BengkelPhoto
	where := models.BengkelPhoto{}
	where.BengkelID = bengkelId
	_, err := First(ctx, where, &bengkelPhoto, nil)
	if err != nil {
		return nil, err
	}
	return &bengkelPhoto, nil
}

// DeleteBengkelPhotoById implements BengkelPhotoRepositoryInterface.
func (*BengkelPhotoRepository) DeleteBengkelPhotoById(ctx context.Context, photoId string) error {
	return db.GetDB().WithContext(ctx).Delete(&models.BengkelPhoto{}, "id = ?", photoId).Error
}

// GetBengkelPhotoByPK implements BengkelPhotoRepositoryInterface.
func (*BengkelPhotoRepository) GetBengkelPhotoByPK(ctx context.Context, photoId string) (*models.BengkelPhoto, error) {
	var photo models.BengkelPhoto
	if err := db.GetDB().WithContext(ctx).First(&photo, "id = ?", photoId).Error; err != nil {
		return nil, err
	}
	return &photo, nil
}

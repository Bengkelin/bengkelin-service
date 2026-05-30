package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/db"
	"github.com/Bengkelin/bengkelin-service/internal/models"
)

var (
	bengkelTestimonialRepository *BengkelTestimonialRepository
	bengkelTestimonialOnce       sync.Once
)

type BengkelTestimonialRepositoryInterface interface {
	CreateBengkelTestimonial(ctx context.Context, bengkelTestimonial models.BengkelTestimonial) (models.BengkelTestimonial, error)
	UpdateBengkelTestimonialById(ctx context.Context, bengkelTestimonialId string, bengkelTestimonial *models.BengkelTestimonial) error
	GetBengkelTestimonialById(ctx context.Context, bengkelTestimonialId string) (*models.BengkelTestimonial, error)
	GetAllBengkelTestimonialPaginate(ctx context.Context, page int, limit int) ([]models.BengkelTestimonial, int, error)
}

type BengkelTestimonialRepository struct{}

func GetBengkelTestimonialRepository() BengkelTestimonialRepositoryInterface {
	bengkelTestimonialOnce.Do(func() {
		bengkelTestimonialRepository = &BengkelTestimonialRepository{}
	})
	return bengkelTestimonialRepository
}

// CreateBengkelTestimonial implements BengkelTestimonialRepositoryInterface.
func (repo *BengkelTestimonialRepository) CreateBengkelTestimonial(ctx context.Context, bengkelTestimonial models.BengkelTestimonial) (models.BengkelTestimonial, error) {
	err := db.GetDB().WithContext(ctx).Create(&bengkelTestimonial).Error
	if err != nil {
		return models.BengkelTestimonial{}, err
	}
	return bengkelTestimonial, nil
}

// UpdateBengkelTestimonialById implements BengkelTestimonialRepositoryInterface.
func (*BengkelTestimonialRepository) UpdateBengkelTestimonialById(ctx context.Context, bengkelTestimonialId string, bengkelTestimonial *models.BengkelTestimonial) error {
	err := db.GetDB().WithContext(ctx).Model(&models.BengkelTestimonial{}).Where("id = ?", bengkelTestimonialId).Updates(bengkelTestimonial).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelTestimonialById implements BengkelTestimonialRepositoryInterface.
func (*BengkelTestimonialRepository) GetBengkelTestimonialById(ctx context.Context, bengkelTestimonialId string) (*models.BengkelTestimonial, error) {
	var bengkelTestimonial models.BengkelTestimonial
	where := models.BengkelTestimonial{}
	where.UserID = bengkelTestimonialId
	err := db.GetDB().WithContext(ctx).First(where, &bengkelTestimonial, nil).Error
	if err != nil {
		return nil, err
	}
	return &bengkelTestimonial, nil
}

// GetAllBengkelTestimonialPaginate implements BengkelTestimonialRepositoryInterface.
func (*BengkelTestimonialRepository) GetAllBengkelTestimonialPaginate(ctx context.Context, page int, limit int) ([]models.BengkelTestimonial, int, error) {
	var bengkelTestimonials []models.BengkelTestimonial
	var total int64
	err := db.GetDB().WithContext(ctx).Model(&models.BengkelTestimonial{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.GetDB().WithContext(ctx).Offset((page - 1) * limit).Limit(limit).Find(&bengkelTestimonials).Error
	if err != nil {
		return nil, 0, err
	}
	return bengkelTestimonials, int(total), nil
}

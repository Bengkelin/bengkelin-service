package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	bengkelOperationalRepository *BengkelOperationalRepository
	bengkelOperationalOnce       sync.Once
)

type BengkelOperationalRepositoryInterface interface {
	CreateBengkelOperational(ctx context.Context, bengkelOperational models.BengkelOperational) (models.BengkelOperational, error)
	UpdateBengkelOperationalById(ctx context.Context, bengkelOperationalId, bengkelOperationalHari string, bengkelOperational *models.BengkelOperational) error
	GetBengkelOperationalById(ctx context.Context, bengkelId string) (*models.BengkelOperational, error)
	GetBengkelOperationalByIdAndDay(ctx context.Context, bengkelId, day string) (*models.BengkelOperational, error)
}

type BengkelOperationalRepository struct{}

func GetBengkelOperationalRepository() BengkelOperationalRepositoryInterface {
	bengkelOperationalOnce.Do(func() {
		bengkelOperationalRepository = &BengkelOperationalRepository{}
	})
	return bengkelOperationalRepository
}

// CreateBengkelOperational implements BengkelOperationalRepositoryInterface.
func (repo *BengkelOperationalRepository) CreateBengkelOperational(ctx context.Context, bengkelOperational models.BengkelOperational) (models.BengkelOperational, error) {
	err := Create(ctx, &bengkelOperational)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.BengkelOperational{}, err
	}
	return bengkelOperational, nil
}

// UpdateBengkelOperationalById implements BengkelOperationalRepositoryInterface.
func (*BengkelOperationalRepository) UpdateBengkelOperationalById(ctx context.Context, bengkelOperationalId, bengkelOperationalHari string, bengkelOperational *models.BengkelOperational) error {
	err := db.GetDB().WithContext(ctx).Model(&models.BengkelOperational{}).Where("bengkel_id = ? and hari = ?", bengkelOperationalId, bengkelOperationalHari).Updates(bengkelOperational).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelOperationalById implements BengkelOperationalRepositoryInterface.
func (*BengkelOperationalRepository) GetBengkelOperationalById(ctx context.Context, bengkelId string) (*models.BengkelOperational, error) {
	var BengkelOperational models.BengkelOperational
	where := models.BengkelOperational{}
	where.BengkelID = bengkelId
	_, err := First(ctx, where, &BengkelOperational, nil)
	if err != nil {
		return nil, err
	}
	return &BengkelOperational, nil
}

// GetBengkelOperationalByIdAndDay implements BengkelOperationalRepositoryInterface.
func (*BengkelOperationalRepository) GetBengkelOperationalByIdAndDay(ctx context.Context, bengkelId, day string) (*models.BengkelOperational, error) {
	var BengkelOperational models.BengkelOperational
	err := db.GetDB().WithContext(ctx).Where("bengkel_id = ? AND hari LIKE ?", bengkelId, "%"+day+"%").First(&BengkelOperational).Error
	if err != nil {
		return nil, err
	}
	return &BengkelOperational, nil
}

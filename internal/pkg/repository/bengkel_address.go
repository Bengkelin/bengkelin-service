package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	bengkelAddressRepository *BengkelAddressRepository
	bengkelAddressOnce       sync.Once
)

type BengkelAddressRepositoryInterface interface {
	CreateBengkelAddress(ctx context.Context, bengkelAddress models.BengkelAddress) (models.BengkelAddress, error)
	UpdateBengkelAddressById(ctx context.Context, bengkelAddressId string, bengkelAddress *models.BengkelAddress) error
	GetBengkelAddressById(ctx context.Context, bengkelAddressId string) (*models.BengkelAddress, error)
}

type BengkelAddressRepository struct{}

func GetBengkelAddressRepository() BengkelAddressRepositoryInterface {
	bengkelAddressOnce.Do(func() {
		bengkelAddressRepository = &BengkelAddressRepository{}
	})
	return bengkelAddressRepository
}

// CreateBengkelAddress implements BengkelAddressRepositoryInterface.

func (repo *BengkelAddressRepository) CreateBengkelAddress(ctx context.Context, bengkelAddress models.BengkelAddress) (models.BengkelAddress, error) {
	err := Create(ctx, &bengkelAddress)
	if err != nil {
		return models.BengkelAddress{}, err
	}
	return bengkelAddress, nil
}

// UpdateBengkelAddressById implements BengkelAddressRepositoryInterface.
func (*BengkelAddressRepository) UpdateBengkelAddressById(ctx context.Context, bengkelAddressId string, bengkelAddress *models.BengkelAddress) error {
	err := db.GetDB().WithContext(ctx).Model(&models.BengkelAddress{}).Where("bengkel_id = ?", bengkelAddressId).Updates(bengkelAddress).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelAddressById implements BengkelAddressRepositoryInterface.

func (*BengkelAddressRepository) GetBengkelAddressById(ctx context.Context, bengkelAddressId string) (*models.BengkelAddress, error) {
	var bengkelAddress models.BengkelAddress
	where := models.BengkelAddress{}
	where.BengkelID = bengkelAddressId
	_, err := First(ctx, where, &bengkelAddress, nil)
	if err != nil {
		return nil, err
	}
	return &bengkelAddress, nil
}

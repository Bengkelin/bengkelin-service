package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	bengkelAddressRepository *BengkelAddressRepository
)

type BengkelAddressRepositoryInterface interface {
	CreateBengkelAddress(bengkelAddress models.BengkelAddress) (models.BengkelAddress, error)
	UpdateBengkelAddressById(bengkelAddressId string, bengkelAddress *models.BengkelAddress) error
	GetBengkelAddressById(bengkelAddressId string) (*models.BengkelAddress, error)
}

type BengkelAddressRepository struct{}

func GetBengkelAddressRepository() BengkelAddressRepositoryInterface {
	if bengkelAddressRepository == nil {
		bengkelAddressRepository = &BengkelAddressRepository{}
	}
	return bengkelAddressRepository
}

// CreateBengkelAddress implements BengkelAddressRepositoryInterface.

func (repo *BengkelAddressRepository) CreateBengkelAddress(bengkelAddress models.BengkelAddress) (models.BengkelAddress, error) {
	err := Create(&bengkelAddress)
	if err != nil {
		return models.BengkelAddress{}, err
	}
	return bengkelAddress, nil
}

// UpdateBengkelAddressById implements BengkelAddressRepositoryInterface.
func (*BengkelAddressRepository) UpdateBengkelAddressById(bengkelAddressId string, bengkelAddress *models.BengkelAddress) error {
	err := db.GetDB().Model(&models.BengkelAddress{}).Where("bengkel_id = ?", bengkelAddressId).Updates(bengkelAddress).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelAddressById implements BengkelAddressRepositoryInterface.

func (*BengkelAddressRepository) GetBengkelAddressById(bengkelAddressId string) (*models.BengkelAddress, error) {
	var bengkelAddress models.BengkelAddress
	where := models.BengkelAddress{}
	where.BengkelID = bengkelAddressId
	_, err := First(where, &bengkelAddress, nil)
	if err != nil {
		return nil, err
	}
	return &bengkelAddress, nil
}

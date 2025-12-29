package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	bengkelOperationalRepository *BengkelOperationalRepository
)

type BengkelOperationalRepositoryInterface interface {
	CreateBengkelOperational(bengkelOperational models.BengkelOperational) (models.BengkelOperational, error)
	UpdateBengkelOperationalById(bengkelOperationalId, bengkelOperationalHari string, bengkelOperational *models.BengkelOperational) error
	GetBengkelOperationalById(bengkelId string) (*models.BengkelOperational, error)
	GetBengkelOperationalByIdAndDay(bengkelId, day string) (*models.BengkelOperational, error)
}

type BengkelOperationalRepository struct{}

func GetBengkelOperationalRepository() BengkelOperationalRepositoryInterface {
	if bengkelOperationalRepository == nil {
		bengkelOperationalRepository = &BengkelOperationalRepository{}
	}
	return bengkelOperationalRepository
}

// CreateBengkelOperational implements BengkelOperationalRepositoryInterface.
func (repo *BengkelOperationalRepository) CreateBengkelOperational(bengkelOperational models.BengkelOperational) (models.BengkelOperational, error) {
	err := Create(&bengkelOperational)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.BengkelOperational{}, err
	}
	return bengkelOperational, nil
}

// UpdateBengkelOperationalById implements BengkelOperationalRepositoryInterface.
func (*BengkelOperationalRepository) UpdateBengkelOperationalById(bengkelOperationalId, bengkelOperationalHari string, bengkelOperational *models.BengkelOperational) error {
	err := db.GetDB().Model(&models.BengkelOperational{}).Where("bengkel_id = ? and hari = ?", bengkelOperationalId, bengkelOperationalHari).Updates(bengkelOperational).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelOperationalById implements BengkelOperationalRepositoryInterface.
func (*BengkelOperationalRepository) GetBengkelOperationalById(bengkelId string) (*models.BengkelOperational, error) {
	var BengkelOperational models.BengkelOperational
	where := models.BengkelOperational{}
	where.BengkelID = bengkelId
	_, err := First(where, &BengkelOperational, nil)
	if err != nil {
		return nil, err
	}
	return &BengkelOperational, nil
}

// GetBengkelOperationalByIdAndDay implements BengkelOperationalRepositoryInterface.
func (*BengkelOperationalRepository) GetBengkelOperationalByIdAndDay(bengkelId, day string) (*models.BengkelOperational, error) {
	var BengkelOperational models.BengkelOperational
	err := db.GetDB().Where("bengkel_id = ? AND hari LIKE ?", bengkelId, "%"+day+"%").First(&BengkelOperational).Error
	if err != nil {
		return nil, err
	}
	return &BengkelOperational, nil
}

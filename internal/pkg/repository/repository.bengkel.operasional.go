package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	bengkelOperasionalRepository *BengkelOperasionalRepository
)

type BengkelOperasionalRepositoryInterface interface {
	CreateBengkelOperasional(bengkelOperasional models.BengkelOperasional) (models.BengkelOperasional, error)
	UpdateBengkelOperasionalById(bengkelOperasionalId string, bengkelOperasional *models.BengkelOperasional) error
	GetBengkelOperasionalById(bengkelId string) (*models.BengkelOperasional, error)
}

type BengkelOperasionalRepository struct{}

func GetBengkelOperasionalRepository() BengkelOperasionalRepositoryInterface {
	if bengkelOperasionalRepository == nil {
		bengkelOperasionalRepository = &BengkelOperasionalRepository{}
	}
	return bengkelOperasionalRepository
}

// CreateBengkelOperasional implements BengkelOperasionalRepositoryInterface.
func (repo *BengkelOperasionalRepository) CreateBengkelOperasional(bengkelOperasional models.BengkelOperasional) (models.BengkelOperasional, error) {
	err := Create(&bengkelOperasional)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.BengkelOperasional{}, err
	}
	return bengkelOperasional, nil
}	

// UpdateBengkelOperasionalById implements BengkelOperasionalRepositoryInterface.
func (*BengkelOperasionalRepository) UpdateBengkelOperasionalById(bengkelOperasionalId string, bengkelOperasional *models.BengkelOperasional) error {
	err := db.GetDB().Model(&models.BengkelOperasional{}).Where("bengkel_id = ?", bengkelOperasionalId).Updates(bengkelOperasional).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelOperasionalById implements BengkelOperasionalRepositoryInterface.
func (*BengkelOperasionalRepository) GetBengkelOperasionalById(bengkelId string) (*models.BengkelOperasional, error) {
	var bengkelOperasional models.BengkelOperasional
	where := models.BengkelOperasional{}
	where.BengkelID = bengkelId
	_, err := First(where, &bengkelOperasional, nil)
	if err != nil {
		return nil, err
	}
	return &bengkelOperasional, nil
}
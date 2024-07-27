package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	bengkelRepository *BengkelRepository
)

type BengkelRepositoryInterface interface {
	CreateBengkel(bengkel models.Bengkel) (models.Bengkel, error)
	UpdateBengkelById(bengkelId string, bengkel *models.Bengkel) error
	GetBengkelById(bengkelId string) (*models.Bengkel, error)
}

type BengkelRepository struct{}

func GetBengkelRepository() BengkelRepositoryInterface {
	if bengkelRepository == nil {
		bengkelRepository = &BengkelRepository{}
	}
	return bengkelRepository
}

// CreateBengkel implements BengkelRepositoryInterface.
func (repo *BengkelRepository) CreateBengkel(bengkel models.Bengkel) (models.Bengkel, error) {
	err := Create(&bengkel)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.Bengkel{}, err
	}
	return bengkel, nil
}

// UpdateBengkelById implements BengkelRepositoryInterface.
func (*BengkelRepository) UpdateBengkelById(bengkelId string, bengkel *models.Bengkel) error {
	err := db.GetDB().Model(&models.Bengkel{}).Where("id = ?", bengkelId).Updates(bengkel).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelById implements BengkelRepositoryInterface.
func (*BengkelRepository) GetBengkelById(bengkelId string) (*models.Bengkel, error) {
	var bengkel models.Bengkel
	where := models.Bengkel{}
	where.ID = bengkelId
	_, err := First(where, &bengkel, []string{"Photos"})
	if err != nil {
		return nil, err
	}
	return &bengkel, nil
}

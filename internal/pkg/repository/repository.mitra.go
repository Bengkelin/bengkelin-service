package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	mitraRepository *MitraRepository
)

type MitraRepositoryInterface interface {
	FindMitraByEmail(email string) (*models.Mitra, error)
	FindMitraByID(mitraID string) (*models.Mitra, error)
	GetMitraByID(mitraID string) (*models.Mitra, error)
	CreateMitra(mitra models.Mitra) (models.Mitra, error)
	UpdateMitra(mitraID string, mitra *models.Mitra) error
}

type MitraRepository struct{}

// CreateMitra implements MitraRepositoryInterface.
func (repo *MitraRepository) CreateMitra(mitra models.Mitra) (models.Mitra, error) {
	err := Create(&mitra)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.Mitra{}, err
	}
	return mitra, nil
}

// FindMitraByEmail implements MitraRepositoryInterface.
func (*MitraRepository) FindMitraByEmail(email string) (*models.Mitra, error) {
	var mitra models.Mitra
	where := models.Mitra{}
	where.Email = email
	_, err := First(where, &mitra, nil)
	if err != nil {
		return nil, err
	}
	return &mitra, nil
}

// FindMitraByID implements MitraRepositoryInterface.
func (*MitraRepository) FindMitraByID(mitraID string) (*models.Mitra, error) {
	var mitra models.Mitra
	where := models.Mitra{}
	where.ID = mitraID
	_, err := First(where, &mitra, []string{"Bengkel", "Bengkel.Photos", "Bengkel.Operasionals", "Bengkel.Services", "Bengkel.Addresses"})
	if err != nil {
		return nil, err
	}
	return &mitra, nil
}

// GetMitraByID implements MitraRepositoryInterface.
func (*MitraRepository) GetMitraByID(mitraID string) (*models.Mitra, error) {
	var mitra models.Mitra
	where := models.Mitra{}
	where.ID = mitraID
	_, err := First(where, &mitra, []string{"Bengkel"})
	if err != nil {
		return nil, err
	}
	return &mitra, nil
}

// UpdateMitra implements MitraRepositoryInterface.
func (*MitraRepository) UpdateMitra(mitraID string, mitra *models.Mitra) error {
	err := db.GetDB().Model(&models.Mitra{}).Where("id = ?", mitraID).Updates(mitra).Error

	if err != nil {
		return err
	}
	return nil
}

func GetMitraRepository() MitraRepositoryInterface {
	if mitraRepository == nil {
		mitraRepository = &MitraRepository{}
	}
	return mitraRepository
}
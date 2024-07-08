package repository

import "github.com/Bengkelin/bengkelin-service/internal/pkg/models"

var (
	mitraRepository *MitraRepository
)

type MitraRepositoryInterface interface {
	FindMitraByEmail(email string) (*models.Mitra, error)
	FindMitraByID(mitraID string) (*models.Mitra, error)
	CreateMitra(mitra models.Mitra) (models.Mitra, error)
	UpdateMitra(mitra *models.Mitra) error
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
	_, err := First(where, &mitra, nil)
	if err != nil {
		return nil, err
	}
	return &mitra, nil
}

// UpdateMitra implements MitraRepositoryInterface.
func (*MitraRepository) UpdateMitra(mitra *models.Mitra) error {
	panic("unimplemented")
}

func GetMitraRepository() MitraRepositoryInterface {
	if mitraRepository == nil {
		mitraRepository = &MitraRepository{}
	}
	return mitraRepository
}
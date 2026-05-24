package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	mitraRepository *MitraRepository
	mitraOnce       sync.Once
)

type MitraRepositoryInterface interface {
	FindMitraByEmail(ctx context.Context, email string) (*models.Mitra, error)
	FindMitraByID(ctx context.Context, mitraID string) (*models.Mitra, error)
	GetMitraByID(ctx context.Context, mitraID string) (*models.Mitra, error)
	CreateMitra(ctx context.Context, mitra models.Mitra) (models.Mitra, error)
	UpdateMitra(ctx context.Context, mitraID string, mitra *models.Mitra) error
}

type MitraRepository struct{}

// CreateMitra implements MitraRepositoryInterface.
func (repo *MitraRepository) CreateMitra(ctx context.Context, mitra models.Mitra) (models.Mitra, error) {
	err := Create(ctx, &mitra)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.Mitra{}, err
	}
	return mitra, nil
}

// FindMitraByEmail implements MitraRepositoryInterface.
func (*MitraRepository) FindMitraByEmail(ctx context.Context, email string) (*models.Mitra, error) {
	var mitra models.Mitra
	where := models.Mitra{}
	where.Email = email
	_, err := First(ctx, where, &mitra, nil)
	if err != nil {
		return nil, err
	}
	return &mitra, nil
}

// FindMitraByID implements MitraRepositoryInterface.
func (*MitraRepository) FindMitraByID(ctx context.Context, mitraID string) (*models.Mitra, error) {
	var mitra models.Mitra
	where := models.Mitra{}
	where.ID = mitraID
	_, err := First(ctx, where, &mitra, []string{"Bengkel", "Bengkel.Photos", "Bengkel.Operasionals", "Bengkel.Services", "Bengkel.Addresses"})
	if err != nil {
		return nil, err
	}
	return &mitra, nil
}

// GetMitraByID implements MitraRepositoryInterface.
func (*MitraRepository) GetMitraByID(ctx context.Context, mitraID string) (*models.Mitra, error) {
	var mitra models.Mitra
	where := models.Mitra{}
	where.ID = mitraID
	_, err := First(ctx, where, &mitra, []string{"Bengkel"})
	if err != nil {
		return nil, err
	}
	return &mitra, nil
}

// UpdateMitra implements MitraRepositoryInterface.
func (*MitraRepository) UpdateMitra(ctx context.Context, mitraID string, mitra *models.Mitra) error {
	err := db.GetDB().WithContext(ctx).Model(&models.Mitra{}).Where("id = ?", mitraID).Updates(mitra).Error

	if err != nil {
		return err
	}
	return nil
}

func GetMitraRepository() MitraRepositoryInterface {
	mitraOnce.Do(func() {
		mitraRepository = &MitraRepository{}
	})
	return mitraRepository
}

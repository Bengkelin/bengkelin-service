package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	adminFeeRepository *AdminFeeRepository
	adminFeeOnce       sync.Once
)

type AdminFeeRepositoryInterface interface {
	CreateAdminFee(ctx context.Context, adminFee models.AdminFee) (models.AdminFee, error)
	UpdateAdminFeeById(ctx context.Context, adminFeeId string, adminFee *models.AdminFee) error
	GetAdminFeeById(ctx context.Context, adminFeeId string) (*models.AdminFee, error)
	GetOneAdminFeeLatest(ctx context.Context) (*models.AdminFee, error)
}

type AdminFeeRepository struct{}

func GetAdminFeeRepository() AdminFeeRepositoryInterface {
	adminFeeOnce.Do(func() {
		adminFeeRepository = &AdminFeeRepository{}
	})
	return adminFeeRepository
}

// CreateAdminFee implements AdminFeeRepositoryInterface.

func (repo *AdminFeeRepository) CreateAdminFee(ctx context.Context, adminFee models.AdminFee) (models.AdminFee, error) {
	err := Create(ctx, &adminFee)
	if err != nil {
		return models.AdminFee{}, err
	}
	return adminFee, nil
}

// UpdateAdminFeeById implements AdminFeeRepositoryInterface.
func (*AdminFeeRepository) UpdateAdminFeeById(ctx context.Context, adminFeeId string, adminFee *models.AdminFee) error {
	err := db.GetDB().WithContext(ctx).Model(&models.AdminFee{}).Where("id = ?", adminFeeId).Updates(adminFee).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAdminFeeById implements AdminFeeRepositoryInterface.

func (*AdminFeeRepository) GetAdminFeeById(ctx context.Context, adminFeeId string) (*models.AdminFee, error) {
	var adminFee models.AdminFee
	where := models.AdminFee{}
	where.ID = adminFeeId
	_, err := First(ctx, where, &adminFee, nil)
	if err != nil {
		return nil, err
	}
	return &adminFee, nil
}

// GetOneAdminFeeLatest implements AdminFeeRepositoryInterface.
func (*AdminFeeRepository) GetOneAdminFeeLatest(ctx context.Context) (*models.AdminFee, error) {
	var adminFee models.AdminFee
	err := db.GetDB().WithContext(ctx).Model(&models.AdminFee{}).Order("created_at desc").Limit(1).Select("id, admin_fee, created_at, updated_at").Scan(&adminFee).Error
	if err != nil {
		return nil, err
	}
	return &adminFee, nil
}

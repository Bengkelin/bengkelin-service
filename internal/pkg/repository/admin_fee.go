package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	adminFeeRepository *AdminFeeRepository
)

type AdminFeeRepositoryInterface interface {
	CreateAdminFee(adminFee models.AdminFee) (models.AdminFee, error)
	UpdateAdminFeeById(adminFeeId string, adminFee *models.AdminFee) error
	GetAdminFeeById(adminFeeId string) (*models.AdminFee, error)
	GetOneAdminFeeLatest() (*models.AdminFee, error)
}

type AdminFeeRepository struct{}

func GetAdminFeeRepository() AdminFeeRepositoryInterface {
	if adminFeeRepository == nil {
		adminFeeRepository = &AdminFeeRepository{}
	}
	return adminFeeRepository
}

// CreateAdminFee implements AdminFeeRepositoryInterface.

func (repo *AdminFeeRepository) CreateAdminFee(adminFee models.AdminFee) (models.AdminFee, error) {
	err := Create(&adminFee)
	if err != nil {
		return models.AdminFee{}, err
	}
	return adminFee, nil
}

// UpdateAdminFeeById implements AdminFeeRepositoryInterface.
func (*AdminFeeRepository) UpdateAdminFeeById(adminFeeId string, adminFee *models.AdminFee) error {
	err := db.GetDB().Model(&models.AdminFee{}).Where("id = ?", adminFeeId).Updates(adminFee).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAdminFeeById implements AdminFeeRepositoryInterface.

func (*AdminFeeRepository) GetAdminFeeById(adminFeeId string) (*models.AdminFee, error) {
	var adminFee models.AdminFee
	where := models.AdminFee{}
	where.ID = adminFeeId
	_, err := First(where, &adminFee, nil)
	if err != nil {
		return nil, err
	}
	return &adminFee, nil
}

// GetOneAdminFeeLatest implements AdminFeeRepositoryInterface.
func (*AdminFeeRepository) GetOneAdminFeeLatest() (*models.AdminFee, error) {
	var adminFee models.AdminFee
	err := db.GetDB().Model(&models.AdminFee{}).Order("created_at desc").Limit(1).Select("id, admin_fee, created_at, updated_at").Scan(&adminFee).Error
	if err != nil {
		return nil, err
	}
	return &adminFee, nil
}
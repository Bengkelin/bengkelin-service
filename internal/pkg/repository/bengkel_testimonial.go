package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	bengkelTestimonialRepository *BengkelTestimonialRepository
)

type BengkelTestimonialRepositoryInterface interface {
	CreateBengkelTestimonial(bengkelTestimonial models.BengkelTestimonial) (models.BengkelTestimonial, error)
	UpdateBengkelTestimonialById(bengkelTestimonialId string, bengkelTestimonial *models.BengkelTestimonial) error
	GetBengkelTestimonialById(bengkelTestimonialId string) (*models.BengkelTestimonial, error)
	GetAllBengkelTestimonialPaginate(page int, limit int) ([]models.BengkelTestimonial, int, error)
}

type BengkelTestimonialRepository struct{}

func GetBengkelTestimonialRepository() BengkelTestimonialRepositoryInterface {
	if bengkelTestimonialRepository == nil {
		bengkelTestimonialRepository = &BengkelTestimonialRepository{}
	}
	return bengkelTestimonialRepository
}

// CreateBengkelTestimonial implements BengkelTestimonialRepositoryInterface.
func (repo *BengkelTestimonialRepository) CreateBengkelTestimonial(bengkelTestimonial models.BengkelTestimonial) (models.BengkelTestimonial, error) {
	err := db.GetDB().Create(&bengkelTestimonial).Error
	if err != nil {
		return models.BengkelTestimonial{}, err
	}
	return bengkelTestimonial, nil
}

// UpdateBengkelTestimonialById implements BengkelTestimonialRepositoryInterface.
func (*BengkelTestimonialRepository) UpdateBengkelTestimonialById(bengkelTestimonialId string, bengkelTestimonial *models.BengkelTestimonial) error {
	err := db.GetDB().Model(&models.BengkelTestimonial{}).Where("id = ?", bengkelTestimonialId).Updates(bengkelTestimonial).Error

	if err != nil {
		return err
	}
	return nil
}

// GetBengkelTestimonialById implements BengkelTestimonialRepositoryInterface.
func (*BengkelTestimonialRepository) GetBengkelTestimonialById(bengkelTestimonialId string) (*models.BengkelTestimonial, error) {
	var bengkelTestimonial models.BengkelTestimonial
	where := models.BengkelTestimonial{}
	where.UserID = bengkelTestimonialId
	err := db.GetDB().First(where, &bengkelTestimonial, nil).Error
	if err != nil {
		return nil, err
	}
	return &bengkelTestimonial, nil
}

// GetAllBengkelTestimonialPaginate implements BengkelTestimonialRepositoryInterface.
func (*BengkelTestimonialRepository) GetAllBengkelTestimonialPaginate(page int, limit int) ([]models.BengkelTestimonial, int, error) {
	var bengkelTestimonials []models.BengkelTestimonial
	var total int64
	err := db.GetDB().Model(&models.BengkelTestimonial{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.GetDB().Offset((page - 1) * limit).Limit(limit).Find(&bengkelTestimonials).Error
	if err != nil {
		return nil, 0, err
	}
	return bengkelTestimonials, int(total), nil
}

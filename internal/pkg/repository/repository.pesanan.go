package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	pesananRepository *PesananRepository
)

type PesananRepositoryInterface interface {
	CreatePesanan(pesanan models.Pesanan) (models.Pesanan, error)
	UpdatePesananById(pesananId string, pesanan *models.Pesanan) error
	GetPesananById(pesananId string) (*models.Pesanan, error)
	GetDetailPesananById(pesananId, userId string) (*models.Pesanan, error)
	GetAllPesananUserPaginate(userId string, page, limit int) ([]models.Pesanan, int, error)
	GetAllPesananMitraPaginate(bengkelId string, page, limit int) ([]models.Pesanan, int, error)
}

type PesananRepository struct{}

func GetPesananRepository() PesananRepositoryInterface {
	if pesananRepository == nil {
		pesananRepository = &PesananRepository{}
	}
	return pesananRepository
}

// CreatePesanan implements PesananRepositoryInterface.
func (repo *PesananRepository) CreatePesanan(pesanan models.Pesanan) (models.Pesanan, error) {
	err := db.GetDB().Create(&pesanan).Error
	if err != nil {
		return models.Pesanan{}, err
	}
	return pesanan, nil
}

// UpdatePesananById implements PesananRepositoryInterface.
func (*PesananRepository) UpdatePesananById(pesananId string, pesanan *models.Pesanan) error {
	err := db.GetDB().Model(&models.Pesanan{}).Where("id = ?", pesananId).Updates(pesanan).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAllPesananUserPaginate implements PesananRepositoryInterface.
func (*PesananRepository) GetAllPesananUserPaginate(userId string, page, limit int) ([]models.Pesanan, int, error) {
	var pesanans []models.Pesanan
	var count int64

	err := db.GetDB().Model(&models.Pesanan{}).Where("user_id = ?", userId).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.GetDB().Preload("PesananService").Preload("User").Preload("Bengkel").Preload("Bengkel.Addresses").Preload("Vehicle").Where("user_id = ?", userId).Offset((page - 1) * limit).Limit(limit).Find(&pesanans).Error

	if err != nil {
		return nil, 0, err
	}

	return pesanans, int(count), nil
}

// GetAllPesananMitraPaginate implements PesananRepositoryInterface.
func (*PesananRepository) GetAllPesananMitraPaginate(bengkelId string, page, limit int) ([]models.Pesanan, int, error) {
	var pesanans []models.Pesanan
	var count int64

	err := db.GetDB().Model(&models.Pesanan{}).Where("bengkel_id = ?", bengkelId).Count(&count).Error

	if err != nil {
		return nil, 0, err
	}

	err = db.GetDB().Preload("PesananService").Preload("User").Preload("Bengkel").Preload("Bengkel.Addresses").Preload("Vehicle").Where("bengkel_id = ?", bengkelId).Offset((page - 1) * limit).Limit(limit).Find(&pesanans).Error

	if err != nil {
		return nil, 0, err
	}

	return pesanans, int(count), nil
}

// GetPesananById implements PesananRepositoryInterface.
func (*PesananRepository) GetPesananById(pesananId string) (*models.Pesanan, error) {
	var pesanan models.Pesanan
	where := models.Pesanan{}
	where.ID = pesananId
	_, err := First(where, &pesanan, []string{"PesananService", "User", "Bengkel", "Bengkel.Addresses", "Vehicle"})
	if err != nil {
		return nil, err
	}
	return &pesanan, nil
}

// GetDetailPesananById implements PesananRepositoryInterface.
func (*PesananRepository) GetDetailPesananById(pesananId, userId string) (*models.Pesanan, error) {
	var pesanan models.Pesanan
	err := db.GetDB().
		Model(&models.Pesanan{}).
		Preload("PesananService").
		Preload("User").
		Preload("Bengkel").
		Preload("Bengkel.Addresses").
		Preload("Vehicle").
		Where("id = ? AND user_id = ?", pesananId, userId).
		Take(&pesanan).Error
	if err != nil {
		return nil, err
	}
	return &pesanan, nil
}

//

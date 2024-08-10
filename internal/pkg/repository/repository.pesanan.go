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
	GetAllPesanan() ([]models.Pesanan, error)
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

// GetAllPesanan implements PesananRepositoryInterface.
func (*PesananRepository) GetAllPesanan() ([]models.Pesanan, error) {
	var pesanan []models.Pesanan
	err := db.GetDB().Find(&pesanan).Error
	if err != nil {
		return nil, err
	}
	return pesanan, nil
}

// GetPesananById implements PesananRepositoryInterface.
func (*PesananRepository) GetPesananById(pesananId string) (*models.Pesanan, error) {
	var pesanan models.Pesanan
	where := models.Pesanan{}
	where.ID = pesananId
	err := db.GetDB().First(where, &pesanan, nil).Error
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

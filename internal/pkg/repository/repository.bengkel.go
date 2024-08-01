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
	GetAllBengkel() ([]models.Bengkel, error)
	GetAllBengkelPaginate(page int, limit int) ([]models.Bengkel, int, error)
	GetBengkelSearch(query string, page int, limit int) ([]models.Bengkel, int, error)
	GetBengkelByFilterService(service string, page int, limit int) ([]models.Bengkel, int, error)
	GetBengkelSearchV2(service, query string, page int, limit int) ([]models.Bengkel, int, error)
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

// GetAllBengkel implements BengkelRepositoryInterface.
func (*BengkelRepository) GetAllBengkel() ([]models.Bengkel, error) {
	var bengkels []models.Bengkel
	err := Find(models.Bengkel{}, &bengkels, []string{"Photos", "Addresses", "Operasionals"})
	if err != nil {
		return nil, err
	}
	return bengkels, nil
}

// GetAllBengkelPaginate implements BengkelRepositoryInterface.
func (*BengkelRepository) GetAllBengkelPaginate(page int, limit int) ([]models.Bengkel, int, error) {
	var bengkels []models.Bengkel
	var count int64
	err := db.GetDB().Model(&models.Bengkel{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.GetDB().Preload("Photos").Preload("Services").Preload("Addresses").Preload("Operasionals").Offset((page - 1) * limit).Limit(limit).Find(&bengkels).Error
	if err != nil {
		return nil, 0, err
	}
	return bengkels, int(count), nil
}

// GetBengkelSearch implements BengkelRepositoryInterface.
func (*BengkelRepository) GetBengkelSearch(query string, page int, limit int) ([]models.Bengkel, int, error) {
	var bengkels []models.Bengkel

	var count int64
	// Create a subquery to filter bengkels based on the services
	subQuery := db.GetDB().Model(&models.Bengkel{}).
		Joins("LEFT JOIN bengkel_services as bs ON bs.bengkel_id = bengkels.id").
		Joins("LEFT JOIN bengkel_addresses as ba ON ba.bengkel_id = bengkels.id").
		Where("bengkel_name LIKE ? OR bs.nama_service LIKE ? OR ba.full_address LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Group("bengkels.id")

	// Count the total number of filtered bengkels
	err := db.GetDB().Model(&models.Bengkel{}).
		Where("id IN (?)", subQuery.Select("bengkels.id")).
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Fetch the filtered bengkels with associations
	err = db.GetDB().Model(&models.Bengkel{}).
		Where("id IN (?)", subQuery.Select("bengkels.id")).
		Preload("Photos").
		Preload("Services").
		Preload("Addresses").
		Preload("Operasionals").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&bengkels).Error
	if err != nil {
		return nil, 0, err
	}
	return bengkels, int(count), nil
}

// GetBengkelByFilterService implements BengkelRepositoryInterface.
func (*BengkelRepository) GetBengkelByFilterService(service string, page int, limit int) ([]models.Bengkel, int, error) {
	var bengkels []models.Bengkel

	var count int64

	query := db.GetDB().Model(&models.Bengkel{})
	if service == "home_service" {
		err := query.Where("home_service = true").Count(&count).Error
		if err != nil {
			return nil, 0, err
		}
		err = query.
			Where("home_service = true").
			Preload("Photos").
			Preload("Services").
			Preload("Addresses").
			Preload("Operasionals").
			Offset((page - 1) * limit).
			Limit(limit).
			Find(&bengkels).Error
		if err != nil {
			return nil, 0, err
		}
		return bengkels, int(count), nil
	}

	err := query.Where("store_service = true").Count(&count).Error

	if err != nil {
		return nil, 0, err
	}
	err = query.
		Where("store_service = true").
		Preload("Photos").
		Preload("Services").
		Preload("Addresses").
		Preload("Operasionals").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&bengkels).Error
	if err != nil {
		return nil, 0, err
	}
	return bengkels, int(count), nil
}

// GetBengkelSearchV2 implements BengkelRepositoryInterface.
func (*BengkelRepository) GetBengkelSearchV2(service, query string, page int, limit int) ([]models.Bengkel, int, error) {
	var bengkels []models.Bengkel
	var count int64

	subQuery := db.GetDB().Model(&models.Bengkel{})

	if service == "home_service" {
		subQuery = subQuery.Where("home_service = true")
	} else if service == "store_service" {
		subQuery = subQuery.Where("store_service = true")
	}

	if query != "" {
		subQuery = db.GetDB().Model(&models.Bengkel{}).
			Joins("LEFT JOIN bengkel_services as bs ON bs.bengkel_id = bengkels.id").
			Joins("LEFT JOIN bengkel_addresses as ba ON ba.bengkel_id = bengkels.id").
			Where("bengkel_name LIKE ? OR bs.nama_service LIKE ? OR ba.full_address LIKE ?",
				"%"+query+"%", "%"+query+"%", "%"+query+"%").
			Group("bengkels.id")
	}

	// Count the total number of filtered bengkels
	err := db.GetDB().Model(&models.Bengkel{}).
		Where("id IN (?)", subQuery.Select("bengkels.id")).
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Fetch the filtered bengkels with associations
	err = db.GetDB().Model(&models.Bengkel{}).
		Where("id IN (?)", subQuery.Select("bengkels.id")).
		Preload("Photos").
		Preload("Services").
		Preload("Addresses").
		Preload("Operasionals").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&bengkels).Error
	if err != nil {
		return nil, 0, err
	}

	return bengkels, int(count), nil
}

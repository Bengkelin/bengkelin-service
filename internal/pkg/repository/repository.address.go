package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	addressRepository *AddressRepository
)

type AddressRepositoryInterface interface {
	CreateAddress(address models.AddressUser) (models.AddressUser, error)
	UpdateAddressById(addressId uint, userId string, address *models.AddressUser) error
	GetAddressById(userId string, addressId uint) (*models.AddressUser, error)
	DeleteAddressById(addressId uint, userId string) error
}

type AddressRepository struct{}

func GetAddressRepository() AddressRepositoryInterface {
	if addressRepository == nil {
		addressRepository = &AddressRepository{}
	}
	return addressRepository
}

// CreateAddress implements AddressRepositoryInterface.
func (repo *AddressRepository) CreateAddress(address models.AddressUser) (models.AddressUser, error) {
	err := Create(&address)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.AddressUser{}, err
	}
	return address, nil
}

// UpdateAddressById implements AddressRepositoryInterface.
func (*AddressRepository) UpdateAddressById(addressId uint, userId string, address *models.AddressUser) error {
	err := db.GetDB().Model(&models.AddressUser{}).Where("id = ? AND user_id = ?", addressId, userId).Updates(address).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAddressById implements AddressRepositoryInterface.
func (*AddressRepository) GetAddressById(userId string, addressId uint) (*models.AddressUser, error) {
	var address models.AddressUser
	where := models.AddressUser{}
	where.UserID = userId
	where.ID = addressId
	_, err := First(where, &address, nil)
	if err != nil {
		return nil, err
	}
	return &address, nil
}

// DeleteAddressById implements AddressRepositoryInterface.
func (*AddressRepository) DeleteAddressById(addressId uint, userId string) error {
	err := db.GetDB().Where("id = ? AND user_id = ?", addressId, userId).Delete(&models.AddressUser{}).Error
	if err != nil {
		return err
	}
	return nil
}

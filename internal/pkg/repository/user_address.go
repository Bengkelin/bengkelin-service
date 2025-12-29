package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	addressRepository *AddressRepository
)

type AddressRepositoryInterface interface {
	CreateAddress(address models.UserAddress) (models.UserAddress, error)
	UpdateAddressById(addressId uint, userId string, address *models.UserAddress) error
	GetAddressById(userId string, addressId uint) (*models.UserAddress, error)
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
func (repo *AddressRepository) CreateAddress(address models.UserAddress) (models.UserAddress, error) {
	err := Create(&address)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.UserAddress{}, err
	}
	return address, nil
}

// UpdateAddressById implements AddressRepositoryInterface.
func (*AddressRepository) UpdateAddressById(addressId uint, userId string, address *models.UserAddress) error {
	err := db.GetDB().Model(&models.UserAddress{}).Where("id = ? AND user_id = ?", addressId, userId).Updates(address).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAddressById implements AddressRepositoryInterface.
func (*AddressRepository) GetAddressById(userId string, addressId uint) (*models.UserAddress, error) {
	var address models.UserAddress
	where := models.UserAddress{}
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
	err := db.GetDB().Where("id = ? AND user_id = ?", addressId, userId).Delete(&models.UserAddress{}).Error
	if err != nil {
		return err
	}
	return nil
}

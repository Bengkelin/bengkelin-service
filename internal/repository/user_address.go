package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/db"
	"github.com/Bengkelin/bengkelin-service/internal/models"
)

var (
	addressRepository *AddressRepository
	addressOnce       sync.Once
)

type AddressRepositoryInterface interface {
	CreateAddress(ctx context.Context, address models.UserAddress) (models.UserAddress, error)
	UpdateAddressById(ctx context.Context, addressId uint, userId string, address *models.UserAddress) error
	GetAddressById(ctx context.Context, userId string, addressId uint) (*models.UserAddress, error)
	DeleteAddressById(ctx context.Context, addressId uint, userId string) error
}

type AddressRepository struct{}

func GetAddressRepository() AddressRepositoryInterface {
	addressOnce.Do(func() {
		addressRepository = &AddressRepository{}
	})
	return addressRepository
}

// CreateAddress implements AddressRepositoryInterface.
func (repo *AddressRepository) CreateAddress(ctx context.Context, address models.UserAddress) (models.UserAddress, error) {
	err := Create(ctx, &address)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.UserAddress{}, err
	}
	return address, nil
}

// UpdateAddressById implements AddressRepositoryInterface.
func (*AddressRepository) UpdateAddressById(ctx context.Context, addressId uint, userId string, address *models.UserAddress) error {
	err := db.GetDB().WithContext(ctx).Model(&models.UserAddress{}).Where("id = ? AND user_id = ?", addressId, userId).Updates(address).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAddressById implements AddressRepositoryInterface.
func (*AddressRepository) GetAddressById(ctx context.Context, userId string, addressId uint) (*models.UserAddress, error) {
	var address models.UserAddress
	where := models.UserAddress{}
	where.UserID = userId
	where.ID = addressId
	_, err := First(ctx, where, &address, nil)
	if err != nil {
		return nil, err
	}
	return &address, nil
}

// DeleteAddressById implements AddressRepositoryInterface.
func (*AddressRepository) DeleteAddressById(ctx context.Context, addressId uint, userId string) error {
	err := db.GetDB().WithContext(ctx).Where("id = ? AND user_id = ?", addressId, userId).Delete(&models.UserAddress{}).Error
	if err != nil {
		return err
	}
	return nil
}

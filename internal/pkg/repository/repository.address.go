package repository

import "github.com/Bengkelin/bengkelin-service/internal/pkg/models"

var (
	addressRepository *AddressRepository
)

type AddressRepositoryInterface interface {
	CreateAddress(address models.Address) (models.Address, error)
}

type AddressRepository struct{}

func GetAddressRepository() AddressRepositoryInterface {
	if addressRepository == nil {
		addressRepository = &AddressRepository{}
	}
	return addressRepository
}

// CreateAddress implements AddressRepositoryInterface.
func (repo *AddressRepository) CreateAddress(address models.Address) (models.Address, error) {
	err := Create(&address)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.Address{}, err
	}
	return address, nil
}

package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	userRepository *UserRepository
)

type UserRepositoryInterface interface {
	FindUserByEmail(email string) (*models.User, error)
	FindUserByID(userID string) (*models.User, error)
	GetDetailUser(userId string) (*models.User, error)
	CreateUser(user models.User) (models.User, error)
	UpdateUser(user *models.User) error
	UpdateUserById(userId string, user *models.User) error
}

type UserRepository struct{}

// CreateUser implements UserRepositoryInterface.
func (repo *UserRepository) CreateUser(user models.User) (models.User, error) {
	err := Create(&user)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// FindUserByEmail implements UserRepositoryInterface.
func (*UserRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	where := models.User{}
	where.Email = email
	_, err := First(where, &user, nil)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByID implements UserRepositoryInterface.
func (*UserRepository) FindUserByID(userID string) (*models.User, error) {
	var user models.User
	where := models.User{}
	where.ID = userID
	_, err := First(where, &user, []string{"Addresses"})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (*UserRepository) GetDetailUser(userId string) (*models.User, error) {
	var user models.User
	where := models.User{}
	where.ID = userId
	_, err := First(where, &user, nil)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser implements UserRepositoryInterface.
func (*UserRepository) UpdateUser(user *models.User) error {
	panic("unimplemented")
}

// UpdateUserById implements UserRepositoryInterface.
func (*UserRepository) UpdateUserById(userId string, user *models.User) error {
	err := db.GetDB().Model(&user).Where("id = ?", userId).Updates(user)
	if err.Error != nil {
		return err.Error
	}
	return nil
}

func GetUserRepository() UserRepositoryInterface {
	if userRepository == nil {
		userRepository = &UserRepository{}
	}
	return userRepository
}

package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/db"
	"github.com/Bengkelin/bengkelin-service/internal/models"
)

var (
	userRepository *UserRepository
	userOnce       sync.Once
)

type UserRepositoryInterface interface {
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	FindUserByID(ctx context.Context, userID string) (*models.User, error)
	GetDetailUser(ctx context.Context, userId string) (*models.User, error)
	CreateUser(ctx context.Context, user models.User) (models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdateUserById(ctx context.Context, userId string, user *models.User) error
}

type UserRepository struct{}

// CreateUser implements UserRepositoryInterface.
func (repo *UserRepository) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	err := Create(ctx, &user)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// FindUserByEmail implements UserRepositoryInterface.
func (*UserRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	where := models.User{}
	where.Email = email
	_, err := First(ctx, where, &user, nil)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByID implements UserRepositoryInterface.
func (*UserRepository) FindUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	where := models.User{}
	where.ID = userID
	_, err := First(ctx, where, &user, []string{"Addresses", "Vehicles", "Vehicles.Photos"})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (*UserRepository) GetDetailUser(ctx context.Context, userId string) (*models.User, error) {
	var user models.User
	where := models.User{}
	where.ID = userId
	_, err := First(ctx, where, &user, []string{"Addresses"})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser implements UserRepositoryInterface.
func (*UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	panic("unimplemented")
}

// UpdateUserById implements UserRepositoryInterface.
func (*UserRepository) UpdateUserById(ctx context.Context, userId string, user *models.User) error {
	err := db.GetDB().WithContext(ctx).Model(&user).Where("id = ?", userId).Updates(user)
	if err.Error != nil {
		return err.Error
	}
	return nil
}

func GetUserRepository() UserRepositoryInterface {
	userOnce.Do(func() {
		userRepository = &UserRepository{}
	})
	return userRepository
}

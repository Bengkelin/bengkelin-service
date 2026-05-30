package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/dto"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/models"
	"github.com/Bengkelin/bengkelin-service/internal/service"
	"github.com/Bengkelin/bengkelin-service/tests/fixtures/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserService() (*mocks.MockUserRepository, *mocks.MockAddressRepository, *mocks.MockVehicleRepository, service.UserServiceInterface) {
	userRepo := new(mocks.MockUserRepository)
	addrRepo := new(mocks.MockAddressRepository)
	vehicleRepo := new(mocks.MockVehicleRepository)
	svc := service.NewUserService(service.ServiceDependencies{
		UserRepo:    userRepo,
		AddressRepo: addrRepo,
		VehicleRepo: vehicleRepo,
	})
	return userRepo, addrRepo, vehicleRepo, svc
}

func boolPtr(b bool) *bool { return &b }

// --- GetUserProfile ---

func TestGetUserProfile_Success(t *testing.T) {
	userRepo, _, _, svc := setupUserService()
	ctx := context.Background()

	user := &appErrors.User{
		ID:          "user-1",
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		PhoneNumber: "081234567890",
		AvatarUrl:   "https://example.com/avatar.jpg",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Addresses: []appErrors.UserAddress{
			{ID: 1, Latitude: -6.2, Longitude: 106.8, AddressLabel: "Home", FullAddress: "Jl. Sudirman", IsPrimary: boolPtr(true)},
		},
		Vehicles: []appErrors.Vehicle{
			{ID: 1, VehicleType: "Motor", VehicleNumber: "B 1234 CD", VehicleColor: "Hitam"},
		},
	}

	userRepo.On("FindUserByID", ctx, "user-1").Return(user, nil)

	result, err := svc.GetUserProfile(ctx, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user-1", result.ID)
	assert.Equal(t, "John", result.FirstName)
	assert.Equal(t, "Doe", result.LastName)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Len(t, result.Addresses, 1)
	assert.Equal(t, "Home", result.Addresses[0].AddressLabel)
	assert.True(t, result.Addresses[0].IsPrimary)
	assert.Len(t, result.Vehicles, 1)
	assert.Equal(t, "Motor", result.Vehicles[0].VehicleType)
	userRepo.AssertExpectations(t)
}

func TestGetUserProfile_NotFound(t *testing.T) {
	userRepo, _, _, svc := setupUserService()
	ctx := context.Background()

	userRepo.On("FindUserByID", ctx, "user-999").Return(nil, errors.New("not found"))

	result, err := svc.GetUserProfile(ctx, "user-999")

	assert.Error(t, err)
	assert.Nil(t, result)
	userRepo.AssertExpectations(t)
}

// --- UpdateUserProfile ---

func TestUpdateUserProfile_Success(t *testing.T) {
	userRepo, _, _, svc := setupUserService()
	ctx := context.Background()

	existingUser := &appErrors.User{
		ID:          "user-1",
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		PhoneNumber: "081234567890",
	}

	updatedUser := &appErrors.User{
		ID:          "user-1",
		FirstName:   "Jane",
		LastName:    "Smith",
		Email:       "john@example.com",
		PhoneNumber: "089999999999",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	userRepo.On("GetDetailUser", ctx, "user-1").Return(existingUser, nil)
	userRepo.On("UpdateUserById", ctx, "user-1", mock.AnythingOfType("*models.User")).Return(nil)
	userRepo.On("FindUserByID", ctx, "user-1").Return(updatedUser, nil)

	req := dto.UpdateUserRequest{
		FirstName:   "Jane",
		LastName:    "Smith",
		PhoneNumber: "089999999999",
	}

	result, err := svc.UpdateUserProfile(ctx, "user-1", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Jane", result.FirstName)
	assert.Equal(t, "Smith", result.LastName)
	userRepo.AssertExpectations(t)
}

func TestUpdateUserProfile_NotFound(t *testing.T) {
	userRepo, _, _, svc := setupUserService()
	ctx := context.Background()

	userRepo.On("GetDetailUser", ctx, "user-999").Return(nil, errors.New("not found"))

	req := dto.UpdateUserRequest{FirstName: "Jane"}
	result, err := svc.UpdateUserProfile(ctx, "user-999", req)

	assert.Error(t, err)
	assert.Nil(t, result)
	userRepo.AssertExpectations(t)
}

// --- UpdateUserAvatar ---

func TestUpdateUserAvatar_Success(t *testing.T) {
	userRepo, _, _, svc := setupUserService()
	ctx := context.Background()

	user := &appErrors.User{ID: "user-1"}
	userRepo.On("FindUserByID", ctx, "user-1").Return(user, nil)
	userRepo.On("UpdateUserById", ctx, "user-1", mock.AnythingOfType("*models.User")).Return(nil)

	err := svc.UpdateUserAvatar(ctx, "user-1", "https://example.com/new-avatar.jpg")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestUpdateUserAvatar_UserNotFound(t *testing.T) {
	userRepo, _, _, svc := setupUserService()
	ctx := context.Background()

	userRepo.On("FindUserByID", ctx, "user-999").Return(nil, errors.New("not found"))

	err := svc.UpdateUserAvatar(ctx, "user-999", "https://example.com/avatar.jpg")

	assert.Error(t, err)
	userRepo.AssertExpectations(t)
}

// --- AddUserAddress ---

func TestAddUserAddress_Success(t *testing.T) {
	userRepo, addrRepo, _, svc := setupUserService()
	ctx := context.Background()

	// User has existing addresses so isPrimary defaults to false (avoids db.WithTransaction)
	user := &appErrors.User{
		ID: "user-1",
		Addresses: []appErrors.UserAddress{
			{ID: 1, IsPrimary: boolPtr(true)},
		},
	}
	userRepo.On("GetDetailUser", ctx, "user-1").Return(user, nil)

	createdAddr := appErrors.UserAddress{
		ID:           2,
		UserID:       "user-1",
		Latitude:     -6.2,
		Longitude:    106.8,
		AddressLabel: "Office",
		FullAddress:  "Jl. Sudirman",
		IsPrimary:    boolPtr(false),
	}
	addrRepo.On("CreateAddress", ctx, mock.AnythingOfType("models.UserAddress")).Return(createdAddr, nil)

	req := dto.AddressRequest{
		Latitude:     -6.2,
		Longitude:    106.8,
		AddressLabel: "Office",
		FullAddress:  "Jl. Sudirman",
	}

	result, err := svc.AddUserAddress(ctx, "user-1", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Office", result.AddressLabel)
	assert.False(t, result.IsPrimary)
	userRepo.AssertExpectations(t)
	addrRepo.AssertExpectations(t)
}

func TestAddUserAddress_UserNotFound(t *testing.T) {
	userRepo, _, _, svc := setupUserService()
	ctx := context.Background()

	userRepo.On("GetDetailUser", ctx, "user-999").Return(nil, errors.New("not found"))

	req := dto.AddressRequest{Latitude: -6.2, Longitude: 106.8}
	result, err := svc.AddUserAddress(ctx, "user-999", req)

	assert.Error(t, err)
	assert.Nil(t, result)
	userRepo.AssertExpectations(t)
}

// --- UpdateAddress ---

func TestUpdateAddress_Success(t *testing.T) {
	_, addrRepo, _, svc := setupUserService()
	ctx := context.Background()

	existing := &appErrors.UserAddress{
		ID:           1,
		UserID:       "user-1",
		Latitude:     -6.2,
		Longitude:    106.8,
		AddressLabel: "Home",
		FullAddress:  "Jl. Sudirman",
		IsPrimary:    boolPtr(true),
	}
	addrRepo.On("GetAddressById", ctx, "user-1", uint(1)).Return(existing, nil)
	addrRepo.On("UpdateAddressById", ctx, uint(1), "user-1", mock.AnythingOfType("*models.UserAddress")).Return(nil)

	updated := &appErrors.UserAddress{
		ID:           1,
		UserID:       "user-1",
		Latitude:     -6.3,
		Longitude:    106.9,
		AddressLabel: "Office",
		FullAddress:  "Jl. Thamrin",
		IsPrimary:    boolPtr(true),
	}
	addrRepo.On("GetAddressById", ctx, "user-1", uint(1)).Return(updated, nil)

	req := dto.AddressRequest{
		Latitude:     -6.3,
		Longitude:    106.9,
		AddressLabel: "Office",
		FullAddress:  "Jl. Thamrin",
	}

	result, err := svc.UpdateAddress(ctx, "user-1", 1, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Office", result.AddressLabel)
	addrRepo.AssertExpectations(t)
}

func TestUpdateAddress_NotFound(t *testing.T) {
	_, addrRepo, _, svc := setupUserService()
	ctx := context.Background()

	addrRepo.On("GetAddressById", ctx, "user-1", uint(999)).Return(nil, errors.New("not found"))

	req := dto.AddressRequest{Latitude: -6.2}
	result, err := svc.UpdateAddress(ctx, "user-1", 999, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	addrRepo.AssertExpectations(t)
}

func TestUpdateAddress_SetPrimary(t *testing.T) {
	t.Skip("Requires db.WithTransaction — needs integration test with real DB")

	userRepo, addrRepo, _, svc := setupUserService()
	ctx := context.Background()

	existing := &appErrors.UserAddress{
		ID:        2,
		UserID:    "user-1",
		IsPrimary: boolPtr(false),
	}
	addrRepo.On("GetAddressById", ctx, "user-1", uint(2)).Return(existing, nil)

	user := &appErrors.User{
		ID: "user-1",
		Addresses: []appErrors.UserAddress{
			{ID: 1, IsPrimary: boolPtr(true)},
			{ID: 2, IsPrimary: boolPtr(false)},
		},
	}
	userRepo.On("GetDetailUser", ctx, "user-1").Return(user, nil)
	addrRepo.On("UpdateAddressById", ctx, uint(1), "user-1", mock.AnythingOfType("*models.UserAddress")).Return(nil)
	addrRepo.On("UpdateAddressById", ctx, uint(2), "user-1", mock.AnythingOfType("*models.UserAddress")).Return(nil)

	updated := &appErrors.UserAddress{ID: 2, UserID: "user-1", IsPrimary: boolPtr(true)}
	addrRepo.On("GetAddressById", ctx, "user-1", uint(2)).Return(updated, nil)

	isPrimary := true
	req := dto.AddressRequest{IsPrimary: &isPrimary}

	result, err := svc.UpdateAddress(ctx, "user-1", 2, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsPrimary)
}

// --- CreateOrUpdateAddress ---

func TestCreateOrUpdateAddress_CreateNew(t *testing.T) {
	userRepo, addrRepo, _, svc := setupUserService()
	ctx := context.Background()

	user := &appErrors.User{
		ID:        "user-1",
		Addresses: []appErrors.UserAddress{},
	}
	userRepo.On("GetDetailUser", ctx, "user-1").Return(user, nil)

	created := appErrors.UserAddress{
		ID:           1,
		UserID:       "user-1",
		Latitude:     -6.2,
		Longitude:    106.8,
		AddressLabel: "Home",
		IsPrimary:    boolPtr(true),
	}
	addrRepo.On("CreateAddress", ctx, mock.AnythingOfType("models.UserAddress")).Return(created, nil)

	req := dto.AddressRequest{
		Latitude:     -6.2,
		Longitude:    106.8,
		AddressLabel: "Home",
	}

	result, err := svc.CreateOrUpdateAddress(ctx, "user-1", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Home", result.AddressLabel)
	assert.True(t, result.IsPrimary)
	userRepo.AssertExpectations(t)
	addrRepo.AssertExpectations(t)
}

func TestCreateOrUpdateAddress_UpdateExisting(t *testing.T) {
	userRepo, addrRepo, _, svc := setupUserService()
	ctx := context.Background()

	user := &appErrors.User{
		ID: "user-1",
		Addresses: []appErrors.UserAddress{
			{ID: 1, Latitude: -6.2, Longitude: 106.8, AddressLabel: "Home", IsPrimary: boolPtr(true)},
		},
	}
	userRepo.On("GetDetailUser", ctx, "user-1").Return(user, nil)
	addrRepo.On("UpdateAddressById", ctx, uint(1), "user-1", mock.AnythingOfType("*models.UserAddress")).Return(nil)

	updated := &appErrors.UserAddress{
		ID:           1,
		UserID:       "user-1",
		Latitude:     -6.3,
		Longitude:    106.9,
		AddressLabel: "Updated Home",
		IsPrimary:    boolPtr(true),
	}
	addrRepo.On("GetAddressById", ctx, "user-1", uint(1)).Return(updated, nil)

	req := dto.AddressRequest{
		Latitude:     -6.3,
		Longitude:    106.9,
		AddressLabel: "Updated Home",
	}

	result, err := svc.CreateOrUpdateAddress(ctx, "user-1", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Home", result.AddressLabel)
}

func TestCreateOrUpdateAddress_UserNotFound(t *testing.T) {
	userRepo, _, _, svc := setupUserService()
	ctx := context.Background()

	userRepo.On("GetDetailUser", ctx, "user-999").Return(nil, errors.New("not found"))

	req := dto.AddressRequest{Latitude: -6.2}
	result, err := svc.CreateOrUpdateAddress(ctx, "user-999", req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- GetUserAddress ---

func TestGetUserAddress_Success(t *testing.T) {
	_, addrRepo, _, svc := setupUserService()
	ctx := context.Background()

	addr := &appErrors.UserAddress{
		ID:           1,
		UserID:       "user-1",
		Latitude:     -6.2,
		Longitude:    106.8,
		AddressLabel: "Home",
		FullAddress:  "Jl. Sudirman",
		IsPrimary:    boolPtr(true),
	}
	addrRepo.On("GetAddressById", ctx, "user-1", uint(1)).Return(addr, nil)

	result, err := svc.GetUserAddress(ctx, "user-1", "1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Home", result.AddressLabel)
	addrRepo.AssertExpectations(t)
}

func TestGetUserAddress_NotFound(t *testing.T) {
	_, addrRepo, _, svc := setupUserService()
	ctx := context.Background()

	addrRepo.On("GetAddressById", ctx, "user-1", uint(999)).Return(nil, errors.New("not found"))

	result, err := svc.GetUserAddress(ctx, "user-1", "999")

	assert.Error(t, err)
	assert.Nil(t, result)
	addrRepo.AssertExpectations(t)
}

func TestGetUserAddress_InvalidID(t *testing.T) {
	_, _, _, svc := setupUserService()
	ctx := context.Background()

	result, err := svc.GetUserAddress(ctx, "user-1", "abc")

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- DeleteUserAddress ---

func TestDeleteUserAddress_Success(t *testing.T) {
	_, addrRepo, _, svc := setupUserService()
	ctx := context.Background()

	addr := &appErrors.UserAddress{ID: 1, UserID: "user-1"}
	addrRepo.On("GetAddressById", ctx, "user-1", uint(1)).Return(addr, nil)
	addrRepo.On("DeleteAddressById", ctx, uint(1), "user-1").Return(nil)

	err := svc.DeleteUserAddress(ctx, "user-1", "1")

	assert.NoError(t, err)
	addrRepo.AssertExpectations(t)
}

func TestDeleteUserAddress_NotFound(t *testing.T) {
	_, addrRepo, _, svc := setupUserService()
	ctx := context.Background()

	addrRepo.On("GetAddressById", ctx, "user-1", uint(999)).Return(nil, errors.New("not found"))

	err := svc.DeleteUserAddress(ctx, "user-1", "999")

	assert.Error(t, err)
	addrRepo.AssertExpectations(t)
}

// --- AddUserVehicle ---

func TestAddUserVehicle_Success(t *testing.T) {
	userRepo, _, vehicleRepo, svc := setupUserService()
	ctx := context.Background()

	user := &appErrors.User{ID: "user-1"}
	userRepo.On("GetDetailUser", ctx, "user-1").Return(user, nil)

	created := appErrors.Vehicle{
		ID:            1,
		UserID:        "user-1",
		VehicleType:   "Motor",
		VehicleNumber: "B 1234 CD",
		VehicleColor:  "Hitam",
	}
	vehicleRepo.On("CreateVehicle", ctx, mock.AnythingOfType("models.Vehicle")).Return(created, nil)

	req := dto.VehicleRequest{
		VehicleType:   "Motor",
		VehicleNumber: "B 1234 CD",
		VehicleColor:  "Hitam",
	}

	result, err := svc.AddUserVehicle(ctx, "user-1", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Motor", result.VehicleType)
	assert.Equal(t, "B 1234 CD", result.VehicleNumber)
	userRepo.AssertExpectations(t)
	vehicleRepo.AssertExpectations(t)
}

func TestAddUserVehicle_UserNotFound(t *testing.T) {
	userRepo, _, _, svc := setupUserService()
	ctx := context.Background()

	userRepo.On("GetDetailUser", ctx, "user-999").Return(nil, errors.New("not found"))

	req := dto.VehicleRequest{VehicleType: "Motor"}
	result, err := svc.AddUserVehicle(ctx, "user-999", req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- GetUserVehicle ---

func TestGetUserVehicle_Success(t *testing.T) {
	_, _, vehicleRepo, svc := setupUserService()
	ctx := context.Background()

	vehicle := &appErrors.Vehicle{
		ID:            1,
		UserID:        "user-1",
		VehicleType:   "Motor",
		VehicleNumber: "B 1234 CD",
		VehicleColor:  "Hitam",
	}
	vehicleRepo.On("GetVehicleById", ctx, "user-1", uint(1)).Return(vehicle, nil)

	result, err := svc.GetUserVehicle(ctx, "user-1", "1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Motor", result.VehicleType)
	vehicleRepo.AssertExpectations(t)
}

func TestGetUserVehicle_NotFound(t *testing.T) {
	_, _, vehicleRepo, svc := setupUserService()
	ctx := context.Background()

	vehicleRepo.On("GetVehicleById", ctx, "user-1", uint(999)).Return(nil, errors.New("not found"))

	result, err := svc.GetUserVehicle(ctx, "user-1", "999")

	assert.Error(t, err)
	assert.Nil(t, result)
	vehicleRepo.AssertExpectations(t)
}

func TestGetUserVehicle_InvalidID(t *testing.T) {
	_, _, _, svc := setupUserService()
	ctx := context.Background()

	result, err := svc.GetUserVehicle(ctx, "user-1", "abc")

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- GetAllUserVehicles ---

func TestGetAllUserVehicles_Success(t *testing.T) {
	_, _, vehicleRepo, svc := setupUserService()
	ctx := context.Background()

	vehicles := []appErrors.Vehicle{
		{ID: 1, UserID: "user-1", VehicleType: "Motor", VehicleNumber: "B 1234 CD", VehicleColor: "Hitam"},
		{ID: 2, UserID: "user-1", VehicleType: "Mobil", VehicleNumber: "B 5678 EF", VehicleColor: "Putih"},
	}
	vehicleRepo.On("GetAllVehiclesByUserId", ctx, "user-1").Return(vehicles, nil)

	result, err := svc.GetAllUserVehicles(ctx, "user-1")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Motor", result[0].VehicleType)
	assert.Equal(t, "Mobil", result[1].VehicleType)
	vehicleRepo.AssertExpectations(t)
}

func TestGetAllUserVehicles_Empty(t *testing.T) {
	_, _, vehicleRepo, svc := setupUserService()
	ctx := context.Background()

	vehicleRepo.On("GetAllVehiclesByUserId", ctx, "user-1").Return([]appErrors.Vehicle{}, nil)

	result, err := svc.GetAllUserVehicles(ctx, "user-1")

	assert.NoError(t, err)
	assert.Len(t, result, 0)
	vehicleRepo.AssertExpectations(t)
}

// --- DeleteUserVehicle ---

func TestDeleteUserVehicle_Success(t *testing.T) {
	_, _, vehicleRepo, svc := setupUserService()
	ctx := context.Background()

	vehicle := &appErrors.Vehicle{ID: 1, UserID: "user-1"}
	vehicleRepo.On("GetVehicleById", ctx, "user-1", uint(1)).Return(vehicle, nil)
	vehicleRepo.On("DeleteVehicleById", ctx, uint(1), "user-1").Return(nil)

	err := svc.DeleteUserVehicle(ctx, "user-1", "1")

	assert.NoError(t, err)
	vehicleRepo.AssertExpectations(t)
}

func TestDeleteUserVehicle_NotFound(t *testing.T) {
	_, _, vehicleRepo, svc := setupUserService()
	ctx := context.Background()

	vehicleRepo.On("GetVehicleById", ctx, "user-1", uint(999)).Return(nil, errors.New("not found"))

	err := svc.DeleteUserVehicle(ctx, "user-1", "999")

	assert.Error(t, err)
	vehicleRepo.AssertExpectations(t)
}

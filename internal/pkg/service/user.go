package service

import (
	"context"
	"strconv"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/pkg/errors"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"gorm.io/gorm"
)

// UserService implements UserServiceInterface
type UserService struct {
	deps ServiceDependencies
}

// NewUserService creates a new user service
func NewUserService(deps ServiceDependencies) UserServiceInterface {
	return &UserService{deps: deps}
}

// GetUserProfile gets user profile with caching
func (s *UserService) GetUserProfile(ctx context.Context, userID string) (*dto.UserProfileResponse, error) {
	// Try cache first
	cacheService := GetProfileCacheService()
	cachedUser, err := cacheService.GetUserProfile(ctx, userID)
	if err == nil && cachedUser != nil {
		return s.mapUserToProfileResponse(cachedUser), nil
	}

	// Fetch from database
	user, err := s.deps.UserRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, appErrors.ErrUserNotFound.WithDetails(err.Error())
	}

	// Store in cache (best effort)
	if cacheErr := cacheService.SetUserProfile(ctx, user); cacheErr != nil {
		applog.WarnCtx(ctx, "Failed to cache user profile", "error", cacheErr.Error(), "user_id", userID)
	}

	return s.mapUserToProfileResponse(user), nil
}

// UpdateUserProfile updates user profile and invalidates cache
func (s *UserService) UpdateUserProfile(ctx context.Context, userID string, req dto.UpdateUserRequest) (*dto.UserProfileResponse, error) {
	// Get existing user
	user, err := s.deps.UserRepo.GetDetailUser(ctx, userID)
	if err != nil {
		return nil, appErrors.ErrUserNotFound.WithDetails(err.Error())
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}

	// Save to database
	if err := s.deps.UserRepo.UpdateUserById(ctx, userID, user); err != nil {
		return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
	}

	// Invalidate cache
	cacheService := GetProfileCacheService()
	if cacheErr := cacheService.InvalidateUserProfile(ctx, userID); cacheErr != nil {
		applog.WarnCtx(ctx, "Failed to invalidate user profile cache", "error", cacheErr.Error(), "user_id", userID)
	}

	return s.GetUserProfile(ctx, userID)
}

// UpdateUserAvatar updates user avatar and invalidates cache
func (s *UserService) UpdateUserAvatar(ctx context.Context, userID string, avatarURL string) error {
	// Verify user exists
	_, err := s.deps.UserRepo.FindUserByID(ctx, userID)
	if err != nil {
		return appErrors.ErrUserNotFound.WithDetails(err.Error())
	}

	// Update avatar
	userUpdate := &models.User{
		AvatarUrl: avatarURL,
	}

	if err := s.deps.UserRepo.UpdateUserById(ctx, userID, userUpdate); err != nil {
		return appErrors.ErrDatabaseError.WithDetails(err.Error())
	}

	// Invalidate cache
	cacheService := GetProfileCacheService()
	if cacheErr := cacheService.InvalidateUserProfile(ctx, userID); cacheErr != nil {
		applog.WarnCtx(ctx, "Failed to invalidate user profile cache", "error", cacheErr.Error(), "user_id", userID)
	}

	return nil
}

// AddUserAddress adds a new address for user
func (s *UserService) AddUserAddress(ctx context.Context, userID string, req dto.AddressRequest) (*dto.AddressResponse, error) {
	// Verify user exists
	_, err := s.deps.UserRepo.GetDetailUser(ctx, userID)
	if err != nil {
		return nil, appErrors.ErrUserNotFound.WithDetails(err.Error())
	}

	// Handle IsPrimary logic
	var isPrimary bool
	if req.IsPrimary != nil {
		isPrimary = *req.IsPrimary
	} else {
		// Check if user has any addresses
		user, err := s.deps.UserRepo.GetDetailUser(ctx, userID)
		if err == nil && len(user.Addresses) == 0 {
			isPrimary = true // First address is primary by default
		} else {
			isPrimary = false
		}
	}

	newAddress := &models.UserAddress{
		UserID:       userID,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		AddressLabel: req.AddressLabel,
		FullAddress:  req.FullAddress,
		Note:         req.Note,
		IsPrimary:    &isPrimary,
	}

	// If setting as primary, unset other addresses + create in a single transaction
	if isPrimary {
		err = db.WithTransaction(func(tx *gorm.DB) error {
			user, err := s.deps.UserRepo.GetDetailUser(ctx, userID)
			if err == nil {
				for i := range user.Addresses {
					nonPrimary := false
					user.Addresses[i].IsPrimary = &nonPrimary
					if err := s.deps.AddressRepo.UpdateAddressById(ctx, user.Addresses[i].ID, userID, &user.Addresses[i]); err != nil {
						applog.WarnCtx(ctx, "Failed to update address primary status", "error", err.Error())
					}
				}
			}
			return nil
		})
		if err != nil {
			return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
		}
	}

	createdAddress, err := s.deps.AddressRepo.CreateAddress(ctx, *newAddress)
	if err != nil {
		return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
	}

	return s.mapAddressToResponse(&createdAddress), nil
}

// UpdateAddress updates an existing address with business logic for primary address management
func (s *UserService) UpdateAddress(ctx context.Context, userID string, addressID uint, req dto.AddressRequest) (*dto.AddressResponse, error) {
	existingAddress, err := s.deps.AddressRepo.GetAddressById(ctx, userID, addressID)
	if err != nil {
		return nil, appErrors.ErrAddressNotFound.WithDetails(err.Error())
	}

	if req.IsPrimary != nil && *req.IsPrimary {
		existingAddress.IsPrimary = req.IsPrimary

		err = db.WithTransaction(func(tx *gorm.DB) error {
			user, err := s.deps.UserRepo.GetDetailUser(ctx, userID)
			if err == nil {
				for i := range user.Addresses {
					if user.Addresses[i].ID != existingAddress.ID {
						nonPrimary := false
						user.Addresses[i].IsPrimary = &nonPrimary
						if err := s.deps.AddressRepo.UpdateAddressById(ctx, user.Addresses[i].ID, userID, &user.Addresses[i]); err != nil {
							applog.WarnCtx(ctx, "Failed to unset primary address", "error", err.Error())
						}
					}
				}
			}
			return nil
		})
		if err != nil {
			return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
		}
	}

	if req.Latitude != 0 {
		existingAddress.Latitude = req.Latitude
	}
	if req.Longitude != 0 {
		existingAddress.Longitude = req.Longitude
	}
	if req.AddressLabel != "" {
		existingAddress.AddressLabel = req.AddressLabel
	}
	if req.FullAddress != "" {
		existingAddress.FullAddress = req.FullAddress
	}
	if req.Note != "" {
		existingAddress.Note = req.Note
	}

	if err := s.deps.AddressRepo.UpdateAddressById(ctx, existingAddress.ID, userID, existingAddress); err != nil {
		return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
	}

	updatedAddress, err := s.deps.AddressRepo.GetAddressById(ctx, userID, addressID)
	if err != nil {
		return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
	}

	return s.mapAddressToResponse(updatedAddress), nil
}

// CreateOrUpdateAddress creates a new address or updates the primary one
func (s *UserService) CreateOrUpdateAddress(ctx context.Context, userID string, req dto.AddressRequest) (*dto.AddressResponse, error) {
	user, err := s.deps.UserRepo.GetDetailUser(ctx, userID)
	if err != nil {
		return nil, appErrors.ErrUserNotFound.WithDetails(err.Error())
	}

	if len(user.Addresses) == 0 {
		isPrimary := true
		newAddress := &models.UserAddress{
			UserID:       userID,
			Latitude:     req.Latitude,
			Longitude:    req.Longitude,
			AddressLabel: req.AddressLabel,
			FullAddress:  req.FullAddress,
			Note:         req.Note,
			IsPrimary:    &isPrimary,
		}
		if newAddress.AddressLabel == "" {
			newAddress.AddressLabel = "Home"
		}
		createdAddress, err := s.deps.AddressRepo.CreateAddress(ctx, *newAddress)
		if err != nil {
			return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
		}
		return s.mapAddressToResponse(&createdAddress), nil
	}

	var targetAddress *models.UserAddress
	for i := range user.Addresses {
		if user.Addresses[i].IsPrimary != nil && *user.Addresses[i].IsPrimary {
			targetAddress = &user.Addresses[i]
			break
		}
	}
	if targetAddress == nil {
		targetAddress = &user.Addresses[0]
		isPrimary := true
		targetAddress.IsPrimary = &isPrimary
	}

	if req.IsPrimary != nil {
		if *req.IsPrimary {
			targetAddress.IsPrimary = req.IsPrimary

			err = db.WithTransaction(func(tx *gorm.DB) error {
				for i := range user.Addresses {
					if user.Addresses[i].ID != targetAddress.ID {
						nonPrimary := false
						user.Addresses[i].IsPrimary = &nonPrimary
						if err := s.deps.AddressRepo.UpdateAddressById(ctx, user.Addresses[i].ID, userID, &user.Addresses[i]); err != nil {
							applog.WarnCtx(ctx, "Failed to unset primary address", "error", err.Error())
						}
					}
				}
				return nil
			})
			if err != nil {
				return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
			}
		} else if len(user.Addresses) > 1 {
			targetAddress.IsPrimary = req.IsPrimary
		}
	}

	addressUpdated := false
	if req.Latitude != 0 {
		targetAddress.Latitude = req.Latitude
		addressUpdated = true
	}
	if req.Longitude != 0 {
		targetAddress.Longitude = req.Longitude
		addressUpdated = true
	}
	if req.AddressLabel != "" {
		targetAddress.AddressLabel = req.AddressLabel
		addressUpdated = true
	}
	if req.FullAddress != "" {
		targetAddress.FullAddress = req.FullAddress
		addressUpdated = true
	}
	if req.Note != "" {
		targetAddress.Note = req.Note
		addressUpdated = true
	}

	if addressUpdated || req.IsPrimary != nil {
		if err := s.deps.AddressRepo.UpdateAddressById(ctx, targetAddress.ID, userID, targetAddress); err != nil {
			return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
		}
	}

	updatedAddress, err := s.deps.AddressRepo.GetAddressById(ctx, userID, targetAddress.ID)
	if err != nil {
		return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
	}

	return s.mapAddressToResponse(updatedAddress), nil
}

// GetUserAddress gets a specific address for user
func (s *UserService) GetUserAddress(ctx context.Context, userID, addressID string) (*dto.AddressResponse, error) {
	addressIdUint, err := parseUint(addressID)
	if err != nil {
		return nil, appErrors.ErrValidationFailed.WithDetails("invalid address ID format")
	}

	address, err := s.deps.AddressRepo.GetAddressById(ctx, userID, uint(addressIdUint))
	if err != nil {
		return nil, appErrors.ErrAddressNotFound.WithDetails(err.Error())
	}

	return s.mapAddressToResponse(address), nil
}

// DeleteUserAddress deletes a user address
func (s *UserService) DeleteUserAddress(ctx context.Context, userID, addressID string) error {
	addressIdUint, err := parseUint(addressID)
	if err != nil {
		return appErrors.ErrValidationFailed.WithDetails("invalid address ID format")
	}

	// Verify address exists
	_, err = s.deps.AddressRepo.GetAddressById(ctx, userID, uint(addressIdUint))
	if err != nil {
		return appErrors.ErrAddressNotFound.WithDetails(err.Error())
	}

	if err := s.deps.AddressRepo.DeleteAddressById(ctx, uint(addressIdUint), userID); err != nil {
		return appErrors.ErrDatabaseError.WithDetails(err.Error())
	}

	return nil
}

// AddUserVehicle adds a new vehicle for user
func (s *UserService) AddUserVehicle(ctx context.Context, userID string, req dto.VehicleRequest) (*dto.VehicleResponse, error) {
	// Verify user exists
	_, err := s.deps.UserRepo.GetDetailUser(ctx, userID)
	if err != nil {
		return nil, appErrors.ErrUserNotFound.WithDetails(err.Error())
	}

	// Create vehicle model
	vehicle := models.Vehicle{
		VehicleType:   req.VehicleType,
		VehicleNumber: req.VehicleNumber,
		VehicleColor:  req.VehicleColor,
		UserID:        userID,
	}

	createdVehicle, err := s.deps.VehicleRepo.CreateVehicle(ctx, vehicle)
	if err != nil {
		return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
	}

	return s.mapVehicleToResponse(&createdVehicle), nil
}

// GetUserVehicle gets a specific vehicle for user
func (s *UserService) GetUserVehicle(ctx context.Context, userID, vehicleID string) (*dto.VehicleResponse, error) {
	vehicleIdUint, err := parseUint(vehicleID)
	if err != nil {
		return nil, appErrors.ErrValidationFailed.WithDetails("invalid vehicle ID format")
	}

	vehicle, err := s.deps.VehicleRepo.GetVehicleById(ctx, userID, uint(vehicleIdUint))
	if err != nil {
		return nil, appErrors.ErrVehicleNotFound.WithDetails(err.Error())
	}

	return s.mapVehicleToResponse(vehicle), nil
}

// GetAllUserVehicles gets all vehicles for a user
func (s *UserService) GetAllUserVehicles(ctx context.Context, userID string) ([]dto.VehicleResponse, error) {
	vehicles, err := s.deps.VehicleRepo.GetAllVehiclesByUserId(ctx, userID)
	if err != nil {
		return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
	}

	responses := make([]dto.VehicleResponse, len(vehicles))
	for i, v := range vehicles {
		responses[i] = *s.mapVehicleToResponse(&v)
	}

	return responses, nil
}

// DeleteUserVehicle deletes a user vehicle
func (s *UserService) DeleteUserVehicle(ctx context.Context, userID, vehicleID string) error {
	vehicleIdUint, err := parseUint(vehicleID)
	if err != nil {
		return appErrors.ErrValidationFailed.WithDetails("invalid vehicle ID format")
	}

	// Verify vehicle exists
	_, err = s.deps.VehicleRepo.GetVehicleById(ctx, userID, uint(vehicleIdUint))
	if err != nil {
		return appErrors.ErrVehicleNotFound.WithDetails(err.Error())
	}

	if err := s.deps.VehicleRepo.DeleteVehicleById(ctx, uint(vehicleIdUint), userID); err != nil {
		return appErrors.ErrDatabaseError.WithDetails(err.Error())
	}

	return nil
}

// Helper methods

func (s *UserService) mapUserToProfileResponse(user *models.User) *dto.UserProfileResponse {
	addresses := make([]dto.AddressResponse, len(user.Addresses))
	for i, addr := range user.Addresses {
		addresses[i] = *s.mapAddressToResponse(&addr)
	}

	vehicles := make([]dto.VehicleResponse, len(user.Vehicles))
	for i, veh := range user.Vehicles {
		vehicles[i] = *s.mapVehicleToResponse(&veh)
	}

	return &dto.UserProfileResponse{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		AvatarURL:   user.AvatarUrl,
		Addresses:   addresses,
		Vehicles:    vehicles,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func (s *UserService) mapAddressToResponse(addr *models.UserAddress) *dto.AddressResponse {
	isPrimary := false
	if addr.IsPrimary != nil {
		isPrimary = *addr.IsPrimary
	}

	return &dto.AddressResponse{
		ID:           addr.ID,
		Latitude:     addr.Latitude,
		Longitude:    addr.Longitude,
		AddressLabel: addr.AddressLabel,
		FullAddress:  addr.FullAddress,
		Note:         addr.Note,
		IsPrimary:    isPrimary,
	}
}

func (s *UserService) mapVehicleToResponse(veh *models.Vehicle) *dto.VehicleResponse {
	return &dto.VehicleResponse{
		ID:            veh.ID,  // uint type as per DTO
		VehicleType:   veh.VehicleType,
		VehicleNumber: veh.VehicleNumber,
		VehicleColor:  veh.VehicleColor,
	}
}

// parseUint helper function
func parseUint(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

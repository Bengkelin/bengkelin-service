package service

import (
	"context"
	"strconv"

	"github.com/Bengkelin/bengkelin-service/internal/dto"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/errors"
	"github.com/Bengkelin/bengkelin-service/internal/models"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
)

// MitraService implements business logic for mitra operations
type MitraService struct {
	deps ServiceDependencies
}

// NewMitraService creates a new mitra service
func NewMitraService(deps ServiceDependencies) MitraServiceInterface {
	return &MitraService{deps: deps}
}

// GetMitraProfile gets mitra profile with caching
func (s *MitraService) GetMitraProfile(ctx context.Context, mitraID string) (*models.Mitra, error) {
	// Try cache first
	cacheService := GetProfileCacheService()
	cachedMitra, err := cacheService.GetMitraProfile(ctx, mitraID)
	if err == nil && cachedMitra != nil {
		return cachedMitra, nil
	}

	// Fetch from database
	mitra, err := s.deps.MitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return nil, appErrors.ErrMitraNotFound.WithDetails(err.Error())
	}

	// Store in cache (best effort)
	if cacheErr := cacheService.SetMitraProfile(ctx, mitra); cacheErr != nil {
		applog.WarnCtx(ctx, "Failed to cache mitra profile", "error", cacheErr.Error(), "mitra_id", mitraID)
	}

	return mitra, nil
}

// UpdateMitraProfile updates mitra profile and invalidates cache
func (s *MitraService) UpdateMitraProfile(ctx context.Context, mitraID string, req dto.UpdateMitraRequest) (*models.Mitra, error) {
	// Get existing mitra
	mitra, err := s.deps.MitraRepo.GetMitraByID(ctx, mitraID)
	if err != nil {
		return nil, appErrors.ErrMitraNotFound.WithDetails(err.Error())
	}

	// Update fields
	mitraUpdated := false
	if req.FirstName != "" {
		mitra.FirstName = req.FirstName
		mitraUpdated = true
	}
	if req.LastName != "" {
		mitra.LastName = req.LastName
		mitraUpdated = true
	}
	if req.PhoneNumber != "" {
		mitra.PhoneNumber = req.PhoneNumber
		mitraUpdated = true
	}

	// Save to database
	if mitraUpdated {
		if err := s.deps.MitraRepo.UpdateMitra(ctx, mitraID, mitra); err != nil {
			return nil, appErrors.ErrDatabaseError.WithDetails(err.Error())
		}

		// Invalidate cache
		cacheService := GetProfileCacheService()
		if cacheErr := cacheService.InvalidateMitraProfile(ctx, mitraID); cacheErr != nil {
			applog.WarnCtx(ctx, "Failed to invalidate mitra profile cache", "error", cacheErr.Error(), "mitra_id", mitraID)
		}
	}

	// Return updated mitra
	return s.GetMitraProfile(ctx, mitraID)
}

// CreateMitraBank creates bank account for mitra
func (s *MitraService) CreateMitraBank(ctx context.Context, mitraID string, req dto.MitraBankRequest) (*dto.MitraBankResponse, error) {
	// Check if mitra exists
	mitra, err := s.deps.MitraRepo.GetMitraByID(ctx, mitraID)
	if err != nil {
		return nil, appErrors.ErrMitraNotFound.WithDetails(err.Error())
	}

	// Check if bank account already exists
	if mitra.BankName != "" || mitra.BankNumber != "" {
		return nil, appErrors.ErrValidationFailed.WithDetails("bank account already exists. Please update existing bank account instead")
	}

	// Update mitra with bank information
	mitraUpdate := &models.Mitra{
		BankName:   req.BankName,
		BankNumber: req.BankNumber,
	}

	err = s.deps.MitraRepo.UpdateMitra(ctx, mitraID, mitraUpdate)
	if err != nil {
		return nil, appErrors.ErrDatabaseError.WithDetails("Failed to create bank account")
	}

	// Invalidate cache
	cacheService := GetProfileCacheService()
	if cacheErr := cacheService.InvalidateMitraProfile(ctx, mitraID); cacheErr != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra profile cache", "error", cacheErr.Error(), "mitra_id", mitraID)
	}

	return &dto.MitraBankResponse{
		BankName:   req.BankName,
		BankNumber: req.BankNumber,
	}, nil
}

// UpdateMitraBank updates bank account for mitra
func (s *MitraService) UpdateMitraBank(ctx context.Context, mitraID string, req dto.MitraBankRequest) (*dto.MitraBankResponse, error) {
	// Check if mitra exists
	mitra, err := s.deps.MitraRepo.GetMitraByID(ctx, mitraID)
	if err != nil {
		return nil, appErrors.ErrMitraNotFound.WithDetails(err.Error())
	}

	// Check if bank account exists
	if mitra.BankName == "" && mitra.BankNumber == "" {
		return nil, appErrors.ErrValidationFailed.WithDetails("bank account not found. Please create a bank account first")
	}

	// Update mitra with new bank information
	mitraUpdate := &models.Mitra{
		BankName:   req.BankName,
		BankNumber: req.BankNumber,
	}

	err = s.deps.MitraRepo.UpdateMitra(ctx, mitraID, mitraUpdate)
	if err != nil {
		return nil, appErrors.ErrDatabaseError.WithDetails("Failed to update bank account")
	}

	// Invalidate cache
	cacheService := GetProfileCacheService()
	if cacheErr := cacheService.InvalidateMitraProfile(ctx, mitraID); cacheErr != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra profile cache", "error", cacheErr.Error(), "mitra_id", mitraID)
	}

	return &dto.MitraBankResponse{
		BankName:   req.BankName,
		BankNumber: req.BankNumber,
	}, nil
}

// Helper function
func parseMitraID(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

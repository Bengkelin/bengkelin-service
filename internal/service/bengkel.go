package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/dto"
	"github.com/Bengkelin/bengkelin-service/internal/models"
	"github.com/Bengkelin/bengkelin-service/internal/repository"
	"github.com/Bengkelin/bengkelin-service/internal/helpers"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
)

type BengkelService struct {
	bengkelRepo      repository.BengkelRepositoryInterface
	mitraRepo        repository.MitraRepositoryInterface
	bengkelOpRepo    repository.BengkelOperationalRepositoryInterface
	bengkelAddrRepo  repository.BengkelAddressRepositoryInterface
	bengkelSvcRepo   repository.BengkelServiceRepositoryInterface
	bengkelPhotoRepo repository.BengkelPhotoRepositoryInterface
	bengkelTestRepo  repository.BengkelTestimonialRepositoryInterface
	orderRepo        repository.OrderRepositoryInterface
	cacheService     *ProfileCacheService
}

func NewBengkelService(deps ServiceDependencies) BengkelServiceInterface {
	return &BengkelService{
		bengkelRepo:      deps.BengkelRepo,
		mitraRepo:        deps.MitraRepo,
		bengkelOpRepo:    deps.BengkelOperationalRepo,
		bengkelAddrRepo:  deps.BengkelAddressRepo,
		bengkelSvcRepo:   deps.BengkelServiceRepo,
		bengkelPhotoRepo: deps.BengkelPhotoRepo,
		bengkelTestRepo:  deps.BengkelTestimonialRepo,
		orderRepo:        deps.OrderRepo,
		cacheService:     GetProfileCacheService(),
	}
}

// CreateBengkel creates a new bengkel with operational hours
func (s *BengkelService) CreateBengkel(ctx context.Context, mitraID string, req dto.CreateBengkelRequest) (*dto.BengkelResponse, error) {
	// Verify mitra exists
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return nil, fmt.Errorf("mitra not found: %w", err)
	}

	// Create bengkel
	bengkel := models.Bengkel{
		ID:           helpers.GenerateUUID(),
		MitraID:      mitra.ID,
		BengkelName:  req.BengkelName,
		BengkelPhone: req.BengkelPhone,
		JumlahMontir: req.JumlahMontir,
	}

	createdBengkel, err := s.bengkelRepo.CreateBengkel(ctx, bengkel)
	if err != nil {
		applog.ErrorCtx(ctx, "Failed to create bengkel", "error", err, "mitra_id", mitraID)
		return nil, fmt.Errorf("failed to create bengkel: %w", err)
	}

	// Create operational hours
	for i, hari := range req.Hari {
		operational := models.BengkelOperational{
			BengkelID: createdBengkel.ID,
			Hari:      hari,
			JamBuka:   req.JamBuka[i],
		}
		if _, err := s.bengkelOpRepo.CreateBengkelOperational(ctx, operational); err != nil {
			applog.ErrorCtx(ctx, "Failed to create operational hours", "error", err, "bengkel_id", createdBengkel.ID)
			// Continue - don't fail the whole operation
		}
	}

	// Invalidate cache
	if err := s.cacheService.InvalidateMitraProfile(ctx, mitraID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra cache", "error", err, "mitra_id", mitraID)
	}

	return s.mapToBengkelResponse(&createdBengkel), nil
}

// UpdateBengkelProfile updates bengkel profile information
func (s *BengkelService) UpdateBengkelProfile(ctx context.Context, mitraID string, req dto.UpdateBengkelRequest) (*dto.BengkelResponse, error) {
	// Get mitra to find bengkel
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return nil, fmt.Errorf("mitra not found: %w", err)
	}

	if len(mitra.Bengkel) == 0 {
		return nil, fmt.Errorf("mitra has no bengkel")
	}

	bengkelID := mitra.Bengkel[0].ID

	// Update bengkel
	updateData := &models.Bengkel{
		BengkelName:  req.BengkelName,
		BengkelPhone: req.BengkelPhone,
		JumlahMontir: req.JumlahMontir,
	}

	if err := s.bengkelRepo.UpdateBengkelById(ctx, bengkelID, updateData); err != nil {
		applog.ErrorCtx(ctx, "Failed to update bengkel", "error", err, "bengkel_id", bengkelID)
		return nil, fmt.Errorf("failed to update bengkel: %w", err)
	}

	// Invalidate caches
	if err := s.cacheService.InvalidateBengkelDetail(ctx, bengkelID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate bengkel cache", "error", err, "bengkel_id", bengkelID)
	}
	if err := s.cacheService.InvalidateMitraProfile(ctx, mitraID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra cache", "error", err, "mitra_id", mitraID)
	}

	// Get updated bengkel
	updated, err := s.bengkelRepo.GetBengkelById(ctx, bengkelID)
	if err != nil {
		return nil, err
	}

	return s.mapToBengkelResponse(updated), nil
}

// GetBengkelProfile gets bengkel profile for a mitra
func (s *BengkelService) GetBengkelProfile(ctx context.Context, mitraID string) (*dto.BengkelResponse, error) {
	// Try cache first
	cached, err := s.cacheService.GetMitraProfile(ctx, mitraID)
	if err == nil && cached != nil && len(cached.Bengkel) > 0 {
		return s.mapToBengkelResponse(&cached.Bengkel[0]), nil
	}

	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return nil, fmt.Errorf("mitra not found: %w", err)
	}

	if len(mitra.Bengkel) == 0 {
		return nil, fmt.Errorf("mitra has no bengkel")
	}

	return s.mapToBengkelResponse(&mitra.Bengkel[0]), nil
}

// UpdateBengkelMontir updates the number of mechanics
func (s *BengkelService) UpdateBengkelMontir(ctx context.Context, mitraID string, jumlahMontir uint) error {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return fmt.Errorf("mitra has no bengkel")
	}

	bengkelID := mitra.Bengkel[0].ID
	updateData := &models.Bengkel{JumlahMontir: jumlahMontir}

	if err := s.bengkelRepo.UpdateBengkelById(ctx, bengkelID, updateData); err != nil {
		applog.ErrorCtx(ctx, "Failed to update bengkel montir", "error", err, "bengkel_id", bengkelID)
		return fmt.Errorf("failed to update bengkel montir: %w", err)
	}

	if err := s.cacheService.InvalidateBengkelDetail(ctx, bengkelID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate bengkel cache", "error", err, "bengkel_id", bengkelID)
	}
	if err := s.cacheService.InvalidateMitraProfile(ctx, mitraID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra cache", "error", err, "mitra_id", mitraID)
	}

	return nil
}

// UpdateBengkelOperational updates operational hours
func (s *BengkelService) UpdateBengkelOperational(ctx context.Context, mitraID string, operationals []dto.OperationalItemRequest) error {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return fmt.Errorf("mitra has no bengkel")
	}

	bengkelID := mitra.Bengkel[0].ID

	for _, op := range operationals {
		operationalModel := models.BengkelOperational{
			BengkelID: bengkelID,
			Hari:      op.Hari,
			JamBuka:   op.JamBuka,
			JamTutup:  op.JamTutup,
			IsActive:  op.IsActive,
		}

		if op.ID > 0 {
			operationalModel.ID = op.ID
			if err := s.bengkelOpRepo.UpdateBengkelOperationalById(ctx, bengkelID, op.Hari, &operationalModel); err != nil {
				applog.ErrorCtx(ctx, "Failed to update operational", "error", err)
				return err
			}
		} else {
			if operationalModel.IsActive == nil {
				defaultActive := true
				operationalModel.IsActive = &defaultActive
			}
			if _, err := s.bengkelOpRepo.CreateBengkelOperational(ctx, operationalModel); err != nil {
				applog.ErrorCtx(ctx, "Failed to create operational", "error", err)
				return err
			}
		}
	}

	if err := s.cacheService.InvalidateBengkelDetail(ctx, bengkelID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate bengkel cache", "error", err, "bengkel_id", bengkelID)
	}
	if err := s.cacheService.InvalidateMitraProfile(ctx, mitraID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra cache", "error", err, "mitra_id", mitraID)
	}

	return nil
}

// CreateBengkelAddress creates a new address for bengkel
func (s *BengkelService) CreateBengkelAddress(ctx context.Context, mitraID string, req dto.BengkelAddressRequest) error {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return fmt.Errorf("mitra has no bengkel")
	}

	addressRepo := s.bengkelAddrRepo
	addressModel := models.BengkelAddress{
		BengkelID:    mitra.Bengkel[0].ID,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		AddressLabel: req.AddressLabel,
		FullAddress:  req.FullAddress,
		Note:         req.Note,
	}

	if _, err := addressRepo.CreateBengkelAddress(ctx, addressModel); err != nil {
		applog.ErrorCtx(ctx, "Failed to create address", "error", err)
		return err
	}

	if err := s.cacheService.InvalidateBengkelDetail(ctx, mitra.Bengkel[0].ID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate bengkel cache", "error", err)
	}
	if err := s.cacheService.InvalidateMitraProfile(ctx, mitraID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra cache", "error", err)
	}

	return nil
}

// CreateBengkelServices creates new services for bengkel
func (s *BengkelService) CreateBengkelServices(ctx context.Context, mitraID string, services []dto.BengkelServiceItemRequest) error {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return fmt.Errorf("mitra has no bengkel")
	}

	bengkelID := mitra.Bengkel[0].ID
	serviceRepo := s.bengkelSvcRepo

	for _, svc := range services {
		if svc.IsAvailable == nil {
			defaultAvailable := true
			svc.IsAvailable = &defaultAvailable
		}
		serviceModel := models.BengkelService{
			BengkelID:   bengkelID,
			NamaService: svc.NamaService,
			Description: svc.Description,
			Price:       svc.Price,
			IsAvailable: svc.IsAvailable,
		}
		if _, err := serviceRepo.CreateBengkelService(ctx, serviceModel); err != nil {
			applog.ErrorCtx(ctx, "Failed to create service", "error", err)
			return err
		}
	}

	if err := s.cacheService.InvalidateBengkelDetail(ctx, bengkelID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate bengkel cache", "error", err)
	}
	if err := s.cacheService.InvalidateMitraProfile(ctx, mitraID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra cache", "error", err)
	}

	return nil
}

// UpdateBengkelServices updates existing services
func (s *BengkelService) UpdateBengkelServices(ctx context.Context, mitraID string, services []dto.BengkelServiceItemRequest) error {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return fmt.Errorf("mitra has no bengkel")
	}

	bengkelID := mitra.Bengkel[0].ID
	serviceRepo := s.bengkelSvcRepo

	for _, svc := range services {
		serviceModel := models.BengkelService{
			BengkelID:   bengkelID,
			NamaService: svc.NamaService,
			Description: svc.Description,
			Price:       svc.Price,
			IsAvailable: svc.IsAvailable,
		}

		if svc.ID > 0 {
			serviceModel.ID = svc.ID
			if err := serviceRepo.UpdateBengkelService(ctx, svc.ID, &serviceModel); err != nil {
				applog.ErrorCtx(ctx, "Failed to update service", "error", err)
				return err
			}
		} else {
			if serviceModel.IsAvailable == nil {
				defaultAvailable := true
				serviceModel.IsAvailable = &defaultAvailable
			}
			if _, err := serviceRepo.CreateBengkelService(ctx, serviceModel); err != nil {
				applog.ErrorCtx(ctx, "Failed to create service", "error", err)
				return err
			}
		}
	}

	if err := s.cacheService.InvalidateBengkelDetail(ctx, bengkelID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate bengkel cache", "error", err)
	}
	if err := s.cacheService.InvalidateMitraProfile(ctx, mitraID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra cache", "error", err)
	}

	return nil
}

// CreateBengkelPhotos creates photo records for bengkel
func (s *BengkelService) CreateBengkelPhotos(ctx context.Context, mitraID string, photoURLs []string) error {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return fmt.Errorf("mitra has no bengkel")
	}

	bengkelID := mitra.Bengkel[0].ID
	photoRepo := s.bengkelPhotoRepo

	for _, url := range photoURLs {
		photoModel := models.BengkelPhoto{
			BengkelID: bengkelID,
			PhotoURL:  url,
		}
		if _, err := photoRepo.CreateBengkelPhoto(ctx, photoModel); err != nil {
			applog.ErrorCtx(ctx, "Failed to create photo", "error", err)
			return err
		}
	}

	if err := s.cacheService.InvalidateBengkelDetail(ctx, bengkelID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate bengkel cache", "error", err)
	}
	if err := s.cacheService.InvalidateMitraProfile(ctx, mitraID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra cache", "error", err)
	}

	// Re-fetch and re-cache fresh data
	if freshMitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID); err == nil {
		_ = s.cacheService.SetMitraProfile(ctx, freshMitra)
		applog.InfoCtx(ctx, "Mitra profile re-cached after photo upload", "mitra_id", mitraID)
	}
	if freshBengkel, err := s.bengkelRepo.GetBengkelByIdFresh(ctx, bengkelID); err == nil {
		_ = s.cacheService.SetBengkelDetail(ctx, freshBengkel)
		applog.InfoCtx(ctx, "Bengkel detail re-cached after photo upload", "bengkel_id", bengkelID)
	}

	return nil
}

// GetBengkelPhotoURL returns the URL of a photo by its ID
func (s *BengkelService) GetBengkelPhotoURL(ctx context.Context, photoID string) (string, error) {
	photo, err := s.bengkelPhotoRepo.GetBengkelPhotoByPK(ctx, photoID)
	if err != nil {
		return "", err
	}
	return photo.PhotoURL, nil
}

// DeleteBengkelPhoto deletes a photo from bengkel and invalidates cache
func (s *BengkelService) DeleteBengkelPhoto(ctx context.Context, mitraID string, photoID string) error {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return fmt.Errorf("mitra has no bengkel")
	}

	bengkelID := mitra.Bengkel[0].ID

	if err := s.bengkelPhotoRepo.DeleteBengkelPhotoById(ctx, photoID); err != nil {
		return fmt.Errorf("failed to delete photo: %w", err)
	}

	// Invalidate and re-cache
	_ = s.cacheService.InvalidateBengkelDetail(ctx, bengkelID)
	_ = s.cacheService.InvalidateMitraProfile(ctx, mitraID)

	if freshMitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID); err == nil {
		_ = s.cacheService.SetMitraProfile(ctx, freshMitra)
	}
	if freshBengkel, err := s.bengkelRepo.GetBengkelByIdFresh(ctx, bengkelID); err == nil {
		_ = s.cacheService.SetBengkelDetail(ctx, freshBengkel)
	}

	return nil
}

// UpdateBengkelStatusOptions updates service availability options
func (s *BengkelService) UpdateBengkelStatusOptions(ctx context.Context, mitraID string, homeService, storeService, isOpen bool) error {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return fmt.Errorf("mitra has no bengkel")
	}

	bengkelID := mitra.Bengkel[0].ID
	updateData := &models.Bengkel{
		HomeService:  &homeService,
		StoreService: &storeService,
		IsOpen:       &isOpen,
	}

	if err := s.bengkelRepo.UpdateBengkelById(ctx, bengkelID, updateData); err != nil {
		applog.ErrorCtx(ctx, "Failed to update status options", "error", err)
		return err
	}

	if err := s.cacheService.InvalidateBengkelDetail(ctx, bengkelID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate bengkel cache", "error", err)
	}
	if err := s.cacheService.InvalidateMitraProfile(ctx, mitraID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra cache", "error", err)
	}

	return nil
}

// UpdateBengkelAvatar updates the bengkel avatar URL
func (s *BengkelService) UpdateBengkelAvatar(ctx context.Context, mitraID string, avatarURL string) error {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return fmt.Errorf("mitra has no bengkel")
	}

	bengkelID := mitra.Bengkel[0].ID
	updateData := &models.Bengkel{AvatarUrl: avatarURL}

	if err := s.bengkelRepo.UpdateBengkelById(ctx, bengkelID, updateData); err != nil {
		applog.ErrorCtx(ctx, "Failed to update avatar", "error", err)
		return err
	}

	if err := s.cacheService.InvalidateBengkelDetail(ctx, bengkelID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate bengkel cache", "error", err)
	}
	if err := s.cacheService.InvalidateMitraProfile(ctx, mitraID); err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra cache", "error", err)
	}

	return nil
}

// SearchBengkels searches for bengkels
func (s *BengkelService) SearchBengkels(ctx context.Context, req dto.SearchBengkelRequest) (*dto.PaginatedBengkelResponse, error) {
	// TODO: Implement search logic
	return nil, nil
}

// GetNearestBengkels finds nearest bengkels
func (s *BengkelService) GetNearestBengkels(ctx context.Context, req dto.NearestBengkelRequest) (*dto.PaginatedBengkelResponse, error) {
	// TODO: Implement nearest search logic
	return nil, nil
}

// CreateTestimonial creates a testimonial for a bengkel after validating user, bengkel, and order
func (s *BengkelService) CreateTestimonial(ctx context.Context, userID, bengkelID, orderID string, testimoni string, rating int) error {
	if _, err := s.orderRepo.GetOrderById(ctx, orderID); err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if _, err := s.bengkelRepo.GetBengkelById(ctx, bengkelID); err != nil {
		return fmt.Errorf("bengkel not found: %w", err)
	}

	testimonial := models.BengkelTestimonial{
		BengkelID: bengkelID,
		UserID:    userID,
		OrderID:   orderID,
		Testimoni: testimoni,
		Rating:    rating,
	}

	if _, err := s.bengkelTestRepo.CreateBengkelTestimonial(ctx, testimonial); err != nil {
		return fmt.Errorf("failed to create testimonial: %w", err)
	}

	return nil
}

// mapToBengkelResponse converts model to DTO
func (s *BengkelService) mapToBengkelResponse(b *models.Bengkel) *dto.BengkelResponse {
	return &dto.BengkelResponse{
		ID:           b.ID,
		MitraID:      b.MitraID,
		BengkelName:  b.BengkelName,
		BengkelPhone: b.BengkelPhone,
		JumlahMontir: b.JumlahMontir,
		AvatarURL:    b.AvatarUrl,
		HomeService:  b.HomeService != nil && *b.HomeService,
		StoreService: b.StoreService != nil && *b.StoreService,
		IsOpen:       b.IsOpen != nil && *b.IsOpen,
		CreatedAt:    b.CreatedAt,
		UpdatedAt:    b.UpdatedAt,
	}
}

// GetMitraWithBengkel gets mitra with bengkel data
func (s *BengkelService) GetMitraWithBengkel(ctx context.Context, mitraID string) (*models.Mitra, error) {
	return s.mitraRepo.FindMitraByID(ctx, mitraID)
}

// GetAllBengkels gets all bengkels
func (s *BengkelService) GetAllBengkels(ctx context.Context) ([]dto.BengkelResponse, error) {
	bengkels, err := s.bengkelRepo.GetAllBengkel(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.BengkelResponse, len(bengkels))
	for i, b := range bengkels {
		result[i] = *s.mapToBengkelResponse(&b)
	}
	return result, nil
}

// GetAllBengkelsPaginate gets paginated bengkels
func (s *BengkelService) GetAllBengkelsPaginate(ctx context.Context, page, limit int) ([]dto.BengkelResponse, int, error) {
	bengkels, count, err := s.bengkelRepo.GetAllBengkelPaginate(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.BengkelResponse, len(bengkels))
	for i, b := range bengkels {
		result[i] = *s.mapToBengkelResponse(&b)
	}
	return result, count, nil
}

// SearchBengkelsV2 searches bengkels with filters
func (s *BengkelService) SearchBengkelsV2(ctx context.Context, serviceType, query string, page, limit int) ([]dto.BengkelResponse, int, error) {
	bengkels, count, err := s.bengkelRepo.GetBengkelSearchV2(ctx, serviceType, query, page, limit)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.BengkelResponse, len(bengkels))
	for i, b := range bengkels {
		result[i] = *s.mapToBengkelResponse(&b)
	}
	return result, count, nil
}

// GetBengkelDetailWithTestimonials gets bengkel with paginated testimonials
func (s *BengkelService) GetBengkelDetailWithTestimonials(ctx context.Context, bengkelID string, page, limit int) (*models.Bengkel, []models.BengkelTestimonial, int, error) {
	bengkel, testimonials, count, err := s.bengkelRepo.FindBengkelById(ctx, bengkelID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return bengkel, testimonials, count, nil
}

// GetBengkelFullDetail gets full bengkel detail with cache support
func (s *BengkelService) GetBengkelFullDetail(ctx context.Context, bengkelID string, forceFresh bool, page, limit int) (*models.Bengkel, []models.BengkelTestimonial, int, error) {
	var bengkel *models.Bengkel
	var err error

	if forceFresh {
		bengkel, err = s.bengkelRepo.GetBengkelByIdFresh(ctx, bengkelID)
	} else {
		bengkel, err = s.bengkelRepo.GetBengkelById(ctx, bengkelID)
	}
	if err != nil {
		return nil, nil, 0, err
	}

	_, testimonials, testimonialCount, err := s.bengkelRepo.FindBengkelById(ctx, bengkelID, page, limit)
	if err != nil {
		testimonials = []models.BengkelTestimonial{}
		testimonialCount = 0
	}

	return bengkel, testimonials, testimonialCount, nil
}

// GetBengkelOperationalTimeSlots gets time slots for a bengkel on a specific day
func (s *BengkelService) GetBengkelOperationalTimeSlots(ctx context.Context, bengkelID, day string) (map[int]string, error) {
	operational, err := s.bengkelOpRepo.GetBengkelOperationalByIdAndDay(ctx, bengkelID, day)
	if err != nil {
		return nil, err
	}

	var dataTimePerHours []string
	if operational.JamBuka != "" {
		dataTimePerHours = strings.Split(operational.JamBuka, " - ")
	}

	if len(dataTimePerHours) < 2 {
		return map[int]string{}, nil
	}

	start, _ := time.Parse("15:04", dataTimePerHours[0])
	end, _ := time.Parse("15:04", dataTimePerHours[1])

	data := make(map[int]string)
	i := 0
	for current := start; current.Before(end); current = current.Add(time.Hour) {
		next := current.Add(time.Hour)
		data[i] = current.Format("15:04") + " - " + next.Format("15:04")
		i++
	}

	return data, nil
}

// SearchBengkelsPublic searches bengkels for public API
func (s *BengkelService) SearchBengkelsPublic(ctx context.Context, criteria map[string]interface{}) ([]models.Bengkel, int, error) {
	return s.bengkelRepo.SearchBengkelPublic(ctx, criteria)
}

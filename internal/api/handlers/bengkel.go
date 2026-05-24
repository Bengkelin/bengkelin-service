package handlers

import (
	"net/http"
	"strconv"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/container"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/pkg/errors"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/service"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/validator"
	"github.com/Bengkelin/bengkelin-service/pkg/response"
	"github.com/gin-gonic/gin"
)

var (
	bengkelHandler *BengkelHandler
)

// BengkelHandler struct
type BengkelHandler struct {
	BaseHandler
	bengkelService  service.BengkelServiceInterface
	orderService    service.OrderServiceInterface
	userService     service.UserServiceInterface
	uploadService   *service.FileUploadService
}

// NewBengkelHandler function
func GetBengkelHandler() *BengkelHandler {
	if bengkelHandler == nil {
		c := container.GetContainer()
		bengkelHandler = &BengkelHandler{
			bengkelService: c.BengkelService,
			orderService:   c.OrderService,
			userService:    c.UserService,
			uploadService:  service.NewFileUploadService(),
		}
	}
	return bengkelHandler
}

// --- DTO Mapping Functions ---

func toBengkelResponse(b *models.Bengkel) dto.BengkelResponse {
	resp := dto.BengkelResponse{
		ID:           b.ID,
		MitraID:      b.MitraID,
		BengkelName:  b.BengkelName,
		BengkelPhone: b.BengkelPhone,
		JumlahMontir: b.JumlahMontir,
		AvatarURL:    b.AvatarUrl,
		CreatedAt:    b.CreatedAt,
		UpdatedAt:    b.UpdatedAt,
	}
	if b.HomeService != nil {
		resp.HomeService = *b.HomeService
	}
	if b.StoreService != nil {
		resp.StoreService = *b.StoreService
	}
	if b.IsOpen != nil {
		resp.IsOpen = *b.IsOpen
	}
	if len(b.Services) > 0 {
		services := make([]string, 0, len(b.Services))
		for _, s := range b.Services {
			services = append(services, s.NamaService)
		}
		resp.Services = services
	}
	if len(b.Photos) > 0 {
		photos := make([]string, 0, len(b.Photos))
		for _, p := range b.Photos {
			photos = append(photos, p.PhotoURL)
		}
		resp.Photos = photos
	}
	if len(b.Operasionals) > 0 {
		ops := make([]dto.OperationalResponse, 0, len(b.Operasionals))
		for _, o := range b.Operasionals {
			ops = append(ops, dto.OperationalResponse{
				ID:        strconv.FormatUint(uint64(o.ID), 10),
				BengkelID: o.BengkelID,
				Hari:      o.Hari,
				JamBuka:   o.JamBuka,
			})
		}
		resp.Operational = ops
	}
	if len(b.Addresses) > 0 {
		resp.Address = &dto.AddressResponse{
			ID:          b.Addresses[0].ID,
			Latitude:    b.Addresses[0].Latitude,
			Longitude:   b.Addresses[0].Longitude,
			FullAddress: b.Addresses[0].FullAddress,
		}
	}
	return resp
}

func toBengkelResponseList(bengkels []models.Bengkel) []dto.BengkelResponse {
	result := make([]dto.BengkelResponse, 0, len(bengkels))
	for _, b := range bengkels {
		result = append(result, toBengkelResponse(&b))
	}
	return result
}

func toMitraWithBengkelResponse(m *models.Mitra) *dto.MitraProfileResponse {
	return &dto.MitraProfileResponse{
		ID:          m.ID,
		FirstName:   m.FirstName,
		LastName:    m.LastName,
		Email:       m.Email,
		PhoneNumber: m.PhoneNumber,
		BankName:    m.BankName,
		BankNumber:  m.BankNumber,
		Bengkel:     toBengkelResponseList(m.Bengkel),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toUserResponse(u *models.User) *dto.UserProfileResponse {
	resp := &dto.UserProfileResponse{
		ID:          u.ID,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Email:       u.Email,
		PhoneNumber: u.PhoneNumber,
		AvatarURL:   u.AvatarUrl,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
	if len(u.Addresses) > 0 {
		addrs := make([]dto.AddressResponse, 0, len(u.Addresses))
		for _, a := range u.Addresses {
			ar := dto.AddressResponse{
				ID:           a.ID,
				Latitude:     a.Latitude,
				Longitude:    a.Longitude,
				AddressLabel: a.AddressLabel,
				FullAddress:  a.FullAddress,
				Note:         a.Note,
			}
			if a.IsPrimary != nil {
				ar.IsPrimary = *a.IsPrimary
			}
			addrs = append(addrs, ar)
		}
		resp.Addresses = addrs
	}
	if len(u.Vehicles) > 0 {
		vehs := make([]dto.VehicleResponse, 0, len(u.Vehicles))
		for _, v := range u.Vehicles {
			vehs = append(vehs, dto.VehicleResponse{
				ID:            v.ID,
				VehicleType:   v.VehicleType,
				VehicleColor:  v.VehicleColor,
				VehicleNumber: v.VehicleNumber,
			})
		}
		resp.Vehicles = vehs
	}
	return resp
}

func toOrderResponse(o *models.Order) dto.OrderResponse {
	resp := dto.OrderResponse{
		ID:         o.ID,
		UserID:     o.UserID,
		BengkelID:  o.BengkelID,
		TotalPrice: o.TotalPrice,
		Notes:      o.Note,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
	if o.IsHomeService != nil && *o.IsHomeService {
		resp.ServiceType = "home"
	} else {
		resp.ServiceType = "store"
	}
	resp.Status = o.Status.String()
	if len(o.OrderServices) > 0 {
		services := make([]dto.OrderServiceResponse, 0, len(o.OrderServices))
		for _, s := range o.OrderServices {
			services = append(services, dto.OrderServiceResponse{
				ID:          strconv.FormatUint(uint64(s.ID), 10),
				OrderID:     s.OrderID,
				ServiceName: s.Title,
				Price:       s.Price,
			})
		}
		resp.Services = services
	}
	return resp
}

// BengkelHandlerInterface interface
type BengkelHandlerInterface interface {
	CreateBengkel(c *gin.Context)
	UpdateBengkelMontir(c *gin.Context)
	UpdateBengkelOperational(c *gin.Context)
	CreateBengkelAddress(c *gin.Context)
	CreateBengkelService(c *gin.Context)
	UpdateBengkelService(c *gin.Context)
	CreateBengkelPhoto(c *gin.Context)
	UpdateBengkelStatusOpsiService(c *gin.Context)
	UpdateBengkel(c *gin.Context)
	GetBengkel(c *gin.Context)
	GetAllBengkel(c *gin.Context)
	GetAllBengkelPaginate(c *gin.Context)
	GetAllBengkelPublic(c *gin.Context) // Public bengkel list (no auth)
	GetBengkelSearchPaginate(c *gin.Context)
	GetBengkelSearchV2Paginate(c *gin.Context)
	GetBengkelSearchPublic(c *gin.Context) // Public bengkel search (no auth)
	CreateBengkelTestimonial(c *gin.Context)
	GetDetailBengkelById(c *gin.Context)
	GetBengkelDetailForUser(c *gin.Context)
	CreateBengkelOrderService(c *gin.Context)
	GetBengkelOrderServiceById(c *gin.Context)
	GetBengkelOrderServiceByIdMitra(c *gin.Context)
	UpdateAvatarBengkel(c *gin.Context)
	GetBengkelOperationalByIdAndDay(c *gin.Context)
	UpdateBengkelOrderServiceById(c *gin.Context)
	GetDetailUserById(c *gin.Context)
	GetAllBengkelOrderServicePaginate(c *gin.Context)
	GetAllOrderUserPaginate(c *gin.Context)
	GetNearestBengkelPaginate(c *gin.Context)
	UpdateStatusOrderService(c *gin.Context)
}

// CreateBengkel function - refactored to use service layer
func (h *BengkelHandler) CreateBengkel(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	var requestDataBengkel validator.BengkelRegisterRequest
	if err := c.ShouldBindJSON(&requestDataBengkel); err != nil {
		h.HandleError(c, err)
		return
	}

	req := dto.CreateBengkelRequest{
		BengkelName:  requestDataBengkel.BengkelName,
		BengkelPhone: requestDataBengkel.BengkelPhone,
		JumlahMontir: requestDataBengkel.JumlahMontir,
		Hari:         requestDataBengkel.Hari,
		JamBuka:      requestDataBengkel.JamBuka,
	}

	if _, err := h.bengkelService.CreateBengkel(c.Request.Context(), mitraId, req); err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success create bengkel", nil)
	c.JSON(http.StatusOK, resp)
}

// UpdateBengkelMontir function - refactored to use service layer
func (h *BengkelHandler) UpdateBengkelMontir(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	var requestDataBengkelMontir validator.BengkelMontirUpdateRequest
	if err := c.ShouldBindJSON(&requestDataBengkelMontir); err != nil {
		h.HandleError(c, err)
		return
	}

	if err := h.bengkelService.UpdateBengkelMontir(c.Request.Context(), mitraId, requestDataBengkelMontir.JumlahMontir); err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success update bengkel montir", nil)
	c.JSON(http.StatusOK, resp)
}

// UpdateBengkelOperational function
func (h *BengkelHandler) UpdateBengkelOperational(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	var requestDataBengkelOperational validator.BengkelOperationalUpdateRequestV2
	if err := c.ShouldBindJSON(&requestDataBengkelOperational); err != nil {
		h.HandleError(c, err)
		return
	}

	// Map validator types to DTO types
	operationals := make([]dto.OperationalItemRequest, len(requestDataBengkelOperational.Operasionals))
	for i, op := range requestDataBengkelOperational.Operasionals {
		operationals[i] = dto.OperationalItemRequest{
			ID:       op.ID,
			Hari:     op.Hari,
			JamBuka:  op.JamBuka,
			JamTutup: op.JamTutup,
			IsActive: op.IsActive,
		}
	}

	if err := h.bengkelService.UpdateBengkelOperational(c.Request.Context(), mitraId, operationals); err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success update bengkel operasional", nil)
	c.JSON(http.StatusOK, resp)
}

// CreateBengkelAddress function
func (h *BengkelHandler) CreateBengkelAddress(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	var requestDataBengkelAddress validator.BengkelAddressRequest
	if err := c.ShouldBindJSON(&requestDataBengkelAddress); err != nil {
		h.HandleError(c, err)
		return
	}

	req := dto.BengkelAddressRequest{
		Latitude:     requestDataBengkelAddress.Latitude,
		Longitude:    requestDataBengkelAddress.Longitude,
		AddressLabel: requestDataBengkelAddress.AddressLabel,
		FullAddress:  requestDataBengkelAddress.FullAddress,
		Note:         requestDataBengkelAddress.Note,
	}

	if err := h.bengkelService.CreateBengkelAddress(c.Request.Context(), mitraId, req); err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success attach bengkel address", nil)
	c.JSON(http.StatusOK, resp)
}

// CreateBengkelServive function
func (h *BengkelHandler) CreateBengkelService(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	// Try new format first
	var requestDataBengkelServiceV2 validator.BengkelServiceCreateRequest
	err := c.ShouldBindJSON(&requestDataBengkelServiceV2)

	if err == nil {
		// New format - map to DTO
		services := make([]dto.BengkelServiceItemRequest, len(requestDataBengkelServiceV2.Services))
		for i, svc := range requestDataBengkelServiceV2.Services {
			services[i] = dto.BengkelServiceItemRequest{
				NamaService: svc.NamaService,
				Description: svc.Description,
				Price:       svc.Price,
				IsAvailable: svc.IsAvailable,
			}
		}

		if err := h.bengkelService.CreateBengkelServices(c.Request.Context(), mitraId, services); err != nil {
			h.HandleError(c, err)
			return
		}

		resp := response.BuildSuccessResponse("success create bengkel services", nil)
		c.JSON(http.StatusCreated, resp)
		return
	}

	// Fallback to legacy format
	var requestDataBengkelService validator.BengkelServiceRequest
	if err := c.ShouldBindJSON(&requestDataBengkelService); err != nil {
		h.HandleError(c, err)
		return
	}

	// Convert legacy format (string array) to structured format
	defaultAvailable := true
	services := make([]dto.BengkelServiceItemRequest, len(requestDataBengkelService.NamaService))
	for i, name := range requestDataBengkelService.NamaService {
		services[i] = dto.BengkelServiceItemRequest{
			NamaService: name,
			IsAvailable: &defaultAvailable,
		}
	}

	if err := h.bengkelService.CreateBengkelServices(c.Request.Context(), mitraId, services); err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success attach bengkel service", nil)
	c.JSON(http.StatusOK, resp)
}

// UpdateBengkelService function
func (h *BengkelHandler) UpdateBengkelService(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	var requestDataBengkelService validator.BengkelServiceUpdateRequest
	if err := c.ShouldBindJSON(&requestDataBengkelService); err != nil {
		h.HandleError(c, err)
		return
	}

	// Map validator types to DTO types
	services := make([]dto.BengkelServiceItemRequest, len(requestDataBengkelService.Services))
	for i, svc := range requestDataBengkelService.Services {
		services[i] = dto.BengkelServiceItemRequest{
			ID:          svc.ID,
			NamaService: svc.NamaService,
			Description: svc.Description,
			Price:       svc.Price,
			IsAvailable: svc.IsAvailable,
		}
	}

	if err := h.bengkelService.UpdateBengkelServices(c.Request.Context(), mitraId, services); err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success update bengkel services", nil)
	c.JSON(http.StatusOK, resp)
}

// CreateBengkelPhoto function
func (h *BengkelHandler) CreateBengkelPhoto(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	form, _ := c.MultipartForm()
	photos := form.File["photos"]

	if len(photos) == 0 {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("photos is empty"))
		return
	}

	// Determine protocol from request
	protocol := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		protocol = "https"
	}

	urlPictures, err := h.uploadService.UploadMultipleFiles(photos, service.PhotoUploadConfig, protocol, c.Request.Host)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("failed to upload file: "+err.Error()))
		return
	}

	if err := h.bengkelService.CreateBengkelPhotos(c.Request.Context(), mitraId, urlPictures); err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success attach new bengkel photo", nil)
	c.JSON(http.StatusOK, resp)
}

// UpdateBengkel function - refactored to use service layer
func (h *BengkelHandler) UpdateBengkel(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	var requestDataBengkel validator.BengkelUpdateRequest
	if err := c.ShouldBindJSON(&requestDataBengkel); err != nil {
		h.HandleError(c, err)
		return
	}

	req := dto.UpdateBengkelRequest{
		BengkelName:  requestDataBengkel.BengkelName,
		BengkelPhone: requestDataBengkel.BengkelPhone,
		JumlahMontir: requestDataBengkel.JumlahMontir,
	}

	_, err := h.bengkelService.UpdateBengkelProfile(c.Request.Context(), mitraId, req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success update bengkel profile", nil)
	c.JSON(http.StatusOK, resp)
}

// UpdateBengkelStatusOpsiService function
func (h *BengkelHandler) UpdateBengkelStatusOpsiService(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	var req validator.BengkelServiceOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, err)
		return
	}

	if err := h.bengkelService.UpdateBengkelStatusOptions(c.Request.Context(), mitraId, req.HomeService, req.StoreService, req.IsOpen); err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success update bengkel status option service", nil)
	c.JSON(http.StatusOK, resp)
}

// GetBengkel function
func (h *BengkelHandler) GetBengkel(c *gin.Context) {
	mitraId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	mitra, err := h.bengkelService.GetMitraWithBengkel(ctx, mitraId)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success get mitra", toMitraWithBengkelResponse(mitra))
	c.JSON(http.StatusOK, resp)
}

// GetAllBengkel function
func (h *BengkelHandler) GetAllBengkel(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if _, err := h.userService.GetUserProfile(ctx, userId); err != nil {
		h.HandleError(c, appErrors.ErrUserNotFound.WithDetails(err.Error()))
		return
	}

	bengkels, err := h.bengkelService.GetAllBengkels(ctx)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success get all bengkel", bengkels)
	c.JSON(http.StatusOK, resp)
}

// GetAllBengkelPaginate function
func (h *BengkelHandler) GetAllBengkelPaginate(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if _, err := h.userService.GetUserProfile(ctx, userId); err != nil {
		h.HandleError(c, appErrors.ErrUserNotFound.WithDetails(err.Error()))
		return
	}

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	bengkels, count, err := h.bengkelService.GetAllBengkelsPaginate(ctx, pageInt, limitInt)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success get all bengkel", map[string]any{
		"bengkels": bengkels,
		"count":    count,
	})
	c.JSON(http.StatusOK, resp)
}

// GetBengkelSearchV2Paginate function
func (h *BengkelHandler) GetBengkelSearchV2Paginate(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	serviceType := c.Query("service")
	query := c.Query("query")
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if _, err := h.userService.GetUserProfile(ctx, userId); err != nil {
		h.HandleError(c, appErrors.ErrUserNotFound.WithDetails(err.Error()))
		return
	}

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	bengkels, count, err := h.bengkelService.SearchBengkelsV2(ctx, serviceType, query, pageInt, limitInt)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success get all bengkel", map[string]any{
		"bengkels": bengkels,
		"count":    count,
	})
	c.JSON(http.StatusOK, resp)
}

// CreateBengkelTestimonial function
func (h *BengkelHandler) CreateBengkelTestimonial(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if _, err := h.userService.GetUserProfile(ctx, userId); err != nil {
		h.HandleError(c, appErrors.ErrUserNotFound.WithDetails(err.Error()))
		return
	}

	bengkelId := c.Param("bengkelId")
	pesananId := c.Param("pesananId")

	var requestDataBengkelTestimonial validator.BengkelTestimoniRequest
	if err := c.ShouldBindJSON(&requestDataBengkelTestimonial); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	if err := h.bengkelService.CreateTestimonial(ctx, userId, bengkelId, pesananId, requestDataBengkelTestimonial.Testimoni, requestDataBengkelTestimonial.Rating); err != nil {
		h.HandleError(c, appErrors.ErrDatabaseError.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success create bengkel testimoni", nil)
	c.JSON(http.StatusOK, resp)
}

// GetAllBengkelTestimonialPaginate function
func (h *BengkelHandler) GetDetailBengkelById(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if _, err := h.userService.GetUserProfile(ctx, userId); err != nil {
		h.HandleError(c, appErrors.ErrUserNotFound.WithDetails(err.Error()))
		return
	}

	bengkelId := c.Param("bengkelId")
	page := c.Query("page")
	limit := c.Query("limit")

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	bengkel, testimonials, count, err := h.bengkelService.GetBengkelDetailWithTestimonials(ctx, bengkelId, pageInt, limitInt)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success get detail data bengkel", map[string]any{
		"bengkel":             bengkel,
		"bengkel_testimonies": testimonials,
		"count":               count,
	})
	c.JSON(http.StatusOK, resp)
}

// GetBengkelDetailForUser function - Comprehensive bengkel details with flexible authentication
func (h *BengkelHandler) GetBengkelDetailForUser(c *gin.Context) {
	bengkelId := c.Param("id")
	ctx := c.Request.Context()

	// Check authentication status
	userType, exists := c.Get("user_type")
	isAuthenticated := exists && userType != "anonymous"

	var userId, mitraId string
	if isAuthenticated {
		if userType == "user" {
			if id, exists := c.Get("user_id"); exists {
				userId = id.(string)
			}
		} else if userType == "mitra" {
			if id, exists := c.Get("mitra_id"); exists {
				mitraId = id.(string)
			}
		}
	}

	forceFresh := c.Query("fresh") == "true"

	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	if !isAuthenticated && limitInt > 5 {
		limitInt = 5
	}

	bengkel, testimonials, testimonialCount, err := h.bengkelService.GetBengkelFullDetail(ctx, bengkelId, forceFresh, pageInt, limitInt)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	if testimonials == nil {
		testimonials = []models.BengkelTestimonial{}
	}

	// Calculate average rating
	var totalRating float64
	for _, t := range testimonials {
		totalRating += float64(t.Rating)
	}
	avgRating := float64(0)
	if len(testimonials) > 0 {
		avgRating = totalRating / float64(len(testimonials))
	}

	availableServices := bengkel.Services
	activeOperationals := bengkel.Operasionals

	bengkelDetail := map[string]interface{}{
		"bengkel_id":    bengkel.ID,
		"bengkel_name":  bengkel.BengkelName,
		"bengkel_phone": bengkel.BengkelPhone,
		"jumlah_montir": bengkel.JumlahMontir,
		"home_service":  bengkel.HomeService,
		"store_service": bengkel.StoreService,
		"is_open":       bengkel.IsOpen,
		"avatar_url":    bengkel.AvatarUrl,
		"created_at":    bengkel.CreatedAt,
		"updated_at":    bengkel.UpdatedAt,
		"services":      availableServices,
		"operasionals":  activeOperationals,
		"photos": func() interface{} {
			if !isAuthenticated && len(bengkel.Photos) > 3 {
				return bengkel.Photos[:3]
			}
			return bengkel.Photos
		}(),
		"addresses": bengkel.Addresses,
		"rating": map[string]interface{}{
			"average_rating":   avgRating,
			"total_reviews":    len(testimonials),
			"total_rating_sum": totalRating,
		},
	}

	userContext := map[string]interface{}{
		"is_authenticated": isAuthenticated,
		"user_type":        userType,
	}
	bengkelDetail["user_context"] = userContext

	if isAuthenticated {
		bengkelDetail["testimonials"] = map[string]interface{}{
			"data":        testimonials,
			"total_count": testimonialCount,
			"page":        pageInt,
			"limit":       limitInt,
			"total_pages": (testimonialCount + limitInt - 1) / limitInt,
		}
		if userId != "" {
			userContext["user_id"] = userId
			userContext["can_book_service"] = true
			userContext["can_leave_review"] = true
		} else if mitraId != "" {
			userContext["is_owner"] = bengkel.MitraID == mitraId
			userContext["management_access"] = bengkel.MitraID == mitraId
		}
		bengkelDetail["analytics"] = map[string]interface{}{
			"total_services":   len(availableServices),
			"total_photos":     len(bengkel.Photos),
			"total_addresses":  len(bengkel.Addresses),
			"operational_days": len(activeOperationals),
		}
	} else {
		limitedTestimonials := testimonials
		if len(limitedTestimonials) > 3 {
			limitedTestimonials = limitedTestimonials[:3]
		}
		bengkelDetail["testimonials"] = map[string]interface{}{
			"data":        limitedTestimonials,
			"total_count": testimonialCount,
			"showing":     len(limitedTestimonials),
			"message":     "Login to see all testimonials",
		}
		bengkelDetail["call_to_action"] = map[string]interface{}{
			"message":      "Login to see full details and book services",
			"login_url":    "/api/v1/users/auth/login",
			"register_url": "/api/v1/users/auth/register",
		}
	}

	message := "success get bengkel detail"
	if !isAuthenticated {
		message = "success get bengkel detail (limited view - login for full access)"
	}

	resp := response.BuildSuccessResponse(message, bengkelDetail)
	c.JSON(http.StatusOK, resp)
}

// CreateBengkelOrderService function
func (h *BengkelHandler) CreateBengkelOrderService(c *gin.Context) {
	// Get authentication context
	userType, exists := c.Get("user_type")
	if !exists {
		h.HandleError(c, appErrors.ErrUnauthorized.WithDetails("user type not found"))
		return
	}

	authenticatedId := c.MustGet("id").(string)

	var userId, mitraId string

	// Handle different authentication scenarios
	if userType == "user" {
		userId = authenticatedId
		if paramMitraId := c.Query("mitraId"); paramMitraId != "" {
			mitraId = paramMitraId
		}
	} else if userType == "mitra" {
		mitraId = authenticatedId
		targetUserId := c.Param("userId")
		if targetUserId == "" {
			h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("userId parameter is required when mitra creates order"))
			return
		}
		userId = targetUserId
	} else {
		h.HandleError(c, appErrors.ErrForbidden.WithDetails("only users and mitras can create orders"))
		return
	}

	// Try new structured format first
	var requestDataBengkelOrderService validator.OrderServiceRequest
	err := c.ShouldBindJSON(&requestDataBengkelOrderService)

	var serviceItems []dto.OrderServiceItem

	if err == nil && len(requestDataBengkelOrderService.Services) > 0 {
		// New structured format
		if mitraId == "" {
			mitraId = requestDataBengkelOrderService.MitraId
		}
		for _, svc := range requestDataBengkelOrderService.Services {
			serviceItems = append(serviceItems, dto.OrderServiceItem{
				Title:  svc.Title,
				Detail: svc.Detail,
				Price:  svc.Price,
			})
		}
	} else {
		// Fallback to legacy format
		var legacyRequest validator.OrderServiceRequestLegacy
		if legacyErr := c.ShouldBindJSON(&legacyRequest); legacyErr != nil {
			h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("invalid request format"))
			return
		}

		if mitraId == "" {
			mitraId = legacyRequest.MitraId
		}

		if mitraId == "" {
			h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("mitra_id must be provided"))
			return
		}

		// Validate arrays have same length
		if len(legacyRequest.Title) != len(legacyRequest.Price) {
			h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("title and price arrays must have the same length"))
			return
		}
		if len(legacyRequest.Detail) > 0 && len(legacyRequest.Detail) != len(legacyRequest.Title) {
			h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("detail array length must match title array length"))
			return
		}

		for i, title := range legacyRequest.Title {
			detail := ""
			if i < len(legacyRequest.Detail) {
				detail = legacyRequest.Detail[i]
			}
			serviceItems = append(serviceItems, dto.OrderServiceItem{
				Title:  title,
				Detail: detail,
				Price:  legacyRequest.Price[i],
			})
		}
	}

	if mitraId == "" {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("mitra_id must be provided"))
		return
	}

	req := dto.CreateOrderWithServicesRequest{
		MitraID:   mitraId,
		Services:  serviceItems,
	}

	result, err := h.orderService.CreateOrderWithServices(c.Request.Context(), userId, mitraId, req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	responseData := map[string]interface{}{
		"pesanan_id":   result.OrderID,
		"user_id":      result.UserID,
		"bengkel_id":   result.BengkelID,
		"bengkel_name": result.BengkelName,
		"total_price":  result.TotalPrice,
		"admin_fee":    result.AdminFee,
		"status":       result.Status,
		"created_by": map[string]interface{}{
			"type": userType,
			"id":   authenticatedId,
			"name": result.CreatedByName,
		},
		"order_context": map[string]interface{}{
			"is_self_order":    userType == "user",
			"is_mitra_created": userType == "mitra",
		},
	}

	message := "success create bengkel pesanan service"
	if userType == "user" {
		message = "success create order for yourself"
	} else {
		message = "success create order for user"
	}

	resp := response.BuildSuccessResponse(message, responseData)
	c.JSON(http.StatusOK, resp)
}

// UpdateAvatarBengkel function
func (h *BengkelHandler) UpdateAvatarBengkel(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	avatar, err := c.FormFile("avatar")
	if err != nil {
		h.HandleError(c, err)
		return
	}

	// Determine protocol from request
	protocol := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		protocol = "https"
	}

	result, err := h.uploadService.UploadFile(avatar, service.AvatarUploadConfig, protocol, c.Request.Host)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("failed to upload file: "+err.Error()))
		return
	}

	if err := h.bengkelService.UpdateBengkelAvatar(c.Request.Context(), mitraId, result.URL); err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success update bengkel avatar", nil)
	c.JSON(http.StatusOK, resp)
}

// GetBenkelOrderServiceById function
func (h *BengkelHandler) GetBengkelOrderServiceById(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	pesananId := c.Param("pesananId")
	if pesananId == "" {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("pesananId params is empty"))
		return
	}

	pesanan, err := h.orderService.GetOrderForUser(ctx, pesananId, userId)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success get pesanan service", pesanan)
	c.JSON(http.StatusOK, resp)
}

// GetBengkelOperationalByIdAndDay function
func (h *BengkelHandler) GetBengkelOperationalByIdAndDay(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if _, err := h.userService.GetUserProfile(ctx, userId); err != nil {
		h.HandleError(c, appErrors.ErrUserNotFound.WithDetails(err.Error()))
		return
	}

	bengkelId := c.Query("bengkelId")
	day := c.Query("day")

	data, err := h.bengkelService.GetBengkelOperationalTimeSlots(ctx, bengkelId, day)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success get bengkel operasional by day", data)
	c.JSON(http.StatusOK, resp)
}

// UpdateBengkelOrderServiceById function
func (h *BengkelHandler) UpdateBengkelOrderServiceById(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	pesananId := c.Param("pesananId")
	if pesananId == "" {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("pesananId params is empty"))
		return
	}

	var requestUpdateBengkelService validator.PesananUpdateRequest
	if err := c.ShouldBindJSON(&requestUpdateBengkelService); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	if err := h.orderService.UpdateOrderDetails(ctx, pesananId, userId, &requestUpdateBengkelService.IsHomeService, requestUpdateBengkelService.HomeServiceSchedule, requestUpdateBengkelService.PaymentMethod); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success update bengkel pesanan service", nil)
	c.JSON(http.StatusOK, resp)
}

// GetDetailUserById function
func (h *BengkelHandler) GetDetailUserById(c *gin.Context) {
	mitraId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	userId := c.Param("userId")

	if _, err := h.bengkelService.GetMitraWithBengkel(ctx, mitraId); err != nil {
		h.HandleError(c, appErrors.ErrMitraNotFound.WithDetails(err.Error()))
		return
	}

	profile, err := h.userService.GetUserProfile(ctx, userId)
	if err != nil {
		h.HandleError(c, appErrors.ErrUserNotFound.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success get user", profile)
	c.JSON(http.StatusOK, resp)
}

// GetBengkelOrderServiceByIdMitra function
func (h *BengkelHandler) GetBengkelOrderServiceByIdMitra(c *gin.Context) {
	mitraId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	pesananId := c.Param("pesananId")
	if pesananId == "" {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("pesananId params is empty"))
		return
	}

	pesanan, err := h.orderService.GetOrderForMitra(ctx, pesananId, mitraId)
	if err != nil {
		h.HandleError(c, appErrors.ErrForbidden.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success get pesanan service", pesanan)
	c.JSON(http.StatusOK, resp)
}

// GetAllBengkelOrderServicePaginate function
func (h *BengkelHandler) GetAllBengkelOrderServicePaginate(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	mitraId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("invalid page"))
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("invalid limit"))
		return
	}

	pesanans, count, err := h.orderService.GetMitraOrdersPaginate(ctx, mitraId, pageInt, limitInt)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success get pesanan service", map[string]any{
		"pesanans": pesanans,
		"count":    count,
	})
	c.JSON(http.StatusOK, resp)
}

// GetAllOrderUserPaginate function
func (h *BengkelHandler) GetAllOrderUserPaginate(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	pesanans, count, err := h.orderService.GetUserOrdersPaginate(ctx, userId, pageInt, limitInt)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success get all pesanan", map[string]any{
		"pesanans": pesanans,
		"count":    count,
	})
	c.JSON(http.StatusOK, resp)
}

// GetNearestBengkelPaginate function
func (h *BengkelHandler) GetNearestBengkelPaginate(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	latitude := c.Query("latitude")
	longitude := c.Query("longitude")
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if _, err := h.userService.GetUserProfile(ctx, userId); err != nil {
		h.HandleError(c, appErrors.ErrUserNotFound.WithDetails(err.Error()))
		return
	}

	floatLatitude, err := strconv.ParseFloat(latitude, 64)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("invalid latitude"))
		return
	}

	floatLongitude, err := strconv.ParseFloat(longitude, 64)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("invalid longitude"))
		return
	}

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	req := dto.NearestBengkelRequest{
		Latitude:  floatLatitude,
		Longitude: floatLongitude,
		PaginationRequest: dto.PaginationRequest{
			Page:  pageInt,
			Limit: limitInt,
		},
	}

	result, err := h.bengkelService.GetNearestBengkels(ctx, req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success get nearest bengkel", result)
	c.JSON(http.StatusOK, resp)
}

// ConfirmOrderService function
func (h *BengkelHandler) UpdateStatusOrderService(c *gin.Context) {
	mitraId := c.MustGet("id").(string)
	pesananId := c.Param("pesananId")

	if pesananId == "" {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("pesananId params is empty"))
		return
	}

	var pesananStatusRequest validator.PesananStatusUpdateRequest
	if err := c.ShouldBindJSON(&pesananStatusRequest); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	updatedOrder, err := h.orderService.UpdateOrderStatusWithValidation(c.Request.Context(), mitraId, pesananId, pesananStatusRequest.Status, pesananStatusRequest.Reason)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success update status pesanan", toOrderResponse(updatedOrder))
	c.JSON(http.StatusOK, resp)
}

// GetAllBengkelPublic - Public endpoint for bengkel list (no authentication required)
func (h *BengkelHandler) GetAllBengkelPublic(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	ctx := c.Request.Context()

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		limitInt = 10
	}

	if limitInt > 50 {
		limitInt = 50
	}

	bengkels, count, err := h.bengkelService.GetAllBengkelsPaginate(ctx, pageInt, limitInt)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	totalPages := (count + limitInt - 1) / limitInt
	hasNext := pageInt < totalPages
	hasPrev := pageInt > 1

	resp := response.BuildSuccessResponse("success get bengkel list", map[string]any{
		"items": bengkels,
		"pagination": map[string]any{
			"page":        pageInt,
			"limit":       limitInt,
			"total":       count,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
	c.JSON(http.StatusOK, resp)
}

// GetBengkelSearchPublic - Public endpoint for bengkel search (no authentication required)
func (h *BengkelHandler) GetBengkelSearchPublic(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	service := c.Query("service")
	query := c.Query("query")
	city := c.Query("city")
	province := c.Query("province")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		limitInt = 10
	}

	if limitInt > 50 {
		limitInt = 50
	}

	searchCriteria := map[string]interface{}{
		"page":  pageInt,
		"limit": limitInt,
	}
	if query != "" {
		searchCriteria["query"] = query
	}
	if service != "" {
		searchCriteria["service"] = service
	}
	if city != "" {
		searchCriteria["city"] = city
	}
	if province != "" {
		searchCriteria["province"] = province
	}

	bengkels, count, err := h.bengkelService.SearchBengkelsPublic(c.Request.Context(), searchCriteria)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	totalPages := (count + limitInt - 1) / limitInt
	hasNext := pageInt < totalPages
	hasPrev := pageInt > 1

	resp := response.BuildSuccessResponse("success search bengkels", map[string]any{
		"items":           bengkels,
		"search_criteria": searchCriteria,
		"pagination": map[string]any{
			"page":        pageInt,
			"limit":       limitInt,
			"total":       count,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
	c.JSON(http.StatusOK, resp)
}

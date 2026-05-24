package handlers

import (
	"net/http"
	"strconv"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/container"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/pkg/errors"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/service"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/validator"
	"github.com/Bengkelin/bengkelin-service/pkg/response"
	"github.com/Bengkelin/bengkelin-service/pkg/validation"
	"github.com/gin-gonic/gin"
)

var (
	userHandler *UserHandler
)

type UserHandler struct {
	BaseHandler
	userService   service.UserServiceInterface
	uploadService *service.FileUploadService
}

func GetUserHandler() UserHandlerInterface {
	if userHandler == nil {
		c := container.GetContainer()
		userHandler = &UserHandler{
			userService:   c.UserService,
			uploadService: service.NewFileUploadService(),
		}
	}
	return userHandler
}

type UserHandlerInterface interface {
	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	CreateAddress(c *gin.Context)
	UpdateAddress(c *gin.Context)
	UpdateOrCreateAddress(c *gin.Context)
	UpdateAvatarUser(c *gin.Context)
	GetDetailAddressUser(c *gin.Context)
	DeleteAddressUser(c *gin.Context)
	CreateVehicle(c *gin.Context)
	GetAllVehiclesUser(c *gin.Context)
	GetDetailVehicleUser(c *gin.Context)
	DeleteVehicleUser(c *gin.Context)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	profile, err := h.userService.GetUserProfile(ctx, userId)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success get profile", profile)
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	var userUpdateRequest validator.UserUpdateRequest
	if err := c.ShouldBind(&userUpdateRequest); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	req := dto.UpdateUserRequest{
		FirstName:   userUpdateRequest.FirstName,
		LastName:    userUpdateRequest.LastName,
		PhoneNumber: userUpdateRequest.PhoneNumber,
	}

	updatedUser, err := h.userService.UpdateUserProfile(ctx, userId, req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success update profile", updatedUser)
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) CreateAddress(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	var addressCreateRequest validator.UserAddressCreateRequest
	if err := c.ShouldBind(&addressCreateRequest); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	req := dto.AddressRequest{
		Latitude:     addressCreateRequest.Latitude,
		Longitude:    addressCreateRequest.Longitude,
		AddressLabel: addressCreateRequest.AddressLabel,
		FullAddress:  addressCreateRequest.FullAddress,
		Note:         addressCreateRequest.Note,
		IsPrimary:    addressCreateRequest.IsPrimary,
	}

	createdAddress, err := h.userService.AddUserAddress(ctx, userId, req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success create address", createdAddress)
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateAddress(c *gin.Context) {
	userId := c.MustGet("id").(string)
	addressId := c.Param("addressId")
	ctx := c.Request.Context()

	if addressId == "" {
		h.HandleError(c, appErrors.ErrMissingField.WithDetails("address id is required"))
		return
	}

	addressIdUint, err := strconv.ParseUint(addressId, 10, 64)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("address id must be a number"))
		return
	}

	var addressUpdateRequest validator.UserAddressUpdateRequest
	if err := c.ShouldBind(&addressUpdateRequest); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	req := dto.AddressRequest{
		Latitude:     addressUpdateRequest.Latitude,
		Longitude:    addressUpdateRequest.Longitude,
		AddressLabel: addressUpdateRequest.AddressLabel,
		FullAddress:  addressUpdateRequest.FullAddress,
		Note:         addressUpdateRequest.Note,
		IsPrimary:    addressUpdateRequest.IsPrimary,
	}

	updatedAddress, err := h.userService.UpdateAddress(ctx, userId, uint(addressIdUint), req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success update address", updatedAddress)
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateOrCreateAddress(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	var addressUpdateRequest validator.UserAddressUpdateRequest
	if err := c.ShouldBind(&addressUpdateRequest); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	req := dto.AddressRequest{
		Latitude:     addressUpdateRequest.Latitude,
		Longitude:    addressUpdateRequest.Longitude,
		AddressLabel: addressUpdateRequest.AddressLabel,
		FullAddress:  addressUpdateRequest.FullAddress,
		Note:         addressUpdateRequest.Note,
		IsPrimary:    addressUpdateRequest.IsPrimary,
	}

	result, err := h.userService.CreateOrUpdateAddress(ctx, userId, req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success update address", result)
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateAvatarUser(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	avatar, err := c.FormFile("avatar")
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("failed to get file: "+err.Error()))
		return
	}

	// Determine protocol from request
	protocol := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		protocol = "https"
	}

	result, err := h.uploadService.UploadFile(avatar, service.AvatarUploadConfig, protocol, c.Request.Host)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	if err := h.userService.UpdateUserAvatar(ctx, userId, result.URL); err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success update avatar", nil)
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) DeleteAddressUser(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	addressId := c.Param("addressId")

	if addressId == "" {
		h.HandleError(c, appErrors.ErrMissingField.WithDetails("address id is required"))
		return
	}

	if err := h.userService.DeleteUserAddress(ctx, userId, addressId); err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success delete address", nil)
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetDetailAddressUser(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	addressId := c.Param("addressId")

	address, err := h.userService.GetUserAddress(ctx, userId, addressId)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success get address", address)
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) DeleteVehicleUser(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	vehicleId := c.Param("vehicleId")
	if vehicleId == "" {
		h.HandleError(c, appErrors.ErrMissingField.WithDetails("vehicle id is required"))
		return
	}

	if err := h.userService.DeleteUserVehicle(ctx, userId, vehicleId); err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success delete vehicle", nil)
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetDetailVehicleUser(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()
	vehicleId := c.Param("vehicleId")

	vehicle, err := h.userService.GetUserVehicle(ctx, userId, vehicleId)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success get vehicle", vehicle)
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetAllVehiclesUser(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	vehicles, err := h.userService.GetAllUserVehicles(ctx, userId)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success get all vehicles", vehicles)
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) CreateVehicle(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	var vehicleRequest validator.VehicleUserRequest
	if err := c.ShouldBind(&vehicleRequest); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	if validationErrors := validation.ValidateStruct(vehicleRequest); len(validationErrors) > 0 {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(validationErrors.Error()))
		return
	}

	req := dto.VehicleRequest{
		VehicleType:   vehicleRequest.VehicleType,
		VehicleNumber: vehicleRequest.VehicleNumber,
		VehicleColor:  vehicleRequest.VehicleColor,
	}

	createdVehicle, err := h.userService.AddUserVehicle(ctx, userId, req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success create vehicle", createdVehicle)
	c.JSON(http.StatusCreated, resp)
}

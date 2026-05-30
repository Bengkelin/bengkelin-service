package handlers

import (
	"net/http"

	"github.com/Bengkelin/bengkelin-service/internal/container"
	"github.com/Bengkelin/bengkelin-service/internal/dto"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/errors"
	"github.com/Bengkelin/bengkelin-service/internal/models"
	"github.com/Bengkelin/bengkelin-service/internal/service"
	"github.com/Bengkelin/bengkelin-service/internal/validator"
	"github.com/Bengkelin/bengkelin-service/internal/response"
	"github.com/Bengkelin/bengkelin-service/internal/validation"
	"github.com/gin-gonic/gin"
)

var (
	mitraHandler *MitraHandler
)

type MitraHandler struct {
	BaseHandler
	mitraService service.MitraServiceInterface
}

func GetMitraHandler() MitraHandlerInterface {
	if mitraHandler == nil {
		c := container.GetContainer()
		mitraHandler = &MitraHandler{
			mitraService: c.MitraService,
		}
	}
	return mitraHandler
}

type MitraHandlerInterface interface {
	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	CreateBank(c *gin.Context)
	UpdateBank(c *gin.Context)
}

func toMitraProfileResponse(m *models.Mitra) *dto.MitraProfileResponse {
	return toMitraWithBengkelResponse(m)
}

func (h *MitraHandler) GetProfile(c *gin.Context) {
	mitraID := c.MustGet("id").(string)
	ctx := c.Request.Context()

	mitra, err := h.mitraService.GetMitraProfile(ctx, mitraID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success get mitra profile", toMitraProfileResponse(mitra))
	c.JSON(http.StatusOK, resp)
}

func (h *MitraHandler) UpdateProfile(c *gin.Context) {
	mitraID := c.MustGet("id").(string)
	ctx := c.Request.Context()

	var mitraUpdateRequest validator.MitraUpdateProfileRequest
	if err := c.ShouldBind(&mitraUpdateRequest); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	// Validate the request using custom validation
	if validationErrors := validation.ValidateStruct(mitraUpdateRequest); validationErrors != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(validationErrors.Error()))
		return
	}

	updateReq := dto.UpdateMitraRequest{
		FirstName:   mitraUpdateRequest.FirstName,
		LastName:    mitraUpdateRequest.LastName,
		PhoneNumber: mitraUpdateRequest.PhoneNumber,
	}

	updatedMitra, err := h.mitraService.UpdateMitraProfile(ctx, mitraID, updateReq)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success update mitra profile", toMitraProfileResponse(updatedMitra))
	c.JSON(http.StatusOK, resp)
}

func (h *MitraHandler) CreateBank(c *gin.Context) {
	mitraID := c.MustGet("id").(string)
	ctx := c.Request.Context()

	var bankRequest validator.MitraBankUpdateRequest
	if err := c.ShouldBind(&bankRequest); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	if validationErrors := validation.ValidateStruct(bankRequest); validationErrors != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(validationErrors.Error()))
		return
	}

	bankReq := dto.MitraBankRequest{
		BankName:   bankRequest.BankName,
		BankNumber: bankRequest.BankNumber,
	}

	result, err := h.mitraService.CreateMitraBank(ctx, mitraID, bankReq)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success create bank account", result)
	c.JSON(http.StatusCreated, resp)
}

func (h *MitraHandler) UpdateBank(c *gin.Context) {
	mitraID := c.MustGet("id").(string)
	ctx := c.Request.Context()

	var bankRequest validator.MitraBankUpdateRequest
	if err := c.ShouldBind(&bankRequest); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	if validationErrors := validation.ValidateStruct(bankRequest); validationErrors != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(validationErrors.Error()))
		return
	}

	bankReq := dto.MitraBankRequest{
		BankName:   bankRequest.BankName,
		BankNumber: bankRequest.BankNumber,
	}

	result, err := h.mitraService.UpdateMitraBank(ctx, mitraID, bankReq)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	resp := response.BuildSuccessResponse("success update bank account", result)
	c.JSON(http.StatusOK, resp)
}

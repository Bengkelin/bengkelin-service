package handlers

import (
	"net/http"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/container"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/pkg/errors"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/service"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/validator"
	"github.com/Bengkelin/bengkelin-service/pkg/response"
	"github.com/gin-gonic/gin"
)

var (
	adminFeeHandler *AdminFeeHandler
)

type AdminFeeHandler struct {
	BaseHandler
	adminFeeService service.AdminFeeServiceInterface
}

type AdminFeeHandlerInterface interface {
	CreateAdminFee(c *gin.Context)
	UpdateAdminFeeById(c *gin.Context)
	GetAdminFeeById(c *gin.Context)
}

// NewAdminFeeHandler is a constructor to create AdminFeeHandler instance
func GetAdminFeeHandler() AdminFeeHandlerInterface {
	if adminFeeHandler == nil {
		c := container.GetContainer()
		adminFeeHandler = &AdminFeeHandler{
			adminFeeService: c.AdminFeeService,
		}
	}
	return adminFeeHandler
}

// CreateAdminFee implements AdminFeeHandlerInterface.
func (h *AdminFeeHandler) CreateAdminFee(c *gin.Context) {
	var createAdminFeeRequest validator.AdminFeeRequest
	if err := c.ShouldBindJSON(&createAdminFeeRequest); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	result, err := h.adminFeeService.CreateAdminFee(c.Request.Context(), createAdminFeeRequest.AdminFee)
	if err != nil {
		h.HandleError(c, appErrors.ErrDatabaseError.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("admin fee created successfully", result)
	c.JSON(http.StatusCreated, resp)
}

// UpdateAdminFeeById implements AdminFeeHandlerInterface.
func (h *AdminFeeHandler) UpdateAdminFeeById(c *gin.Context) {
}

// GetAdminFeeById implements AdminFeeHandlerInterface.
func (h *AdminFeeHandler) GetAdminFeeById(c *gin.Context) {
}

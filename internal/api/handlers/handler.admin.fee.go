package handlers

import (
	"net/http"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/validator"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	"github.com/Bengkelin/bengkelin-service/pkg/response"
	"github.com/gin-gonic/gin"
)

var (
	adminFeeHandler *AdminFeeHandler
)

type AdminFeeHandler struct{}

type AdminFeeHandlerInterface interface {
	CreateAdminFee(c *gin.Context)
	UpdateAdminFeeById(c *gin.Context)
	GetAdminFeeById(c *gin.Context)
}

// NewAdminFeeHandler is a constructor to create AdminFeeHandler instance
func GetAdminFeeHandler() AdminFeeHandlerInterface {
	if adminFeeHandler == nil {
		adminFeeHandler = &AdminFeeHandler{}
	}
	return adminFeeHandler
}

// CreateAdminFee implements AdminFeeHandlerInterface.
func (*AdminFeeHandler) CreateAdminFee(c *gin.Context) {
	secret := c.GetHeader("secret")

	serverConfiguration := config.GetConfig().Server

	if secret != serverConfiguration.ApiSecret {
		response := response.BuildFailedResponse("unauthorized", "secret is not valid")
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	var createAdminFeeRequest validator.AdminFeeRequest

	err := c.ShouldBindJSON(&createAdminFeeRequest)
	if err != nil {
		response := response.BuildFailedResponse("bad request when create new admin fee", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	adminFee := &models.AdminFee{
		ID:       helpers.GenerateUUID(),
		AdminFee: createAdminFeeRequest.AdminFee,
	}

	adminFeeRepository := repository.GetAdminFeeRepository()

	_, err = adminFeeRepository.CreateAdminFee(*adminFee)
	if err != nil {
		response := response.BuildFailedResponse("failed to create admin fee", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	data, err := adminFeeRepository.GetAdminFeeById(adminFee.ID)
	if err != nil {
		response := response.BuildFailedResponse("failed to get admin fee by id", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := response.BuildSuccessResponse("admin fee created successfully", data)
	c.JSON(http.StatusCreated, response)
}

// UpdateAdminFeeById implements AdminFeeHandlerInterface.
func (*AdminFeeHandler) UpdateAdminFeeById(c *gin.Context) {
}

// GetAdminFeeById implements AdminFeeHandlerInterface.
func (*AdminFeeHandler) GetAdminFeeById(c *gin.Context) {
}

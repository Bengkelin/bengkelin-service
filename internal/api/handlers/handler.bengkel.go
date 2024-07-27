package handlers

import (
	"net/http"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/validator"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	"github.com/Bengkelin/bengkelin-service/pkg/response"
	"github.com/gin-gonic/gin"
)

var (
	bengkelHandler *BengkelHandler
)

// BengkelHandler struct
type BengkelHandler struct{}

// NewBengkelHandler function
func GetBengkelHandler() *BengkelHandler {
	if bengkelHandler == nil {
		bengkelHandler = &BengkelHandler{}
	}
	return bengkelHandler
}

// BengkelHandlerInterface interface
type BengkelHandlerInterface interface {
	CreateBengkel(c *gin.Context)
	UpdateBengkel(c *gin.Context)
	GetBengkel(c *gin.Context)
}

// CreateBengkel function
func (handler *BengkelHandler) CreateBengkel(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	mitraRepo := repository.GetMitraRepository()

	mitra, err := mitraRepo.FindMitraByID(mitraId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get mitra", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	var requestDataBengkel validator.RegisterBengkelRequest

	err = c.ShouldBindJSON(&requestDataBengkel)
	if err != nil {
		response := response.BuildFailedResponse("failed to bind json", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	bengkelRepo := repository.GetBengkelRepository()

	bengkelModel := &models.Bengkel{
		ID:           helpers.GenerateUUID(),
		MitraID:      mitra.ID,
		BengkelName:  requestDataBengkel.BengkelName,
		BengkelPhone: requestDataBengkel.BengkelPhone,
		JumlahMontir: requestDataBengkel.JumlahMontir,
	}

	_, err = bengkelRepo.CreateBengkel(*bengkelModel)
	if err != nil {
		response := response.BuildFailedResponse("failed to create bengkel", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	bengkelOperasionalRepo := repository.GetBengkelOperasionalRepository()

	bengkelOperasionalModel := []models.BengkelOperasional{}
	for i, v := range requestDataBengkel.Hari {
		bengkelOperasionalModel = append(bengkelOperasionalModel, models.BengkelOperasional{
			BengkelID: bengkelModel.ID,
			Hari:      v,
			JamBuka:   requestDataBengkel.JamBuka[i],
		})
	}

	for _, v := range bengkelOperasionalModel {
		_, err = bengkelOperasionalRepo.CreateBengkelOperasional(v)
		if err != nil {
			response := response.BuildFailedResponse("failed to create bengkel operasional", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}

	response := response.BuildSuccessResponse("success create bengkel", nil)
	c.JSON(http.StatusOK, response)
}

// UpdateBengkel function
func (handler *BengkelHandler) UpdateBengkel(c *gin.Context) {

}

// GetBengkel function
func (handler *BengkelHandler) GetBengkel(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	mitraRepo := repository.GetMitraRepository()

	mitra, err := mitraRepo.FindMitraByID(mitraId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get mitra", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success get mitra", mitra)
	c.JSON(http.StatusOK, response)
}

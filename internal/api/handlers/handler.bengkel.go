package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
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
	CreateBengkelAddress(c *gin.Context)
	CreateBengkelService(c *gin.Context)
	CreateBengkelPhoto(c *gin.Context)
	UpdateBengkelStatusOpsiService(c *gin.Context)
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

	var requestDataBengkel validator.BengkelRegisterRequest

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

// CreateBengkelAddress function
func (handler *BengkelHandler) CreateBengkelAddress(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	mitraRepo := repository.GetMitraRepository()

	mitra, err := mitraRepo.GetMitraByID(mitraId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get mitra", err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	var requestDataBengkelAddress validator.BengkelAddressRequest

	err = c.ShouldBindJSON(&requestDataBengkelAddress)

	if err != nil {
		response := response.BuildFailedResponse("failed to bind json", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	mitraAddressRepo := repository.GetBengkelAddressRepository()

	mitraAddressModel := &models.BengkelAddress{
		BengkelID:    mitra.Bengkel[0].ID,
		Latitude:     requestDataBengkelAddress.Latitude,
		Longitude:    requestDataBengkelAddress.Longitude,
		AddressLabel: requestDataBengkelAddress.AddressLabel,
		FullAddress:  requestDataBengkelAddress.FullAddress,
		Note:         requestDataBengkelAddress.Note,
	}

	_, err = mitraAddressRepo.CreateBengkelAddress(*mitraAddressModel)
	if err != nil {
		response := response.BuildFailedResponse("failed to attach bengkel address", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success attach bengkel address", nil)
	c.JSON(http.StatusOK, response)
}

// CreateBengkelServive function
func (handler *BengkelHandler) CreateBengkelService(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	mitraRepo := repository.GetMitraRepository()

	mitra, err := mitraRepo.GetMitraByID(mitraId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get mitra", err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	var requestDataBengkelService validator.BengkelServiceRequest

	err = c.ShouldBindJSON(&requestDataBengkelService)

	if err != nil {
		response := response.BuildFailedResponse("failed to bind json", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	mitraServiceRepo := repository.GetBengkelServiceRepository()

	mitraServiceModel := []models.BengkelService{}

	for _, v := range requestDataBengkelService.NamaService {
		mitraServiceModel = append(mitraServiceModel, models.BengkelService{
			BengkelID:   mitra.Bengkel[0].ID,
			NamaService: v,
		})
	}

	for _, v := range mitraServiceModel {
		_, err = mitraServiceRepo.CreateBengkelService(v)
		if err != nil {
			response := response.BuildFailedResponse("failed to attach bengkel service", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}

	response := response.BuildSuccessResponse("success attach bengkel service", nil)
	c.JSON(http.StatusOK, response)
}

// CreateBengkelPhoto function
func (handler *BengkelHandler) CreateBengkelPhoto(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	mitraRepo := repository.GetMitraRepository()

	bengkelPhotoRepo := repository.GetBengkelPhotoRepository()

	mitra, err := mitraRepo.GetMitraByID(mitraId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get mitra", err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	form, _ := c.MultipartForm()
	photos := form.File["photos"]

	urlPictures := []string{}

	serverConfiguration := config.GetConfig().Server

	for _, file := range photos {
		fileExt := filepath.Ext(file.Filename)

		originalFileName := strings.TrimSuffix(filepath.Base(file.Filename), filepath.Ext(file.Filename))
		now := time.Now()
		fileName := strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + "-" + fmt.Sprintf("%v", now.Unix()) + fileExt
		res, err := os.Stat("public/bengkels")
		if os.IsNotExist(err) || !res.IsDir() {
			os.Mkdir("public/bengkels", os.ModePerm)
		}

		var directoryName string = "./public/bengkels/" + fileName

		fileName = strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + "-" + fmt.Sprintf("%v", now.Unix()) + fileExt

		if err := c.SaveUploadedFile(file, directoryName); err != nil {
			response := response.BuildFailedResponse("failed to upload file", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}

		var reqHost string = "true"

		if serverConfiguration.DevMode == "false" {
			reqHost = serverConfiguration.Host
		} else {
			reqHost = serverConfiguration.Host + ":" + serverConfiguration.Port
		}

		urlPicture := fmt.Sprintf("http://%s/api/v1/static/%s", reqHost, fileName)
		urlPictures = append(urlPictures, urlPicture)
	}

	for _, urlLink := range urlPictures {
		vehiclePhotoModel := &models.BengkelPhoto{
			BengkelID: mitra.Bengkel[0].ID,
			PhotoURL:  urlLink,
		}
		_, err = bengkelPhotoRepo.CreateBengkelPhoto(*vehiclePhotoModel)
		if err != nil {
			response := response.BuildFailedResponse("failed to attach new vehicle photo", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}

	response := response.BuildSuccessResponse("success attach new bengkel photo", nil)
	c.JSON(http.StatusOK, response)
}

// UpdateBengkel function
func (handler *BengkelHandler) UpdateBengkel(c *gin.Context) {

}

// UpdateBengkelStatusOpsiService function
func (handler *BengkelHandler) UpdateBengkelStatusOpsiService(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	mitraRepo := repository.GetMitraRepository()

	mitra, err := mitraRepo.FindMitraByID(mitraId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get mitra", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	var requestOptionStatusBengkelService validator.BengkelServiceOptionRequest

	err = c.ShouldBindJSON(&requestOptionStatusBengkelService)
	if err != nil {
		response := response.BuildFailedResponse("failed to bind json", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	bengkelRepo := repository.GetBengkelRepository()

	bengkelModel := &models.Bengkel{
		HomeService:  &requestOptionStatusBengkelService.HomeService,
		StoreService: &requestOptionStatusBengkelService.StoreService,
	}

	err = bengkelRepo.UpdateBengkelById(mitra.Bengkel[0].ID, bengkelModel)
	if err != nil {
		response := response.BuildFailedResponse("failed to update bengkel status option service", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success update bengkel status option service", nil)
	c.JSON(http.StatusOK, response)
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

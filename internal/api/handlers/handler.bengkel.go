package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	GetAllBengkel(c *gin.Context)
	GetAllBengkelPaginate(c *gin.Context)
	GetBengkelSearchPaginate(c *gin.Context)
	CreateBengkelTestimoni(c *gin.Context)
	GetDetailBengkelById(c *gin.Context)
	CreateBengkelPesananService(c *gin.Context)
	GetAllBengkelPesananServicePaginate(c *gin.Context)
	GetBengkelPesananServiceById(c *gin.Context)
	UpdateAvatarBengkel(c *gin.Context)
	GetBengkelOperasionalByIdAndDay(c *gin.Context)
	UpdateBengkelPesananServiceById(c *gin.Context)
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
		c.AbortWithStatusJSON(http.StatusNotFound, response)
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

	// check if photos is empty
	if len(photos) == 0 {
		response := response.BuildFailedResponse("failed to upload file", "photos is empty")
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

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

		urlPicture := fmt.Sprintf("https://%s/api/v1/static/bengkel/%s", reqHost, fileName)
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
		IsOpen:       &requestOptionStatusBengkelService.IsOpen,
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

// GetAllBengkel function
func (handler *BengkelHandler) GetAllBengkel(c *gin.Context) {
	userId := c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()

	_, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	bengkelRepo := repository.GetBengkelRepository()

	bengkels, err := bengkelRepo.GetAllBengkel()
	if err != nil {
		response := response.BuildFailedResponse("failed to get all bengkel", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success get all bengkel", bengkels)
	c.JSON(http.StatusOK, response)
}

// GetAllBengkelPaginate function
func (handler *BengkelHandler) GetAllBengkelPaginate(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	userId := c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()

	_, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	bengkelRepo := repository.GetBengkelRepository()

	bengkels, count, err := bengkelRepo.GetAllBengkelPaginate(pageInt, limitInt)
	if err != nil {
		response := response.BuildFailedResponse("failed to get all bengkel", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success get all bengkel", map[string]any{
		"bengkels": bengkels,
		"count":    count,
	})
	c.JSON(http.StatusOK, response)
}

// GetBengkelSearchV2Paginate function
func (handler *BengkelHandler) GetBengkelSearchV2Paginate(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	service := c.Query("service")
	query := c.Query("query")
	userId := c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()

	_, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	bengkelRepo := repository.GetBengkelRepository()

	bengkels, count, err := bengkelRepo.GetBengkelSearchV2(service, query, pageInt, limitInt)
	if err != nil {
		response := response.BuildFailedResponse("failed to get all bengkel", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success get all bengkel", map[string]any{
		"bengkels": bengkels,
		"count":    count,
	})
	c.JSON(http.StatusOK, response)
}

// CreateBengkelTestimoni function
func (handler *BengkelHandler) CreateBengkelTestimoni(c *gin.Context) {
	userId := c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()

	_, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	bengkelId := c.Param("bengkelId")
	pesananId := c.Param("pesananId")

	bengkelRepo := repository.GetBengkelRepository()

	bengkel, err := bengkelRepo.GetBengkelById(bengkelId)
	if err != nil {
		response := response.BuildFailedResponse("bengkel not found", err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	pesananRepo := repository.GetPesananRepository()

	pesananData, err := pesananRepo.GetPesananById(pesananId)
	if err != nil {
		response := response.BuildFailedResponse("pesanan not found", err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	var requestDataBengkelTestimoni validator.BengkelTestimoniRequest

	err = c.ShouldBindJSON(&requestDataBengkelTestimoni)
	if err != nil {
		response := response.BuildFailedResponse("failed to bind json", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	bengkelTestimoniRepo := repository.GetBengkelTestimoniRepository()

	bengkelTestimoniModel := &models.BengkelTestimoni{
		BengkelID: bengkel.ID,
		UserID:    userId,
		PesananID: pesananData.ID,
		Testimoni: requestDataBengkelTestimoni.Testimoni,
		Rating:    requestDataBengkelTestimoni.Rating,
	}

	_, err = bengkelTestimoniRepo.CreateBengkelTestimoni(*bengkelTestimoniModel)
	if err != nil {
		response := response.BuildFailedResponse("failed to create bengkel testimoni", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success create bengkel testimoni", nil)
	c.JSON(http.StatusOK, response)
}

// GetAllBengkelTestimoniPaginate function
func (handler *BengkelHandler) GetDetailBengkelById(c *gin.Context) {
	userId := c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()

	_, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	bengkelId := c.Param("bengkelId")
	page := c.Query("page")
	limit := c.Query("limit")

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	bengkelTestimoniRepo := repository.GetBengkelRepository()

	bengkel, bengkelTestimonies, count, err := bengkelTestimoniRepo.FindBengkelById(bengkelId, pageInt, limitInt)
	if err != nil {
		response := response.BuildFailedResponse("failed to get detail data bengkel", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success get detail data bengkel", map[string]any{
		"bengkel":             bengkel,
		"bengkel_testimonies": bengkelTestimonies,
		"count":               count,
	})
	c.JSON(http.StatusOK, response)
}

// CreateBengkelPesananService function
func (handler *BengkelHandler) CreateBengkelPesananService(c *gin.Context) {
	mitraId := c.MustGet("id").(string)
	userId := c.Param("userId")

	userRepo := repository.GetUserRepository()

	user, err := userRepo.FindUserByID(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	mitraRepo := repository.GetMitraRepository()

	mitra, err := mitraRepo.FindMitraByID(mitraId)
	if err != nil {
		response := response.BuildFailedResponse("mitras not found", err.Error())
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	var requestDataBengkelPesananService validator.PesananServiceRequest

	err = c.ShouldBindJSON(&requestDataBengkelPesananService)
	if err != nil {
		response := response.BuildFailedResponse("failed to bind json", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	adminFeeRepo := repository.GetAdminFeeRepository()
	adminFeeData, err := adminFeeRepo.GetOneAdminFeeLatest()
	if err != nil {
		response := response.BuildFailedResponse("failed to get admin fee", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	bengkelPesananRepo := repository.GetPesananRepository()

	isHomeService := false

	pesananModel := &models.Pesanan{
		ID:            helpers.GenerateUUID(),
		UserID:        userId,
		BengkelID:     mitra.Bengkel[0].ID,
		Status:        0,
		VehicleID:     user.Vehicles[0].ID,
		IsHomeService: &isHomeService,
		AdminFee:      adminFeeData.AdminFee,
	}

	_, err = bengkelPesananRepo.CreatePesanan(*pesananModel)
	if err != nil {
		response := response.BuildFailedResponse("failed to create pesanan", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	bengkelPesananServiceModel := []models.PesananService{}

	for i, v := range requestDataBengkelPesananService.ServiceName {
		bengkelPesananServiceModel = append(bengkelPesananServiceModel, models.PesananService{
			PesananID:   pesananModel.ID,
			ServiceName: v,
			Note:        requestDataBengkelPesananService.Note[i],
			Price:       requestDataBengkelPesananService.Price[i],
		})
		requestDataBengkelPesananService.TotalPrice += requestDataBengkelPesananService.Price[i]
	}

	bengkelPesananServiceRepo := repository.GetPesananServiceRepository()

	for _, v := range bengkelPesananServiceModel {
		_, err = bengkelPesananServiceRepo.CreatePesananService(v)
		if err != nil {
			response := response.BuildFailedResponse("failed to create pesanan service", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}

	pesananModel.TotalPrice = requestDataBengkelPesananService.TotalPrice

	err = bengkelPesananRepo.UpdatePesananById(pesananModel.ID, pesananModel)
	if err != nil {
		response := response.BuildFailedResponse("failed to update pesanan", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success create bengkel pesanan service", nil)
	c.JSON(http.StatusOK, response)
}

// UpdateAvatarBengkel function
func (handler *BengkelHandler) UpdateAvatarBengkel(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	mitraRepo := repository.GetMitraRepository()

	mitra, err := mitraRepo.GetMitraByID(mitraId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get mitra", err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	avatar, err := c.FormFile("avatar")
	if err != nil {
		response := response.BuildFailedResponse("failed to get avatar url", err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	serverConfiguration := config.GetConfig().Server

	fileExt := filepath.Ext(avatar.Filename)

	originalFileName := strings.TrimSuffix(filepath.Base(avatar.Filename), filepath.Ext(avatar.Filename))
	now := time.Now()
	fileName := strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + "-" + fmt.Sprintf("%v", now.Unix()) + fileExt
	res, err := os.Stat("public/avatars")
	if os.IsNotExist(err) || !res.IsDir() {
		os.Mkdir("public/avatars", os.ModePerm)
	}

	var directoryName string = "./public/avatars/" + fileName

	fileName = strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + "-" + fmt.Sprintf("%v", now.Unix()) + fileExt

	if err := c.SaveUploadedFile(avatar, directoryName); err != nil {
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

	urlPicture := fmt.Sprintf("https://%s/api/v1/static/avatar/%s", reqHost, fileName)

	bengkelModel := &models.Bengkel{
		AvatarUrl: urlPicture,
	}

	bengkelRepo := repository.GetBengkelRepository()

	err = bengkelRepo.UpdateBengkelById(mitra.Bengkel[0].ID, bengkelModel)
	if err != nil {
		response := response.BuildFailedResponse("failed to update bengkel avatar", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success update bengkel avatar", nil)
	c.JSON(http.StatusOK, response)
}

// GetBenkelPesananServiceById function
func (handler *BengkelHandler) GetBengkelPesananServiceById(c *gin.Context) {
	userId := c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()

	_, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	pesananId := c.Param("pesananId")

	if pesananId == "" {
		response := response.BuildFailedResponse("failed to get pesanan", "pesananId params is empty")
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	bengkelPesananRepo := repository.GetPesananRepository()

	pesanan, err := bengkelPesananRepo.GetDetailPesananById(pesananId, userId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get pesanan", err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	response := response.BuildSuccessResponse("success get pesanan service", pesanan)
	c.JSON(http.StatusOK, response)
}

// GetBengkelOperasionalByIdAndDay function
func (handler *BengkelHandler) GetBengkelOperasionalByIdAndDay(c *gin.Context) {
	userId := c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()

	_, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	bengkelId := c.Query("bengkelId")
	day := c.Query("day")

	bengkelOperasionalRepo := repository.GetBengkelOperasionalRepository()

	bengkelOperasional, err := bengkelOperasionalRepo.GetBengkelOperasionalByIdAndDay(bengkelId, day)

	var dataTimePerHours []string
	var timePerHoursOpen []string
	var timePerHoursClose []string

	if bengkelOperasional.JamBuka != "" {
		dataTimePerHours = strings.Split(bengkelOperasional.JamBuka, " - ")
	}

	start, _ := time.Parse("15:04", dataTimePerHours[0])
	end, _ := time.Parse("15:04", dataTimePerHours[1])

	for current := start; current.Before(end); current = current.Add(time.Hour) {
		next := current.Add(time.Hour)
		timePerHoursOpen = append(timePerHoursOpen, current.Format("15:04"))
		timePerHoursClose = append(timePerHoursClose, next.Format("15:04"))
	}

	if err != nil {
		response := response.BuildFailedResponse("failed to get bengkel operasional by day", err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	data := make(map[int]string)

	for i := 0; i < len(timePerHoursOpen); i++ {
		data[i] = timePerHoursOpen[i] + " - " + timePerHoursClose[i]
	}

	response := response.BuildSuccessResponse("success get bengkel operasional by day", data)
	c.JSON(http.StatusOK, response)
}

// UpdateBengkelPesananServiceById function
func (handler *BengkelHandler) UpdateBengkelPesananServiceById(c *gin.Context) {
	userId := c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()

	_, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	pesananId := c.Param("pesananId")

	if pesananId == "" {
		response := response.BuildFailedResponse("failed to get pesanan", "pesananId params is empty")
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	var requestUpdateBengkelService validator.PesananUpdateRequest

	err = c.ShouldBindJSON(&requestUpdateBengkelService)
	if err != nil {
		response := response.BuildFailedResponse("failed to bind json", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	bengkelPesananRepo := repository.GetPesananRepository()

	pesanan, err := bengkelPesananRepo.GetDetailPesananById(pesananId, userId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get pesanan", err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	err = bengkelPesananRepo.UpdatePesananById(pesanan.ID,
		&models.Pesanan{
			IsHomeService:       &requestUpdateBengkelService.IsHomeService,
			HomeServiceSchedule: requestUpdateBengkelService.HomeServiceSchedule,
			PaymentMethod:       requestUpdateBengkelService.PaymentMethod,
		})
	if err != nil {
		response := response.BuildFailedResponse("failed to update pesanan", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success update bengkel pesanan service", nil)
	c.JSON(http.StatusOK, response)
}

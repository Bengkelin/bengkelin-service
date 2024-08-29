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
	"github.com/Bengkelin/bengkelin-service/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/mashingan/smapping"
)

var (
	userHandler *UserHandler
)

type UserHandler struct{}

func GetUserHandler() UserHandlerInterface {
	if userHandler == nil {
		userHandler = &UserHandler{}
	}
	return userHandler
}

type UserHandlerInterface interface {
	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	UpdateAvatarUser(c *gin.Context)
	GetDetailAddressUser(c *gin.Context)
	DeleteAddressUser(c *gin.Context)
	GetDetailVehicleUser(c *gin.Context)
	DeleteVehicleUser(c *gin.Context)
}

func (handler *UserHandler) GetProfile(c *gin.Context) {
	userId := c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()

	user, err := userRepo.FindUserByID(userId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get profile", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success get profile", user)
	c.JSON(http.StatusOK, response)
}

func (handler *UserHandler) UpdateProfile(c *gin.Context) {
	userId := c.MustGet("id").(string)

	var userUpdateRequest validator.UserUpdateRequest

	userRepo := repository.GetUserRepository()

	user, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	err = c.ShouldBind(&userUpdateRequest)
	if err != nil {
		response := response.BuildFailedResponse("request doesn't match with validator", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	if userUpdateRequest.PhoneNumber != "" {
		user.PhoneNumber = userUpdateRequest.PhoneNumber

		err = userRepo.UpdateUserById(userId, user)
	} else {
		addressRepo := repository.GetAddressRepository()

		addressUser, err := addressRepo.GetAddressById(userId, user.Addresses[0].ID)

		if err != nil {
			response := response.BuildFailedResponse("address user not found", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		smapping.FillStruct(addressUser, smapping.MapFields(&userUpdateRequest))

		err = addressRepo.UpdateAddressById(addressUser.ID, userId, addressUser)
		if err != nil {
			response := response.BuildFailedResponse("failed to update address user", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}

	if err != nil {
		response := response.BuildFailedResponse("failed to update user profile", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	response := response.BuildSuccessResponse("success update profile", user)
	c.JSON(http.StatusOK, response)
}

func (handler *UserHandler) UpdateAvatarUser(c *gin.Context) {
	userId := c.MustGet("id").(string)

	avatar, err := c.FormFile("avatar")
	if err != nil {
		response := response.BuildFailedResponse("failed to get file", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	userRepo := repository.GetUserRepository()

	_, err = userRepo.FindUserByID(userId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get user", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
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

	userModel := &models.User{
		AvatarUrl: urlPicture,
	}

	err = userRepo.UpdateUserById(userId, userModel)
	if err != nil {
		response := response.BuildFailedResponse("failed to update avatar", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success update avatar", nil)
	c.JSON(http.StatusOK, response)
}

func (handler *UserHandler) DeleteAddressUser(c *gin.Context) {
	userId := c.MustGet("id").(string)

	addressId := c.Param("addressId")

	if addressId == "" {
		response := response.BuildFailedResponse("address id is required", nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	addressIdUint, err := strconv.ParseUint(addressId, 10, 64)
	if err != nil {
		response := response.BuildFailedResponse("address id must be a number", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	addressRepo := repository.GetAddressRepository()

	_, err = addressRepo.GetAddressById(userId, uint(addressIdUint))
	if err != nil {
		response := response.BuildFailedResponse("failed to get address", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	err = addressRepo.DeleteAddressById(uint(addressIdUint), userId)
	if err != nil {
		response := response.BuildFailedResponse("failed to delete address", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success delete address", nil)
	c.JSON(http.StatusOK, response)
}

func (handler *UserHandler) GetDetailAddressUser(c *gin.Context) {
	userId := c.MustGet("id").(string)

	addressId := c.Param("addressId")

	addressIdUint, err := strconv.Atoi(addressId)

	if err != nil {
		response := response.BuildFailedResponse("address id must be a number", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	addressRepo := repository.GetAddressRepository()

	address, err := addressRepo.GetAddressById(userId, uint(addressIdUint))

	if err != nil {
		response := response.BuildFailedResponse("failed to get address", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success get address", address)
	c.JSON(http.StatusOK, response)
}

func (handler *UserHandler) DeleteVehicleUser(c *gin.Context) {
	userId := c.MustGet("id").(string)

	vehicleRepo := repository.GetVehicleRepository()

	vehicleId := c.Param("vehicleId")

	if vehicleId == "" {
		response := response.BuildFailedResponse("vehicle id is required", nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	vehicleIdUint, err := strconv.ParseUint(vehicleId, 10, 64)
	if err != nil {
		response := response.BuildFailedResponse("vehicle id must be a number", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	_, err = vehicleRepo.GetVehicleById(userId, uint(vehicleIdUint))
	if err != nil {
		response := response.BuildFailedResponse("failed to get vehicle", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	err = vehicleRepo.DeleteVehicleById(uint(vehicleIdUint), userId)
	if err != nil {
		response := response.BuildFailedResponse("failed to delete vehicle", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success delete vehicle", nil)
	c.JSON(http.StatusOK, response)
}

func (handler *UserHandler) GetDetailVehicleUser(c *gin.Context) {
	userId := c.MustGet("id").(string)
	vehicleId := c.Param("vehicleId")

	vehicleIdUint, err := strconv.Atoi(vehicleId)
	if err != nil {
		response := response.BuildFailedResponse("vehicle id must be a number", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	vehicleRepo := repository.GetVehicleRepository()

	vehicle, err := vehicleRepo.GetVehicleById(userId, uint(vehicleIdUint))

	if err != nil {
		response := response.BuildFailedResponse("failed to get vehicle", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success get vehicle", vehicle)
	c.JSON(http.StatusOK, response)
}

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
	DeleteAddressUser(c *gin.Context)
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

		addressUser, err := addressRepo.GetAddressById(userId)

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

	addressRepo := repository.GetAddressRepository()

	data, err := addressRepo.GetAddressById(userId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get address", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	err = addressRepo.DeleteAddressById(data.ID, userId)
	if err != nil {
		response := response.BuildFailedResponse("failed to delete address", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success delete address", nil)
	c.JSON(http.StatusOK, response)
}

func (handler *UserHandler) DeleteVehicleUser(c *gin.Context) {
	userId := c.MustGet("id").(string)

	vehicleRepo := repository.GetVehicleRepository()

	data, err := vehicleRepo.GetVehicleById(userId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get vehicle", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	err = vehicleRepo.DeleteVehicleById(data.ID, userId)
	if err != nil {
		response := response.BuildFailedResponse("failed to delete vehicle", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success delete vehicle", nil)
	c.JSON(http.StatusOK, response)
}
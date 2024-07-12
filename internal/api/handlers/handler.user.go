package handlers

import (
	"net/http"

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

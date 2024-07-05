package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/validator"
	"github.com/Bengkelin/bengkelin-service/pkg/crypto"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	"github.com/Bengkelin/bengkelin-service/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/mashingan/smapping"
)

var (
	authHandler *AuthHandler
)

type AuthHandler struct{}

type AuthHandlerInterface interface {
	AuthLogin(c *gin.Context)
	AuthRegister(c *gin.Context)
}

func GetAuthHandler() AuthHandlerInterface {
	if authHandler == nil {
		authHandler = &AuthHandler{}
	}
	return authHandler
}

func (handler *AuthHandler) AuthLogin(c *gin.Context) {
	var loginRequest validator.LoginRequest
	err := c.ShouldBind(&loginRequest)

	// Error when binding in validator
	if err != nil {
		response := response.BuildFailedResponse("failed to login", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	userRepo := repository.GetUserRepository()
	// If user doesn't exist
	user, err := userRepo.FindUserByEmail(loginRequest.Email)
	if err != nil {
		response := response.BuildFailedResponse("failed to login", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// if user is nil
	if user == nil {
		response := response.BuildFailedResponse("failed to login", errors.New("user doesn't exist").Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	// Wrong password
	passwordHelper := crypto.GetPasswordCryptoHelper()
	if !passwordHelper.ComparePassword(user.Password, []byte(loginRequest.Password)) {
		response := response.BuildFailedResponse("failed to login", errors.New("wrong credential").Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Correct password
	tokenHelper := crypto.GetJWTCrypto()
	token, err := tokenHelper.GenerateToken(fmt.Sprint(user.ID))
	if err != nil {
		response := response.BuildFailedResponse("wrong credential", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	response := response.BuildSuccessResponse("success login", map[string]interface{}{
		"token": token,
	})
	c.JSON(http.StatusOK, response)
}

func (handler *AuthHandler) AuthRegister(c *gin.Context) {
	var registerRequest validator.RegisterNewUserRequest
	err := c.ShouldBind(&registerRequest)

	if err != nil {
		response := response.BuildFailedResponse("failed to register", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	if registerRequest.Password != registerRequest.ConfirmPassword {
		response := response.BuildFailedResponse("failed to register", errors.New("password and confirm password doesn't match").Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	userRepo := repository.GetUserRepository()
	passwordHelper := crypto.GetPasswordCryptoHelper()
	userModel := &models.User{
		ID: helpers.GenerateUUID(),
	}

	// smapping the struct
	smapping.FillStruct(userModel, smapping.MapFields(&registerRequest))
	userModel.Password, _ = passwordHelper.HashAndSalt([]byte(registerRequest.Password))

	if newUser, err := userRepo.CreateUser(*userModel); err != nil {
		response := response.BuildFailedResponse("failed to register", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	} else {
		if err != nil {
			response := response.BuildFailedResponse("failed to generate token", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}
		response := response.BuildSuccessResponse("success register new user", newUser)
		c.JSON(http.StatusCreated, response)
		return
	}
}

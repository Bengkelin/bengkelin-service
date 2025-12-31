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
	CreateAddress(c *gin.Context)
	UpdateAddress(c *gin.Context)
	UpdateOrCreateAddress(c *gin.Context)
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

	// Update user profile fields only
	userUpdated := false
	if userUpdateRequest.FirstName != "" {
		user.FirstName = userUpdateRequest.FirstName
		userUpdated = true
	}
	if userUpdateRequest.LastName != "" {
		user.LastName = userUpdateRequest.LastName
		userUpdated = true
	}
	if userUpdateRequest.PhoneNumber != "" {
		user.PhoneNumber = userUpdateRequest.PhoneNumber
		userUpdated = true
	}

	// Update user if any user fields were changed
	if userUpdated {
		err = userRepo.UpdateUserById(userId, user)
		if err != nil {
			response := response.BuildFailedResponse("failed to update user profile", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}

	// Get updated user data to return
	updatedUser, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("failed to get updated user", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success update profile", updatedUser)
	c.JSON(http.StatusOK, response)
}

func (handler *UserHandler) CreateAddress(c *gin.Context) {
	userId := c.MustGet("id").(string)

	var addressCreateRequest validator.UserAddressCreateRequest

	err := c.ShouldBind(&addressCreateRequest)
	if err != nil {
		response := response.BuildFailedResponse("request doesn't match with validator", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	userRepo := repository.GetUserRepository()
	addressRepo := repository.GetAddressRepository()

	// Check if user exists
	_, err = userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("user not found", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Handle IsPrimary logic
	var isPrimary bool
	if addressCreateRequest.IsPrimary != nil {
		isPrimary = *addressCreateRequest.IsPrimary
	} else {
		// If not specified, check if user has any addresses
		user, err := userRepo.GetDetailUser(userId)
		if err == nil && len(user.Addresses) == 0 {
			// If this is the first address, make it primary by default
			isPrimary = true
		} else {
			// If user already has addresses, make this non-primary by default
			isPrimary = false
		}
	}

	// If setting this address as primary, unset all other addresses as non-primary
	if isPrimary {
		user, err := userRepo.GetDetailUser(userId)
		if err == nil {
			for i := range user.Addresses {
				nonPrimary := false
				user.Addresses[i].IsPrimary = &nonPrimary
				err = addressRepo.UpdateAddressById(user.Addresses[i].ID, userId, &user.Addresses[i])
				if err != nil {
					response := response.BuildFailedResponse("failed to update existing addresses", err.Error())
					c.AbortWithStatusJSON(http.StatusBadRequest, response)
					return
				}
			}
		}
	}

	// Create new address
	newAddress := &models.UserAddress{
		UserID:       userId,
		Latitude:     addressCreateRequest.Latitude,
		Longitude:    addressCreateRequest.Longitude,
		AddressLabel: addressCreateRequest.AddressLabel,
		FullAddress:  addressCreateRequest.FullAddress,
		Note:         addressCreateRequest.Note,
		IsPrimary:    &isPrimary,
	}

	createdAddress, err := addressRepo.CreateAddress(*newAddress)
	if err != nil {
		response := response.BuildFailedResponse("failed to create address", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success create address", createdAddress)
	c.JSON(http.StatusOK, response)
}

func (handler *UserHandler) UpdateAddress(c *gin.Context) {
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

	var addressUpdateRequest validator.UserAddressUpdateRequest

	err = c.ShouldBind(&addressUpdateRequest)
	if err != nil {
		response := response.BuildFailedResponse("request doesn't match with validator", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	addressRepo := repository.GetAddressRepository()

	// Check if address exists and belongs to user
	existingAddress, err := addressRepo.GetAddressById(userId, uint(addressIdUint))
	if err != nil {
		response := response.BuildFailedResponse("address not found", err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	// Update address fields
	addressUpdated := false
	
	// Handle IsPrimary field changes first
	if addressUpdateRequest.IsPrimary != nil {
		if *addressUpdateRequest.IsPrimary {
			// If setting this address as primary, unset all other addresses as non-primary
			userRepo := repository.GetUserRepository()
			user, err := userRepo.GetDetailUser(userId)
			if err == nil {
				for i := range user.Addresses {
					if user.Addresses[i].ID != existingAddress.ID {
						nonPrimary := false
						user.Addresses[i].IsPrimary = &nonPrimary
						addressRepo.UpdateAddressById(user.Addresses[i].ID, userId, &user.Addresses[i])
					}
				}
			}
		}
		existingAddress.IsPrimary = addressUpdateRequest.IsPrimary
		addressUpdated = true
	}
	
	if addressUpdateRequest.Latitude != 0 {
		existingAddress.Latitude = addressUpdateRequest.Latitude
		addressUpdated = true
	}
	if addressUpdateRequest.Longitude != 0 {
		existingAddress.Longitude = addressUpdateRequest.Longitude
		addressUpdated = true
	}
	if addressUpdateRequest.AddressLabel != "" {
		existingAddress.AddressLabel = addressUpdateRequest.AddressLabel
		addressUpdated = true
	}
	if addressUpdateRequest.FullAddress != "" {
		existingAddress.FullAddress = addressUpdateRequest.FullAddress
		addressUpdated = true
	}
	if addressUpdateRequest.Note != "" {
		existingAddress.Note = addressUpdateRequest.Note
		addressUpdated = true
	}

	// Update address if any fields were changed
	if addressUpdated {
		err = addressRepo.UpdateAddressById(existingAddress.ID, userId, existingAddress)
		if err != nil {
			response := response.BuildFailedResponse("failed to update address", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}

	// Get updated address to return
	updatedAddress, err := addressRepo.GetAddressById(userId, uint(addressIdUint))
	if err != nil {
		response := response.BuildFailedResponse("failed to get updated address", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success update address", updatedAddress)
	c.JSON(http.StatusOK, response)
}

func (handler *UserHandler) UpdateOrCreateAddress(c *gin.Context) {
	userId := c.MustGet("id").(string)

	var addressUpdateRequest validator.UserAddressUpdateRequest

	err := c.ShouldBind(&addressUpdateRequest)
	if err != nil {
		response := response.BuildFailedResponse("request doesn't match with validator", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	userRepo := repository.GetUserRepository()
	addressRepo := repository.GetAddressRepository()

	// Get user with addresses
	user, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("user not found", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check if user has any addresses
	if len(user.Addresses) == 0 {
		// Create new address - first address is always primary
		isPrimary := true
		newAddress := &models.UserAddress{
			UserID:       userId,
			Latitude:     addressUpdateRequest.Latitude,
			Longitude:    addressUpdateRequest.Longitude,
			AddressLabel: addressUpdateRequest.AddressLabel,
			FullAddress:  addressUpdateRequest.FullAddress,
			Note:         addressUpdateRequest.Note,
			IsPrimary:    &isPrimary, // First address is always primary
		}

		// Set default label if not provided
		if newAddress.AddressLabel == "" {
			newAddress.AddressLabel = "Home"
		}

		createdAddress, err := addressRepo.CreateAddress(*newAddress)
		if err != nil {
			response := response.BuildFailedResponse("failed to create address", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}

		response := response.BuildSuccessResponse("success create address", createdAddress)
		c.JSON(http.StatusOK, response)
		return
	}

	// Find primary address or use first address
	var targetAddress *models.UserAddress
	for i := range user.Addresses {
		if user.Addresses[i].IsPrimary != nil && *user.Addresses[i].IsPrimary {
			targetAddress = &user.Addresses[i]
			break
		}
	}
	
	// If no primary address found, use the first one and make it primary
	if targetAddress == nil {
		targetAddress = &user.Addresses[0]
		isPrimary := true
		targetAddress.IsPrimary = &isPrimary
	}

	// Handle IsPrimary field changes
	if addressUpdateRequest.IsPrimary != nil {
		if *addressUpdateRequest.IsPrimary {
			// If setting this address as primary, unset all other addresses as non-primary
			for i := range user.Addresses {
				if user.Addresses[i].ID != targetAddress.ID {
					nonPrimary := false
					user.Addresses[i].IsPrimary = &nonPrimary
					err = addressRepo.UpdateAddressById(user.Addresses[i].ID, userId, &user.Addresses[i])
					if err != nil {
						response := response.BuildFailedResponse("failed to update other addresses", err.Error())
						c.AbortWithStatusJSON(http.StatusBadRequest, response)
						return
					}
				}
			}
			targetAddress.IsPrimary = addressUpdateRequest.IsPrimary
		} else {
			// If unsetting primary, we need to ensure at least one address remains primary
			// Only allow unsetting if there are other addresses that can be primary
			if len(user.Addresses) > 1 {
				targetAddress.IsPrimary = addressUpdateRequest.IsPrimary
			}
			// If this is the only address, ignore the request to unset primary
		}
	}

	// Update other address fields
	addressUpdated := false
	if addressUpdateRequest.Latitude != 0 {
		targetAddress.Latitude = addressUpdateRequest.Latitude
		addressUpdated = true
	}
	if addressUpdateRequest.Longitude != 0 {
		targetAddress.Longitude = addressUpdateRequest.Longitude
		addressUpdated = true
	}
	if addressUpdateRequest.AddressLabel != "" {
		targetAddress.AddressLabel = addressUpdateRequest.AddressLabel
		addressUpdated = true
	}
	if addressUpdateRequest.FullAddress != "" {
		targetAddress.FullAddress = addressUpdateRequest.FullAddress
		addressUpdated = true
	}
	if addressUpdateRequest.Note != "" {
		targetAddress.Note = addressUpdateRequest.Note
		addressUpdated = true
	}

	// Update address if any fields were changed or IsPrimary was set
	if addressUpdated || addressUpdateRequest.IsPrimary != nil {
		err = addressRepo.UpdateAddressById(targetAddress.ID, userId, targetAddress)
		if err != nil {
			response := response.BuildFailedResponse("failed to update address", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}

	// Get updated address to return
	updatedAddress, err := addressRepo.GetAddressById(userId, targetAddress.ID)
	if err != nil {
		response := response.BuildFailedResponse("failed to get updated address", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success update address", updatedAddress)
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

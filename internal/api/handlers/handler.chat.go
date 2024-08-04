package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	rtmtokenbuilder2 "github.com/AgoraIO-Community/go-tokenbuilder/rtmtokenbuilder"
	"github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/chatTokenBuilder"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/validator"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	"github.com/Bengkelin/bengkelin-service/pkg/response"
	"github.com/gin-gonic/gin"
)

var (
	chatHandler *ChatHandler
)

type ChatHandler struct{}

func GetChatHandler() ChatHandlerInterface {
	if chatHandler == nil {
		chatHandler = &ChatHandler{}
	}
	return chatHandler
}

type ChatHandlerInterface interface {
	CreateRtmToken(c *gin.Context)
	CreateAppToken(c *gin.Context)
	CreateChatToken(c *gin.Context)
	CreateChatHistoryUser(c *gin.Context)
	CreateChatHistoryBengkel(c *gin.Context)
}

func (handler *ChatHandler) CreateRtmToken(c *gin.Context) {
	userId, expireTimestamp, err := parseRtmParams(c)
	if err != nil {
		response := response.BuildFailedResponse("failed to create rtm token", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	config := config.GetConfig()

	rtmToken, tokenErr := rtmtokenbuilder2.BuildToken(config.Agore.AppID, config.Agore.AppCertificate, userId, expireTimestamp, "")

	if tokenErr != nil {
		response := response.BuildFailedResponse("failed to create rtm token", tokenErr.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success create rtm token", map[string]string{
		"rtm_token": rtmToken,
	})
	c.JSON(http.StatusOK, response)
}

func parseRtmParams(c *gin.Context) (userId string, expireTimestamp uint32, errorParse error) {
	userId = c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()
	_, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	config := config.GetConfig()
	expireTime := config.Agore.ExpiryTime

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		parseErr = fmt.Errorf("failed to parse expireTime: %s, causing error: %s", expireTime, parseErr)
		return "", 0, parseErr
	}

	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp = currentTimestamp + expireTimeInSeconds

	return userId, expireTimestamp, parseErr
}

func (handler *ChatHandler) CreateAppToken(c *gin.Context) {
	userId := c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()
	_, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	config := config.GetConfig()

	expire := config.Agore.ExpiryTime

	expireUint, err := strconv.ParseUint(expire, 10, 32)
	if err != nil {
		response := response.BuildFailedResponse("failed to parse expire time to uint32", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	appToken, err := chatTokenBuilder.BuildChatAppToken(config.Agore.AppID, config.Agore.AppCertificate, uint32(expireUint))

	if err != nil {
		response := response.BuildFailedResponse("failed to get app token", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success create app token", map[string]string{
		"app_token": appToken,
	})
	c.JSON(http.StatusOK, response)
}

func (handler *ChatHandler) CreateChatToken(c *gin.Context) {
	userId := c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()
	_, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	agoraUserId := c.Query("agoraId")

	if agoraUserId == "" {
		response := response.BuildFailedResponse("agora user id is required", nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	config := config.GetConfig()

	expire := config.Agore.ExpiryTime

	expireUint, err := strconv.ParseUint(expire, 10, 32)
	if err != nil {
		response := response.BuildFailedResponse("failed to parse expire time to uint32", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	chatToken, err := chatTokenBuilder.BuildChatUserToken(config.Agore.AppID, config.Agore.AppCertificate, agoraUserId, uint32(expireUint))

	if err != nil {
		response := response.BuildFailedResponse("failed to get chat token", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success create chat token", map[string]string{
		"chat_token": chatToken,
	})
	c.JSON(http.StatusOK, response)
}

func (handler *ChatHandler) CreateChatHistoryUser(c *gin.Context) {
	userId := c.MustGet("id").(string)

	userRepo := repository.GetUserRepository()
	_, err := userRepo.GetDetailUser(userId)
	if err != nil {
		response := response.BuildFailedResponse("users not found", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	var chatRequest validator.ChatRequest

	err = c.ShouldBindJSON(&chatRequest)
	if err != nil {
		response := response.BuildFailedResponse("request chat history doesn't match with validator", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	chatRepo := repository.GetChatHistoryRepository()

	chatHistoryModel := &models.ChatHistory{
		ID:             helpers.GenerateUUID(),
		MessageText:    chatRequest.MessageText,
		Type:           chatRequest.Type,
		ImageUrl:       chatRequest.ImageUrl,
		SenderUserId:   chatRequest.SenderUserId,
		ReceiverUserId: chatRequest.ReceiverUserId,
	}

	res, err := chatRepo.CreateChatHistory(*chatHistoryModel)

	if err != nil {
		response := response.BuildFailedResponse("failed to create chat history user", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success create chat history user", res)
	c.JSON(http.StatusOK, response)
}

func (handler *ChatHandler) CreateChatHistoryBengkel(c *gin.Context) {
	mitraId := c.MustGet("id").(string)

	mitraRepo := repository.GetMitraRepository()
	_, err := mitraRepo.FindMitraByID(mitraId)

	if err != nil {
		response := response.BuildFailedResponse("mitra not found", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	var chatRequest validator.ChatRequest

	err = c.ShouldBindJSON(&chatRequest)
	if err != nil {
		response := response.BuildFailedResponse("request chat history doesn't match with validator", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	chatRepo := repository.GetChatHistoryRepository()

	chatHistoryModel := &models.ChatHistory{
		ID:             helpers.GenerateUUID(),
		MessageText:    chatRequest.MessageText,
		Type:           chatRequest.Type,
		ImageUrl:       chatRequest.ImageUrl,
		SenderUserId:   chatRequest.SenderUserId,
		ReceiverUserId: chatRequest.ReceiverUserId,
	}

	res, err := chatRepo.CreateChatHistory(*chatHistoryModel)

	if err != nil {
		response := response.BuildFailedResponse("failed to create chat history bengkel", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := response.BuildSuccessResponse("success create chat history bengkel", res)
	c.JSON(http.StatusOK, response)
}

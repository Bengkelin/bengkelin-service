package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	rtmtokenbuilder2 "github.com/AgoraIO-Community/go-tokenbuilder/rtmtokenbuilder"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
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

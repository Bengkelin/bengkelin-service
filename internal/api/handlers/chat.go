package handlers

import (
	"net/http"
	"strconv"

	rtmtokenbuilder2 "github.com/AgoraIO-Community/go-tokenbuilder/rtmtokenbuilder"
	"github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/chatTokenBuilder"
	"github.com/Bengkelin/bengkelin-service/internal/config"
	"github.com/Bengkelin/bengkelin-service/internal/container"
	"github.com/Bengkelin/bengkelin-service/internal/dto"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/errors"
	"github.com/Bengkelin/bengkelin-service/internal/service"
	"github.com/Bengkelin/bengkelin-service/internal/validator"
	"github.com/Bengkelin/bengkelin-service/internal/response"
	"github.com/gin-gonic/gin"
)

var (
	chatHandler *ChatHandler
)

type ChatHandler struct {
	BaseHandler
	chatService service.ChatServiceInterface
}

func GetChatHandler() ChatHandlerInterface {
	if chatHandler == nil {
		c := container.GetContainer()
		chatHandler = &ChatHandler{
			chatService: c.ChatService,
		}
	}
	return chatHandler
}

type ChatHandlerInterface interface {
	CreateRtmToken(c *gin.Context)
	CreateAppToken(c *gin.Context)
	CreateChatToken(c *gin.Context)
	CreateAppTokenMitra(c *gin.Context)
	CreateChatTokenMitra(c *gin.Context)
	CreateChatHistoryUser(c *gin.Context)
	CreateChatHistoryBengkel(c *gin.Context)
	GetChatHistoryUser(c *gin.Context)
	GetChatHistoryBengkel(c *gin.Context)
}

func (h *ChatHandler) CreateRtmToken(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	rtmUserId, expireTimestamp, err := h.chatService.GenerateRtmParams(ctx, userId)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	cfg := config.GetConfig()

	rtmToken, tokenErr := rtmtokenbuilder2.BuildToken(cfg.Agore.AppID, cfg.Agore.AppCertificate, rtmUserId, expireTimestamp, "")

	if tokenErr != nil {
		h.HandleError(c, appErrors.ErrExternalAPI.WithDetails("failed to create rtm token: "+tokenErr.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success create rtm token", map[string]string{
		"rtm_token": rtmToken,
	})
	c.JSON(http.StatusOK, resp)
}

func (h *ChatHandler) CreateAppToken(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if err := h.chatService.ValidateUserExists(ctx, userId); err != nil {
		h.HandleError(c, appErrors.ErrUserNotFound.WithDetails(err.Error()))
		return
	}

	cfg := config.GetConfig()

	expire := cfg.Agore.ExpiryTime

	expireUint, err := strconv.ParseUint(expire, 10, 32)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("failed to parse expire time: "+err.Error()))
		return
	}

	appToken, err := chatTokenBuilder.BuildChatAppToken(cfg.Agore.AppID, cfg.Agore.AppCertificate, uint32(expireUint))

	if err != nil {
		h.HandleError(c, appErrors.ErrExternalAPI.WithDetails("failed to get app token: "+err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success create app token", map[string]string{
		"app_token": appToken,
	})
	c.JSON(http.StatusOK, resp)
}

func (h *ChatHandler) CreateChatToken(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if err := h.chatService.ValidateUserExists(ctx, userId); err != nil {
		h.HandleError(c, appErrors.ErrUserNotFound.WithDetails(err.Error()))
		return
	}

	agoraUserId := c.Query("agoraId")

	if agoraUserId == "" {
		h.HandleError(c, appErrors.ErrMissingField.WithDetails("agora user id is required"))
		return
	}

	cfg := config.GetConfig()

	expire := cfg.Agore.ExpiryTime

	expireUint, err := strconv.ParseUint(expire, 10, 32)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("failed to parse expire time: "+err.Error()))
		return
	}

	chatToken, err := chatTokenBuilder.BuildChatUserToken(cfg.Agore.AppID, cfg.Agore.AppCertificate, agoraUserId, uint32(expireUint))

	if err != nil {
		h.HandleError(c, appErrors.ErrExternalAPI.WithDetails("failed to get chat token: "+err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success create chat token", map[string]string{
		"chat_token": chatToken,
	})
	c.JSON(http.StatusOK, resp)
}

func (h *ChatHandler) CreateAppTokenMitra(c *gin.Context) {
	mitraId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if err := h.chatService.ValidateMitraExists(ctx, mitraId); err != nil {
		h.HandleError(c, appErrors.ErrMitraNotFound.WithDetails(err.Error()))
		return
	}

	cfg := config.GetConfig()

	expire := cfg.Agore.ExpiryTime

	expireUint, err := strconv.ParseUint(expire, 10, 32)
	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("failed to parse expire time: "+err.Error()))
		return
	}

	appToken, err := chatTokenBuilder.BuildChatAppToken(cfg.Agore.AppID, cfg.Agore.AppCertificate, uint32(expireUint))

	if err != nil {
		h.HandleError(c, appErrors.ErrExternalAPI.WithDetails("failed to get app token: "+err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success create app token mitra", map[string]string{
		"app_token": appToken,
	})
	c.JSON(http.StatusOK, resp)
}

func (h *ChatHandler) CreateChatTokenMitra(c *gin.Context) {
	mitraId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if err := h.chatService.ValidateMitraExists(ctx, mitraId); err != nil {
		h.HandleError(c, appErrors.ErrMitraNotFound.WithDetails(err.Error()))
		return
	}

	agoraUserId := c.Query("agoraId")

	if agoraUserId == "" {
		h.HandleError(c, appErrors.ErrMissingField.WithDetails("agora user id is required"))
		return
	}

	cfg := config.GetConfig()

	expire := cfg.Agore.ExpiryTime

	expireUint, err := strconv.ParseUint(expire, 10, 32)

	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("failed to parse expire time: "+err.Error()))
		return
	}

	chatToken, err := chatTokenBuilder.BuildChatUserToken(cfg.Agore.AppID, cfg.Agore.AppCertificate, agoraUserId, uint32(expireUint))

	if err != nil {
		h.HandleError(c, appErrors.ErrExternalAPI.WithDetails("failed to get chat token: "+err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success create chat token mitra", map[string]string{
		"chat_token": chatToken,
	})
	c.JSON(http.StatusOK, resp)
}

func (h *ChatHandler) CreateChatHistoryUser(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if err := h.chatService.ValidateUserExists(ctx, userId); err != nil {
		h.HandleError(c, appErrors.ErrUserNotFound.WithDetails(err.Error()))
		return
	}

	var chatRequest validator.ChatRequest
	if err := c.ShouldBindJSON(&chatRequest); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	req := dto.ChatHistoryRequest{
		UserID:    chatRequest.SenderUserId,
		BengkelID: chatRequest.ReceiverUserId,
		Message:   chatRequest.MessageText,
		UserType:  "user",
	}

	if err := h.chatService.SaveChatHistory(ctx, req); err != nil {
		h.HandleError(c, appErrors.ErrDatabaseError.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success create chat history user", nil)
	c.JSON(http.StatusOK, resp)
}

func (h *ChatHandler) CreateChatHistoryBengkel(c *gin.Context) {
	mitraId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if err := h.chatService.ValidateMitraExists(ctx, mitraId); err != nil {
		h.HandleError(c, appErrors.ErrMitraNotFound.WithDetails(err.Error()))
		return
	}

	var chatRequest validator.ChatRequest
	if err := c.ShouldBindJSON(&chatRequest); err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails(err.Error()))
		return
	}

	req := dto.ChatHistoryRequest{
		UserID:    chatRequest.SenderUserId,
		BengkelID: chatRequest.ReceiverUserId,
		Message:   chatRequest.MessageText,
		UserType:  "mitra",
	}

	if err := h.chatService.SaveChatHistory(ctx, req); err != nil {
		h.HandleError(c, appErrors.ErrDatabaseError.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success create chat history bengkel", nil)
	c.JSON(http.StatusOK, resp)
}

func (h *ChatHandler) GetChatHistoryUser(c *gin.Context) {
	userId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if err := h.chatService.ValidateUserExists(ctx, userId); err != nil {
		h.HandleError(c, appErrors.ErrUserNotFound.WithDetails(err.Error()))
		return
	}

	page := c.Query("page")
	limit := c.Query("limit")
	senderId := c.Query("senderId")
	receiverId := c.Query("receiverId")

	pageInt, err := strconv.Atoi(page)

	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("failed to parse page: "+err.Error()))
		return
	}
	limitInt, err := strconv.Atoi(limit)

	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("failed to parse limit: "+err.Error()))
		return
	}

	res, count, err := h.chatService.GetChatHistoryPaginate(ctx, pageInt, limitInt, senderId, receiverId)

	if err != nil {
		h.HandleError(c, appErrors.ErrDatabaseError.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success get chat history user", map[string]interface{}{
		"chat_history": res,
		"total":        count,
	})
	c.JSON(http.StatusOK, resp)
}

func (h *ChatHandler) GetChatHistoryBengkel(c *gin.Context) {
	mitraId := c.MustGet("id").(string)
	ctx := c.Request.Context()

	if err := h.chatService.ValidateMitraExists(ctx, mitraId); err != nil {
		h.HandleError(c, appErrors.ErrMitraNotFound.WithDetails(err.Error()))
		return
	}

	page := c.Query("page")
	limit := c.Query("limit")
	senderId := c.Query("senderId")
	receiverId := c.Query("receiverId")

	pageInt, err := strconv.Atoi(page)

	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("failed to parse page: "+err.Error()))
		return
	}
	limitInt, err := strconv.Atoi(limit)

	if err != nil {
		h.HandleError(c, appErrors.ErrValidationFailed.WithDetails("failed to parse limit: "+err.Error()))
		return
	}

	res, count, err := h.chatService.GetChatHistoryPaginate(ctx, pageInt, limitInt, senderId, receiverId)

	if err != nil {
		h.HandleError(c, appErrors.ErrDatabaseError.WithDetails(err.Error()))
		return
	}

	resp := response.BuildSuccessResponse("success get chat history bengkel", map[string]interface{}{
		"chat_history": res,
		"total":        count,
	})
	c.JSON(http.StatusOK, resp)
}

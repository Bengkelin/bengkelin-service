package handlers

import (
	"net/http"
	"strconv"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/service"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ChatV2Handler struct {
	chatService service.ChatV2ServiceInterface
	validator   *validator.Validate
}

func NewChatV2Handler(chatService service.ChatV2ServiceInterface) *ChatV2Handler {
	return &ChatV2Handler{
		chatService: chatService,
		validator:   validator.New(),
	}
}

// CreateOrGetChatRoom creates a new chat room or returns existing one
// @Summary Create or get chat room
// @Description Create a new chat room between user and bengkel or return existing one
// @Tags Chat V2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateChatRoomRequest true "Create chat room request"
// @Success 200 {object} dto.ChatRoomResponse
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/v2/chat/rooms [post]
func (h *ChatV2Handler) CreateOrGetChatRoom(c *gin.Context) {
	var req dto.CreateChatRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Invalid request body")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Validation failed")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "User ID not found in context")
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	response, err := h.chatService.CreateOrGetChatRoom(c.Request.Context(), userID, req.BengkelID)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to create or get chat room")
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create or get chat room", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Chat room created or retrieved successfully", response)
}

// GetUserChatRooms gets all chat rooms for a user
// @Summary Get user chat rooms
// @Description Get all chat rooms for the authenticated user with pagination
// @Tags Chat V2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.PaginatedRoomsResponse
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/v2/chat/rooms [get]
func (h *ChatV2Handler) GetUserChatRooms(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "User ID not found in context")
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	req := dto.GetRoomsRequest{
		Page:  page,
		Limit: limit,
	}

	response, err := h.chatService.GetUserChatRooms(c.Request.Context(), userID, req)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to get user chat rooms")
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get chat rooms", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Chat rooms retrieved successfully", response)
}

// GetBengkelChatRooms gets all chat rooms for a bengkel (mitra)
// @Summary Get bengkel chat rooms
// @Description Get all chat rooms for the authenticated mitra's bengkel with pagination
// @Tags Chat V2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.PaginatedRoomsResponse
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/v2/chat/bengkel/rooms [get]
func (h *ChatV2Handler) GetBengkelChatRooms(c *gin.Context) {
	mitraID := c.GetString("mitra_id")
	if mitraID == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "Mitra ID not found in context")
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "Mitra ID not found")
		return
	}

	// Get bengkel ID from mitra ID (this would need to be implemented in the service)
	// For now, we'll assume the bengkel ID is passed as a query parameter
	bengkelID := c.Query("bengkel_id")
	if bengkelID == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "Bengkel ID not provided")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Bad request", "Bengkel ID is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	req := dto.GetRoomsRequest{
		Page:  page,
		Limit: limit,
	}

	response, err := h.chatService.GetBengkelChatRooms(c.Request.Context(), bengkelID, req)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to get bengkel chat rooms")
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get chat rooms", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Chat rooms retrieved successfully", response)
}

// GetChatRoom gets a specific chat room by ID
// @Summary Get chat room
// @Description Get a specific chat room by ID
// @Tags Chat V2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "Room ID"
// @Success 200 {object} dto.ChatRoomResponse
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/v2/chat/rooms/{roomId} [get]
func (h *ChatV2Handler) GetChatRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Bad request", "Room ID is required")
		return
	}

	userID := c.GetString("user_id")
	userType := "user"
	
	// Check if it's a mitra request
	if mitraID := c.GetString("mitra_id"); mitraID != "" {
		userID = mitraID
		userType = "mitra"
	}

	if userID == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "User ID not found in context")
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	response, err := h.chatService.GetChatRoomByID(c.Request.Context(), roomID, userID, userType)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to get chat room")
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get chat room", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Chat room retrieved successfully", response)
}

// SendMessage sends a new message to a chat room
// @Summary Send message
// @Description Send a new text message to a chat room
// @Tags Chat V2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.SendMessageRequest true "Send message request"
// @Success 201 {object} dto.ChatMessageResponse
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/v2/chat/messages [post]
func (h *ChatV2Handler) SendMessage(c *gin.Context) {
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Invalid request body")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Validation failed")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	userID := c.GetString("user_id")
	userType := "user"
	
	// Check if it's a mitra request
	if mitraID := c.GetString("mitra_id"); mitraID != "" {
		userID = mitraID
		userType = "mitra"
	}

	if userID == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "User ID not found in context")
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	response, err := h.chatService.SendMessage(c.Request.Context(), userID, userType, req)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to send message")
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to send message", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, "Message sent successfully", response)
}

// SendFileMessage sends a file message to a chat room
// @Summary Send file message
// @Description Send a file message to a chat room
// @Tags Chat V2
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param room_id formData string true "Room ID"
// @Param file formData file true "File to upload"
// @Param reply_to_id formData string false "Reply to message ID"
// @Success 201 {object} dto.ChatMessageResponse
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/v2/chat/messages/file [post]
func (h *ChatV2Handler) SendFileMessage(c *gin.Context) {
	roomID := c.PostForm("room_id")
	if roomID == "" {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Bad request", "Room ID is required")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to get file from request")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Bad request", "File is required")
		return
	}
	defer file.Close()

	userID := c.GetString("user_id")
	userType := "user"
	
	// Check if it's a mitra request
	if mitraID := c.GetString("mitra_id"); mitraID != "" {
		userID = mitraID
		userType = "mitra"
	}

	if userID == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "User ID not found in context")
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	// TODO: Upload file to storage service (Cloudinary, S3, etc.)
	// For now, we'll use a placeholder URL
	fileURL := "https://example.com/files/" + header.Filename

	var replyToID *string
	if replyTo := c.PostForm("reply_to_id"); replyTo != "" {
		replyToID = &replyTo
	}

	req := dto.SendFileMessageRequest{
		RoomID:    roomID,
		FileName:  header.Filename,
		FileSize:  header.Size,
		ReplyToID: replyToID,
	}

	if err := h.validator.Struct(req); err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Validation failed")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	response, err := h.chatService.SendFileMessage(c.Request.Context(), userID, userType, req, fileURL)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to send file message")
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to send file message", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, "File message sent successfully", response)
}

// GetRoomMessages gets messages from a chat room
// @Summary Get room messages
// @Description Get messages from a chat room with pagination
// @Tags Chat V2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "Room ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Param before query string false "Get messages before this message ID"
// @Success 200 {object} dto.PaginatedMessagesResponse
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/v2/chat/rooms/{roomId}/messages [get]
func (h *ChatV2Handler) GetRoomMessages(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Bad request", "Room ID is required")
		return
	}

	userID := c.GetString("user_id")
	userType := "user"
	
	// Check if it's a mitra request
	if mitraID := c.GetString("mitra_id"); mitraID != "" {
		userID = mitraID
		userType = "mitra"
	}

	if userID == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "User ID not found in context")
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	var before *string
	if beforeParam := c.Query("before"); beforeParam != "" {
		before = &beforeParam
	}

	req := dto.GetMessagesRequest{
		RoomID: roomID,
		Page:   page,
		Limit:  limit,
		Before: before,
	}

	if err := h.validator.Struct(req); err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Validation failed")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	response, err := h.chatService.GetRoomMessages(c.Request.Context(), roomID, userID, userType, req)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to get room messages")
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get messages", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Messages retrieved successfully", response)
}

// EditMessage edits an existing message
// @Summary Edit message
// @Description Edit an existing message (only text messages can be edited)
// @Tags Chat V2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param messageId path string true "Message ID"
// @Param request body dto.EditMessageRequest true "Edit message request"
// @Success 200 {object} dto.ChatMessageResponse
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/v2/chat/messages/{messageId} [patch]
func (h *ChatV2Handler) EditMessage(c *gin.Context) {
	messageID := c.Param("messageId")
	if messageID == "" {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Bad request", "Message ID is required")
		return
	}

	var req dto.EditMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Invalid request body")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Validation failed")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	userID := c.GetString("user_id")
	
	// Check if it's a mitra request
	if mitraID := c.GetString("mitra_id"); mitraID != "" {
		userID = mitraID
	}

	if userID == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "User ID not found in context")
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	response, err := h.chatService.EditMessage(c.Request.Context(), messageID, userID, req)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to edit message")
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to edit message", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Message edited successfully", response)
}

// DeleteMessage deletes a message
// @Summary Delete message
// @Description Delete a message (only sender can delete their own messages)
// @Tags Chat V2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param messageId path string true "Message ID"
// @Success 200 {object} helpers.SuccessResponse
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/v2/chat/messages/{messageId} [delete]
func (h *ChatV2Handler) DeleteMessage(c *gin.Context) {
	messageID := c.Param("messageId")
	if messageID == "" {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Bad request", "Message ID is required")
		return
	}

	userID := c.GetString("user_id")
	
	// Check if it's a mitra request
	if mitraID := c.GetString("mitra_id"); mitraID != "" {
		userID = mitraID
	}

	if userID == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "User ID not found in context")
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	err := h.chatService.DeleteMessage(c.Request.Context(), messageID, userID)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to delete message")
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete message", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Message deleted successfully", nil)
}

// MarkMessagesAsRead marks messages as read
// @Summary Mark messages as read
// @Description Mark multiple messages as read
// @Tags Chat V2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.MessageReadRequest true "Mark messages as read request"
// @Success 200 {object} []dto.MessageReadResponse
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/v2/chat/messages/read [post]
func (h *ChatV2Handler) MarkMessagesAsRead(c *gin.Context) {
	var req dto.MessageReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Invalid request body")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Validation failed")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	userID := c.GetString("user_id")
	
	// Check if it's a mitra request
	if mitraID := c.GetString("mitra_id"); mitraID != "" {
		userID = mitraID
	}

	if userID == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "User ID not found in context")
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	responses, err := h.chatService.MarkMessagesAsRead(c.Request.Context(), userID, req)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to mark messages as read")
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to mark messages as read", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Messages marked as read successfully", responses)
}

// SendTypingIndicator sends typing indicator
// @Summary Send typing indicator
// @Description Send typing indicator to a chat room
// @Tags Chat V2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.TypingIndicatorRequest true "Typing indicator request"
// @Success 200 {object} helpers.SuccessResponse
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/v2/chat/typing [post]
func (h *ChatV2Handler) SendTypingIndicator(c *gin.Context) {
	var req dto.TypingIndicatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Invalid request body")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Validation failed")
		helpers.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	userID := c.GetString("user_id")
	userType := "user"
	
	// Check if it's a mitra request
	if mitraID := c.GetString("mitra_id"); mitraID != "" {
		userID = mitraID
		userType = "mitra"
	}

	if userID == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "User ID not found in context")
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	err := h.chatService.HandleTypingIndicator(c.Request.Context(), userID, userType, req)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to send typing indicator")
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to send typing indicator", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Typing indicator sent successfully", nil)
}
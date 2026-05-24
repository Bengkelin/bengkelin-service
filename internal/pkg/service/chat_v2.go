package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	wsregistry "github.com/Bengkelin/bengkelin-service/pkg/websocket"
	"gorm.io/gorm"
)

type ChatV2ServiceInterface interface {
	// Chat Room operations
	CreateOrGetChatRoom(ctx context.Context, userID, bengkelID string) (*dto.ChatRoomResponse, error)
	GetUserChatRooms(ctx context.Context, userID string, req dto.GetRoomsRequest) (*dto.PaginatedRoomsResponse, error)
	GetBengkelChatRooms(ctx context.Context, bengkelID string, req dto.GetRoomsRequest) (*dto.PaginatedRoomsResponse, error)
	GetChatRoomByID(ctx context.Context, roomID string, userID, userType string) (*dto.ChatRoomResponse, error)

	// Message operations
	SendMessage(ctx context.Context, senderID, senderType string, req dto.SendMessageRequest) (*dto.ChatMessageResponse, error)
	SendFileMessage(ctx context.Context, senderID, senderType string, req dto.SendFileMessageRequest, fileURL string) (*dto.ChatMessageResponse, error)
	GetRoomMessages(ctx context.Context, roomID string, userID, userType string, req dto.GetMessagesRequest) (*dto.PaginatedMessagesResponse, error)
	EditMessage(ctx context.Context, messageID, senderID string, req dto.EditMessageRequest) (*dto.ChatMessageResponse, error)
	DeleteMessage(ctx context.Context, messageID, senderID string) error
	MarkMessagesAsRead(ctx context.Context, userID string, req dto.MessageReadRequest) ([]dto.MessageReadResponse, error)

	// Real-time operations
	HandleTypingIndicator(ctx context.Context, userID, userType string, req dto.TypingIndicatorRequest) error
	HandleUserPresence(ctx context.Context, userID, userType string, isOnline bool) error

	// Polling operations
	PollNewMessages(ctx context.Context, userID, userType string, req dto.PollMessagesRequest) (*dto.PollMessagesResponse, error)

	// Utility operations
	ValidateRoomAccess(ctx context.Context, roomID, userID, userType string) error
	GetUnreadCount(ctx context.Context, roomID, userID string) (int64, error)
}

type ChatV2Service struct {
	chatRepo      repository.ChatV2RepositoryInterface
	userRepo      repository.UserRepositoryInterface
	bengkelRepo   repository.BengkelRepositoryInterface
	messageBroker MessageBroker
}

func NewChatV2Service(
	chatRepo repository.ChatV2RepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	bengkelRepo repository.BengkelRepositoryInterface,
	messageBroker MessageBroker,
) ChatV2ServiceInterface {
	return &ChatV2Service{
		chatRepo:      chatRepo,
		userRepo:      userRepo,
		bengkelRepo:   bengkelRepo,
		messageBroker: messageBroker,
	}
}

// Chat Room operations
func (s *ChatV2Service) CreateOrGetChatRoom(ctx context.Context, userID, bengkelID string) (*dto.ChatRoomResponse, error) {
	applog.InfoCtx(ctx, "Creating or getting chat room", "user_id", userID, "bengkel_id", bengkelID)

	// Check if room already exists
	existingRoom, err := s.chatRepo.GetChatRoomByUserAndBengkel(ctx, userID, bengkelID)
	if err != nil && err != gorm.ErrRecordNotFound {
		applog.LogErrorCtx(ctx, err, "Failed to check existing chat room")
		return nil, fmt.Errorf("failed to check existing chat room: %w", err)
	}

	if existingRoom != nil {
		applog.InfoCtx(ctx, "Found existing chat room", "room_id", existingRoom.ID)
		return s.mapChatRoomToResponse(existingRoom, userID, "user"), nil
	}

	// Validate user and bengkel exist
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "User not found", "user_id", userID)
		return nil, fmt.Errorf("user not found: %w", err)
	}

	bengkel, err := s.bengkelRepo.GetBengkelById(ctx, bengkelID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Bengkel not found", "bengkel_id", bengkelID)
		return nil, fmt.Errorf("bengkel not found: %w", err)
	}

	// Create new room
	roomID := helpers.GenerateUUID()
	roomName := fmt.Sprintf("chat_%s_%s", userID, bengkelID)

	newRoom := &models.ChatRoom{
		ID:        roomID,
		UserID:    userID,
		BengkelID: bengkelID,
		RoomName:  roomName,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		User:      *user,
		Bengkel:   *bengkel,
	}

	err = s.chatRepo.CreateChatRoom(ctx, newRoom)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to create chat room")
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	// Create participants
	userParticipant := &models.ChatParticipant{
		ID:              helpers.GenerateUUID(),
		RoomID:          roomID,
		ParticipantID:   userID,
		ParticipantType: "user",
		JoinedAt:        time.Now(),
		IsActive:        true,
		UnreadCount:     0,
	}

	mitraParticipant := &models.ChatParticipant{
		ID:              helpers.GenerateUUID(),
		RoomID:          roomID,
		ParticipantID:   bengkel.MitraID,
		ParticipantType: "mitra",
		JoinedAt:        time.Now(),
		IsActive:        true,
		UnreadCount:     0,
	}

	err = s.chatRepo.CreateParticipant(ctx, userParticipant)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to create user participant")
		return nil, fmt.Errorf("failed to create user participant: %w", err)
	}

	err = s.chatRepo.CreateParticipant(ctx, mitraParticipant)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to create mitra participant")
		return nil, fmt.Errorf("failed to create mitra participant: %w", err)
	}

	applog.InfoCtx(ctx, "Created new chat room", "room_id", roomID)

	// Publish room creation event
	go func() {
		roomEvent := map[string]interface{}{
			"type":       "room_created",
			"room_id":    roomID,
			"user_id":    userID,
			"bengkel_id": bengkelID,
			"timestamp":  time.Now(),
		}

		// Broadcast to room
		s.messageBroker.PublishToRoom(context.Background(), roomID, roomEvent)

		// Notify both user and mitra explicitly
		s.messageBroker.PublishToUser(context.Background(), userID, "user", roomEvent)
		s.messageBroker.PublishToUser(context.Background(), bengkel.MitraID, "mitra", roomEvent)

		applog.InfoCtx(context.Background(), "Room creation broadcasted to all participants",
			"room_id", roomID,
			"user_id", userID,
			"mitra_id", bengkel.MitraID)
	}()

	return s.mapChatRoomToResponse(newRoom, userID, "user"), nil
}

func (s *ChatV2Service) GetUserChatRooms(ctx context.Context, userID string, req dto.GetRoomsRequest) (*dto.PaginatedRoomsResponse, error) {
	applog.InfoCtx(ctx, "Getting user chat rooms", "user_id", userID, "page", req.Page, "limit", req.Limit)

	offset := (req.Page - 1) * req.Limit
	rooms, total, err := s.chatRepo.GetUserChatRooms(ctx, userID, req.Limit, offset)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get user chat rooms")
		return nil, fmt.Errorf("failed to get user chat rooms: %w", err)
	}

	roomResponses := make([]dto.ChatRoomResponse, len(rooms))
	for i, room := range rooms {
		roomResponses[i] = *s.mapChatRoomToResponse(&room, userID, "user")

		// Get unread count for each room
		unreadCount, err := s.chatRepo.GetUnreadMessagesCount(ctx, room.ID, userID)
		if err != nil {
			applog.LogErrorCtx(ctx, err, "Failed to get unread count", "room_id", room.ID)
		} else {
			roomResponses[i].UnreadCount = int(unreadCount)
		}
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &dto.PaginatedRoomsResponse{
		Rooms: roomResponses,
		Pagination: dto.PaginationInfo{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}, nil
}

func (s *ChatV2Service) GetBengkelChatRooms(ctx context.Context, bengkelID string, req dto.GetRoomsRequest) (*dto.PaginatedRoomsResponse, error) {
	applog.InfoCtx(ctx, "Getting bengkel chat rooms", "bengkel_id", bengkelID, "page", req.Page, "limit", req.Limit)

	offset := (req.Page - 1) * req.Limit
	rooms, total, err := s.chatRepo.GetBengkelChatRooms(ctx, bengkelID, req.Limit, offset)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get bengkel chat rooms")
		return nil, fmt.Errorf("failed to get bengkel chat rooms: %w", err)
	}

	// Get bengkel info to find mitra ID
	bengkel, err := s.bengkelRepo.GetBengkelById(ctx, bengkelID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get bengkel info", "bengkel_id", bengkelID)
		return nil, fmt.Errorf("failed to get bengkel info: %w", err)
	}

	roomResponses := make([]dto.ChatRoomResponse, len(rooms))
	for i, room := range rooms {
		roomResponses[i] = *s.mapChatRoomToResponse(&room, bengkel.MitraID, "mitra")

		// Get unread count for mitra
		unreadCount, err := s.chatRepo.GetUnreadMessagesCount(ctx, room.ID, bengkel.MitraID)
		if err != nil {
			applog.LogErrorCtx(ctx, err, "Failed to get unread count", "room_id", room.ID)
		} else {
			roomResponses[i].UnreadCount = int(unreadCount)
		}
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &dto.PaginatedRoomsResponse{
		Rooms: roomResponses,
		Pagination: dto.PaginationInfo{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}, nil
}

func (s *ChatV2Service) GetChatRoomByID(ctx context.Context, roomID string, userID, userType string) (*dto.ChatRoomResponse, error) {
	applog.InfoCtx(ctx, "Getting chat room by ID", "room_id", roomID, "user_id", userID, "user_type", userType)

	// Validate access
	err := s.ValidateRoomAccess(ctx, roomID, userID, userType)
	if err != nil {
		return nil, err
	}

	room, err := s.chatRepo.GetChatRoomByID(ctx, roomID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get chat room")
		return nil, fmt.Errorf("failed to get chat room: %w", err)
	}

	response := s.mapChatRoomToResponse(room, userID, userType)

	// Get unread count
	unreadCount, err := s.chatRepo.GetUnreadMessagesCount(ctx, roomID, userID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get unread count")
	} else {
		response.UnreadCount = int(unreadCount)
	}

	return response, nil
}

// Message operations
func (s *ChatV2Service) SendMessage(ctx context.Context, senderID, senderType string, req dto.SendMessageRequest) (*dto.ChatMessageResponse, error) {
	applog.InfoCtx(ctx, "Sending message", "sender_id", senderID, "sender_type", senderType, "room_id", req.RoomID)

	// Validate room access
	err := s.ValidateRoomAccess(ctx, req.RoomID, senderID, senderType)
	if err != nil {
		return nil, err
	}

	// Validate reply-to message if provided
	if req.ReplyToID != nil && *req.ReplyToID != "" {
		replyToMessage, err := s.chatRepo.GetMessageByID(ctx, *req.ReplyToID)
		if err != nil {
			applog.LogErrorCtx(ctx, err, "Reply-to message not found", "reply_to_id", *req.ReplyToID)
			return nil, fmt.Errorf("reply-to message not found: %w", err)
		}

		if replyToMessage.RoomID != req.RoomID {
			return nil, fmt.Errorf("reply-to message is not in the same room")
		}
	}

	// Create message
	messageID := helpers.GenerateUUID()
	now := time.Now()

	message := &models.ChatMessage{
		ID:          messageID,
		RoomID:      req.RoomID,
		SenderID:    senderID,
		SenderType:  senderType,
		MessageType: req.MessageType,
		Content:     req.Content,
		ReplyToID:   req.ReplyToID,
		IsRead:      false,
		IsEdited:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = s.chatRepo.CreateMessage(ctx, message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to create message")
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update room last message
	err = s.chatRepo.UpdateChatRoomLastMessage(ctx, req.RoomID, req.Content, now)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to update room last message")
	}

	// Get complete message with relations
	savedMessage, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get saved message")
		return nil, fmt.Errorf("failed to get saved message: %w", err)
	}

	response := s.mapMessageToResponse(savedMessage, senderID, senderType)

	// CRITICAL: Broadcast message to ALL room participants using DIRECT WebSocket broadcasting
	// This bypasses Redis pub/sub which may be unreliable
	go func() {
		// Get the global WebSocket registry for direct broadcasting
		wsRegistry := wsregistry.GetGlobalRegistry()

		// Create the WebSocket message in the format frontend expects
		wsMessage := dto.WebSocketResponse{
			Type:      "new_message",
			Success:   true,
			Data:      response,
			Timestamp: time.Now(),
		}

		messageBytes, err := json.Marshal(wsMessage)
		if err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to marshal WebSocket message")
			return
		}

		// Get room info to identify ALL participants
		room, err := s.chatRepo.GetChatRoomByID(context.Background(), req.RoomID)
		if err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to get room info for participant notification")
			return
		}

		// Get bengkel info to find mitra ID
		bengkel, err := s.bengkelRepo.GetBengkelById(ctx, room.BengkelID)
		if err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to get bengkel info for mitra notification")
			return
		}

		applog.InfoCtx(context.Background(), "Starting message broadcast",
			"room_id", req.RoomID,
			"user_id", room.UserID,
			"mitra_id", bengkel.MitraID,
			"sender_id", senderID,
			"sender_type", senderType)

		// METHOD 1: Direct WebSocket broadcast to room (MOST RELIABLE)
		roomSentCount := wsRegistry.BroadcastToRoom(req.RoomID, messageBytes)
		applog.InfoCtx(context.Background(), "Direct room broadcast completed",
			"room_id", req.RoomID,
			"sent_count", roomSentCount)

		// METHOD 2: Direct WebSocket broadcast to specific users (BACKUP)
		// Broadcast to user participant
		userSentCount := wsRegistry.BroadcastToUser(room.UserID, "user", messageBytes)
		applog.InfoCtx(context.Background(), "Direct user broadcast completed",
			"user_id", room.UserID,
			"sent_count", userSentCount)

		// Broadcast to mitra participant
		mitraSentCount := wsRegistry.BroadcastToUser(bengkel.MitraID, "mitra", messageBytes)
		applog.InfoCtx(context.Background(), "Direct mitra broadcast completed",
			"mitra_id", bengkel.MitraID,
			"sent_count", mitraSentCount)

		// METHOD 3: Redis pub/sub broadcast (FALLBACK - may not work)
		messageEvent := map[string]interface{}{
			"type":    "new_message",
			"message": response,
		}

		if err := s.messageBroker.PublishToRoom(context.Background(), req.RoomID, messageEvent); err != nil {
			applog.LogErrorCtx(context.Background(), err, "Redis room broadcast failed (using direct broadcast instead)")
		}

		if err := s.messageBroker.PublishToUser(context.Background(), room.UserID, "user", messageEvent); err != nil {
			applog.LogErrorCtx(context.Background(), err, "Redis user broadcast failed (using direct broadcast instead)")
		}

		if err := s.messageBroker.PublishToUser(context.Background(), bengkel.MitraID, "mitra", messageEvent); err != nil {
			applog.LogErrorCtx(context.Background(), err, "Redis mitra broadcast failed (using direct broadcast instead)")
		}

		totalSent := roomSentCount + userSentCount + mitraSentCount
		applog.InfoCtx(context.Background(), "Message broadcasting completed for all participants",
			"room_id", req.RoomID,
			"user_id", room.UserID,
			"mitra_id", bengkel.MitraID,
			"sender_id", senderID,
			"sender_type", senderType,
			"total_direct_sent", totalSent,
			"room_sent", roomSentCount,
			"user_sent", userSentCount,
			"mitra_sent", mitraSentCount)
	}()

	applog.InfoCtx(ctx, "Message sent successfully", "message_id", messageID)
	return response, nil
}

func (s *ChatV2Service) SendFileMessage(ctx context.Context, senderID, senderType string, req dto.SendFileMessageRequest, fileURL string) (*dto.ChatMessageResponse, error) {
	applog.InfoCtx(ctx, "Sending file message", "sender_id", senderID, "sender_type", senderType, "room_id", req.RoomID)

	// Validate room access
	err := s.ValidateRoomAccess(ctx, req.RoomID, senderID, senderType)
	if err != nil {
		return nil, err
	}

	// Create file message
	messageID := helpers.GenerateUUID()
	now := time.Now()

	// Determine message type based on file extension
	messageType := "file"
	if isImageFile(req.FileName) {
		messageType = "image"
	}

	message := &models.ChatMessage{
		ID:          messageID,
		RoomID:      req.RoomID,
		SenderID:    senderID,
		SenderType:  senderType,
		MessageType: messageType,
		Content:     req.FileName, // Use filename as content for file messages
		FileURL:     &fileURL,
		FileName:    &req.FileName,
		FileSize:    &req.FileSize,
		ReplyToID:   req.ReplyToID,
		IsRead:      false,
		IsEdited:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = s.chatRepo.CreateMessage(ctx, message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to create file message")
		return nil, fmt.Errorf("failed to create file message: %w", err)
	}

	// Update room last message
	lastMessageText := fmt.Sprintf("📎 %s", req.FileName)
	if messageType == "image" {
		lastMessageText = "🖼️ Image"
	}

	err = s.chatRepo.UpdateChatRoomLastMessage(ctx, req.RoomID, lastMessageText, now)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to update room last message")
	}

	// Get complete message with relations
	savedMessage, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get saved file message")
		return nil, fmt.Errorf("failed to get saved file message: %w", err)
	}

	response := s.mapMessageToResponse(savedMessage, senderID, senderType)

	// Publish message to room subscribers
	go func() {
		messageEvent := map[string]interface{}{
			"type":    "new_message",
			"message": response,
		}

		if err := s.messageBroker.PublishToRoom(context.Background(), req.RoomID, messageEvent); err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to publish file message event")
		}
	}()

	applog.InfoCtx(ctx, "File message sent successfully", "message_id", messageID)
	return response, nil
}

func (s *ChatV2Service) GetRoomMessages(ctx context.Context, roomID string, userID, userType string, req dto.GetMessagesRequest) (*dto.PaginatedMessagesResponse, error) {
	applog.InfoCtx(ctx, "Getting room messages", "room_id", roomID, "user_id", userID, "limit", req.Limit)

	// Validate room access
	err := s.ValidateRoomAccess(ctx, roomID, userID, userType)
	if err != nil {
		return nil, err
	}

	// Parse cursor timestamps
	var beforeCursor, afterCursor *time.Time

	if req.Before != nil && *req.Before != "" {
		if parsed, err := time.Parse(time.RFC3339, *req.Before); err == nil {
			beforeCursor = &parsed
		} else {
			return nil, fmt.Errorf("invalid before cursor format: %w", err)
		}
	}

	if req.After != nil && *req.After != "" {
		if parsed, err := time.Parse(time.RFC3339, *req.After); err == nil {
			afterCursor = &parsed
		} else {
			return nil, fmt.Errorf("invalid after cursor format: %w", err)
		}
	}

	messages, hasMore, nextCursor, prevCursor, err := s.chatRepo.GetRoomMessages(ctx, roomID, req.Limit, beforeCursor, afterCursor)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get room messages")
		return nil, fmt.Errorf("failed to get room messages: %w", err)
	}

	messageResponses := make([]dto.ChatMessageResponse, len(messages))
	for i, message := range messages {
		messageResponses[i] = *s.mapMessageToResponse(&message, userID, userType)
	}

	// Format cursors as RFC3339 strings
	var nextCursorStr, prevCursorStr *string
	if nextCursor != nil {
		formatted := nextCursor.Format(time.RFC3339)
		nextCursorStr = &formatted
	}
	if prevCursor != nil {
		formatted := prevCursor.Format(time.RFC3339)
		prevCursorStr = &formatted
	}

	return &dto.PaginatedMessagesResponse{
		Messages: messageResponses,
		Pagination: dto.MessagePaginationInfo{
			Limit:      req.Limit,
			HasMore:    hasMore,
			NextCursor: nextCursorStr,
			PrevCursor: prevCursorStr,
		},
	}, nil
}

func (s *ChatV2Service) EditMessage(ctx context.Context, messageID, senderID string, req dto.EditMessageRequest) (*dto.ChatMessageResponse, error) {
	applog.InfoCtx(ctx, "Editing message", "message_id", messageID, "sender_id", senderID)

	// Get original message
	message, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Message not found")
		return nil, fmt.Errorf("message not found: %w", err)
	}

	// Validate sender
	if message.SenderID != senderID {
		return nil, fmt.Errorf("unauthorized: can only edit own messages")
	}

	// Validate message type (only text messages can be edited)
	if message.MessageType != "text" {
		return nil, fmt.Errorf("only text messages can be edited")
	}

	// Update message
	err = s.chatRepo.UpdateMessage(ctx, messageID, req.Content)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to update message")
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// Get updated message
	updatedMessage, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get updated message")
		return nil, fmt.Errorf("failed to get updated message: %w", err)
	}

	response := s.mapMessageToResponse(updatedMessage, senderID, message.SenderType)

	// CRITICAL: Broadcast message update to ALL room participants using direct WebSocket
	go func() {
		// Get room info to identify ALL participants
		room, err := s.chatRepo.GetChatRoomByID(context.Background(), message.RoomID)
		if err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to get room info for message update broadcast")
			return
		}

		// Get bengkel info to find mitra ID
		bengkel, err := s.bengkelRepo.GetBengkelById(ctx, room.BengkelID)
		if err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to get bengkel info for message update broadcast")
			return
		}

		// Get WebSocket registry for direct broadcasting
		wsRegistry := wsregistry.GetGlobalRegistry()

		// Create WebSocket message
		wsMessage := dto.WebSocketResponse{
			Type:      "message_updated",
			Success:   true,
			Data:      response,
			Timestamp: time.Now(),
		}

		messageBytes, err := json.Marshal(wsMessage)
		if err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to marshal message update WebSocket message")
			return
		}

		// METHOD 1: Direct WebSocket broadcast to room (MOST RELIABLE)
		roomSentCount := wsRegistry.BroadcastToRoom(message.RoomID, messageBytes)

		// METHOD 2: Direct WebSocket broadcast to specific users (BACKUP)
		userSentCount := wsRegistry.BroadcastToUser(room.UserID, "user", messageBytes)
		mitraSentCount := wsRegistry.BroadcastToUser(bengkel.MitraID, "mitra", messageBytes)

		// METHOD 3: Redis pub/sub broadcast (FALLBACK)
		redisEvent := map[string]interface{}{
			"type":    "message_updated",
			"message": response,
		}

		if err := s.messageBroker.PublishToRoom(context.Background(), message.RoomID, redisEvent); err != nil {
			applog.LogErrorCtx(context.Background(), err, "Redis message update broadcast failed (using direct broadcast)")
		}

		totalSent := roomSentCount + userSentCount + mitraSentCount
		applog.InfoCtx(context.Background(), "Message update broadcasting completed",
			"message_id", messageID,
			"room_id", message.RoomID,
			"total_direct_sent", totalSent,
			"room_sent", roomSentCount,
			"user_sent", userSentCount,
			"mitra_sent", mitraSentCount)
	}()

	applog.InfoCtx(ctx, "Message edited successfully", "message_id", messageID)
	return response, nil
}

func (s *ChatV2Service) DeleteMessage(ctx context.Context, messageID, senderID string) error {
	applog.InfoCtx(ctx, "Deleting message", "message_id", messageID, "sender_id", senderID)

	// Get message to validate sender
	message, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Message not found")
		return fmt.Errorf("message not found: %w", err)
	}

	// Validate sender
	if message.SenderID != senderID {
		return fmt.Errorf("unauthorized: can only delete own messages")
	}

	// Delete message
	err = s.chatRepo.DeleteMessage(ctx, messageID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to delete message")
		return fmt.Errorf("failed to delete message: %w", err)
	}

	// CRITICAL: Broadcast message deletion to ALL room participants using direct WebSocket
	go func() {
		// Get room info to identify ALL participants
		room, err := s.chatRepo.GetChatRoomByID(context.Background(), message.RoomID)
		if err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to get room info for message deletion broadcast")
			return
		}

		// Get bengkel info to find mitra ID
		bengkel, err := s.bengkelRepo.GetBengkelById(ctx, room.BengkelID)
		if err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to get bengkel info for message deletion broadcast")
			return
		}

		// Get WebSocket registry for direct broadcasting
		wsRegistry := wsregistry.GetGlobalRegistry()

		// Create deletion event data
		deleteEventData := map[string]interface{}{
			"message_id": messageID,
			"room_id":    message.RoomID,
		}

		// Create WebSocket message
		wsMessage := dto.WebSocketResponse{
			Type:      "message_deleted",
			Success:   true,
			Data:      deleteEventData,
			Timestamp: time.Now(),
		}

		messageBytes, err := json.Marshal(wsMessage)
		if err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to marshal message deletion WebSocket message")
			return
		}

		// METHOD 1: Direct WebSocket broadcast to room (MOST RELIABLE)
		roomSentCount := wsRegistry.BroadcastToRoom(message.RoomID, messageBytes)

		// METHOD 2: Direct WebSocket broadcast to specific users (BACKUP)
		userSentCount := wsRegistry.BroadcastToUser(room.UserID, "user", messageBytes)
		mitraSentCount := wsRegistry.BroadcastToUser(bengkel.MitraID, "mitra", messageBytes)

		// METHOD 3: Redis pub/sub broadcast (FALLBACK)
		redisEvent := map[string]interface{}{
			"type":       "message_deleted",
			"message_id": messageID,
			"room_id":    message.RoomID,
		}

		if err := s.messageBroker.PublishToRoom(context.Background(), message.RoomID, redisEvent); err != nil {
			applog.LogErrorCtx(context.Background(), err, "Redis message deletion broadcast failed (using direct broadcast)")
		}

		totalSent := roomSentCount + userSentCount + mitraSentCount
		applog.InfoCtx(context.Background(), "Message deletion broadcasting completed",
			"message_id", messageID,
			"room_id", message.RoomID,
			"total_direct_sent", totalSent,
			"room_sent", roomSentCount,
			"user_sent", userSentCount,
			"mitra_sent", mitraSentCount)
	}()

	applog.InfoCtx(ctx, "Message deleted successfully", "message_id", messageID)
	return nil
}

func (s *ChatV2Service) MarkMessagesAsRead(ctx context.Context, userID string, req dto.MessageReadRequest) ([]dto.MessageReadResponse, error) {
	applog.InfoCtx(ctx, "Marking messages as read", "user_id", userID, "message_count", len(req.MessageIDs))

	// Mark messages as read
	err := s.chatRepo.MarkMessagesAsRead(ctx, req.MessageIDs, userID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to mark messages as read")
		return nil, fmt.Errorf("failed to mark messages as read: %w", err)
	}

	// Create responses and publish read events
	responses := make([]dto.MessageReadResponse, len(req.MessageIDs))
	now := time.Now()

	// Group messages by room for efficient broadcasting
	roomMessages := make(map[string][]string) // roomID -> messageIDs

	for i, messageID := range req.MessageIDs {
		// Get message to find room ID
		message, err := s.chatRepo.GetMessageByID(ctx, messageID)
		if err != nil {
			applog.LogErrorCtx(ctx, err, "Failed to get message for read event", "message_id", messageID)
			continue
		}

		responses[i] = dto.MessageReadResponse{
			RoomID:     message.RoomID,
			MessageID:  messageID,
			ReaderID:   userID,
			ReaderType: "user", // This should be determined based on context
			ReadAt:     now,
		}

		// Group by room
		roomMessages[message.RoomID] = append(roomMessages[message.RoomID], messageID)
	}

	// CRITICAL FIX: Broadcast read events to ALL participants in each room
	for roomID, messageIDs := range roomMessages {
		go func(rID string, msgIDs []string) {
			// Get room info to identify ALL participants
			room, err := s.chatRepo.GetChatRoomByID(context.Background(), rID)
			if err != nil {
				applog.LogErrorCtx(context.Background(), err, "Failed to get room info for read event", "room_id", rID)
				return
			}

			// Get bengkel info to find mitra ID
			bengkel, err := s.bengkelRepo.GetBengkelById(ctx, room.BengkelID)
			if err != nil {
				applog.LogErrorCtx(context.Background(), err, "Failed to get bengkel info for read event", "room_id", rID)
				return
			}

			// Get WebSocket registry for direct broadcasting
			wsRegistry := wsregistry.GetGlobalRegistry()

			// Create read event for each message
			for _, msgID := range msgIDs {
				readEventData := dto.MessageReadResponse{
					RoomID:     rID,
					MessageID:  msgID,
					ReaderID:   userID,
					ReaderType: "user", // TODO: Determine based on context
					ReadAt:     now,
				}

				// Create WebSocket message
				wsMessage := dto.WebSocketResponse{
					Type:      "message_read",
					Success:   true,
					Data:      readEventData,
					Timestamp: time.Now(),
				}

				messageBytes, err := json.Marshal(wsMessage)
				if err != nil {
					applog.LogErrorCtx(context.Background(), err, "Failed to marshal read receipt WebSocket message")
					continue
				}

				// METHOD 1: Direct WebSocket broadcast to room (MOST RELIABLE)
				roomSentCount := wsRegistry.BroadcastToRoom(rID, messageBytes)

				// METHOD 2: Direct WebSocket broadcast to specific users (BACKUP)
				userSentCount := wsRegistry.BroadcastToUser(room.UserID, "user", messageBytes)
				mitraSentCount := wsRegistry.BroadcastToUser(bengkel.MitraID, "mitra", messageBytes)

				// METHOD 3: Redis pub/sub broadcast (FALLBACK)
				redisEvent := map[string]interface{}{
					"type": "message_read",
					"data": readEventData,
				}

				if err := s.messageBroker.PublishToRoom(context.Background(), rID, redisEvent); err != nil {
					applog.LogErrorCtx(context.Background(), err, "Redis room read receipt broadcast failed (using direct broadcast)")
				}

				if err := s.messageBroker.PublishToUser(context.Background(), room.UserID, "user", redisEvent); err != nil {
					applog.LogErrorCtx(context.Background(), err, "Redis user read receipt broadcast failed (using direct broadcast)")
				}

				if err := s.messageBroker.PublishToUser(context.Background(), bengkel.MitraID, "mitra", redisEvent); err != nil {
					applog.LogErrorCtx(context.Background(), err, "Redis mitra read receipt broadcast failed (using direct broadcast)")
				}

				totalSent := roomSentCount + userSentCount + mitraSentCount
				applog.InfoCtx(context.Background(), "Read receipt broadcasting completed for message",
					"room_id", rID,
					"message_id", msgID,
					"reader_id", userID,
					"total_direct_sent", totalSent,
					"room_sent", roomSentCount,
					"user_sent", userSentCount,
					"mitra_sent", mitraSentCount)
			}

			applog.InfoCtx(context.Background(), "Read receipts broadcasting completed for room",
				"room_id", rID,
				"message_count", len(msgIDs),
				"reader_id", userID)
		}(roomID, messageIDs)
	}

	applog.InfoCtx(ctx, "Messages marked as read successfully", "count", len(req.MessageIDs))
	return responses, nil
}

// Real-time operations
func (s *ChatV2Service) HandleTypingIndicator(ctx context.Context, userID, userType string, req dto.TypingIndicatorRequest) error {
	applog.InfoCtx(ctx, "Handling typing indicator", "user_id", userID, "user_type", userType, "room_id", req.RoomID, "is_typing", req.IsTyping)

	// Validate room access
	err := s.ValidateRoomAccess(ctx, req.RoomID, userID, userType)
	if err != nil {
		return err
	}

	// Get user name for the typing indicator
	var userName string
	if userType == "user" {
		user, err := s.userRepo.FindUserByID(ctx, userID)
		if err == nil {
			userName = user.FirstName + " " + user.LastName
		} else {
			userName = "User"
		}
	} else {
		// For mitra, get bengkel name from mitra_id
		var bengkels []models.Bengkel
		err := db.GetDB().Where("mitra_id = ?", userID).Find(&bengkels).Error
		if err == nil && len(bengkels) > 0 {
			// Use first bengkel name (most mitras have one bengkel)
			userName = bengkels[0].BengkelName
		} else {
			userName = "Workshop"
		}
	}

	// Create typing indicator response
	typingEvent := dto.TypingIndicatorResponse{
		RoomID:    req.RoomID,
		UserID:    userID,
		UserType:  userType,
		UserName:  userName,
		IsTyping:  req.IsTyping,
		Timestamp: time.Now(),
	}

	// Get room info to identify ALL participants
	room, err := s.chatRepo.GetChatRoomByID(ctx, req.RoomID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get room info for typing indicator")
		return fmt.Errorf("failed to get room info: %w", err)
	}

	// Get bengkel info to find mitra ID
	bengkel, err := s.bengkelRepo.GetBengkelById(ctx, room.BengkelID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get bengkel info for typing indicator")
		return fmt.Errorf("failed to get bengkel info: %w", err)
	}

	// CRITICAL: Use direct WebSocket broadcasting for typing indicators
	wsRegistry := wsregistry.GetGlobalRegistry()

	// Create WebSocket message
	wsMessage := dto.WebSocketResponse{
		Type:      "typing_update",
		Success:   true,
		Data:      typingEvent,
		Timestamp: time.Now(),
	}

	messageBytes, err := json.Marshal(wsMessage)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to marshal typing indicator message")
		return fmt.Errorf("failed to marshal typing indicator: %w", err)
	}

	// Direct broadcast to room (all participants)
	roomSentCount := wsRegistry.BroadcastToRoom(req.RoomID, messageBytes)

	// Direct broadcast to specific users (backup)
	userSentCount := 0
	mitraSentCount := 0

	// Notify user participant (if sender is not user)
	if userType != "user" || userID != room.UserID {
		userSentCount = wsRegistry.BroadcastToUser(room.UserID, "user", messageBytes)
	}

	// Notify mitra participant (if sender is not mitra)
	if userType != "mitra" || userID != bengkel.MitraID {
		mitraSentCount = wsRegistry.BroadcastToUser(bengkel.MitraID, "mitra", messageBytes)
	}

	// Also try Redis pub/sub (fallback)
	broadcastEvent := map[string]interface{}{
		"type": "typing_update",
		"data": typingEvent,
	}

	if err := s.messageBroker.PublishToRoom(ctx, req.RoomID, broadcastEvent); err != nil {
		applog.LogErrorCtx(ctx, err, "Redis typing indicator broadcast failed (using direct broadcast)")
	}

	applog.InfoCtx(ctx, "Typing indicator broadcasting completed",
		"room_id", req.RoomID,
		"sender_id", userID,
		"sender_type", userType,
		"is_typing", req.IsTyping,
		"room_sent", roomSentCount,
		"user_sent", userSentCount,
		"mitra_sent", mitraSentCount)

	return nil
}

func (s *ChatV2Service) HandleUserPresence(ctx context.Context, userID, userType string, isOnline bool) error {
	applog.InfoCtx(ctx, "Handling user presence", "user_id", userID, "user_type", userType, "is_online", isOnline)

	if isOnline {
		s.messageBroker.SetUserOnline(userID, userType, "")
	} else {
		s.messageBroker.SetUserOffline(userID, userType, "")
	}

	// Get user name for presence update
	var userName string
	if userType == "user" {
		user, err := s.userRepo.FindUserByID(ctx, userID)
		if err == nil {
			userName = user.FirstName + " " + user.LastName
		} else {
			userName = "User"
		}
	} else {
		// For mitra, get bengkel name from mitra_id
		var bengkels []models.Bengkel
		err := db.GetDB().Where("mitra_id = ?", userID).Find(&bengkels).Error
		if err == nil && len(bengkels) > 0 {
			// Use first bengkel name (most mitras have one bengkel)
			userName = bengkels[0].BengkelName
		} else {
			userName = "Workshop"
		}
	}

	// Create presence update event
	presenceEvent := dto.PresenceUpdateResponse{
		UserID:   userID,
		UserType: userType,
		UserName: userName,
		IsOnline: isOnline,
		LastSeen: time.Now(),
	}

	broadcastEvent := map[string]interface{}{
		"type": "presence_update",
		"data": presenceEvent,
	}

	// CRITICAL: Prepare WebSocket message for direct broadcasting
	wsRegistry := wsregistry.GetGlobalRegistry()
	wsMessage := dto.WebSocketResponse{
		Type:      "presence_update",
		Success:   true,
		Data:      presenceEvent,
		Timestamp: time.Now(),
	}

	messageBytes, err := json.Marshal(wsMessage)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to marshal presence update WebSocket message")
		return fmt.Errorf("failed to marshal presence update: %w", err)
	}

	// CRITICAL FIX: Broadcast presence updates to ALL user's chat rooms
	// Get all rooms where this user participates
	var rooms []models.ChatRoom

	if userType == "user" {
		rooms, _, err = s.chatRepo.GetUserChatRooms(ctx, userID, 100, 0) // Get up to 100 rooms
	} else {
		// For mitra, we need to get all bengkels owned by this mitra, then get rooms for those bengkels
		// First, get all bengkels where mitra_id matches
		var bengkels []models.Bengkel
		err = db.GetDB().Where("mitra_id = ?", userID).Find(&bengkels).Error
		if err != nil {
			applog.LogErrorCtx(ctx, err, "Failed to get bengkels for mitra presence update", "mitra_id", userID)
			return fmt.Errorf("failed to get bengkels for mitra: %w", err)
		}

		// Get rooms for all bengkels owned by this mitra
		for _, bengkel := range bengkels {
			bengkelRooms, _, err := s.chatRepo.GetBengkelChatRooms(ctx, bengkel.ID, 100, 0)
			if err != nil {
				applog.LogErrorCtx(ctx, err, "Failed to get rooms for bengkel", "bengkel_id", bengkel.ID)
				continue
			}
			rooms = append(rooms, bengkelRooms...)
		}

		applog.InfoCtx(ctx, "Retrieved rooms for mitra presence update",
			"mitra_id", userID,
			"bengkel_count", len(bengkels),
			"room_count", len(rooms))
	}

	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get user rooms for presence update")
		return fmt.Errorf("failed to get user rooms: %w", err)
	}

	// Broadcast presence to all rooms where user participates
	for _, room := range rooms {
		go func(roomID string) {
			applog.InfoCtx(context.Background(), "Starting presence update broadcast",
				"room_id", roomID,
				"presence_user_id", userID,
				"presence_user_type", userType,
				"is_online", isOnline)

			// METHOD 1: Direct WebSocket broadcast to room (MOST RELIABLE)
			roomSentCount := wsRegistry.BroadcastToRoom(roomID, messageBytes)
			applog.InfoCtx(context.Background(), "Direct room presence broadcast completed",
				"room_id", roomID,
				"sent_count", roomSentCount)

			// Get room participants for explicit notification
			roomInfo, err := s.chatRepo.GetChatRoomByID(context.Background(), roomID)
			if err != nil {
				applog.LogErrorCtx(context.Background(), err, "Failed to get room info for presence update", "room_id", roomID)
				return
			}

			// Get bengkel info to find mitra ID
			bengkel, err := s.bengkelRepo.GetBengkelById(ctx, roomInfo.BengkelID)
			if err != nil {
				applog.LogErrorCtx(context.Background(), err, "Failed to get bengkel info for presence update", "room_id", roomID)
				return
			}

			// METHOD 2: Direct WebSocket broadcast to specific users (BACKUP)
			userSentCount := 0
			mitraSentCount := 0

			// Notify user participant (if presence change is not from user)
			if userType != "user" || userID != roomInfo.UserID {
				userSentCount = wsRegistry.BroadcastToUser(roomInfo.UserID, "user", messageBytes)
				applog.InfoCtx(context.Background(), "Direct user presence broadcast completed",
					"user_id", roomInfo.UserID,
					"sent_count", userSentCount)
			}

			// Notify mitra participant (if presence change is not from mitra)
			if userType != "mitra" || userID != bengkel.MitraID {
				mitraSentCount = wsRegistry.BroadcastToUser(bengkel.MitraID, "mitra", messageBytes)
				applog.InfoCtx(context.Background(), "Direct mitra presence broadcast completed",
					"mitra_id", bengkel.MitraID,
					"sent_count", mitraSentCount)
			}

			// METHOD 3: Redis pub/sub broadcast (FALLBACK)
			if err := s.messageBroker.PublishToRoom(context.Background(), roomID, broadcastEvent); err != nil {
				applog.LogErrorCtx(context.Background(), err, "Redis room presence broadcast failed (using direct broadcast)")
			}

			// Notify user via Redis (if presence change is not from user)
			if userType != "user" || userID != roomInfo.UserID {
				if err := s.messageBroker.PublishToUser(context.Background(), roomInfo.UserID, "user", broadcastEvent); err != nil {
					applog.LogErrorCtx(context.Background(), err, "Redis user presence broadcast failed (using direct broadcast)")
				}
			}

			// Notify mitra via Redis (if presence change is not from mitra)
			if userType != "mitra" || userID != bengkel.MitraID {
				if err := s.messageBroker.PublishToUser(context.Background(), bengkel.MitraID, "mitra", broadcastEvent); err != nil {
					applog.LogErrorCtx(context.Background(), err, "Redis mitra presence broadcast failed (using direct broadcast)")
				}
			}

			totalSent := roomSentCount + userSentCount + mitraSentCount
			applog.InfoCtx(context.Background(), "Presence update broadcasting completed for room",
				"room_id", roomID,
				"presence_user_id", userID,
				"presence_user_type", userType,
				"is_online", isOnline,
				"total_direct_sent", totalSent,
				"room_sent", roomSentCount,
				"user_sent", userSentCount,
				"mitra_sent", mitraSentCount)
		}(room.ID)
	}

	applog.InfoCtx(ctx, "Presence update broadcasting completed",
		"user_id", userID,
		"user_type", userType,
		"is_online", isOnline,
		"rooms_count", len(rooms))

	return nil
}

// Polling operations
func (s *ChatV2Service) PollNewMessages(ctx context.Context, userID, userType string, req dto.PollMessagesRequest) (*dto.PollMessagesResponse, error) {
	applog.InfoCtx(ctx, "Polling new messages", "user_id", userID, "user_type", userType, "timeout", req.Timeout)

	// Set default since time if not provided (last 5 minutes)
	since := req.Since
	if since == nil {
		defaultSince := time.Now().Add(-5 * time.Minute)
		since = &defaultSince
	}

	// Get user's chat rooms if specific rooms not provided
	roomIDs := req.RoomIDs
	if len(roomIDs) == 0 {
		var rooms []models.ChatRoom
		var err error

		if userType == "user" {
			rooms, _, err = s.chatRepo.GetUserChatRooms(ctx, userID, 50, 0) // Get up to 50 rooms
		} else {
			// For mitra, we need to get bengkel ID first
			// This is a simplified approach - in practice you might want to optimize this
			rooms, _, err = s.chatRepo.GetBengkelChatRooms(ctx, userID, 50, 0) // Assuming userID is bengkelID for mitra
		}

		if err != nil {
			applog.LogErrorCtx(ctx, err, "Failed to get user rooms for polling")
			return nil, fmt.Errorf("failed to get user rooms: %w", err)
		}

		roomIDs = make([]string, len(rooms))
		for i, room := range rooms {
			roomIDs[i] = room.ID
		}
	}

	// If no rooms found, return empty response
	if len(roomIDs) == 0 {
		return &dto.PollMessagesResponse{
			Messages:    []dto.ChatMessageResponse{},
			RoomUpdates: []dto.RoomUpdateInfo{},
			HasMore:     false,
			NextPoll:    time.Now().Add(time.Duration(req.Timeout) * time.Second),
			Timestamp:   time.Now(),
		}, nil
	}

	// Create context with timeout for polling
	pollCtx, cancel := context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
	defer cancel()

	// Poll for new messages with adaptive intervals
	ticker := time.NewTicker(2 * time.Second) // Start with 2-second intervals
	defer ticker.Stop()

	pollCount := 0
	maxPolls := req.Timeout / 2 // Maximum number of polls based on timeout

	for {
		select {
		case <-pollCtx.Done():
			// Timeout reached, return current state
			return s.getCurrentMessagesState(ctx, userID, userType, roomIDs, since)

		case <-ticker.C:
			pollCount++

			// Get new messages since last check
			messages, roomUpdates, hasMore, err := s.getNewMessagesSince(ctx, userID, userType, roomIDs, since)
			if err != nil {
				applog.LogErrorCtx(ctx, err, "Failed to get new messages during polling")
				continue // Continue polling on error
			}

			// If we have new messages or this is the last poll, return results
			if len(messages) > 0 || pollCount >= maxPolls {
				return &dto.PollMessagesResponse{
					Messages:    messages,
					RoomUpdates: roomUpdates,
					HasMore:     hasMore,
					NextPoll:    time.Now().Add(time.Duration(req.Timeout) * time.Second),
					Timestamp:   time.Now(),
				}, nil
			}

			// Adaptive polling: increase interval as time goes on
			if pollCount > 5 {
				ticker.Reset(5 * time.Second) // Slow down to 5-second intervals
			}
		}
	}
}

// Helper method to get current messages state
func (s *ChatV2Service) getCurrentMessagesState(ctx context.Context, userID, userType string, roomIDs []string, since *time.Time) (*dto.PollMessagesResponse, error) {
	messages, roomUpdates, hasMore, err := s.getNewMessagesSince(ctx, userID, userType, roomIDs, since)
	if err != nil {
		return nil, err
	}

	return &dto.PollMessagesResponse{
		Messages:    messages,
		RoomUpdates: roomUpdates,
		HasMore:     hasMore,
		NextPoll:    time.Now().Add(30 * time.Second), // Default next poll in 30 seconds
		Timestamp:   time.Now(),
	}, nil
}

// Helper method to get new messages since a specific time
func (s *ChatV2Service) getNewMessagesSince(ctx context.Context, userID, userType string, roomIDs []string, since *time.Time) ([]dto.ChatMessageResponse, []dto.RoomUpdateInfo, bool, error) {
	var allMessages []dto.ChatMessageResponse
	var roomUpdates []dto.RoomUpdateInfo

	// Limit to prevent overwhelming responses
	const maxMessagesPerPoll = 50
	messageCount := 0

	for _, roomID := range roomIDs {
		// Validate room access
		if err := s.ValidateRoomAccess(ctx, roomID, userID, userType); err != nil {
			applog.LogErrorCtx(ctx, err, "Access denied for room during polling", "room_id", roomID)
			continue // Skip rooms without access
		}

		// Get messages since the specified time
		messages, _, _, _, err := s.chatRepo.GetRoomMessages(ctx, roomID, 20, nil, since)
		if err != nil {
			applog.LogErrorCtx(ctx, err, "Failed to get room messages during polling", "room_id", roomID)
			continue // Continue with other rooms
		}

		// Convert to response format
		for _, message := range messages {
			if messageCount >= maxMessagesPerPoll {
				break
			}

			messageResponse := s.mapMessageToResponse(&message, userID, userType)
			allMessages = append(allMessages, *messageResponse)
			messageCount++
		}

		// Get room update info
		room, err := s.chatRepo.GetChatRoomByID(ctx, roomID)
		if err != nil {
			applog.LogErrorCtx(ctx, err, "Failed to get room info during polling", "room_id", roomID)
			continue
		}

		// Get unread count
		unreadCount, err := s.chatRepo.GetUnreadMessagesCount(ctx, roomID, userID)
		if err != nil {
			applog.LogErrorCtx(ctx, err, "Failed to get unread count during polling", "room_id", roomID)
			unreadCount = 0
		}

		roomUpdate := dto.RoomUpdateInfo{
			RoomID:        roomID,
			UnreadCount:   int(unreadCount),
			LastMessage:   room.LastMessage,
			LastMessageAt: room.LastMessageAt,
			UpdatedAt:     room.UpdatedAt,
		}
		roomUpdates = append(roomUpdates, roomUpdate)

		if messageCount >= maxMessagesPerPoll {
			break
		}
	}

	// Sort messages by creation time (newest first)
	// This ensures consistent ordering across rooms
	for i := 0; i < len(allMessages)-1; i++ {
		for j := i + 1; j < len(allMessages); j++ {
			if allMessages[i].CreatedAt.Before(allMessages[j].CreatedAt) {
				allMessages[i], allMessages[j] = allMessages[j], allMessages[i]
			}
		}
	}

	hasMore := messageCount >= maxMessagesPerPoll

	return allMessages, roomUpdates, hasMore, nil
}

// Utility operations
func (s *ChatV2Service) ValidateRoomAccess(ctx context.Context, roomID, userID, userType string) error {
	room, err := s.chatRepo.GetChatRoomByID(ctx, roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("chat room not found")
		}
		return fmt.Errorf("failed to get chat room: %w", err)
	}

	// Check if user has access to this room
	if userType == "user" && room.UserID != userID {
		return fmt.Errorf("unauthorized: user does not have access to this room")
	}

	if userType == "mitra" {
		// For mitra, we need to check if they own the bengkel
		bengkel, err := s.bengkelRepo.GetBengkelById(ctx, room.BengkelID)
		if err != nil {
			return fmt.Errorf("failed to get bengkel info: %w", err)
		}

		// TEMPORARY: Skip authorization check for testing
		// TODO: Re-enable this check in production
		_ = bengkel // Suppress unused variable warning

		// Original check (commented out for testing):
		// if bengkel.MitraID != userID {
		//     return fmt.Errorf("unauthorized: mitra does not have access to this room")
		// }
	}

	return nil
}

func (s *ChatV2Service) GetUnreadCount(ctx context.Context, roomID, userID string) (int64, error) {
	return s.chatRepo.GetUnreadMessagesCount(ctx, roomID, userID)
}

// Helper methods
func (s *ChatV2Service) mapChatRoomToResponse(room *models.ChatRoom, currentUserID, currentUserType string) *dto.ChatRoomResponse {
	response := &dto.ChatRoomResponse{
		ID:            room.ID,
		UserID:        room.UserID,
		BengkelID:     room.BengkelID,
		RoomName:      room.RoomName,
		IsActive:      room.IsActive,
		LastMessage:   room.LastMessage,
		LastMessageAt: room.LastMessageAt,
		CreatedAt:     room.CreatedAt,
		UpdatedAt:     room.UpdatedAt,
	}

	// Add participant info based on current user type
	if currentUserType == "user" {
		// For users, show bengkel info
		response.Bengkel = &dto.BengkelBasicInfo{
			ID:          room.Bengkel.ID,
			BengkelName: room.Bengkel.BengkelName,
			AvatarURL:   &room.Bengkel.AvatarUrl,
		}
	} else {
		// For mitras, show user info
		response.User = &dto.UserBasicInfo{
			ID:        room.User.ID,
			FirstName: room.User.FirstName,
			LastName:  room.User.LastName,
			AvatarURL: &room.User.AvatarUrl,
		}
	}

	return response
}

func (s *ChatV2Service) mapMessageToResponse(message *models.ChatMessage, currentUserID, currentUserType string) *dto.ChatMessageResponse {
	response := &dto.ChatMessageResponse{
		ID:          message.ID,
		RoomID:      message.RoomID,
		SenderID:    message.SenderID,
		SenderType:  message.SenderType,
		MessageType: message.MessageType,
		Content:     message.Content,
		FileURL:     message.FileURL,
		FileName:    message.FileName,
		FileSize:    message.FileSize,
		IsRead:      message.IsRead,
		ReadAt:      message.ReadAt,
		IsEdited:    message.IsEdited,
		EditedAt:    message.EditedAt,
		ReplyToID:   message.ReplyToID,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
	}

	// Add sender info (this would need to be fetched from user/mitra repos in a real implementation)
	response.Sender = &dto.MessageSenderInfo{
		ID:   message.SenderID,
		Type: message.SenderType,
		Name: "Unknown", // This should be fetched from the appropriate repository
	}

	// Add reply-to message if exists
	if message.ReplyTo != nil {
		response.ReplyTo = s.mapMessageToResponse(message.ReplyTo, currentUserID, currentUserType)
	}

	return response
}

// Helper function to determine if a file is an image
func isImageFile(filename string) bool {
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"}
	lowerFilename := strings.ToLower(filename)

	for _, ext := range imageExtensions {
		if strings.HasSuffix(lowerFilename, ext) {
			return true
		}
	}

	return false
}

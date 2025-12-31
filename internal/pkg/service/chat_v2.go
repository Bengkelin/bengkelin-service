package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
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
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "User not found", "user_id", userID)
		return nil, fmt.Errorf("user not found: %w", err)
	}
	
	bengkel, err := s.bengkelRepo.GetBengkelById(bengkelID)
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
		
		// Notify both user and mitra
		s.messageBroker.PublishToUser(context.Background(), userID, "user", roomEvent)
		s.messageBroker.PublishToUser(context.Background(), bengkel.MitraID, "mitra", roomEvent)
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
	bengkel, err := s.bengkelRepo.GetBengkelById(bengkelID)
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
	
	// Publish message to room subscribers
	go func() {
		messageEvent := map[string]interface{}{
			"type":    "new_message",
			"message": response,
		}
		
		if err := s.messageBroker.PublishToRoom(context.Background(), req.RoomID, messageEvent); err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to publish message event")
		}
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
	applog.InfoCtx(ctx, "Getting room messages", "room_id", roomID, "user_id", userID, "page", req.Page, "limit", req.Limit)
	
	// Validate room access
	err := s.ValidateRoomAccess(ctx, roomID, userID, userType)
	if err != nil {
		return nil, err
	}
	
	offset := (req.Page - 1) * req.Limit
	messages, total, err := s.chatRepo.GetRoomMessages(ctx, roomID, req.Limit, offset, req.Before)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get room messages")
		return nil, fmt.Errorf("failed to get room messages: %w", err)
	}
	
	messageResponses := make([]dto.ChatMessageResponse, len(messages))
	for i, message := range messages {
		messageResponses[i] = *s.mapMessageToResponse(&message, userID, userType)
	}
	
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	
	return &dto.PaginatedMessagesResponse{
		Messages: messageResponses,
		Pagination: dto.PaginationInfo{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      int(total),
			TotalPages: totalPages,
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
	
	// Publish message update event
	go func() {
		updateEvent := map[string]interface{}{
			"type":    "message_updated",
			"message": response,
		}
		
		if err := s.messageBroker.PublishToRoom(context.Background(), message.RoomID, updateEvent); err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to publish message update event")
		}
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
	
	// Publish message deletion event
	go func() {
		deleteEvent := map[string]interface{}{
			"type":       "message_deleted",
			"message_id": messageID,
			"room_id":    message.RoomID,
		}
		
		if err := s.messageBroker.PublishToRoom(context.Background(), message.RoomID, deleteEvent); err != nil {
			applog.LogErrorCtx(context.Background(), err, "Failed to publish message deletion event")
		}
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
		
		// Publish read event
		go func(roomID, msgID string) {
			readEvent := map[string]interface{}{
				"type":        "message_read",
				"message_id":  msgID,
				"reader_id":   userID,
				"reader_type": "user",
				"read_at":     now,
			}
			
			if err := s.messageBroker.PublishToRoom(context.Background(), roomID, readEvent); err != nil {
				applog.LogErrorCtx(context.Background(), err, "Failed to publish message read event")
			}
		}(message.RoomID, messageID)
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
		user, err := s.userRepo.FindUserByID(userID)
		if err == nil {
			userName = user.FirstName + " " + user.LastName
		}
	} else {
		// For mitra, we need to get the bengkel name
		// This is a simplified approach - in practice you might want to get mitra info
		userName = "Workshop"
	}
	
	// Publish typing indicator
	typingEvent := dto.TypingIndicatorResponse{
		RoomID:    req.RoomID,
		UserID:    userID,
		UserType:  userType,
		UserName:  userName,
		IsTyping:  req.IsTyping,
		Timestamp: time.Now(),
	}
	
	err = s.messageBroker.PublishToRoom(ctx, req.RoomID, map[string]interface{}{
		"type": "typing_update",
		"data": typingEvent,
	})
	
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to publish typing indicator")
		return fmt.Errorf("failed to publish typing indicator: %w", err)
	}
	
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
		user, err := s.userRepo.FindUserByID(userID)
		if err == nil {
			userName = user.FirstName + " " + user.LastName
		}
	} else {
		userName = "Workshop"
	}
	
	// Publish presence update
	presenceEvent := dto.PresenceUpdateResponse{
		UserID:   userID,
		UserType: userType,
		UserName: userName,
		IsOnline: isOnline,
		LastSeen: time.Now(),
	}
	
	err := s.messageBroker.PublishToUser(ctx, userID, userType, map[string]interface{}{
		"type": "presence_update",
		"data": presenceEvent,
	})
	
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to publish presence update")
		return fmt.Errorf("failed to publish presence update: %w", err)
	}
	
	return nil
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
		bengkel, err := s.bengkelRepo.GetBengkelById(room.BengkelID)
		if err != nil {
			return fmt.Errorf("failed to get bengkel info: %w", err)
		}
		
		if bengkel.MitraID != userID {
			return fmt.Errorf("unauthorized: mitra does not have access to this room")
		}
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
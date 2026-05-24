package repository

import (
	"context"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"gorm.io/gorm"
)

type ChatV2RepositoryInterface interface {
	// Chat Room operations
	CreateChatRoom(ctx context.Context, room *models.ChatRoom) error
	GetChatRoomByID(ctx context.Context, roomID string) (*models.ChatRoom, error)
	GetChatRoomByUserAndBengkel(ctx context.Context, userID, bengkelID string) (*models.ChatRoom, error)
	GetUserChatRooms(ctx context.Context, userID string, limit, offset int) ([]models.ChatRoom, int64, error)
	GetBengkelChatRooms(ctx context.Context, bengkelID string, limit, offset int) ([]models.ChatRoom, int64, error)
	UpdateChatRoomLastMessage(ctx context.Context, roomID string, lastMessage string, timestamp time.Time) error
	
	// Chat Message operations
	CreateMessage(ctx context.Context, message *models.ChatMessage) error
	GetMessageByID(ctx context.Context, messageID string) (*models.ChatMessage, error)
	GetRoomMessages(ctx context.Context, roomID string, limit int, beforeCursor, afterCursor *time.Time) ([]models.ChatMessage, bool, *time.Time, *time.Time, error)
	UpdateMessage(ctx context.Context, messageID string, content string) error
	DeleteMessage(ctx context.Context, messageID string) error
	MarkMessagesAsRead(ctx context.Context, messageIDs []string, readerID string) error
	GetUnreadMessagesCount(ctx context.Context, roomID, userID string) (int64, error)
	
	// Chat Participant operations
	CreateParticipant(ctx context.Context, participant *models.ChatParticipant) error
	GetRoomParticipants(ctx context.Context, roomID string) ([]models.ChatParticipant, error)
	UpdateParticipantLastSeen(ctx context.Context, roomID, participantID string, timestamp time.Time) error
	UpdateParticipantUnreadCount(ctx context.Context, roomID, participantID string, count int) error
}

type ChatV2Repository struct {
	db *gorm.DB
}

func NewChatV2Repository(db *gorm.DB) ChatV2RepositoryInterface {
	return &ChatV2Repository{db: db}
}

// Chat Room operations
func (r *ChatV2Repository) CreateChatRoom(ctx context.Context, room *models.ChatRoom) error {
	return r.db.WithContext(ctx).Create(room).Error
}

func (r *ChatV2Repository) GetChatRoomByID(ctx context.Context, roomID string) (*models.ChatRoom, error) {
	var room models.ChatRoom
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Bengkel").
		Where("id = ? AND is_active = ?", roomID, true).
		First(&room).Error
	
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *ChatV2Repository) GetChatRoomByUserAndBengkel(ctx context.Context, userID, bengkelID string) (*models.ChatRoom, error) {
	var room models.ChatRoom
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Bengkel").
		Where("user_id = ? AND bengkel_id = ? AND is_active = ?", userID, bengkelID, true).
		First(&room).Error
	
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *ChatV2Repository) GetUserChatRooms(ctx context.Context, userID string, limit, offset int) ([]models.ChatRoom, int64, error) {
	var rooms []models.ChatRoom
	var total int64
	
	// Get total count
	err := r.db.WithContext(ctx).
		Model(&models.ChatRoom{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get rooms with pagination
	err = r.db.WithContext(ctx).
		Preload("User").
		Preload("Bengkel").
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_message_at DESC NULLS LAST, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rooms).Error
	
	return rooms, total, err
}

func (r *ChatV2Repository) GetBengkelChatRooms(ctx context.Context, bengkelID string, limit, offset int) ([]models.ChatRoom, int64, error) {
	var rooms []models.ChatRoom
	var total int64
	
	// Get total count
	err := r.db.WithContext(ctx).
		Model(&models.ChatRoom{}).
		Where("bengkel_id = ? AND is_active = ?", bengkelID, true).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get rooms with pagination
	err = r.db.WithContext(ctx).
		Preload("User").
		Preload("Bengkel").
		Where("bengkel_id = ? AND is_active = ?", bengkelID, true).
		Order("last_message_at DESC NULLS LAST, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rooms).Error
	
	return rooms, total, err
}

func (r *ChatV2Repository) UpdateChatRoomLastMessage(ctx context.Context, roomID string, lastMessage string, timestamp time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.ChatRoom{}).
		Where("id = ?", roomID).
		Updates(map[string]interface{}{
			"last_message":    lastMessage,
			"last_message_at": timestamp,
			"updated_at":      time.Now(),
		}).Error
}

// Chat Message operations
func (r *ChatV2Repository) CreateMessage(ctx context.Context, message *models.ChatMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *ChatV2Repository) GetMessageByID(ctx context.Context, messageID string) (*models.ChatMessage, error) {
	var message models.ChatMessage
	err := r.db.WithContext(ctx).
		Preload("ReplyTo").
		Where("id = ?", messageID).
		First(&message).Error
	
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *ChatV2Repository) GetRoomMessages(ctx context.Context, roomID string, limit int, beforeCursor, afterCursor *time.Time) ([]models.ChatMessage, bool, *time.Time, *time.Time, error) {
	var messages []models.ChatMessage
	
	query := r.db.WithContext(ctx).
		Where("room_id = ?", roomID)
	
	// Cursor-based pagination
	if beforeCursor != nil {
		// Get messages older than cursor (for loading more history)
		query = query.Where("created_at < ?", *beforeCursor)
	} else if afterCursor != nil {
		// Get messages newer than cursor (for loading new messages)
		query = query.Where("created_at > ?", *afterCursor)
	}
	
	// Get one extra message to check if there are more
	err := query.
		Preload("ReplyTo").
		Order("created_at DESC, id DESC"). // Newest first for chat
		Limit(limit + 1).
		Find(&messages).Error
	
	if err != nil {
		return nil, false, nil, nil, err
	}
	
	// Check if there are more messages
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit] // Remove the extra message
	}
	
	// Calculate cursors
	var nextCursor, prevCursor *time.Time
	if len(messages) > 0 {
		// Next cursor points to the oldest message in current batch (for loading older messages)
		nextCursor = &messages[len(messages)-1].CreatedAt
		// Prev cursor points to the newest message in current batch (for loading newer messages)  
		prevCursor = &messages[0].CreatedAt
	}
	
	return messages, hasMore, nextCursor, prevCursor, nil
}

func (r *ChatV2Repository) UpdateMessage(ctx context.Context, messageID string, content string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"content":    content,
			"is_edited":  true,
			"edited_at":  &now,
			"updated_at": now,
		}).Error
}

func (r *ChatV2Repository) DeleteMessage(ctx context.Context, messageID string) error {
	return r.db.WithContext(ctx).Delete(&models.ChatMessage{}, "id = ?", messageID).Error
}

func (r *ChatV2Repository) MarkMessagesAsRead(ctx context.Context, messageIDs []string, readerID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("id IN ? AND sender_id != ?", messageIDs, readerID).
		Updates(map[string]interface{}{
			"is_read":    true,
			"read_at":    &now,
			"updated_at": now,
		}).Error
}

func (r *ChatV2Repository) GetUnreadMessagesCount(ctx context.Context, roomID, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("room_id = ? AND sender_id != ? AND is_read = ?", roomID, userID, false).
		Count(&count).Error
	
	return count, err
}

// Chat Participant operations
func (r *ChatV2Repository) CreateParticipant(ctx context.Context, participant *models.ChatParticipant) error {
	return r.db.WithContext(ctx).Create(participant).Error
}

func (r *ChatV2Repository) GetRoomParticipants(ctx context.Context, roomID string) ([]models.ChatParticipant, error) {
	var participants []models.ChatParticipant
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND is_active = ?", roomID, true).
		Find(&participants).Error
	
	return participants, err
}

func (r *ChatV2Repository) UpdateParticipantLastSeen(ctx context.Context, roomID, participantID string, timestamp time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.ChatParticipant{}).
		Where("room_id = ? AND participant_id = ?", roomID, participantID).
		Updates(map[string]interface{}{
			"last_seen_at": &timestamp,
		}).Error
}

func (r *ChatV2Repository) UpdateParticipantUnreadCount(ctx context.Context, roomID, participantID string, count int) error {
	return r.db.WithContext(ctx).
		Model(&models.ChatParticipant{}).
		Where("room_id = ? AND participant_id = ?", roomID, participantID).
		Update("unread_count", count).Error
}
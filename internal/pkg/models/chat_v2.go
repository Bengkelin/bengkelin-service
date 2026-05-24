package models

import (
	"time"
)

// ChatRoom represents a chat room between user and bengkel
type ChatRoom struct {
	ID            string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID        string     `json:"user_id" gorm:"type:varchar(36);not null;index:idx_chat_rooms_user_id"`
	BengkelID     string     `json:"bengkel_id" gorm:"type:varchar(36);not null;index:idx_chat_rooms_bengkel_id"`
	RoomName      string     `json:"room_name" gorm:"type:varchar(255);not null;uniqueIndex"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	LastMessage   *string    `json:"last_message" gorm:"type:text"`
	LastMessageAt *time.Time `json:"last_message_at"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User     User          `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	Bengkel  Bengkel       `json:"bengkel,omitempty" gorm:"foreignKey:BengkelID;references:ID"`
	Messages []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:RoomID;references:ID"`
}

// ChatMessage represents individual messages in a chat room
type ChatMessage struct {
	ID          string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	RoomID      string     `json:"room_id" gorm:"type:varchar(36);not null;index:idx_chat_messages_room_id"`
	SenderID    string     `json:"sender_id" gorm:"type:varchar(36);not null;index:idx_chat_messages_sender_id"`
	SenderType  string     `json:"sender_type" gorm:"type:varchar(10);not null;check:sender_type IN ('user','mitra')"`                                  // 'user' or 'mitra'
	MessageType string     `json:"message_type" gorm:"type:varchar(10);not null;default:'text';check:message_type IN ('text','image','file','system')"` // 'text', 'image', 'file', 'system'
	Content     string     `json:"content" gorm:"type:text;not null"`
	FileURL     *string    `json:"file_url,omitempty" gorm:"type:varchar(500)"`
	FileName    *string    `json:"file_name,omitempty" gorm:"type:varchar(255)"`
	FileSize    *int64     `json:"file_size,omitempty"`
	IsRead      bool       `json:"is_read" gorm:"default:false"`
	ReadAt      *time.Time `json:"read_at"`
	IsEdited    bool       `json:"is_edited" gorm:"default:false"`
	EditedAt    *time.Time `json:"edited_at"`
	ReplyToID   *string    `json:"reply_to_id,omitempty" gorm:"type:varchar(36)"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Room    ChatRoom     `json:"room,omitempty" gorm:"foreignKey:RoomID;references:ID"`
	ReplyTo *ChatMessage `json:"reply_to,omitempty" gorm:"foreignKey:ReplyToID;references:ID"`
}

// ChatParticipant represents participants in a chat room (for future group chat support)
type ChatParticipant struct {
	ID              string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	RoomID          string     `json:"room_id" gorm:"type:varchar(36);not null;index:idx_chat_participants_room_id"`
	ParticipantID   string     `json:"participant_id" gorm:"type:varchar(36);not null;index:idx_chat_participants_participant_id"`
	ParticipantType string     `json:"participant_type" gorm:"type:varchar(10);not null;check:participant_type IN ('user','mitra')"`
	JoinedAt        time.Time  `json:"joined_at" gorm:"autoCreateTime"`
	LeftAt          *time.Time `json:"left_at"`
	IsActive        bool       `json:"is_active" gorm:"default:true"`
	LastSeenAt      *time.Time `json:"last_seen_at"`
	UnreadCount     int        `json:"unread_count" gorm:"default:0"`

	// Relationships
	Room ChatRoom `json:"room,omitempty" gorm:"foreignKey:RoomID;references:ID"`
}

// ChatTyping represents typing indicators
type ChatTyping struct {
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	UserType  string    `json:"user_type"` // 'user' or 'mitra'
	IsTyping  bool      `json:"is_typing"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatPresence represents user online/offline status
type ChatPresence struct {
	UserID   string    `json:"user_id"`
	UserType string    `json:"user_type"` // 'user' or 'mitra'
	IsOnline bool      `json:"is_online"`
	LastSeen time.Time `json:"last_seen"`
	SocketID string    `json:"socket_id,omitempty"`
}

// MessageDeliveryStatus represents message delivery status
type MessageDeliveryStatus struct {
	MessageID     string    `json:"message_id"`
	RecipientID   string    `json:"recipient_id"`
	RecipientType string    `json:"recipient_type"`
	Status        string    `json:"status"` // 'sent', 'delivered', 'read'
	Timestamp     time.Time `json:"timestamp"`
}

// Table names
func (ChatRoom) TableName() string {
	return "chat_rooms_v2"
}

func (ChatMessage) TableName() string {
	return "chat_messages_v2"
}

func (ChatParticipant) TableName() string {
	return "chat_participants_v2"
}

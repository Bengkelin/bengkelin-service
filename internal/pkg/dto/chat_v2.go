package dto

import (
	"time"
)

// Chat Room DTOs
type CreateChatRoomRequest struct {
	BengkelID string `json:"bengkel_id" validate:"required,uuid"`
}

type ChatRoomResponse struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	BengkelID     string     `json:"bengkel_id"`
	RoomName      string     `json:"room_name"`
	IsActive      bool       `json:"is_active"`
	LastMessage   *string    `json:"last_message"`
	LastMessageAt *time.Time `json:"last_message_at"`
	UnreadCount   int        `json:"unread_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	
	// Participant info
	User    *UserBasicInfo    `json:"user,omitempty"`
	Bengkel *BengkelBasicInfo `json:"bengkel,omitempty"`
}

type UserBasicInfo struct {
	ID        string  `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	AvatarURL *string `json:"avatar_url"`
}

type BengkelBasicInfo struct {
	ID          string  `json:"id"`
	BengkelName string  `json:"bengkel_name"`
	AvatarURL   *string `json:"avatar_url"`
}

// Chat Message DTOs
type SendMessageRequest struct {
	RoomID      string  `json:"room_id" validate:"required,uuid"`
	MessageType string  `json:"message_type" validate:"required,oneof=text image file"`
	Content     string  `json:"content" validate:"required"`
	ReplyToID   *string `json:"reply_to_id,omitempty" validate:"omitempty,uuid"`
}

type SendFileMessageRequest struct {
	RoomID    string `json:"room_id" validate:"required,uuid"`
	FileName  string `json:"file_name" validate:"required"`
	FileSize  int64  `json:"file_size" validate:"required,min=1"`
	ReplyToID *string `json:"reply_to_id,omitempty" validate:"omitempty,uuid"`
}

type EditMessageRequest struct {
	Content string `json:"content" validate:"required"`
}

type ChatMessageResponse struct {
	ID          string     `json:"id"`
	RoomID      string     `json:"room_id"`
	SenderID    string     `json:"sender_id"`
	SenderType  string     `json:"sender_type"`
	MessageType string     `json:"message_type"`
	Content     string     `json:"content"`
	FileURL     *string    `json:"file_url,omitempty"`
	FileName    *string    `json:"file_name,omitempty"`
	FileSize    *int64     `json:"file_size,omitempty"`
	IsRead      bool       `json:"is_read"`
	ReadAt      *time.Time `json:"read_at"`
	IsEdited    bool       `json:"is_edited"`
	EditedAt    *time.Time `json:"edited_at"`
	ReplyToID   *string    `json:"reply_to_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	
	// Sender info
	Sender   *MessageSenderInfo `json:"sender,omitempty"`
	ReplyTo  *ChatMessageResponse `json:"reply_to,omitempty"`
}

type MessageSenderInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url"`
	Type      string  `json:"type"` // 'user' or 'mitra'
}

// Pagination DTOs
type GetMessagesRequest struct {
	RoomID string `json:"room_id" validate:"required,uuid"`
	Limit  int    `json:"limit" validate:"min=1,max=100"`
	Before *string `json:"before,omitempty"` // Cursor: timestamp in RFC3339 format for cursor-based pagination
	After  *string `json:"after,omitempty"`  // Cursor: timestamp in RFC3339 format for loading newer messages
}

type GetRoomsRequest struct {
	Page  int `json:"page" validate:"min=1"`
	Limit int `json:"limit" validate:"min=1,max=50"`
}

type PaginatedMessagesResponse struct {
	Messages   []ChatMessageResponse `json:"messages"`
	Pagination MessagePaginationInfo `json:"pagination"`
}

type PaginatedRoomsResponse struct {
	Rooms      []ChatRoomResponse `json:"rooms"`
	Pagination PaginationInfo     `json:"pagination"`
}

type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Optimized pagination info for messages (cursor-based)
type MessagePaginationInfo struct {
	Limit      int     `json:"limit"`
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor,omitempty"` // Timestamp for loading older messages
	PrevCursor *string `json:"prev_cursor,omitempty"` // Timestamp for loading newer messages
}

// Real-time DTOs
type TypingIndicatorRequest struct {
	RoomID   string `json:"room_id" validate:"required,uuid"`
	IsTyping bool   `json:"is_typing"`
}

type TypingIndicatorResponse struct {
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	UserType  string    `json:"user_type"`
	UserName  string    `json:"user_name"`
	IsTyping  bool      `json:"is_typing"`
	Timestamp time.Time `json:"timestamp"`
}

type PresenceUpdateResponse struct {
	UserID     string    `json:"user_id"`
	UserType   string    `json:"user_type"`
	UserName   string    `json:"user_name"`
	IsOnline   bool      `json:"is_online"`
	LastSeen   time.Time `json:"last_seen"`
}

type MessageReadRequest struct {
	MessageIDs []string `json:"message_ids" validate:"required,min=1"`
}

type MessageReadResponse struct {
	RoomID      string    `json:"room_id"`
	MessageID   string    `json:"message_id"`
	ReaderID    string    `json:"reader_id"`
	ReaderType  string    `json:"reader_type"`
	ReadAt      time.Time `json:"read_at"`
}

// WebSocket DTOs
type WebSocketMessage struct {
	Type      string      `json:"type"`
	RoomID    string      `json:"room_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

type WebSocketResponse struct {
	Type      string      `json:"type"`
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Message types for WebSocket
const (
	// Incoming message types
	WSMsgTypeJoinRoom     = "join_room"
	WSMsgTypeLeaveRoom    = "leave_room"
	WSMsgTypeSendMessage  = "send_message"
	WSMsgTypeTyping       = "typing"
	WSMsgTypeMarkRead     = "mark_read"
	WSMsgTypeGetMessages  = "get_messages"
	
	// Outgoing message types
	WSMsgTypeNewMessage   = "new_message"
	WSMsgTypeMessageRead  = "message_read"
	WSMsgTypeTypingUpdate = "typing_update"
	WSMsgTypePresenceUpdate = "presence_update"
	WSMsgTypeRoomUpdate   = "room_update"
	WSMsgTypeError        = "error"
	WSMsgTypeSuccess      = "success"
)

// Connection info
type ConnectionInfo struct {
	UserID     string    `json:"user_id"`
	UserType   string    `json:"user_type"`
	SocketID   string    `json:"socket_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastPing   time.Time `json:"last_ping"`
}

// Room join/leave DTOs
type JoinRoomRequest struct {
	RoomID string `json:"room_id" validate:"required,uuid"`
}

type LeaveRoomRequest struct {
	RoomID string `json:"room_id" validate:"required,uuid"`
}

// Polling DTOs
type PollMessagesRequest struct {
	Since   *time.Time `json:"since,omitempty"`
	RoomIDs []string   `json:"room_ids,omitempty" validate:"omitempty,dive,uuid"`
	Timeout int        `json:"timeout" validate:"min=1,max=60"`
}

type PollMessagesResponse struct {
	Messages    []ChatMessageResponse `json:"messages"`
	RoomUpdates []RoomUpdateInfo      `json:"room_updates"`
	HasMore     bool                  `json:"has_more"`
	NextPoll    time.Time             `json:"next_poll"`
	Timestamp   time.Time             `json:"timestamp"`
}

type RoomUpdateInfo struct {
	RoomID        string    `json:"room_id"`
	UnreadCount   int       `json:"unread_count"`
	LastMessage   *string   `json:"last_message"`
	LastMessageAt *time.Time `json:"last_message_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
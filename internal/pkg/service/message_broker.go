package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/redis"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	goredis "github.com/redis/go-redis/v9"
)

// MessageBroker handles pub/sub messaging for real-time chat
type MessageBroker interface {
	// Publisher methods
	PublishMessage(ctx context.Context, channel string, message interface{}) error
	PublishToRoom(ctx context.Context, roomID string, message interface{}) error
	PublishToUser(ctx context.Context, userID, userType string, message interface{}) error
	
	// Subscriber methods
	Subscribe(ctx context.Context, channels ...string) (<-chan *Message, error)
	SubscribeToRoom(ctx context.Context, roomID string) (<-chan *Message, error)
	SubscribeToUser(ctx context.Context, userID, userType string) (<-chan *Message, error)
	
	// Connection management
	AddConnection(userID, userType, socketID string)
	RemoveConnection(socketID string)
	GetUserConnections(userID, userType string) []string
	GetConnectionInfo(socketID string) *ConnectionInfo
	
	// Presence management
	SetUserOnline(userID, userType string, socketID string)
	SetUserOffline(userID, userType string, socketID string)
	IsUserOnline(userID, userType string) bool
	GetOnlineUsers() map[string]*PresenceInfo
	
	// Room management
	JoinRoom(socketID, roomID string)
	LeaveRoom(socketID, roomID string)
	GetRoomConnections(roomID string) []string
	
	// Cleanup
	Cleanup(ctx context.Context) error
}

type RedisBroker struct {
	redisClient *redis.RedisCache
	pubsub      *goredis.PubSub
	connections map[string]*ConnectionInfo // socketID -> connection info
	userSockets map[string][]string        // userID:userType -> []socketID
	roomSockets map[string][]string        // roomID -> []socketID
	presence    map[string]*PresenceInfo   // userID:userType -> presence info
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

type Message struct {
	Channel   string      `json:"channel"`
	Pattern   string      `json:"pattern,omitempty"`
	Payload   string      `json:"payload"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ConnectionInfo struct {
	SocketID    string    `json:"socket_id"`
	UserID      string    `json:"user_id"`
	UserType    string    `json:"user_type"`
	ConnectedAt time.Time `json:"connected_at"`
	LastPing    time.Time `json:"last_ping"`
	Rooms       []string  `json:"rooms"`
}

type PresenceInfo struct {
	UserID      string    `json:"user_id"`
	UserType    string    `json:"user_type"`
	IsOnline    bool      `json:"is_online"`
	LastSeen    time.Time `json:"last_seen"`
	SocketIDs   []string  `json:"socket_ids"`
	ConnectedAt time.Time `json:"connected_at"`
}

// Channel patterns
const (
	ChannelRoomPrefix     = "chat:room:"
	ChannelUserPrefix     = "chat:user:"
	ChannelPresencePrefix = "chat:presence:"
	ChannelTypingPrefix   = "chat:typing:"
	ChannelSystemPrefix   = "chat:system:"
)

func NewRedisBroker(redisClient *redis.RedisCache) MessageBroker {
	ctx, cancel := context.WithCancel(context.Background())
	
	broker := &RedisBroker{
		redisClient: redisClient,
		connections: make(map[string]*ConnectionInfo),
		userSockets: make(map[string][]string),
		roomSockets: make(map[string][]string),
		presence:    make(map[string]*PresenceInfo),
		ctx:         ctx,
		cancel:      cancel,
	}
	
	// Start cleanup routine
	go broker.startCleanupRoutine()
	
	return broker
}

// Publisher methods
func (b *RedisBroker) PublishMessage(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// Use the underlying Redis client directly for pub/sub
	client := b.getRedisClient()
	return client.Publish(ctx, channel, data).Err()
}

func (b *RedisBroker) PublishToRoom(ctx context.Context, roomID string, message interface{}) error {
	channel := ChannelRoomPrefix + roomID
	return b.PublishMessage(ctx, channel, message)
}

func (b *RedisBroker) PublishToUser(ctx context.Context, userID, userType string, message interface{}) error {
	channel := ChannelUserPrefix + userID + ":" + userType
	return b.PublishMessage(ctx, channel, message)
}

// Subscriber methods
func (b *RedisBroker) Subscribe(ctx context.Context, channels ...string) (<-chan *Message, error) {
	client := b.getRedisClient()
	pubsub := client.Subscribe(ctx, channels...)
	
	msgChan := make(chan *Message, 100)
	
	go func() {
		defer close(msgChan)
		defer pubsub.Close()
		
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					applog.LogErrorCtx(ctx, err, "Error receiving message from Redis")
					continue
				}
				
				var data interface{}
				if err := json.Unmarshal([]byte(msg.Payload), &data); err != nil {
					applog.LogErrorCtx(ctx, err, "Error unmarshaling message payload")
					continue
				}
				
				message := &Message{
					Channel:   msg.Channel,
					Pattern:   msg.Pattern,
					Payload:   msg.Payload,
					Data:      data,
					Timestamp: time.Now(),
				}
				
				select {
				case msgChan <- message:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	
	return msgChan, nil
}

func (b *RedisBroker) SubscribeToRoom(ctx context.Context, roomID string) (<-chan *Message, error) {
	channel := ChannelRoomPrefix + roomID
	return b.Subscribe(ctx, channel)
}

func (b *RedisBroker) SubscribeToUser(ctx context.Context, userID, userType string) (<-chan *Message, error) {
	channel := ChannelUserPrefix + userID + ":" + userType
	return b.Subscribe(ctx, channel)
}

// Connection management
func (b *RedisBroker) AddConnection(userID, userType, socketID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	now := time.Now()
	userKey := userID + ":" + userType
	
	// Add connection info
	b.connections[socketID] = &ConnectionInfo{
		SocketID:    socketID,
		UserID:      userID,
		UserType:    userType,
		ConnectedAt: now,
		LastPing:    now,
		Rooms:       make([]string, 0),
	}
	
	// Add to user sockets
	b.userSockets[userKey] = append(b.userSockets[userKey], socketID)
	
	// Update presence
	b.setUserOnlineInternal(userID, userType, socketID, now)
	
	applog.InfoCtx(b.ctx, "Connection added", 
		"user_id", userID, 
		"user_type", userType, 
		"socket_id", socketID)
}

func (b *RedisBroker) RemoveConnection(socketID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	conn, exists := b.connections[socketID]
	if !exists {
		return
	}
	
	userKey := conn.UserID + ":" + conn.UserType
	
	// Remove from user sockets
	if sockets, exists := b.userSockets[userKey]; exists {
		for i, s := range sockets {
			if s == socketID {
				b.userSockets[userKey] = append(sockets[:i], sockets[i+1:]...)
				break
			}
		}
		
		// If no more sockets for this user, set offline
		if len(b.userSockets[userKey]) == 0 {
			delete(b.userSockets, userKey)
			b.setUserOfflineInternal(conn.UserID, conn.UserType, time.Now())
		}
	}
	
	// Remove from all rooms
	for _, roomID := range conn.Rooms {
		b.leaveRoomInternal(socketID, roomID)
	}
	
	// Remove connection
	delete(b.connections, socketID)
	
	applog.InfoCtx(b.ctx, "Connection removed", 
		"user_id", conn.UserID, 
		"user_type", conn.UserType, 
		"socket_id", socketID)
}

func (b *RedisBroker) GetUserConnections(userID, userType string) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	userKey := userID + ":" + userType
	if sockets, exists := b.userSockets[userKey]; exists {
		result := make([]string, len(sockets))
		copy(result, sockets)
		return result
	}
	return []string{}
}

func (b *RedisBroker) GetConnectionInfo(socketID string) *ConnectionInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	if conn, exists := b.connections[socketID]; exists {
		// Return a copy
		connCopy := *conn
		connCopy.Rooms = make([]string, len(conn.Rooms))
		copy(connCopy.Rooms, conn.Rooms)
		return &connCopy
	}
	return nil
}

// Presence management
func (b *RedisBroker) SetUserOnline(userID, userType string, socketID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.setUserOnlineInternal(userID, userType, socketID, time.Now())
}

func (b *RedisBroker) setUserOnlineInternal(userID, userType, socketID string, timestamp time.Time) {
	userKey := userID + ":" + userType
	
	if presence, exists := b.presence[userKey]; exists {
		presence.IsOnline = true
		presence.LastSeen = timestamp
		if !contains(presence.SocketIDs, socketID) {
			presence.SocketIDs = append(presence.SocketIDs, socketID)
		}
	} else {
		b.presence[userKey] = &PresenceInfo{
			UserID:      userID,
			UserType:    userType,
			IsOnline:    true,
			LastSeen:    timestamp,
			SocketIDs:   []string{socketID},
			ConnectedAt: timestamp,
		}
	}
	
	// Publish presence update
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		presenceUpdate := map[string]interface{}{
			"user_id":   userID,
			"user_type": userType,
			"is_online": true,
			"last_seen": timestamp,
		}
		
		channel := ChannelPresencePrefix + userID + ":" + userType
		if err := b.PublishMessage(ctx, channel, presenceUpdate); err != nil {
			applog.LogErrorCtx(ctx, err, "Failed to publish presence update")
		}
	}()
}

func (b *RedisBroker) SetUserOffline(userID, userType string, socketID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.setUserOfflineInternal(userID, userType, time.Now())
}

func (b *RedisBroker) setUserOfflineInternal(userID, userType string, timestamp time.Time) {
	userKey := userID + ":" + userType
	
	if presence, exists := b.presence[userKey]; exists {
		presence.IsOnline = false
		presence.LastSeen = timestamp
		presence.SocketIDs = []string{}
	}
	
	// Publish presence update
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		presenceUpdate := map[string]interface{}{
			"user_id":   userID,
			"user_type": userType,
			"is_online": false,
			"last_seen": timestamp,
		}
		
		channel := ChannelPresencePrefix + userID + ":" + userType
		if err := b.PublishMessage(ctx, channel, presenceUpdate); err != nil {
			applog.LogErrorCtx(ctx, err, "Failed to publish presence update")
		}
	}()
}

func (b *RedisBroker) IsUserOnline(userID, userType string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	userKey := userID + ":" + userType
	if presence, exists := b.presence[userKey]; exists {
		return presence.IsOnline
	}
	return false
}

func (b *RedisBroker) GetOnlineUsers() map[string]*PresenceInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	result := make(map[string]*PresenceInfo)
	for key, presence := range b.presence {
		if presence.IsOnline {
			// Return a copy
			presenceCopy := *presence
			presenceCopy.SocketIDs = make([]string, len(presence.SocketIDs))
			copy(presenceCopy.SocketIDs, presence.SocketIDs)
			result[key] = &presenceCopy
		}
	}
	return result
}

// Room management
func (b *RedisBroker) JoinRoom(socketID, roomID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.joinRoomInternal(socketID, roomID)
}

func (b *RedisBroker) joinRoomInternal(socketID, roomID string) {
	// Add socket to room
	if !contains(b.roomSockets[roomID], socketID) {
		b.roomSockets[roomID] = append(b.roomSockets[roomID], socketID)
	}
	
	// Add room to connection
	if conn, exists := b.connections[socketID]; exists {
		if !contains(conn.Rooms, roomID) {
			conn.Rooms = append(conn.Rooms, roomID)
		}
	}
	
	applog.InfoCtx(b.ctx, "Socket joined room", 
		"socket_id", socketID, 
		"room_id", roomID)
}

func (b *RedisBroker) LeaveRoom(socketID, roomID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.leaveRoomInternal(socketID, roomID)
}

func (b *RedisBroker) leaveRoomInternal(socketID, roomID string) {
	// Remove socket from room
	if sockets, exists := b.roomSockets[roomID]; exists {
		for i, s := range sockets {
			if s == socketID {
				b.roomSockets[roomID] = append(sockets[:i], sockets[i+1:]...)
				break
			}
		}
		
		// Clean up empty room
		if len(b.roomSockets[roomID]) == 0 {
			delete(b.roomSockets, roomID)
		}
	}
	
	// Remove room from connection
	if conn, exists := b.connections[socketID]; exists {
		for i, r := range conn.Rooms {
			if r == roomID {
				conn.Rooms = append(conn.Rooms[:i], conn.Rooms[i+1:]...)
				break
			}
		}
	}
	
	applog.InfoCtx(b.ctx, "Socket left room", 
		"socket_id", socketID, 
		"room_id", roomID)
}

func (b *RedisBroker) GetRoomConnections(roomID string) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	if sockets, exists := b.roomSockets[roomID]; exists {
		result := make([]string, len(sockets))
		copy(result, sockets)
		return result
	}
	return []string{}
}

// Cleanup
func (b *RedisBroker) Cleanup(ctx context.Context) error {
	b.cancel()
	
	b.mu.Lock()
	defer b.mu.Unlock()
	
	// Clear all data structures
	b.connections = make(map[string]*ConnectionInfo)
	b.userSockets = make(map[string][]string)
	b.roomSockets = make(map[string][]string)
	b.presence = make(map[string]*PresenceInfo)
	
	if b.pubsub != nil {
		return b.pubsub.Close()
	}
	
	return nil
}

// Helper methods
func (b *RedisBroker) getRedisClient() *goredis.Client {
	// Access the underlying Redis client from RedisCache
	return b.redisClient.GetClient()
}

func (b *RedisBroker) startCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.cleanupStaleConnections()
		}
	}
}

func (b *RedisBroker) cleanupStaleConnections() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	now := time.Now()
	staleThreshold := 10 * time.Minute
	
	var staleConnections []string
	
	for socketID, conn := range b.connections {
		if now.Sub(conn.LastPing) > staleThreshold {
			staleConnections = append(staleConnections, socketID)
		}
	}
	
	for _, socketID := range staleConnections {
		applog.InfoCtx(b.ctx, "Cleaning up stale connection", "socket_id", socketID)
		// Remove without lock since we already have it
		if conn := b.connections[socketID]; conn != nil {
			userKey := conn.UserID + ":" + conn.UserType
			
			// Remove from user sockets
			if sockets, exists := b.userSockets[userKey]; exists {
				for i, s := range sockets {
					if s == socketID {
						b.userSockets[userKey] = append(sockets[:i], sockets[i+1:]...)
						break
					}
				}
				
				if len(b.userSockets[userKey]) == 0 {
					delete(b.userSockets, userKey)
					b.setUserOfflineInternal(conn.UserID, conn.UserType, now)
				}
			}
			
			// Remove from rooms
			for _, roomID := range conn.Rooms {
				b.leaveRoomInternal(socketID, roomID)
			}
		}
		
		delete(b.connections, socketID)
	}
}

// Utility functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
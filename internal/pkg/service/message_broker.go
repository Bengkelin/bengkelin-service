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

// RedisBrokerDebugger provides debugging methods for Redis broker
type RedisBrokerDebugger interface {
	TestRedisPubSub(ctx context.Context) error
	LogRedisState(ctx context.Context)
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
	applog.InfoCtx(ctx, "Redis: Attempting to publish message", 
		"channel", channel, 
		"message_type", getMessageType(message))
	
	data, err := json.Marshal(message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Redis: Failed to marshal message for publishing", "channel", channel)
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// Use the underlying Redis client directly for pub/sub
	client := b.getRedisClient()
	
	// CRITICAL: Test Redis connection before publishing
	if err := b.testRedisConnection(ctx); err != nil {
		applog.LogErrorCtx(ctx, err, "Redis: Connection test failed before publishing", "channel", channel)
		return fmt.Errorf("redis connection failed: %w", err)
	}
	
	// Publish message
	result := client.Publish(ctx, channel, data)
	if err := result.Err(); err != nil {
		applog.LogErrorCtx(ctx, err, "Redis: Failed to publish message", 
			"channel", channel, 
			"data_size", len(data))
		return fmt.Errorf("failed to publish to Redis: %w", err)
	}
	
	// Get number of subscribers who received the message
	subscriberCount := result.Val()
	applog.InfoCtx(ctx, "Redis: Message published successfully", 
		"channel", channel, 
		"subscriber_count", subscriberCount,
		"data_size", len(data),
		"message_type", getMessageType(message))
	
	// CRITICAL: Log if no subscribers received the message
	if subscriberCount == 0 {
		applog.InfoCtx(ctx, "Redis: WARNING - No subscribers received message", 
			"channel", channel,
			"message_type", getMessageType(message))
	}
	
	return nil
}

func (b *RedisBroker) PublishToRoom(ctx context.Context, roomID string, message interface{}) error {
	channel := ChannelRoomPrefix + roomID
	
	applog.InfoCtx(ctx, "Redis: Publishing message to room", 
		"room_id", roomID, 
		"channel", channel,
		"message_type", getMessageType(message))
	
	// CRITICAL: Check if room has any connections before publishing
	roomConnections := b.GetRoomConnections(roomID)
	applog.InfoCtx(ctx, "Redis: Room connection status", 
		"room_id", roomID,
		"local_connections", len(roomConnections),
		"connections", roomConnections)
	
	err := b.PublishMessage(ctx, channel, message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Redis: Failed to publish to room", "room_id", roomID)
		return err
	}
	
	applog.InfoCtx(ctx, "Redis: Room message published successfully", 
		"room_id", roomID, 
		"channel", channel)
	
	return nil
}

func (b *RedisBroker) PublishToUser(ctx context.Context, userID, userType string, message interface{}) error {
	channel := ChannelUserPrefix + userID + ":" + userType
	
	applog.InfoCtx(ctx, "Redis: Publishing message to user", 
		"user_id", userID, 
		"user_type", userType,
		"channel", channel,
		"message_type", getMessageType(message))
	
	// CRITICAL: Check if user has any connections before publishing
	userConnections := b.GetUserConnections(userID, userType)
	applog.InfoCtx(ctx, "Redis: User connection status", 
		"user_id", userID,
		"user_type", userType,
		"local_connections", len(userConnections),
		"connections", userConnections)
	
	err := b.PublishMessage(ctx, channel, message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Redis: Failed to publish to user", 
			"user_id", userID, 
			"user_type", userType)
		return err
	}
	
	applog.InfoCtx(ctx, "Redis: User message published successfully", 
		"user_id", userID, 
		"user_type", userType,
		"channel", channel)
	
	return nil
}

// Subscriber methods
func (b *RedisBroker) Subscribe(ctx context.Context, channels ...string) (<-chan *Message, error) {
	applog.InfoCtx(ctx, "Redis: Starting subscription", 
		"channels", channels, 
		"channel_count", len(channels))
	
	client := b.getRedisClient()
	
	// CRITICAL: Test Redis connection before subscribing
	if err := b.testRedisConnection(ctx); err != nil {
		applog.LogErrorCtx(ctx, err, "Redis: Connection test failed before subscribing", "channels", channels)
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	
	pubsub := client.Subscribe(ctx, channels...)
	
	// CRITICAL: Test subscription immediately with timeout
	testCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := pubsub.Receive(testCtx)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Redis: Failed to establish subscription", "channels", channels)
		pubsub.Close()
		return nil, fmt.Errorf("failed to establish Redis subscription: %w", err)
	}
	
	applog.InfoCtx(ctx, "Redis: Subscription established successfully", "channels", channels)
	
	msgChan := make(chan *Message, 100)
	
	go func() {
		defer func() {
			applog.InfoCtx(ctx, "Redis: Subscription goroutine ending", "channels", channels)
			close(msgChan)
			pubsub.Close()
		}()
		
		applog.InfoCtx(ctx, "Redis: Subscription goroutine started", "channels", channels)
		
		messageCount := 0
		for {
			select {
			case <-ctx.Done():
				applog.InfoCtx(ctx, "Redis: Subscription context cancelled", 
					"channels", channels, 
					"messages_received", messageCount)
				return
			default:
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					applog.LogErrorCtx(ctx, err, "Redis: Error receiving message", 
						"channels", channels, 
						"messages_received", messageCount)
					continue
				}
				
				messageCount++
				applog.InfoCtx(ctx, "Redis: Message received from subscription", 
					"channel", msg.Channel,
					"pattern", msg.Pattern,
					"payload_size", len(msg.Payload),
					"message_count", messageCount)
				
				var data interface{}
				if err := json.Unmarshal([]byte(msg.Payload), &data); err != nil {
					applog.LogErrorCtx(ctx, err, "Redis: Error unmarshaling message payload", 
						"channel", msg.Channel,
						"payload", msg.Payload)
					continue
				}
				
				message := &Message{
					Channel:   msg.Channel,
					Pattern:   msg.Pattern,
					Payload:   msg.Payload,
					Data:      data,
					Timestamp: time.Now(),
				}
				
				applog.InfoCtx(ctx, "Redis: Forwarding message to channel", 
					"redis_channel", msg.Channel,
					"message_type", getMessageType(data))
				
				select {
				case msgChan <- message:
					applog.InfoCtx(ctx, "Redis: Message forwarded successfully", 
						"redis_channel", msg.Channel)
				case <-ctx.Done():
					applog.InfoCtx(ctx, "Redis: Context cancelled while forwarding message", 
						"redis_channel", msg.Channel)
					return
				default:
					applog.InfoCtx(ctx, "Redis: Message channel full, dropping message", 
						"redis_channel", msg.Channel)
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
	
	applog.InfoCtx(b.ctx, "Redis: Adding connection", 
		"user_id", userID, 
		"user_type", userType, 
		"socket_id", socketID,
		"user_key", userKey)
	
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
	
	// CRITICAL: Log current state after adding connection
	applog.InfoCtx(b.ctx, "Redis: Connection added - Current state", 
		"user_id", userID, 
		"user_type", userType, 
		"socket_id", socketID,
		"total_connections", len(b.connections),
		"user_sockets_count", len(b.userSockets[userKey]),
		"total_users", len(b.userSockets))
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
	
	applog.InfoCtx(b.ctx, "Redis: Setting user online", 
		"user_id", userID, 
		"user_type", userType,
		"socket_id", socketID,
		"user_key", userKey)
	
	if presence, exists := b.presence[userKey]; exists {
		presence.IsOnline = true
		presence.LastSeen = timestamp
		if !contains(presence.SocketIDs, socketID) {
			presence.SocketIDs = append(presence.SocketIDs, socketID)
		}
		applog.InfoCtx(b.ctx, "Redis: Updated existing presence", 
			"user_key", userKey,
			"socket_count", len(presence.SocketIDs))
	} else {
		b.presence[userKey] = &PresenceInfo{
			UserID:      userID,
			UserType:    userType,
			IsOnline:    true,
			LastSeen:    timestamp,
			SocketIDs:   []string{socketID},
			ConnectedAt: timestamp,
		}
		applog.InfoCtx(b.ctx, "Redis: Created new presence", 
			"user_key", userKey)
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
		applog.InfoCtx(ctx, "Redis: Publishing presence update", 
			"user_id", userID,
			"user_type", userType,
			"channel", channel,
			"is_online", true)
		
		if err := b.PublishMessage(ctx, channel, presenceUpdate); err != nil {
			applog.LogErrorCtx(ctx, err, "Redis: Failed to publish presence update")
		} else {
			applog.InfoCtx(ctx, "Redis: Presence update published successfully", 
				"user_id", userID,
				"user_type", userType)
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

// CRITICAL: Add Redis connection testing method
func (b *RedisBroker) testRedisConnection(ctx context.Context) error {
	client := b.getRedisClient()
	
	// Test with shorter timeout to avoid blocking
	testCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	result := client.Ping(testCtx)
	if err := result.Err(); err != nil {
		applog.LogErrorCtx(ctx, err, "Redis: Connection test failed")
		return err
	}
	
	response := result.Val()
	applog.InfoCtx(ctx, "Redis: Connection test successful", "response", response)
	return nil
}

// CRITICAL: Add Redis state inspection methods
func (b *RedisBroker) LogRedisState(ctx context.Context) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	applog.InfoCtx(ctx, "Redis: Current broker state", 
		"total_connections", len(b.connections),
		"total_users", len(b.userSockets),
		"total_rooms", len(b.roomSockets),
		"online_users", len(b.presence))
	
	// Log connection details
	for socketID, conn := range b.connections {
		applog.InfoCtx(ctx, "Redis: Connection detail", 
			"socket_id", socketID,
			"user_id", conn.UserID,
			"user_type", conn.UserType,
			"rooms", conn.Rooms,
			"connected_at", conn.ConnectedAt,
			"last_ping", conn.LastPing)
	}
	
	// Log room details
	for roomID, sockets := range b.roomSockets {
		applog.InfoCtx(ctx, "Redis: Room detail", 
			"room_id", roomID,
			"socket_count", len(sockets),
			"sockets", sockets)
	}
	
	// Log user presence
	for userKey, presence := range b.presence {
		applog.InfoCtx(ctx, "Redis: Presence detail", 
			"user_key", userKey,
			"is_online", presence.IsOnline,
			"socket_ids", presence.SocketIDs,
			"last_seen", presence.LastSeen)
	}
}

// CRITICAL: Add method to test Redis pub/sub functionality
func (b *RedisBroker) TestRedisPubSub(ctx context.Context) error {
	testChannel := "test:redis:pubsub:" + time.Now().Format("20060102150405")
	testMessage := map[string]interface{}{
		"type": "test",
		"data": "Redis pub/sub test message",
		"timestamp": time.Now(),
	}
	
	applog.InfoCtx(ctx, "Redis: Starting pub/sub test", "test_channel", testChannel)
	
	// Subscribe to test channel
	msgChan, err := b.Subscribe(ctx, testChannel)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Redis: Failed to subscribe to test channel")
		return err
	}
	
	// Create timeout context for test
	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	// Publish test message
	go func() {
		time.Sleep(100 * time.Millisecond) // Small delay to ensure subscription is ready
		if err := b.PublishMessage(testCtx, testChannel, testMessage); err != nil {
			applog.LogErrorCtx(testCtx, err, "Redis: Failed to publish test message")
		}
	}()
	
	// Wait for message or timeout
	select {
	case msg := <-msgChan:
		applog.InfoCtx(ctx, "Redis: Pub/sub test successful", 
			"test_channel", testChannel,
			"received_channel", msg.Channel,
			"payload_size", len(msg.Payload))
		return nil
	case <-testCtx.Done():
		applog.LogErrorCtx(ctx, testCtx.Err(), "Redis: Pub/sub test timed out", "test_channel", testChannel)
		return fmt.Errorf("Redis pub/sub test timed out")
	}
}

func (b *RedisBroker) startCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	applog.InfoCtx(b.ctx, "Redis: Cleanup routine started")
	
	for {
		select {
		case <-b.ctx.Done():
			applog.InfoCtx(b.ctx, "Redis: Cleanup routine stopping")
			return
		case <-ticker.C:
			applog.InfoCtx(b.ctx, "Redis: Running cleanup routine")
			b.cleanupStaleConnections()
			b.LogRedisState(b.ctx)
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

// Helper function to extract message type for logging
func getMessageType(message interface{}) string {
	if msgData, ok := message.(map[string]interface{}); ok {
		if msgType, exists := msgData["type"]; exists {
			if typeStr, ok := msgType.(string); ok {
				return typeStr
			}
		}
	}
	return "unknown"
}
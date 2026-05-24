package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
)

// InMemoryBroker is a simple in-memory implementation of MessageBroker
// Used as a fallback when Redis is not available
type InMemoryBroker struct {
	connections map[string]*ConnectionInfo // socketID -> connection info
	userSockets map[string][]string        // userID:userType -> []socketID
	roomSockets map[string][]string        // roomID -> []socketID
	presence    map[string]*PresenceInfo   // userID:userType -> presence info
	subscribers map[string][]chan *Message // channel -> []subscriber channels
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewInMemoryBroker() MessageBroker {
	ctx, cancel := context.WithCancel(context.Background())
	
	broker := &InMemoryBroker{
		connections: make(map[string]*ConnectionInfo),
		userSockets: make(map[string][]string),
		roomSockets: make(map[string][]string),
		presence:    make(map[string]*PresenceInfo),
		subscribers: make(map[string][]chan *Message),
		ctx:         ctx,
		cancel:      cancel,
	}
	
	// Start cleanup routine
	go broker.startCleanupRoutine()
	
	return broker
}

// Publisher methods
func (b *InMemoryBroker) PublishMessage(ctx context.Context, channel string, message interface{}) error {
	b.mu.RLock()
	subscribers, exists := b.subscribers[channel]
	b.mu.RUnlock()
	
	applog.InfoCtx(ctx, "Publishing message to channel (in-memory)", 
		"channel", channel, 
		"subscribers_count", len(subscribers),
		"message_type", getMessageType(message))
	
	if !exists || len(subscribers) == 0 {
		applog.InfoCtx(ctx, "No subscribers for channel (in-memory)", "channel", channel)
		return nil
	}
	
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	msg := &Message{
		Channel:   channel,
		Payload:   string(data),
		Data:      message,
		Timestamp: time.Now(),
	}
	
	// Send to all subscribers
	successCount := 0
	for _, subscriber := range subscribers {
		select {
		case subscriber <- msg:
			successCount++
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Subscriber channel is full, skip
			applog.InfoCtx(ctx, "Subscriber channel full, skipping", "channel", channel)
		}
	}
	
	applog.InfoCtx(ctx, "Message sent to subscribers (in-memory)", 
		"channel", channel, 
		"total_subscribers", len(subscribers),
		"successful_sends", successCount)
	
	return nil
}

func (b *InMemoryBroker) PublishToRoom(ctx context.Context, roomID string, message interface{}) error {
	channel := ChannelRoomPrefix + roomID
	
	// Log the room publishing attempt with more details
	applog.InfoCtx(ctx, "Publishing message to room (in-memory)", 
		"room_id", roomID, 
		"channel", channel,
		"message_type", getMessageType(message))
	
	// CRITICAL FIX: Get room connections and log them
	roomConnections := b.GetRoomConnections(roomID)
	applog.InfoCtx(ctx, "Room connections found", 
		"room_id", roomID,
		"connection_count", len(roomConnections),
		"connections", roomConnections)
	
	// Log connection details for debugging
	b.mu.RLock()
	for _, socketID := range roomConnections {
		if conn, exists := b.connections[socketID]; exists {
			applog.InfoCtx(ctx, "Room connection details", 
				"room_id", roomID,
				"socket_id", socketID,
				"user_id", conn.UserID,
				"user_type", conn.UserType)
		}
	}
	b.mu.RUnlock()
	
	return b.PublishMessage(ctx, channel, message)
}

func (b *InMemoryBroker) PublishToUser(ctx context.Context, userID, userType string, message interface{}) error {
	channel := ChannelUserPrefix + userID + ":" + userType
	return b.PublishMessage(ctx, channel, message)
}

// Subscriber methods
func (b *InMemoryBroker) Subscribe(ctx context.Context, channels ...string) (<-chan *Message, error) {
	msgChan := make(chan *Message, 100)
	
	b.mu.Lock()
	for _, channel := range channels {
		b.subscribers[channel] = append(b.subscribers[channel], msgChan)
	}
	b.mu.Unlock()
	
	// Clean up subscription when context is done
	go func() {
		<-ctx.Done()
		b.mu.Lock()
		defer b.mu.Unlock()
		
		for _, channel := range channels {
			if subs, exists := b.subscribers[channel]; exists {
				for i, sub := range subs {
					if sub == msgChan {
						b.subscribers[channel] = append(subs[:i], subs[i+1:]...)
						break
					}
				}
				
				// Clean up empty subscriber list
				if len(b.subscribers[channel]) == 0 {
					delete(b.subscribers, channel)
				}
			}
		}
		close(msgChan)
	}()
	
	return msgChan, nil
}

func (b *InMemoryBroker) SubscribeToRoom(ctx context.Context, roomID string) (<-chan *Message, error) {
	channel := ChannelRoomPrefix + roomID
	return b.Subscribe(ctx, channel)
}

func (b *InMemoryBroker) SubscribeToUser(ctx context.Context, userID, userType string) (<-chan *Message, error) {
	channel := ChannelUserPrefix + userID + ":" + userType
	return b.Subscribe(ctx, channel)
}

// Connection management
func (b *InMemoryBroker) AddConnection(userID, userType, socketID string) {
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
	
	applog.InfoCtx(b.ctx, "Connection added (in-memory)", 
		"user_id", userID, 
		"user_type", userType, 
		"socket_id", socketID)
}

func (b *InMemoryBroker) RemoveConnection(socketID string) {
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
	
	applog.InfoCtx(b.ctx, "Connection removed (in-memory)", 
		"user_id", conn.UserID, 
		"user_type", conn.UserType, 
		"socket_id", socketID)
}

func (b *InMemoryBroker) GetUserConnections(userID, userType string) []string {
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

func (b *InMemoryBroker) GetConnectionInfo(socketID string) *ConnectionInfo {
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
func (b *InMemoryBroker) SetUserOnline(userID, userType string, socketID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.setUserOnlineInternal(userID, userType, socketID, time.Now())
}

func (b *InMemoryBroker) setUserOnlineInternal(userID, userType, socketID string, timestamp time.Time) {
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
}

func (b *InMemoryBroker) SetUserOffline(userID, userType string, socketID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.setUserOfflineInternal(userID, userType, time.Now())
}

func (b *InMemoryBroker) setUserOfflineInternal(userID, userType string, timestamp time.Time) {
	userKey := userID + ":" + userType
	
	if presence, exists := b.presence[userKey]; exists {
		presence.IsOnline = false
		presence.LastSeen = timestamp
		presence.SocketIDs = []string{}
	}
}

func (b *InMemoryBroker) IsUserOnline(userID, userType string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	userKey := userID + ":" + userType
	if presence, exists := b.presence[userKey]; exists {
		return presence.IsOnline
	}
	return false
}

func (b *InMemoryBroker) GetOnlineUsers() map[string]*PresenceInfo {
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
func (b *InMemoryBroker) JoinRoom(socketID, roomID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.joinRoomInternal(socketID, roomID)
}

func (b *InMemoryBroker) joinRoomInternal(socketID, roomID string) {
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
	
	applog.InfoCtx(b.ctx, "Socket joined room (in-memory)", 
		"socket_id", socketID, 
		"room_id", roomID)
}

func (b *InMemoryBroker) LeaveRoom(socketID, roomID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.leaveRoomInternal(socketID, roomID)
}

func (b *InMemoryBroker) leaveRoomInternal(socketID, roomID string) {
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
	
	applog.InfoCtx(b.ctx, "Socket left room (in-memory)", 
		"socket_id", socketID, 
		"room_id", roomID)
}

func (b *InMemoryBroker) GetRoomConnections(roomID string) []string {
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
func (b *InMemoryBroker) Cleanup(ctx context.Context) error {
	b.cancel()
	
	b.mu.Lock()
	defer b.mu.Unlock()
	
	// Clear all data structures
	b.connections = make(map[string]*ConnectionInfo)
	b.userSockets = make(map[string][]string)
	b.roomSockets = make(map[string][]string)
	b.presence = make(map[string]*PresenceInfo)
	b.subscribers = make(map[string][]chan *Message)
	
	return nil
}

func (b *InMemoryBroker) startCleanupRoutine() {
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

func (b *InMemoryBroker) cleanupStaleConnections() {
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
		applog.InfoCtx(b.ctx, "Cleaning up stale connection (in-memory)", "socket_id", socketID)
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
package websocket

import (
	"context"
	"sync"

	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
)

// WebSocketClient interface to avoid circular dependencies
type WebSocketClient interface {
	GetSocketID() string
	GetUserID() string
	GetUserType() string
	GetRooms() map[string]bool
	SendMessage(message []byte) bool
}

// GlobalWebSocketRegistry manages all connected WebSocket clients
type GlobalWebSocketRegistry struct {
	clients     map[string]WebSocketClient              // socketID -> client
	userClients map[string]map[string]WebSocketClient   // userID:userType -> socketID -> client
	roomClients map[string]map[string]WebSocketClient   // roomID -> socketID -> client
	mu          sync.RWMutex
}

var (
	// Global instance
	globalRegistry = &GlobalWebSocketRegistry{
		clients:     make(map[string]WebSocketClient),
		userClients: make(map[string]map[string]WebSocketClient),
		roomClients: make(map[string]map[string]WebSocketClient),
	}
)

// GetGlobalRegistry returns the global WebSocket registry
func GetGlobalRegistry() *GlobalWebSocketRegistry {
	return globalRegistry
}

// RegisterClient adds a client to the registry
func (r *GlobalWebSocketRegistry) RegisterClient(client WebSocketClient) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	socketID := client.GetSocketID()
	userID := client.GetUserID()
	userType := client.GetUserType()
	
	r.clients[socketID] = client
	
	userKey := userID + ":" + userType
	if r.userClients[userKey] == nil {
		r.userClients[userKey] = make(map[string]WebSocketClient)
	}
	r.userClients[userKey][socketID] = client
	
	applog.InfoCtx(context.Background(), "WebSocket client registered in global registry", 
		"socket_id", socketID,
		"user_id", userID,
		"user_type", userType,
		"total_clients", len(r.clients))
}

// UnregisterClient removes a client from the registry
func (r *GlobalWebSocketRegistry) UnregisterClient(socketID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	client, exists := r.clients[socketID]
	if !exists {
		return
	}
	
	userID := client.GetUserID()
	userType := client.GetUserType()
	
	// Remove from user clients
	userKey := userID + ":" + userType
	if r.userClients[userKey] != nil {
		delete(r.userClients[userKey], socketID)
		if len(r.userClients[userKey]) == 0 {
			delete(r.userClients, userKey)
		}
	}
	
	// Remove from room clients
	for roomID := range client.GetRooms() {
		if r.roomClients[roomID] != nil {
			delete(r.roomClients[roomID], socketID)
			if len(r.roomClients[roomID]) == 0 {
				delete(r.roomClients, roomID)
			}
		}
	}
	
	delete(r.clients, socketID)
	
	applog.InfoCtx(context.Background(), "WebSocket client unregistered from global registry", 
		"socket_id", socketID,
		"total_clients", len(r.clients))
}

// JoinRoom adds a client to a room
func (r *GlobalWebSocketRegistry) JoinRoom(socketID, roomID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	client, exists := r.clients[socketID]
	if !exists {
		return
	}
	
	if r.roomClients[roomID] == nil {
		r.roomClients[roomID] = make(map[string]WebSocketClient)
	}
	r.roomClients[roomID][socketID] = client
	
	applog.InfoCtx(context.Background(), "WebSocket client joined room in registry", 
		"socket_id", socketID,
		"room_id", roomID,
		"room_clients", len(r.roomClients[roomID]))
}

// LeaveRoom removes a client from a room
func (r *GlobalWebSocketRegistry) LeaveRoom(socketID, roomID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.roomClients[roomID] != nil {
		delete(r.roomClients[roomID], socketID)
		if len(r.roomClients[roomID]) == 0 {
			delete(r.roomClients, roomID)
		}
	}
}

// BroadcastToRoom sends a message to all clients in a room
func (r *GlobalWebSocketRegistry) BroadcastToRoom(roomID string, message []byte) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	clients := r.roomClients[roomID]
	if clients == nil {
		applog.InfoCtx(context.Background(), "No clients in room for direct broadcast", 
			"room_id", roomID)
		return 0
	}
	
	sentCount := 0
	for socketID, client := range clients {
		if client.SendMessage(message) {
			sentCount++
			applog.InfoCtx(context.Background(), "Direct broadcast to room client successful", 
				"room_id", roomID,
				"socket_id", socketID,
				"user_id", client.GetUserID(),
				"user_type", client.GetUserType())
		} else {
			applog.InfoCtx(context.Background(), "Direct broadcast to room client failed", 
				"room_id", roomID,
				"socket_id", socketID)
		}
	}
	
	applog.InfoCtx(context.Background(), "Direct room broadcast completed", 
		"room_id", roomID,
		"total_clients", len(clients),
		"sent_count", sentCount)
	
	return sentCount
}

// BroadcastToUser sends a message to all clients of a specific user
func (r *GlobalWebSocketRegistry) BroadcastToUser(userID, userType string, message []byte) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	userKey := userID + ":" + userType
	clients := r.userClients[userKey]
	if clients == nil {
		applog.InfoCtx(context.Background(), "No clients for user for direct broadcast", 
			"user_id", userID,
			"user_type", userType)
		return 0
	}
	
	sentCount := 0
	for socketID, client := range clients {
		if client.SendMessage(message) {
			sentCount++
			applog.InfoCtx(context.Background(), "Direct broadcast to user client successful", 
				"user_id", userID,
				"user_type", userType,
				"socket_id", socketID)
		} else {
			applog.InfoCtx(context.Background(), "Direct broadcast to user client failed", 
				"user_id", userID,
				"user_type", userType,
				"socket_id", socketID)
		}
	}
	
	applog.InfoCtx(context.Background(), "Direct user broadcast completed", 
		"user_id", userID,
		"user_type", userType,
		"total_clients", len(clients),
		"sent_count", sentCount)
	
	return sentCount
}

// GetRoomClients returns the number of clients in a room
func (r *GlobalWebSocketRegistry) GetRoomClients(roomID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.roomClients[roomID] == nil {
		return 0
	}
	return len(r.roomClients[roomID])
}

// GetUserClients returns the number of clients for a user
func (r *GlobalWebSocketRegistry) GetUserClients(userID, userType string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	userKey := userID + ":" + userType
	if r.userClients[userKey] == nil {
		return 0
	}
	return len(r.userClients[userKey])
}
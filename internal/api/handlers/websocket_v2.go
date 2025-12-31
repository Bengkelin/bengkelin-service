package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/service"
	"github.com/Bengkelin/bengkelin-service/pkg/crypto"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketV2Handler struct {
	chatService   service.ChatV2ServiceInterface
	messageBroker service.MessageBroker
	upgrader      websocket.Upgrader
}

func NewWebSocketV2Handler(chatService service.ChatV2ServiceInterface, messageBroker service.MessageBroker) *WebSocketV2Handler {
	return &WebSocketV2Handler{
		chatService:   chatService,
		messageBroker: messageBroker,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// WebSocket connection handler
func (h *WebSocketV2Handler) HandleWebSocket(c *gin.Context) {
	// Extract JWT token from query parameter or header
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}
	}

	if token == "" {
		applog.LogErrorCtx(c.Request.Context(), nil, "WebSocket connection rejected: missing token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authentication token"})
		return
	}

	// Validate JWT token
	claims, err := crypto.ValidateJWT(token)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "WebSocket connection rejected: invalid token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authentication token"})
		return
	}

	userID := claims.UserID
	userType := "user"
	if claims.MitraID != "" {
		userID = claims.MitraID
		userType = "mitra"
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		applog.LogErrorCtx(c.Request.Context(), err, "Failed to upgrade WebSocket connection")
		return
	}

	// Generate unique socket ID
	socketID := generateSocketID(userID, userType)
	
	applog.InfoCtx(c.Request.Context(), "WebSocket connection established", 
		"user_id", userID, 
		"user_type", userType, 
		"socket_id", socketID)

	// Create WebSocket client
	client := &WebSocketClient{
		conn:          conn,
		send:          make(chan []byte, 256),
		handler:       h,
		userID:        userID,
		userType:      userType,
		socketID:      socketID,
		rooms:         make(map[string]bool),
		lastPing:      time.Now(),
		ctx:           c.Request.Context(),
	}

	// Register client with message broker
	h.messageBroker.AddConnection(userID, userType, socketID)

	// Set user online
	h.messageBroker.SetUserOnline(userID, userType, socketID)

	// Start client goroutines
	go client.writePump()
	go client.readPump()

	// Subscribe to user-specific messages
	go client.subscribeToUserMessages()
}

type WebSocketClient struct {
	conn     *websocket.Conn
	send     chan []byte
	handler  *WebSocketV2Handler
	userID   string
	userType string
	socketID string
	rooms    map[string]bool
	lastPing time.Time
	ctx      context.Context
}

// Read messages from WebSocket connection
func (c *WebSocketClient) readPump() {
	defer func() {
		c.cleanup()
		c.conn.Close()
	}()

	// Set read deadline and pong handler
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.lastPing = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				applog.LogErrorCtx(c.ctx, err, "WebSocket read error", "socket_id", c.socketID)
			}
			break
		}

		c.lastPing = time.Now()
		c.handleMessage(messageBytes)
	}
}

// Write messages to WebSocket connection
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				applog.LogErrorCtx(c.ctx, err, "WebSocket write error", "socket_id", c.socketID)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Handle incoming WebSocket messages
func (c *WebSocketClient) handleMessage(messageBytes []byte) {
	var wsMsg dto.WebSocketMessage
	if err := json.Unmarshal(messageBytes, &wsMsg); err != nil {
		c.sendError("invalid message format")
		return
	}

	ctx := context.WithValue(c.ctx, "user_id", c.userID)
	ctx = context.WithValue(ctx, "user_type", c.userType)

	switch wsMsg.Type {
	case dto.WSMsgTypeJoinRoom:
		c.handleJoinRoom(ctx, wsMsg)
	case dto.WSMsgTypeLeaveRoom:
		c.handleLeaveRoom(ctx, wsMsg)
	case dto.WSMsgTypeSendMessage:
		c.handleSendMessage(ctx, wsMsg)
	case dto.WSMsgTypeTyping:
		c.handleTyping(ctx, wsMsg)
	case dto.WSMsgTypeMarkRead:
		c.handleMarkRead(ctx, wsMsg)
	case dto.WSMsgTypeGetMessages:
		c.handleGetMessages(ctx, wsMsg)
	default:
		c.sendError("unknown message type")
	}
}

// Handle join room request
func (c *WebSocketClient) handleJoinRoom(ctx context.Context, wsMsg dto.WebSocketMessage) {
	var req dto.JoinRoomRequest
	if err := c.parseMessageData(wsMsg.Data, &req); err != nil {
		c.sendError("invalid join room request")
		return
	}

	// Validate room access
	if err := c.handler.chatService.ValidateRoomAccess(ctx, req.RoomID, c.userID, c.userType); err != nil {
		c.sendError("unauthorized room access")
		return
	}

	// Join room in message broker
	c.handler.messageBroker.JoinRoom(c.socketID, req.RoomID)
	c.rooms[req.RoomID] = true

	// Subscribe to room messages
	go c.subscribeToRoomMessages(req.RoomID)

	c.sendSuccess("joined room", map[string]interface{}{
		"room_id": req.RoomID,
	})

	applog.InfoCtx(ctx, "User joined room", 
		"user_id", c.userID, 
		"room_id", req.RoomID, 
		"socket_id", c.socketID)
}

// Handle leave room request
func (c *WebSocketClient) handleLeaveRoom(ctx context.Context, wsMsg dto.WebSocketMessage) {
	var req dto.LeaveRoomRequest
	if err := c.parseMessageData(wsMsg.Data, &req); err != nil {
		c.sendError("invalid leave room request")
		return
	}

	// Leave room in message broker
	c.handler.messageBroker.LeaveRoom(c.socketID, req.RoomID)
	delete(c.rooms, req.RoomID)

	c.sendSuccess("left room", map[string]interface{}{
		"room_id": req.RoomID,
	})

	applog.InfoCtx(ctx, "User left room", 
		"user_id", c.userID, 
		"room_id", req.RoomID, 
		"socket_id", c.socketID)
}

// Handle send message request
func (c *WebSocketClient) handleSendMessage(ctx context.Context, wsMsg dto.WebSocketMessage) {
	var req dto.SendMessageRequest
	if err := c.parseMessageData(wsMsg.Data, &req); err != nil {
		c.sendError("invalid send message request")
		return
	}

	// Send message through service
	response, err := c.handler.chatService.SendMessage(ctx, c.userID, c.userType, req)
	if err != nil {
		c.sendError("failed to send message: " + err.Error())
		return
	}

	c.sendSuccess("message sent", response)
}

// Handle typing indicator
func (c *WebSocketClient) handleTyping(ctx context.Context, wsMsg dto.WebSocketMessage) {
	var req dto.TypingIndicatorRequest
	if err := c.parseMessageData(wsMsg.Data, &req); err != nil {
		c.sendError("invalid typing request")
		return
	}

	// Handle typing through service
	if err := c.handler.chatService.HandleTypingIndicator(ctx, c.userID, c.userType, req); err != nil {
		c.sendError("failed to handle typing: " + err.Error())
		return
	}
}

// Handle mark messages as read
func (c *WebSocketClient) handleMarkRead(ctx context.Context, wsMsg dto.WebSocketMessage) {
	var req dto.MessageReadRequest
	if err := c.parseMessageData(wsMsg.Data, &req); err != nil {
		c.sendError("invalid mark read request")
		return
	}

	// Mark messages as read through service
	responses, err := c.handler.chatService.MarkMessagesAsRead(ctx, c.userID, req)
	if err != nil {
		c.sendError("failed to mark messages as read: " + err.Error())
		return
	}

	c.sendSuccess("messages marked as read", responses)
}

// Handle get messages request
func (c *WebSocketClient) handleGetMessages(ctx context.Context, wsMsg dto.WebSocketMessage) {
	var req dto.GetMessagesRequest
	if err := c.parseMessageData(wsMsg.Data, &req); err != nil {
		c.sendError("invalid get messages request")
		return
	}

	// Get messages through service
	response, err := c.handler.chatService.GetRoomMessages(ctx, req.RoomID, c.userID, c.userType, req)
	if err != nil {
		c.sendError("failed to get messages: " + err.Error())
		return
	}

	c.sendSuccess("messages retrieved", response)
}

// Subscribe to user-specific messages
func (c *WebSocketClient) subscribeToUserMessages() {
	msgChan, err := c.handler.messageBroker.SubscribeToUser(c.ctx, c.userID, c.userType)
	if err != nil {
		applog.LogErrorCtx(c.ctx, err, "Failed to subscribe to user messages", "socket_id", c.socketID)
		return
	}

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			c.forwardMessage(msg)
		}
	}
}

// Subscribe to room-specific messages
func (c *WebSocketClient) subscribeToRoomMessages(roomID string) {
	msgChan, err := c.handler.messageBroker.SubscribeToRoom(c.ctx, roomID)
	if err != nil {
		applog.LogErrorCtx(c.ctx, err, "Failed to subscribe to room messages", 
			"socket_id", c.socketID, 
			"room_id", roomID)
		return
	}

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			// Only forward if still in room
			if c.rooms[roomID] {
				c.forwardMessage(msg)
			}
		}
	}
}

// Forward message to WebSocket client
func (c *WebSocketClient) forwardMessage(msg *service.Message) {
	wsResponse := dto.WebSocketResponse{
		Type:      msg.Channel,
		Success:   true,
		Data:      msg.Data,
		Timestamp: msg.Timestamp,
	}

	responseBytes, err := json.Marshal(wsResponse)
	if err != nil {
		applog.LogErrorCtx(c.ctx, err, "Failed to marshal WebSocket response", "socket_id", c.socketID)
		return
	}

	select {
	case c.send <- responseBytes:
	default:
		// Channel is full, close connection
		close(c.send)
	}
}

// Send success response
func (c *WebSocketClient) sendSuccess(message string, data interface{}) {
	response := dto.WebSocketResponse{
		Type:      dto.WSMsgTypeSuccess,
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		applog.LogErrorCtx(c.ctx, err, "Failed to marshal success response", "socket_id", c.socketID)
		return
	}

	select {
	case c.send <- responseBytes:
	default:
		close(c.send)
	}
}

// Send error response
func (c *WebSocketClient) sendError(errorMsg string) {
	response := dto.WebSocketResponse{
		Type:      dto.WSMsgTypeError,
		Success:   false,
		Error:     errorMsg,
		Timestamp: time.Now(),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		applog.LogErrorCtx(c.ctx, err, "Failed to marshal error response", "socket_id", c.socketID)
		return
	}

	select {
	case c.send <- responseBytes:
	default:
		close(c.send)
	}
}

// Parse message data
func (c *WebSocketClient) parseMessageData(data interface{}, dest interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(dataBytes, dest)
}

// Cleanup client resources
func (c *WebSocketClient) cleanup() {
	// Remove connection from message broker
	c.handler.messageBroker.RemoveConnection(c.socketID)

	// Set user offline
	c.handler.messageBroker.SetUserOffline(c.userID, c.userType, c.socketID)

	applog.InfoCtx(c.ctx, "WebSocket connection closed", 
		"user_id", c.userID, 
		"socket_id", c.socketID)
}

// Generate unique socket ID
func generateSocketID(userID, userType string) string {
	return userID + ":" + userType + ":" + time.Now().Format("20060102150405")
}
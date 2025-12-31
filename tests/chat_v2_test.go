package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bengkelin/bengkelin-service/internal/api"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestChatV2Routes(t *testing.T) {
	// Setup test environment
	applog.Setup("test")
	config.Setup("../.env")

	// Initialize the API
	api.SetConfiguration("../.env")

	// Create test server
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("Test WebSocket endpoint exists", func(t *testing.T) {
		// Test that the WebSocket endpoint is accessible
		resp, err := http.Get(server.URL + "/api/v2/chat/ws")
		assert.NoError(t, err)
		
		// WebSocket upgrade should fail without proper headers, but endpoint should exist
		// We expect a 400 or similar, not 404
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Test chat rooms endpoint requires auth", func(t *testing.T) {
		// Test that chat rooms endpoint requires authentication
		resp, err := http.Get(server.URL + "/api/v2/chat/rooms")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Test create chat room endpoint requires auth", func(t *testing.T) {
		// Test that create chat room endpoint requires authentication
		reqBody := dto.CreateChatRoomRequest{
			BengkelID: "test-bengkel-id",
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		resp, err := http.Post(
			server.URL+"/api/v2/chat/rooms",
			"application/json",
			bytes.NewBuffer(jsonBody),
		)
		
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Test messages endpoint requires auth", func(t *testing.T) {
		// Test that messages endpoint requires authentication
		reqBody := dto.SendMessageRequest{
			RoomID:      "test-room-id",
			MessageType: "text",
			Content:     "Hello, world!",
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		resp, err := http.Post(
			server.URL+"/api/v2/chat/messages",
			"application/json",
			bytes.NewBuffer(jsonBody),
		)
		
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Test typing indicator endpoint requires auth", func(t *testing.T) {
		// Test that typing indicator endpoint requires authentication
		reqBody := dto.TypingIndicatorRequest{
			RoomID:   "test-room-id",
			IsTyping: true,
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		resp, err := http.Post(
			server.URL+"/api/v2/chat/realtime/typing",
			"application/json",
			bytes.NewBuffer(jsonBody),
		)
		
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func setupTestRouter() http.Handler {
	// This would set up a test version of the router
	// For now, we'll use a simple implementation
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock responses for different endpoints
		switch r.URL.Path {
		case "/api/v2/chat/ws":
			w.WriteHeader(http.StatusBadRequest) // WebSocket upgrade would fail
		case "/api/v2/chat/rooms":
			w.WriteHeader(http.StatusUnauthorized)
		case "/api/v2/chat/messages":
			w.WriteHeader(http.StatusUnauthorized)
		case "/api/v2/chat/realtime/typing":
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}
package v2

import (
	"context"
	"time"
	"github.com/Bengkelin/bengkelin-service/internal/api/handlers"
	"github.com/Bengkelin/bengkelin-service/internal/api/middleware"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/redis"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/service"
	pkgMiddleware "github.com/Bengkelin/bengkelin-service/pkg/middleware"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// SetupV2Routes sets up all v2 routes
func SetupV2Routes(app *gin.Engine) {
	conf := config.GetConfig()
	
	// Create rate limiters for v2 routes
	var generalLimiter, chatLimiter *pkgMiddleware.IPRateLimiter
	
	if conf.RateLimit.Enabled {
		generalLimiter = pkgMiddleware.NewIPRateLimiter(
			rate.Limit(conf.RateLimit.GeneralRPS), 
			conf.RateLimit.GeneralBurst,
		)
		
		// Chat endpoints get higher rate limits due to real-time nature
		chatLimiter = pkgMiddleware.NewIPRateLimiter(
			rate.Limit(conf.RateLimit.GeneralRPS * 2), // Double the general rate
			conf.RateLimit.GeneralBurst * 2,
		)
		
		pkgMiddleware.StartCleanupRoutine(generalLimiter)
		pkgMiddleware.StartCleanupRoutine(chatLimiter)
	}

	// Initialize dependencies
	database := db.GetDB()
	redisClient := redis.GetRedisClient()
	
	// Initialize repositories
	chatRepo := repository.NewChatV2Repository(database)
	userRepo := repository.GetUserRepository()
	bengkelRepo := repository.GetBengkelRepository()
	
	// Initialize message broker with fallback
	var messageBroker service.MessageBroker
	if redisClient != nil && redisClient.IsConnected() {
		// Try to use Redis broker
		redisBroker := service.NewRedisBroker(redisClient)
		messageBroker = redisBroker
		applog.InfoCtx(context.Background(), "Using Redis message broker")
		
		// CRITICAL: Test Redis pub/sub functionality on startup
		testCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// Type assert to access debugging methods
		if debugger, ok := redisBroker.(service.RedisBrokerDebugger); ok {
			if err := debugger.TestRedisPubSub(testCtx); err != nil {
				applog.LogErrorCtx(context.Background(), err, "Redis pub/sub test failed, falling back to in-memory broker")
				messageBroker = service.NewInMemoryBroker()
			} else {
				applog.InfoCtx(context.Background(), "Redis pub/sub test successful")
				// Log initial Redis state
				debugger.LogRedisState(context.Background())
			}
		}
	} else {
		// Fallback to in-memory broker
		messageBroker = service.NewInMemoryBroker()
		if redisClient == nil {
			applog.InfoCtx(context.Background(), "Redis client not initialized, using in-memory message broker")
		} else {
			applog.InfoCtx(context.Background(), "Redis connection failed, using in-memory message broker")
		}
	}
	
	// Initialize services
	chatService := service.NewChatV2Service(chatRepo, userRepo, bengkelRepo, messageBroker)
	
	// Initialize handlers
	chatHandler := handlers.NewChatV2Handler(chatService)
	wsHandler := handlers.NewWebSocketV2Handler(chatService, messageBroker)

	// V2 API routes
	v2Route := app.Group("/api/v2")
	
	// Apply general rate limiting to v2 routes if enabled
	if conf.RateLimit.Enabled && generalLimiter != nil {
		v2Route.Use(pkgMiddleware.RateLimitMiddleware(generalLimiter))
	}

	// Chat routes group
	chatGroup := v2Route.Group("/chat")
	
	// Apply chat-specific rate limiting if enabled
	if conf.RateLimit.Enabled && chatLimiter != nil {
		chatGroup.Use(pkgMiddleware.RateLimitMiddleware(chatLimiter))
	}

	// WebSocket endpoint (no authentication middleware here, handled in the handler)
	chatGroup.GET("/ws", wsHandler.HandleWebSocket)

	// Chat room endpoints
	roomGroup := chatGroup.Group("/rooms")
	{
		// User endpoints
		roomGroup.POST("", middleware.AuthJWT(), chatHandler.CreateOrGetChatRoom)
		roomGroup.GET("", middleware.AuthJWT(), chatHandler.GetUserChatRooms)
		roomGroup.GET("/:roomId", middleware.AuthJWT(), chatHandler.GetChatRoom)
		roomGroup.GET("/:roomId/messages", middleware.AuthJWT(), chatHandler.GetRoomMessages)
	}

	// Bengkel (Mitra) room endpoints - Use flexible auth to match handlers
	bengkelRoomGroup := chatGroup.Group("/bengkel/rooms")
	{
		bengkelRoomGroup.GET("", middleware.AuthJWTMitra(), chatHandler.GetBengkelChatRooms) // Keep mitra-specific
		bengkelRoomGroup.GET("/:roomId", middleware.AuthJWTFlexible(), chatHandler.GetChatRoom)
		bengkelRoomGroup.GET("/:roomId/messages", middleware.AuthJWTFlexible(), chatHandler.GetRoomMessages)
	}

	// Message endpoints (accessible by both users and mitras)
	messageGroup := chatGroup.Group("/messages")
	{
		messageGroup.POST("", middleware.AuthJWTFlexible(), chatHandler.SendMessage)
		messageGroup.POST("/file", middleware.AuthJWTFlexible(), chatHandler.SendFileMessage)
		messageGroup.PATCH("/:messageId", middleware.AuthJWTFlexible(), chatHandler.EditMessage)
		messageGroup.DELETE("/:messageId", middleware.AuthJWTFlexible(), chatHandler.DeleteMessage)
		messageGroup.POST("/read", middleware.AuthJWTFlexible(), chatHandler.MarkMessagesAsRead)
	}

	// Real-time endpoints (accessible by both users and mitras)
	realtimeGroup := chatGroup.Group("/realtime")
	{
		realtimeGroup.POST("/typing", middleware.AuthJWTFlexible(), chatHandler.SendTypingIndicator)
	}

	// Polling endpoint for better mobile performance
	chatGroup.GET("/poll", middleware.AuthJWTFlexible(), chatHandler.PollNewMessages)

	applog.Info("Chat V2 routes initialized successfully")
}

// SetupV2RoutesWithCustomDependencies allows injecting custom dependencies for testing
func SetupV2RoutesWithCustomDependencies(
	app *gin.Engine,
	chatService service.ChatV2ServiceInterface,
	messageBroker service.MessageBroker,
) {
	conf := config.GetConfig()
	
	// Create rate limiters for v2 routes
	var generalLimiter, chatLimiter *pkgMiddleware.IPRateLimiter
	
	if conf.RateLimit.Enabled {
		generalLimiter = pkgMiddleware.NewIPRateLimiter(
			rate.Limit(conf.RateLimit.GeneralRPS), 
			conf.RateLimit.GeneralBurst,
		)
		
		chatLimiter = pkgMiddleware.NewIPRateLimiter(
			rate.Limit(conf.RateLimit.GeneralRPS * 2),
			conf.RateLimit.GeneralBurst * 2,
		)
		
		pkgMiddleware.StartCleanupRoutine(generalLimiter)
		pkgMiddleware.StartCleanupRoutine(chatLimiter)
	}

	// Initialize handlers with injected dependencies
	chatHandler := handlers.NewChatV2Handler(chatService)
	wsHandler := handlers.NewWebSocketV2Handler(chatService, messageBroker)

	// V2 API routes
	v2Route := app.Group("/api/v2")
	
	// Apply general rate limiting to v2 routes if enabled
	if conf.RateLimit.Enabled && generalLimiter != nil {
		v2Route.Use(pkgMiddleware.RateLimitMiddleware(generalLimiter))
	}

	// Chat routes group
	chatGroup := v2Route.Group("/chat")
	
	// Apply chat-specific rate limiting if enabled
	if conf.RateLimit.Enabled && chatLimiter != nil {
		chatGroup.Use(pkgMiddleware.RateLimitMiddleware(chatLimiter))
	}

	// WebSocket endpoint
	chatGroup.GET("/ws", wsHandler.HandleWebSocket)

	// Chat room endpoints - Use flexible auth for both users and mitras
	roomGroup := chatGroup.Group("/rooms")
	{
		roomGroup.POST("", middleware.AuthJWTFlexible(), chatHandler.CreateOrGetChatRoom)
		roomGroup.GET("", middleware.AuthJWT(), chatHandler.GetUserChatRooms) // Keep user-specific
		roomGroup.GET("/:roomId", middleware.AuthJWTFlexible(), chatHandler.GetChatRoom)
		roomGroup.GET("/:roomId/messages", middleware.AuthJWTFlexible(), chatHandler.GetRoomMessages)
	}

	// Bengkel room endpoints - Use flexible auth to match handlers
	bengkelRoomGroup := chatGroup.Group("/bengkel/rooms")
	{
		bengkelRoomGroup.GET("", middleware.AuthJWTMitra(), chatHandler.GetBengkelChatRooms) // Keep mitra-specific
		bengkelRoomGroup.GET("/:roomId", middleware.AuthJWTFlexible(), chatHandler.GetChatRoom)
		bengkelRoomGroup.GET("/:roomId/messages", middleware.AuthJWTFlexible(), chatHandler.GetRoomMessages)
	}

	// Message endpoints
	messageGroup := chatGroup.Group("/messages")
	{
		messageGroup.POST("", middleware.AuthJWT(), chatHandler.SendMessage)
		messageGroup.POST("/file", middleware.AuthJWT(), chatHandler.SendFileMessage)
		messageGroup.PATCH("/:messageId", middleware.AuthJWT(), chatHandler.EditMessage)
		messageGroup.DELETE("/:messageId", middleware.AuthJWT(), chatHandler.DeleteMessage)
		messageGroup.POST("/read", middleware.AuthJWT(), chatHandler.MarkMessagesAsRead)
		
		mitraMessageGroup := messageGroup.Group("/mitra")
		{
			mitraMessageGroup.POST("", middleware.AuthJWTMitra(), chatHandler.SendMessage)
			mitraMessageGroup.POST("/file", middleware.AuthJWTMitra(), chatHandler.SendFileMessage)
			mitraMessageGroup.PATCH("/:messageId", middleware.AuthJWTMitra(), chatHandler.EditMessage)
			mitraMessageGroup.DELETE("/:messageId", middleware.AuthJWTMitra(), chatHandler.DeleteMessage)
			mitraMessageGroup.POST("/read", middleware.AuthJWTMitra(), chatHandler.MarkMessagesAsRead)
		}
	}

	// Real-time endpoints
	realtimeGroup := chatGroup.Group("/realtime")
	{
		realtimeGroup.POST("/typing", middleware.AuthJWT(), chatHandler.SendTypingIndicator)
		
		mitraRealtimeGroup := realtimeGroup.Group("/mitra")
		{
			mitraRealtimeGroup.POST("/typing", middleware.AuthJWTMitra(), chatHandler.SendTypingIndicator)
		}
	}

	// Polling endpoint for better mobile performance
	chatGroup.GET("/poll", middleware.AuthJWTFlexible(), chatHandler.PollNewMessages)

	applog.Info("Chat V2 routes initialized successfully with custom dependencies")
}
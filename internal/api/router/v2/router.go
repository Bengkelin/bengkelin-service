package v2

import (
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
	
	// Initialize message broker
	messageBroker := service.NewRedisBroker(redisClient)
	
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

	// Bengkel (Mitra) room endpoints
	bengkelRoomGroup := chatGroup.Group("/bengkel/rooms")
	{
		bengkelRoomGroup.GET("", middleware.AuthJWTMitra(), chatHandler.GetBengkelChatRooms)
		bengkelRoomGroup.GET("/:roomId", middleware.AuthJWTMitra(), chatHandler.GetChatRoom)
		bengkelRoomGroup.GET("/:roomId/messages", middleware.AuthJWTMitra(), chatHandler.GetRoomMessages)
	}

	// Message endpoints (accessible by both users and mitras)
	messageGroup := chatGroup.Group("/messages")
	{
		// User message endpoints
		messageGroup.POST("", middleware.AuthJWT(), chatHandler.SendMessage)
		messageGroup.POST("/file", middleware.AuthJWT(), chatHandler.SendFileMessage)
		messageGroup.PATCH("/:messageId", middleware.AuthJWT(), chatHandler.EditMessage)
		messageGroup.DELETE("/:messageId", middleware.AuthJWT(), chatHandler.DeleteMessage)
		messageGroup.POST("/read", middleware.AuthJWT(), chatHandler.MarkMessagesAsRead)
		
		// Mitra message endpoints (same handlers, different middleware)
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
		// User real-time endpoints
		realtimeGroup.POST("/typing", middleware.AuthJWT(), chatHandler.SendTypingIndicator)
		
		// Mitra real-time endpoints
		mitraRealtimeGroup := realtimeGroup.Group("/mitra")
		{
			mitraRealtimeGroup.POST("/typing", middleware.AuthJWTMitra(), chatHandler.SendTypingIndicator)
		}
	}

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

	// Chat room endpoints
	roomGroup := chatGroup.Group("/rooms")
	{
		roomGroup.POST("", middleware.AuthJWT(), chatHandler.CreateOrGetChatRoom)
		roomGroup.GET("", middleware.AuthJWT(), chatHandler.GetUserChatRooms)
		roomGroup.GET("/:roomId", middleware.AuthJWT(), chatHandler.GetChatRoom)
		roomGroup.GET("/:roomId/messages", middleware.AuthJWT(), chatHandler.GetRoomMessages)
	}

	// Bengkel room endpoints
	bengkelRoomGroup := chatGroup.Group("/bengkel/rooms")
	{
		bengkelRoomGroup.GET("", middleware.AuthJWTMitra(), chatHandler.GetBengkelChatRooms)
		bengkelRoomGroup.GET("/:roomId", middleware.AuthJWTMitra(), chatHandler.GetChatRoom)
		bengkelRoomGroup.GET("/:roomId/messages", middleware.AuthJWTMitra(), chatHandler.GetRoomMessages)
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

	applog.Info("Chat V2 routes initialized successfully with custom dependencies")
}
package v1

import (
	"net/http"

	"github.com/Bengkelin/bengkelin-service/internal/api/handlers"
	"github.com/Bengkelin/bengkelin-service/internal/api/middleware"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	pkgMiddleware "github.com/Bengkelin/bengkelin-service/pkg/middleware"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
)

func Setup() *gin.Engine {
	app := gin.New()

	// Get configuration
	conf := config.GetConfig()
	
	// Create rate limiters based on configuration
	var generalLimiter, authLimiter, strictLimiter *pkgMiddleware.IPRateLimiter
	
	if conf.RateLimit.Enabled {
		applog.Info("Rate limiting enabled", 
			"general_rps", conf.RateLimit.GeneralRPS,
			"auth_rps", conf.RateLimit.AuthRPS,
			"strict_rps", conf.RateLimit.StrictRPS,
		)
		
		generalLimiter = pkgMiddleware.NewIPRateLimiter(
			rate.Limit(conf.RateLimit.GeneralRPS), 
			conf.RateLimit.GeneralBurst,
		)
		
		authLimiter = pkgMiddleware.NewIPRateLimiter(
			rate.Limit(conf.RateLimit.AuthRPS), 
			conf.RateLimit.AuthBurst,
		)
		
		strictLimiter = pkgMiddleware.NewIPRateLimiter(
			rate.Limit(conf.RateLimit.StrictRPS), 
			conf.RateLimit.StrictBurst,
		)
		
		// Start cleanup routines
		pkgMiddleware.StartCleanupRoutine(generalLimiter)
		pkgMiddleware.StartCleanupRoutine(authLimiter)
		pkgMiddleware.StartCleanupRoutine(strictLimiter)
	} else {
		applog.Info("Rate limiting disabled")
	}

	// Global middlewares
	app.Use(gin.Recovery())
	app.Use(gin.Logger()) // Add colorful Gin logger
	
	// Enhanced logging middlewares
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.LoggingMiddleware())
	app.Use(middleware.UserContextMiddleware())
	app.Use(middleware.ErrorLoggingMiddleware())
	app.Use(middleware.SecurityLoggingMiddleware())
	app.Use(middleware.PerformanceLoggingMiddleware())
	
	// Metrics middleware for Prometheus
	app.Use(middleware.PrometheusMiddleware())
	
	// Add security headers
	app.Use(middleware.SecurityHeadersMiddleware())
	
	// CORS middleware
	app.Use(middleware.CORS())
	
	// Apply general rate limiting to all routes if enabled
	if conf.RateLimit.Enabled && generalLimiter != nil {
		app.Use(pkgMiddleware.RateLimitMiddleware(generalLimiter))
	}
	
	app.NoMethod(middleware.NoMethodHandler())
	app.NoRoute(middleware.NoRouteHandler())

	// Swagger documentation (only in development and staging)
	if conf.App.Environment != "production" {
		app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		applog.Info("Swagger documentation enabled", "url", "/swagger/index.html")
	}

	// Health check endpoints (always available)
	healthHandler := handlers.GetHealthHandler()
	app.GET("/health", healthHandler.HealthCheck)
	app.GET("/ready", healthHandler.ReadinessCheck)
	app.GET("/live", healthHandler.LivenessCheck)

	// Metrics endpoints
	metricsHandler := handlers.GetMetricsHandler()
	app.GET("/metrics", metricsHandler.PrometheusMetrics)
	app.GET("/metrics/app", metricsHandler.ApplicationMetrics)

	// Routes for v1
	v1Route := app.Group("/api/v1")

	v1Route.StaticFS("/static/vehicle", http.Dir("public/vehicles"))
	v1Route.StaticFS("/static/bengkel", http.Dir("public/bengkels"))
	v1Route.StaticFS("/static/avatar", http.Dir("public/avatars"))
	
	// AuthGroup with "auth" prefix - Apply stricter rate limiting
	authGroup := v1Route.Group("users/auth")
	if conf.RateLimit.Enabled && authLimiter != nil {
		authGroup.Use(pkgMiddleware.RateLimitMiddleware(authLimiter)) // Additional auth rate limiting
	}
	authHandler := handlers.GetAuthHandler()
	{
		// Login and register get even stricter limits
		if conf.RateLimit.Enabled && strictLimiter != nil {
			authGroup.POST("login", pkgMiddleware.RateLimitMiddleware(strictLimiter), authHandler.UsersAuthLogin)
			authGroup.POST("register", pkgMiddleware.RateLimitMiddleware(strictLimiter), authHandler.UsersAuthRegister)
		} else {
			authGroup.POST("login", authHandler.UsersAuthLogin)
			authGroup.POST("register", authHandler.UsersAuthRegister)
		}
		authGroup.POST("google", authHandler.UsersAuthGoogle)
		authGroup.POST("refresh", authHandler.UsersRefreshToken)
		authGroup.POST("logout", authHandler.UsersLogout)
		authGroup.POST("logout-all", middleware.AuthJWT(), authHandler.UsersLogoutAll)
		authGroup.POST("address", middleware.AuthJWT(), authHandler.UsersNewAddress)
		authGroup.POST("vehicle", middleware.AuthJWT(), authHandler.UsersNewVehicle)
	}

	// auth mitra group with "auth/mitra" prefix - Apply stricter rate limiting
	authMitraGroup := v1Route.Group("mitras/auth")
	if conf.RateLimit.Enabled && authLimiter != nil {
		authMitraGroup.Use(pkgMiddleware.RateLimitMiddleware(authLimiter)) // Additional auth rate limiting
	}
	{
		// Login and register get even stricter limits
		if conf.RateLimit.Enabled && strictLimiter != nil {
			authMitraGroup.POST("login", pkgMiddleware.RateLimitMiddleware(strictLimiter), authHandler.MitrasAuthLogin)
			authMitraGroup.POST("register", pkgMiddleware.RateLimitMiddleware(strictLimiter), authHandler.MitrasAuthRegister)
		} else {
			authMitraGroup.POST("login", authHandler.MitrasAuthLogin)
			authMitraGroup.POST("register", authHandler.MitrasAuthRegister)
		}
		authMitraGroup.POST("google", authHandler.MitrasAuthGoogle)
		authMitraGroup.POST("refresh", authHandler.MitrasRefreshToken)
		authMitraGroup.POST("logout", authHandler.MitrasLogout)
		authMitraGroup.POST("logout-all", middleware.AuthJWTMitra(), authHandler.MitrasLogoutAll)
		authMitraGroup.POST("bank", middleware.AuthJWTMitra(), authHandler.MitrasNewBank)
		authMitraGroup.PATCH("bank", middleware.AuthJWTMitra(), authHandler.MitrasUpdateBank)
		authMitraGroup.PATCH("profile", middleware.AuthJWTMitra(), authHandler.MitrasUpdateProfile)
	}
	// UserGroup with "user" prefix
	userGroup := v1Route.Group("users")
	userHandler := handlers.GetUserHandler()
	{
		userGroup.GET("profile", middleware.AuthJWT(), userHandler.GetProfile)
		userGroup.PATCH("profile", middleware.AuthJWT(), userHandler.UpdateProfile)
		userGroup.PATCH("avatar", middleware.AuthJWT(), userHandler.UpdateAvatarUser)
		userGroup.POST("address", middleware.AuthJWT(), userHandler.CreateAddress)
		userGroup.GET("address/:addressId", middleware.AuthJWT(), userHandler.GetDetailAddressUser)
		userGroup.PATCH("address/:addressId", middleware.AuthJWT(), userHandler.UpdateAddress)
		userGroup.PATCH("address", middleware.AuthJWT(), userHandler.UpdateOrCreateAddress)
		userGroup.DELETE("address/:addressId", middleware.AuthJWT(), userHandler.DeleteAddressUser)
		userGroup.POST("vehicle", middleware.AuthJWT(), userHandler.CreateVehicle)
		userGroup.GET("vehicle", middleware.AuthJWT(), userHandler.GetAllVehiclesUser)
		userGroup.GET("vehicle/:vehicleId", middleware.AuthJWT(), userHandler.GetDetailVehicleUser)
		userGroup.DELETE("vehicle/:vehicleId", middleware.AuthJWT(), userHandler.DeleteVehicleUser)
	}

	// MitraGroup with "mitra" prefix
	mitraGroup := v1Route.Group("bengkels")
	mitraHandler := handlers.GetBengkelHandler()
	{
		mitraGroup.POST("new", middleware.AuthJWTMitra(), mitraHandler.CreateBengkel)
		mitraGroup.GET("profile", middleware.AuthJWTMitra(), mitraHandler.GetBengkel)
		mitraGroup.PATCH("profile", middleware.AuthJWTMitra(), mitraHandler.UpdateBengkel)
		mitraGroup.GET("", middleware.AuthJWT(), mitraHandler.GetAllBengkelPaginate)
		mitraGroup.POST("address", middleware.AuthJWTMitra(), mitraHandler.CreateBengkelAddress)
		mitraGroup.POST("service", middleware.AuthJWTMitra(), mitraHandler.CreateBengkelService)
		mitraGroup.POST("photo", middleware.AuthJWTMitra(), mitraHandler.CreateBengkelPhoto)
		mitraGroup.PATCH("service/opsi", middleware.AuthJWTMitra(), mitraHandler.UpdateBengkelStatusOpsiService)
		mitraGroup.PATCH("montir", middleware.AuthJWTMitra(), mitraHandler.UpdateBengkelMontir)
		mitraGroup.PATCH("operasional", middleware.AuthJWTMitra(), mitraHandler.UpdateBengkelOperational)
		mitraGroup.GET("search", middleware.AuthJWT(), mitraHandler.GetBengkelSearchV2Paginate)
		mitraGroup.POST("testimoni/:bengkelId", middleware.AuthJWT(), mitraHandler.CreateBengkelTestimonial)
		mitraGroup.GET("testimoni/:bengkelId", middleware.AuthJWT(), mitraHandler.GetDetailBengkelById)
		mitraGroup.PATCH("avatar", middleware.AuthJWTMitra(), mitraHandler.UpdateAvatarBengkel)
		mitraGroup.POST("order/service/:userId", middleware.AuthJWTMitra(), mitraHandler.CreateBengkelOrderService)
		mitraGroup.GET("order/service/:pesananId", middleware.AuthJWT(), mitraHandler.GetBengkelOrderServiceById)
		mitraGroup.GET("orders/list/user", middleware.AuthJWT(), mitraHandler.GetAllOrderUserPaginate)
		mitraGroup.GET("orders/list/mitra", middleware.AuthJWTMitra(), mitraHandler.GetAllBengkelOrderServicePaginate)
		mitraGroup.GET("order/mitra/service/:pesananId", middleware.AuthJWTMitra(), mitraHandler.GetBengkelOrderServiceByIdMitra)
		mitraGroup.GET("order/schedule", middleware.AuthJWT(), mitraHandler.GetBengkelOperationalByIdAndDay)
		mitraGroup.PATCH("order/service/:pesananId", middleware.AuthJWT(), mitraHandler.UpdateBengkelOrderServiceById)
		mitraGroup.PATCH("order/status/:pesananId", middleware.AuthJWTMitra(), mitraHandler.UpdateStatusOrderService)
		mitraGroup.GET("order/user/:userId", middleware.AuthJWTMitra(), mitraHandler.GetDetailUserById)
		mitraGroup.GET("nearest", middleware.AuthJWT(), mitraHandler.GetNearestBengkelPaginate)
	}

	// ChatGroup with "chat" prefix
	chatGroup := v1Route.Group("chats")
	chatHandler := handlers.GetChatHandler()
	{
		chatGroup.GET("appToken", middleware.AuthJWT(), chatHandler.CreateAppToken)
		chatGroup.GET("chatToken", middleware.AuthJWT(), chatHandler.CreateChatToken)
		chatGroup.GET("chatTokenMitra", middleware.AuthJWTMitra(), chatHandler.CreateChatTokenMitra)
		chatGroup.GET("appTokenMitra", middleware.AuthJWTMitra(), chatHandler.CreateAppTokenMitra)
		chatGroup.POST("user/history", middleware.AuthJWT(), chatHandler.CreateChatHistoryUser)
		chatGroup.POST("bengkel/history", middleware.AuthJWTMitra(), chatHandler.CreateChatHistoryBengkel)
		chatGroup.GET("user/history", middleware.AuthJWT(), chatHandler.GetChatHistoryUser)
		chatGroup.GET("bengkel/history", middleware.AuthJWTMitra(), chatHandler.GetChatHistoryBengkel)
	}

	// admin group with "admin" prefix
	adminGroup := v1Route.Group("admins")
	adminHandler := handlers.GetAdminFeeHandler()
	{
		adminGroup.POST("fee", adminHandler.CreateAdminFee)
	}

	return app
}

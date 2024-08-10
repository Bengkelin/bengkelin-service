package v1

import (
	"fmt"
	"net/http"

	"github.com/Bengkelin/bengkelin-service/internal/api/handlers"
	"github.com/Bengkelin/bengkelin-service/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	app := gin.New()

	// Middlewares
	app.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - - [%s] \"%s %s %s %d %s \" \" %s\" \" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	app.Use(gin.Recovery())
	app.NoMethod(middleware.NoMethodHandler())
	app.NoRoute(middleware.NoRouteHandler())

	// Routes for v1
	v1Route := app.Group("/api/v1")

	v1Route.StaticFS("/static/vehicle", http.Dir("public/vehicles"))
	v1Route.StaticFS("/static/bengkel", http.Dir("public/bengkels"))
	v1Route.StaticFS("/static/avatar", http.Dir("public/avatars"))
	// AuthGroup with "auth" prefix
	authGroup := v1Route.Group("users/auth")
	authHandler := handlers.GetAuthHandler()
	{
		authGroup.POST("login", authHandler.UsersAuthLogin)
		authGroup.POST("register", authHandler.UsersAuthRegister)
		authGroup.POST("address", middleware.AuthJWT(), authHandler.UsersNewAddress)
		authGroup.POST("vehicle", middleware.AuthJWT(), authHandler.UsersNewVehicle)
	}

	// auth mitra group with "auth/mitra" prefix
	authMitraGroup := v1Route.Group("mitras/auth")
	{
		authMitraGroup.POST("login", authHandler.MitrasAuthLogin)
		authMitraGroup.POST("register", authHandler.MitrasAuthRegister)
		authMitraGroup.POST("bank", middleware.AuthJWTMitra(), authHandler.MitrasNewBank)
	}
	// UserGroup with "user" prefix
	userGroup := v1Route.Group("users")
	userHandler := handlers.GetUserHandler()
	{
		userGroup.GET("profile", middleware.AuthJWT(), userHandler.GetProfile)
		userGroup.PATCH("profile", middleware.AuthJWT(), userHandler.UpdateProfile)
		userGroup.PATCH("avatar", middleware.AuthJWT(), userHandler.UpdateAvatarUser)
	}

	// MitraGroup with "mitra" prefix
	mitraGroup := v1Route.Group("bengkels")
	mitraHandler := handlers.GetBengkelHandler()
	{
		mitraGroup.POST("new", middleware.AuthJWTMitra(), mitraHandler.CreateBengkel)
		mitraGroup.GET("profile", middleware.AuthJWTMitra(), mitraHandler.GetBengkel)
		mitraGroup.GET("", middleware.AuthJWT(), mitraHandler.GetAllBengkelPaginate)
		mitraGroup.POST("address", middleware.AuthJWTMitra(), mitraHandler.CreateBengkelAddress)
		mitraGroup.POST("service", middleware.AuthJWTMitra(), mitraHandler.CreateBengkelService)
		mitraGroup.POST("photo", middleware.AuthJWTMitra(), mitraHandler.CreateBengkelPhoto)
		mitraGroup.PATCH("service/opsi", middleware.AuthJWTMitra(), mitraHandler.UpdateBengkelStatusOpsiService)
		mitraGroup.GET("search", middleware.AuthJWT(), mitraHandler.GetBengkelSearchV2Paginate)
		mitraGroup.POST("testimoni/:bengkelId", middleware.AuthJWT(), mitraHandler.CreateBengkelTestimoni)
		mitraGroup.GET("testimoni/:bengkelId", middleware.AuthJWT(), mitraHandler.GetDetailBengkelById)
		mitraGroup.PATCH("avatar", middleware.AuthJWTMitra(), mitraHandler.UpdateAvatarBengkel)
		mitraGroup.POST("order/service/:userId", middleware.AuthJWTMitra(), mitraHandler.CreateBengkelPesananService)
		mitraGroup.GET("order/service/:pesananId", middleware.AuthJWT(), mitraHandler.GetBengkelPesananServiceById)
		mitraGroup.GET("order/schedule", middleware.AuthJWT(), mitraHandler.GetBengkelOperasionalByIdAndDay)
		mitraGroup.PATCH("order/service/:pesananId", middleware.AuthJWT(), mitraHandler.UpdateBengkelPesananServiceById)
	}

	// ChatGroup with "chat" prefix
	chatGroup := v1Route.Group("chats")
	chatHandler := handlers.GetChatHandler()
	{
		chatGroup.GET("appToken", middleware.AuthJWT(), chatHandler.CreateAppToken)
		chatGroup.GET("chatToken", middleware.AuthJWT(), chatHandler.CreateChatToken)
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

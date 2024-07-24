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

	v1Route.StaticFS("/static", http.Dir("public/vehicles"))
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
	}
	// UserGroup with "user" prefix
	userGroup := v1Route.Group("users")
	userHandler := handlers.GetUserHandler()
	{
		userGroup.GET("profile", middleware.AuthJWT(), userHandler.GetProfile)
		userGroup.PATCH("profile", middleware.AuthJWT(), userHandler.UpdateProfile)
	}

	return app
}

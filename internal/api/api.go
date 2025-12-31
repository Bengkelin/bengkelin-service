package api

import (
	v1 "github.com/Bengkelin/bengkelin-service/internal/api/router/v1"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	redisClient "github.com/Bengkelin/bengkelin-service/internal/pkg/redis"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/service"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"github.com/gin-gonic/gin"
)

// Set configuration
// Change this func to "exported"  to make Test package can access it
func SetConfiguration(configPath string) {
	// Setup logger first
	applog.Setup("development")

	applog.Info("Initializing application configuration")
	// Setup config from path
	// Default is .env in root folder
	config.Setup(configPath)
	applog.Debug("Configuration loaded successfully")

	// Calling setup db
	db.SetupDB()
	applog.Info("Database setup completed")

	// Setup Redis connection
	conf := config.GetConfig()
	if conf.Redis.Enabled {
		if conf.Redis.URL != "" {
			// Use Redis URL format (redis://user:pass@host:port/db)
			redisClient.SetupRedisFromURL(conf.Redis.URL)
			applog.Info("Redis setup completed using URL", "url", conf.Redis.URL)
		} else {
			// Use individual parameters
			redisAddr := conf.Redis.Host + ":" + conf.Redis.Port
			redisClient.SetupRedis(redisAddr, conf.Redis.Password, conf.Redis.DB)
			applog.Info("Redis setup completed", "host", conf.Redis.Host, "port", conf.Redis.Port, "db", conf.Redis.DB)
		}
	} else {
		applog.Info("Redis is disabled")
	}

	// Start cleanup service for expired refresh tokens
	cleanupService := service.GetCleanupService()
	cleanupService.StartPeriodicCleanup()
	applog.Info("Cleanup service started")

	gin.SetMode(config.GetConfig().Server.Mode)
	applog.Debug("Gin mode set", "mode", config.GetConfig().Server.Mode)
}

// Run the new API with designated configuration
func Run(configPath string) {
	if configPath == "" {
		configPath = ".env"
	}
	SetConfiguration(configPath)
	conf := config.GetConfig()

	// Routing
	web := v1.Setup()
	applog.Info("Starting API server", "port", conf.Server.Port, "mode", conf.Server.Mode)
	applog.Info("==================>")
	_ = web.Run(":" + conf.Server.Port)
}

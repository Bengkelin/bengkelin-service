package api

import (
	v1 "github.com/Bengkelin/bengkelin-service/internal/api/router/v1"
	v2 "github.com/Bengkelin/bengkelin-service/internal/api/router/v2"
	"github.com/Bengkelin/bengkelin-service/internal/config"
	"github.com/Bengkelin/bengkelin-service/internal/container"
	"github.com/Bengkelin/bengkelin-service/internal/db"
	"github.com/Bengkelin/bengkelin-service/internal/rabbitmq"
	redisClient "github.com/Bengkelin/bengkelin-service/internal/redis"
	"github.com/Bengkelin/bengkelin-service/internal/service"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
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
	cleanupService := container.GetContainer().CleanupService
	cleanupService.StartPeriodicCleanup()
	applog.Info("Cleanup service started")

	// Setup RabbitMQ connection
	if conf.RabbitMQ.Enabled {
		err := rabbitmq.Setup()
		if err != nil {
			applog.LogError(err, "Failed to setup RabbitMQ")
		} else {
			applog.Info("RabbitMQ setup completed")
			
			// Start message consumers
			go func() {
				if err := service.StartNotificationConsumers(); err != nil {
					applog.LogError(err, "Failed to start notification consumers")
				}
			}()
			
			go func() {
				if err := service.StartOrderProcessingConsumers(); err != nil {
					applog.LogError(err, "Failed to start order processing consumers")
				}
			}()
			
			go func() {
				if err := service.StartFileProcessingConsumers(); err != nil {
					applog.LogError(err, "Failed to start file processing consumers")
				}
			}()
			
			go func() {
				if err := service.StartAuditConsumers(); err != nil {
					applog.LogError(err, "Failed to start audit consumers")
				}
			}()
			
			applog.Info("All RabbitMQ consumers started successfully")
		}
	} else {
		applog.Info("RabbitMQ is disabled")
	}

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
	
	// Setup V2 routes
	v2.SetupV2Routes(web)
	
	applog.Info("Starting API server", "port", conf.Server.Port, "mode", conf.Server.Mode)
	applog.Info("==================>")
	_ = web.Run(":" + conf.Server.Port)
}

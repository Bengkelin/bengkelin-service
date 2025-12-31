package handlers

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	redisClient "github.com/Bengkelin/bengkelin-service/internal/pkg/redis"
	"github.com/Bengkelin/bengkelin-service/pkg/response"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	BaseHandler
}

// HealthHandlerInterface defines health handler methods
type HealthHandlerInterface interface {
	HealthCheck(c *gin.Context)
	ReadinessCheck(c *gin.Context)
	LivenessCheck(c *gin.Context)
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() HealthHandlerInterface {
	return &HealthHandler{
		BaseHandler: BaseHandler{},
	}
}

// HealthStatus represents the health status response
type HealthStatus struct {
	Status      string                 `json:"status" example:"healthy"`
	Timestamp   time.Time              `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	Version     string                 `json:"version" example:"1.0.0"`
	Environment string                 `json:"environment" example:"production"`
	Uptime      string                 `json:"uptime" example:"1h30m45s"`
	Checks      map[string]CheckResult `json:"checks"`
}

// CheckResult represents individual health check result
type CheckResult struct {
	Status    string `json:"status" example:"healthy"`
	Message   string `json:"message,omitempty" example:"Connection successful"`
	Duration  string `json:"duration" example:"5ms"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T00:00:00Z"`
}

var startTime = time.Now()

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Get the health status of the application and its dependencies
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=HealthStatus} "Health status"
// @Failure 503 {object} response.Response "Service unavailable"
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	start := time.Now()
	
	applog.InfoCtx(ctx, "Health check requested")
	
	status := &HealthStatus{
		Status:      "healthy",
		Timestamp:   time.Now(),
		Version:     getAppVersion(),
		Environment: config.GetConfig().App.Environment,
		Uptime:      time.Since(startTime).String(),
		Checks:      make(map[string]CheckResult),
	}
	
	// Check database connectivity
	dbCheck := h.checkDatabase(ctx)
	status.Checks["database"] = dbCheck
	
	// Check Redis connectivity
	redisCheck := h.checkRedis(ctx)
	status.Checks["redis"] = redisCheck
	
	// Check system resources
	systemCheck := h.checkSystem(ctx)
	status.Checks["system"] = systemCheck
	
	// Determine overall status
	overallHealthy := true
	for _, check := range status.Checks {
		if check.Status != "healthy" {
			overallHealthy = false
			break
		}
	}
	
	if !overallHealthy {
		status.Status = "unhealthy"
		applog.WarnCtx(ctx, "Health check failed", 
			"duration", time.Since(start),
			"checks", status.Checks,
		)
		resp := response.BuildFailedResponse("Service unhealthy", status)
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}
	
	applog.InfoCtx(ctx, "Health check completed successfully", 
		"duration", time.Since(start),
	)
	
	resp := response.BuildSuccessResponse("Service healthy", status)
	c.JSON(http.StatusOK, resp)
}

// ReadinessCheck godoc
// @Summary Readiness check endpoint
// @Description Check if the application is ready to serve requests
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=HealthStatus} "Service ready"
// @Failure 503 {object} response.Response "Service not ready"
// @Router /ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx := c.Request.Context()
	start := time.Now()
	
	status := &HealthStatus{
		Status:      "ready",
		Timestamp:   time.Now(),
		Version:     getAppVersion(),
		Environment: config.GetConfig().App.Environment,
		Uptime:      time.Since(startTime).String(),
		Checks:      make(map[string]CheckResult),
	}
	
	// Check critical dependencies for readiness
	dbCheck := h.checkDatabase(ctx)
	status.Checks["database"] = dbCheck
	
	redisCheck := h.checkRedis(ctx)
	status.Checks["redis"] = redisCheck
	
	// Readiness requires all critical services to be healthy or disabled
	ready := dbCheck.Status == "healthy" && (redisCheck.Status == "healthy" || redisCheck.Status == "disabled")
	
	if !ready {
		status.Status = "not_ready"
		applog.WarnCtx(ctx, "Readiness check failed", 
			"duration", time.Since(start),
		)
		resp := response.BuildFailedResponse("Service not ready", status)
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}
	
	applog.DebugCtx(ctx, "Readiness check completed", 
		"duration", time.Since(start),
	)
	
	resp := response.BuildSuccessResponse("Service ready", status)
	c.JSON(http.StatusOK, resp)
}

// LivenessCheck godoc
// @Summary Liveness check endpoint
// @Description Check if the application is alive and responding
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=HealthStatus} "Service alive"
// @Router /live [get]
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	ctx := c.Request.Context()
	
	status := &HealthStatus{
		Status:      "alive",
		Timestamp:   time.Now(),
		Version:     getAppVersion(),
		Environment: config.GetConfig().App.Environment,
		Uptime:      time.Since(startTime).String(),
		Checks:      make(map[string]CheckResult),
	}
	
	// Basic system check for liveness
	systemCheck := h.checkSystem(ctx)
	status.Checks["system"] = systemCheck
	
	applog.DebugCtx(ctx, "Liveness check completed")
	
	resp := response.BuildSuccessResponse("Service alive", status)
	c.JSON(http.StatusOK, resp)
}

// checkDatabase checks database connectivity
func (h *HealthHandler) checkDatabase(ctx context.Context) CheckResult {
	start := time.Now()
	
	db := db.GetDB()
	if db == nil {
		return CheckResult{
			Status:    "unhealthy",
			Message:   "Database connection not initialized",
			Duration:  time.Since(start).String(),
			Timestamp: time.Now(),
		}
	}
	
	// Get underlying sql.DB for ping
	sqlDB, err := db.DB()
	if err != nil {
		return CheckResult{
			Status:    "unhealthy",
			Message:   "Failed to get database instance: " + err.Error(),
			Duration:  time.Since(start).String(),
			Timestamp: time.Now(),
		}
	}
	
	// Create context with timeout for database ping
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return CheckResult{
			Status:    "unhealthy",
			Message:   "Database ping failed: " + err.Error(),
			Duration:  time.Since(start).String(),
			Timestamp: time.Now(),
		}
	}
	
	// Check database stats
	stats := sqlDB.Stats()
	if stats.OpenConnections == 0 {
		return CheckResult{
			Status:    "unhealthy",
			Message:   "No open database connections",
			Duration:  time.Since(start).String(),
			Timestamp: time.Now(),
		}
	}
	
	return CheckResult{
		Status:    "healthy",
		Message:   "Database connection successful",
		Duration:  time.Since(start).String(),
		Timestamp: time.Now(),
	}
}

// checkRedis checks Redis connectivity
func (h *HealthHandler) checkRedis(ctx context.Context) CheckResult {
	start := time.Now()
	
	// Check if Redis is enabled in configuration
	conf := config.GetConfig()
	if !conf.Redis.Enabled {
		return CheckResult{
			Status:    "disabled",
			Message:   "Redis is disabled in configuration",
			Duration:  time.Since(start).String(),
			Timestamp: time.Now(),
		}
	}
	
	redisCache := redisClient.GetRedisClient()
	if redisCache == nil {
		return CheckResult{
			Status:    "unhealthy",
			Message:   "Redis client not initialized",
			Duration:  time.Since(start).String(),
			Timestamp: time.Now(),
		}
	}
	
	// Test Redis connectivity by setting and getting a test key
	testKey := "health_check_test"
	testValue := "ok"
	
	// Test set operation with shorter timeout for health check
	testCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	if err := redisCache.SetWithContext(testCtx, testKey, testValue, time.Minute); err != nil {
		return CheckResult{
			Status:    "unhealthy",
			Message:   "Redis set operation failed: " + err.Error(),
			Duration:  time.Since(start).String(),
			Timestamp: time.Now(),
		}
	}
	
	// Test get operation
	var result string
	if err := redisCache.GetWithContext(testCtx, testKey, &result); err != nil {
		return CheckResult{
			Status:    "unhealthy",
			Message:   "Redis get operation failed: " + err.Error(),
			Duration:  time.Since(start).String(),
			Timestamp: time.Now(),
		}
	}
	
	// Clean up test key
	redisCache.DeleteWithContext(testCtx, testKey)
	
	if result != testValue {
		return CheckResult{
			Status:    "unhealthy",
			Message:   "Redis returned unexpected value",
			Duration:  time.Since(start).String(),
			Timestamp: time.Now(),
		}
	}
	
	return CheckResult{
		Status:    "healthy",
		Message:   "Redis connection successful",
		Duration:  time.Since(start).String(),
		Timestamp: time.Now(),
	}
}

// checkSystem checks system resources and health
func (h *HealthHandler) checkSystem(ctx context.Context) CheckResult {
	start := time.Now()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Check if memory usage is reasonable (less than 1GB for this example)
	const maxMemoryMB = 1024
	currentMemoryMB := m.Alloc / 1024 / 1024
	
	if currentMemoryMB > maxMemoryMB {
		return CheckResult{
			Status:    "unhealthy",
			Message:   "High memory usage detected",
			Duration:  time.Since(start).String(),
			Timestamp: time.Now(),
		}
	}
	
	// Check goroutine count (should be reasonable)
	goroutines := runtime.NumGoroutine()
	if goroutines > 1000 {
		return CheckResult{
			Status:    "degraded",
			Message:   "High goroutine count detected",
			Duration:  time.Since(start).String(),
			Timestamp: time.Now(),
		}
	}
	
	return CheckResult{
		Status:    "healthy",
		Message:   "System resources normal",
		Duration:  time.Since(start).String(),
		Timestamp: time.Now(),
	}
}

// getAppVersion returns the application version
func getAppVersion() string {
	if version := config.GetConfig().App.Version; version != "" {
		return version
	}
	return "dev"
}
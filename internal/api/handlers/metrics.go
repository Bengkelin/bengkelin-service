package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	redisClient "github.com/Bengkelin/bengkelin-service/internal/pkg/redis"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler handles metrics endpoints
type MetricsHandler struct {
	BaseHandler
}

// MetricsHandlerInterface defines metrics handler methods
type MetricsHandlerInterface interface {
	PrometheusMetrics(c *gin.Context)
	ApplicationMetrics(c *gin.Context)
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() MetricsHandlerInterface {
	return &MetricsHandler{
		BaseHandler: BaseHandler{},
	}
}

// ApplicationMetrics represents custom application metrics
type ApplicationMetrics struct {
	Timestamp   time.Time              `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	Uptime      string                 `json:"uptime" example:"1h30m45s"`
	Version     string                 `json:"version" example:"1.0.0"`
	Environment string                 `json:"environment" example:"production"`
	System      SystemMetrics          `json:"system"`
	Database    DatabaseMetrics        `json:"database"`
	Redis       RedisMetrics           `json:"redis"`
	HTTP        HTTPMetrics            `json:"http"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	GoVersion      string  `json:"go_version" example:"go1.21.0"`
	Goroutines     int     `json:"goroutines" example:"25"`
	MemoryAlloc    uint64  `json:"memory_alloc_bytes" example:"1048576"`
	MemoryTotal    uint64  `json:"memory_total_bytes" example:"2097152"`
	MemorySys      uint64  `json:"memory_sys_bytes" example:"4194304"`
	GCRuns         uint32  `json:"gc_runs" example:"10"`
	CPUCount       int     `json:"cpu_count" example:"4"`
}

// DatabaseMetrics represents database connection metrics
type DatabaseMetrics struct {
	OpenConnections     int `json:"open_connections" example:"5"`
	InUseConnections    int `json:"in_use_connections" example:"2"`
	IdleConnections     int `json:"idle_connections" example:"3"`
	MaxOpenConnections  int `json:"max_open_connections" example:"25"`
	WaitCount           int64 `json:"wait_count" example:"0"`
	WaitDuration        string `json:"wait_duration" example:"0s"`
	MaxIdleTimeClosed   int64 `json:"max_idle_time_closed" example:"0"`
	MaxLifetimeClosed   int64 `json:"max_lifetime_closed" example:"0"`
}

// RedisMetrics represents Redis connection metrics
type RedisMetrics struct {
	PoolSize     int `json:"pool_size" example:"10"`
	MinIdleConns int `json:"min_idle_conns" example:"5"`
	IdleConns    int `json:"idle_conns" example:"8"`
	StaleConns   int `json:"stale_conns" example:"0"`
	TotalConns   int `json:"total_conns" example:"10"`
	Hits         uint64 `json:"hits" example:"1000"`
	Misses       uint64 `json:"misses" example:"50"`
}

// HTTPMetrics represents HTTP request metrics
type HTTPMetrics struct {
	TotalRequests    uint64            `json:"total_requests" example:"5000"`
	RequestsByStatus map[string]uint64 `json:"requests_by_status"`
	AverageLatency   string            `json:"average_latency" example:"50ms"`
}

// PrometheusMetrics godoc
// @Summary Prometheus metrics endpoint
// @Description Get Prometheus-formatted metrics for monitoring
// @Tags Metrics
// @Produce text/plain
// @Success 200 {string} string "Prometheus metrics"
// @Router /metrics [get]
func (h *MetricsHandler) PrometheusMetrics(c *gin.Context) {
	// Use Prometheus handler directly
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

// ApplicationMetrics godoc
// @Summary Application metrics endpoint
// @Description Get detailed application metrics in JSON format
// @Tags Metrics
// @Accept json
// @Produce json
// @Success 200 {object} ApplicationMetrics "Application metrics"
// @Router /metrics/app [get]
func (h *MetricsHandler) ApplicationMetrics(c *gin.Context) {
	ctx := c.Request.Context()
	start := time.Now()
	
	applog.DebugCtx(ctx, "Application metrics requested")
	
	metrics := &ApplicationMetrics{
		Timestamp:   time.Now(),
		Uptime:      time.Since(startTime).String(),
		Version:     getAppVersion(),
		Environment: config.GetConfig().App.Environment,
		System:      h.getSystemMetrics(),
		Database:    h.getDatabaseMetrics(),
		Redis:       h.getRedisMetrics(),
		HTTP:        h.getHTTPMetrics(),
	}
	
	applog.DebugCtx(ctx, "Application metrics collected", 
		"duration", time.Since(start),
	)
	
	c.JSON(http.StatusOK, metrics)
}

// getSystemMetrics collects system-level metrics
func (h *MetricsHandler) getSystemMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return SystemMetrics{
		GoVersion:   runtime.Version(),
		Goroutines:  runtime.NumGoroutine(),
		MemoryAlloc: m.Alloc,
		MemoryTotal: m.TotalAlloc,
		MemorySys:   m.Sys,
		GCRuns:      m.NumGC,
		CPUCount:    runtime.NumCPU(),
	}
}

// getDatabaseMetrics collects database connection metrics
func (h *MetricsHandler) getDatabaseMetrics() DatabaseMetrics {
	database := db.GetDB()
	if database == nil {
		return DatabaseMetrics{}
	}
	
	sqlDB, err := database.DB()
	if err != nil {
		return DatabaseMetrics{}
	}
	
	stats := sqlDB.Stats()
	
	return DatabaseMetrics{
		OpenConnections:     stats.OpenConnections,
		InUseConnections:    stats.InUse,
		IdleConnections:     stats.Idle,
		MaxOpenConnections:  stats.MaxOpenConnections,
		WaitCount:           stats.WaitCount,
		WaitDuration:        stats.WaitDuration.String(),
		MaxIdleTimeClosed:   stats.MaxIdleTimeClosed,
		MaxLifetimeClosed:   stats.MaxLifetimeClosed,
	}
}

// getRedisMetrics collects Redis connection metrics
func (h *MetricsHandler) getRedisMetrics() RedisMetrics {
	redisCache := redisClient.GetRedisClient()
	if redisCache == nil {
		return RedisMetrics{}
	}
	
	// Since the Redis package doesn't expose pool stats directly,
	// we'll return basic metrics
	return RedisMetrics{
		PoolSize:     10, // Default pool size
		MinIdleConns: 0,
		IdleConns:    0,
		StaleConns:   0,
		TotalConns:   1, // Assume 1 connection for simplicity
		Hits:         0, // Would need to be tracked separately
		Misses:       0, // Would need to be tracked separately
	}
}

// getHTTPMetrics collects HTTP request metrics
func (h *MetricsHandler) getHTTPMetrics() HTTPMetrics {
	// This would typically come from middleware that tracks requests
	// For now, return placeholder data
	return HTTPMetrics{
		TotalRequests: 0, // Would be tracked by middleware
		RequestsByStatus: map[string]uint64{
			"2xx": 0,
			"4xx": 0,
			"5xx": 0,
		},
		AverageLatency: "0ms",
	}
}
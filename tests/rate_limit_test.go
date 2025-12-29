package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	pkgMiddleware "github.com/Bengkelin/bengkelin-service/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func setupRateLimitTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	
	// Create a test rate limiter: 2 requests per second, burst of 1
	limiter := pkgMiddleware.NewIPRateLimiter(rate.Limit(2), 1)
	
	router := gin.New()
	router.Use(pkgMiddleware.RateLimitMiddleware(limiter))
	
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	return router
}

func TestRateLimitMiddleware(t *testing.T) {
	router := setupRateLimitTest()
	
	// First request should succeed
	req1, _ := http.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	
	assert.Equal(t, http.StatusOK, w1.Code)
	
	// Second request should succeed (within burst limit)
	req2, _ := http.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	
	assert.Equal(t, http.StatusOK, w2.Code)
	
	// Third request should be rate limited (exceeds burst)
	req3, _ := http.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
	
	// Check response body for rate limit error
	var response map[string]interface{}
	err := json.Unmarshal(w3.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, "rate limit exceeded", response["message"])
	
	// Check rate limit headers
	assert.Equal(t, "60", w3.Header().Get("Retry-After"))
	assert.Equal(t, "0", w3.Header().Get("X-RateLimit-Remaining"))
}

func TestRateLimitRecovery(t *testing.T) {
	router := setupRateLimitTest()
	
	// Exhaust rate limit
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
	
	// Wait for rate limit to reset (1 second for 2 RPS)
	time.Sleep(1 * time.Second)
	
	// Request should succeed again
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitDifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create a stricter rate limiter for this test
	limiter := pkgMiddleware.NewIPRateLimiter(rate.Limit(1), 1) // 1 RPS, burst 1
	
	router := gin.New()
	router.Use(pkgMiddleware.RateLimitMiddleware(limiter))
	
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Request from first IP
	req1, _ := http.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Forwarded-For", "192.168.1.1")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)
	
	// Second request from same IP should be rate limited
	req2, _ := http.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Forwarded-For", "192.168.1.1")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	
	// Request from different IP should succeed
	req3, _ := http.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-Forwarded-For", "192.168.1.2")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

func TestRateLimitConfiguration(t *testing.T) {
	// Test with rate limiting disabled
	config.Config = &config.Configuration{
		RateLimit: config.RateLimitConfiguration{
			Enabled: false,
		},
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// No rate limiting should be applied
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Multiple rapid requests should all succeed
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestRateLimitHeaders(t *testing.T) {
	router := setupRateLimitTest()
	
	// Make a successful request
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Check that rate limit headers are present
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

func BenchmarkRateLimitMiddleware(b *testing.B) {
	router := setupRateLimitTest()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}
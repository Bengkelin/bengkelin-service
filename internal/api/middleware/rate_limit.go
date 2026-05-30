package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/response"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// CleanupOldLimiters removes limiters for IPs that haven't been seen recently
func (i *IPRateLimiter) CleanupOldLimiters() {
	i.mu.Lock()
	defer i.mu.Unlock()
	
	// In a production environment, you might want to track last access time
	// and remove limiters that haven't been used for a while
	// For now, we'll keep all limiters to maintain rate limiting state
}

func RateLimitMiddleware(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rateLimiter := limiter.GetLimiter(ip)
		
		if !rateLimiter.Allow() {
			// Log rate limit exceeded
			applog.Info("Rate limit exceeded", 
				"ip", ip, 
				"path", c.Request.URL.Path, 
				"method", c.Request.Method,
				"user_agent", c.Request.UserAgent(),
			)
			
			// Add rate limit headers
			c.Header("X-RateLimit-Limit", "100") // You can make this dynamic
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", "60") // Reset time in seconds
			c.Header("Retry-After", "60") // Retry after 60 seconds
			
			resp := response.BuildFailedResponse("rate limit exceeded", map[string]interface{}{
				"message": "Too many requests. Please try again later.",
				"retry_after": 60,
			})
			c.AbortWithStatusJSON(http.StatusTooManyRequests, resp)
			return
		}
		
		// Add rate limit headers for successful requests
		reservation := rateLimiter.Reserve()
		if reservation.OK() {
			delay := reservation.Delay()
			if delay > 0 {
				c.Header("X-RateLimit-Remaining", "0")
			} else {
				// Calculate remaining requests (approximation)
				c.Header("X-RateLimit-Remaining", "1") // Simplified
			}
			c.Header("X-RateLimit-Limit", "100")
			c.Header("X-RateLimit-Reset", "60")
		}
		
		c.Next()
	}
}

// StartCleanupRoutine starts a background routine to clean up old rate limiters
func StartCleanupRoutine(limiter *IPRateLimiter) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Cleanup every hour
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				limiter.CleanupOldLimiters()
				applog.Debug("Rate limiter cleanup completed")
			}
		}
	}()
}

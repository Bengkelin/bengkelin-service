package middleware

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"github.com/gin-gonic/gin"
)

// LoggingMiddleware creates a middleware for comprehensive request logging
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Extract additional context
		var userID, requestID string
		if param.Keys != nil {
			if uid, exists := param.Keys["user_id"]; exists {
				if id, ok := uid.(string); ok {
					userID = id
				}
			}
			if rid, exists := param.Keys["request_id"]; exists {
				if id, ok := rid.(string); ok {
					requestID = id
				}
			}
		}
		
		// Log the HTTP request
		applog.LogHTTPRequest(
			param.Method,
			param.Path,
			param.Request.UserAgent(),
			param.ClientIP,
			param.StatusCode,
			param.Latency,
			"request_id", requestID,
			"user_id", userID,
			"query_params", param.Request.URL.RawQuery,
			"protocol", param.Request.Proto,
			"content_length", param.Request.ContentLength,
		)
		
		// Return empty string as we're handling logging ourselves
		return ""
	})
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = helpers.GenerateUUID()
		}
		
		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		
		// Add to request context for service layer
		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)
		
		c.Next()
	}
}

// UserContextMiddleware adds user information to logging context
func UserContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// After processing, add user ID to context if available
		if userID, exists := c.Get("id"); exists {
			if id, ok := userID.(string); ok {
				c.Set("user_id", id)
				
				// Add to request context for service layer
				ctx := context.WithValue(c.Request.Context(), "user_id", id)
				c.Request = c.Request.WithContext(ctx)
			}
		}
	}
}

// ErrorLoggingMiddleware logs detailed error information
func ErrorLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Log errors if any occurred
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				applog.LogErrorCtx(
					c.Request.Context(),
					err.Err,
					"Request processing error",
					"error_type", err.Type,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"client_ip", c.ClientIP(),
					"user_agent", c.Request.UserAgent(),
				)
			}
		}
	}
}

// SecurityLoggingMiddleware logs security-related events
func SecurityLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log suspicious activities
		userAgent := c.Request.UserAgent()
		clientIP := c.ClientIP()
		
		// Check for suspicious patterns
		if isSuspiciousRequest(c) {
			applog.LogSecurityEvent(
				"suspicious_request",
				getUserIDFromContext(c),
				clientIP,
				userAgent,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"query", c.Request.URL.RawQuery,
				"headers", getSanitizedHeaders(c),
			)
		}
		
		c.Next()
		
		// Log authentication failures
		if c.Writer.Status() == 401 {
			applog.LogAuthEvent(
				"authentication_failed",
				getUserIDFromContext(c),
				"",
				clientIP,
				false,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"user_agent", userAgent,
			)
		}
	}
}

// PerformanceLoggingMiddleware logs performance metrics
func PerformanceLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start)
		
		// Log slow requests
		if duration > 1*time.Second {
			applog.WarnCtx(
				c.Request.Context(),
				"Slow HTTP request detected",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"duration_ms", duration.Milliseconds(),
				"status_code", c.Writer.Status(),
				"client_ip", c.ClientIP(),
			)
		}
		
		// Log performance metrics
		applog.DebugCtx(
			c.Request.Context(),
			"Request performance metrics",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"duration_ms", duration.Milliseconds(),
			"status_code", c.Writer.Status(),
			"response_size", c.Writer.Size(),
		)
	}
}

// RequestBodyLoggingMiddleware logs request bodies for debugging (use carefully)
func RequestBodyLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only log in development environment
		if gin.Mode() != gin.DebugMode {
			c.Next()
			return
		}
		
		// Only log for specific content types and methods
		if !shouldLogRequestBody(c) {
			c.Next()
			return
		}
		
		// Read and restore request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		
		// Log sanitized request body
		if len(bodyBytes) > 0 && len(bodyBytes) < 1024 { // Only log small bodies
			sanitizedBody := sanitizeRequestBody(string(bodyBytes))
			applog.DebugCtx(
				c.Request.Context(),
				"Request body logged",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"body", sanitizedBody,
				"content_type", c.Request.Header.Get("Content-Type"),
			)
		}
		
		c.Next()
	}
}

// Helper functions

func isSuspiciousRequest(c *gin.Context) bool {
	userAgent := c.Request.UserAgent()
	path := c.Request.URL.Path
	
	// Check for common attack patterns
	suspiciousPatterns := []string{
		"sqlmap", "nikto", "nmap", "masscan",
		"<script", "javascript:", "onload=", "onerror=",
		"../", "..\\", "/etc/passwd", "/proc/",
		"UNION SELECT", "DROP TABLE", "INSERT INTO",
	}
	
	for _, pattern := range suspiciousPatterns {
		if contains(userAgent, pattern) || contains(path, pattern) || contains(c.Request.URL.RawQuery, pattern) {
			return true
		}
	}
	
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
				s[len(s)-len(substr):] == substr || 
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	if userID, exists := c.Get("id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

func getSanitizedHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	
	// Only log safe headers
	safeHeaders := []string{
		"Content-Type", "Accept", "Accept-Language", 
		"Accept-Encoding", "Cache-Control", "Connection",
		"Host", "Referer", "X-Forwarded-For", "X-Real-IP",
	}
	
	for _, header := range safeHeaders {
		if value := c.Request.Header.Get(header); value != "" {
			headers[header] = value
		}
	}
	
	return headers
}

func shouldLogRequestBody(c *gin.Context) bool {
	method := c.Request.Method
	contentType := c.Request.Header.Get("Content-Type")
	
	// Only log for POST, PUT, PATCH methods
	if method != "POST" && method != "PUT" && method != "PATCH" {
		return false
	}
	
	// Only log JSON content
	if !contains(contentType, "application/json") {
		return false
	}
	
	return true
}

func sanitizeRequestBody(body string) string {
	// Remove sensitive fields from JSON body
	sensitiveFields := []string{
		"password", "token", "secret", "key", "auth",
		"credential", "private", "confidential",
	}
	
	sanitized := body
	for _, field := range sensitiveFields {
		// Simple regex replacement for JSON fields
		// In production, use proper JSON parsing
		if contains(sanitized, `"`+field+`"`) {
			sanitized = replaceJSONField(sanitized, field, "[REDACTED]")
		}
	}
	
	return sanitized
}

func replaceJSONField(json, field, replacement string) string {
	// Simple replacement - in production use proper JSON parsing
	start := `"` + field + `":`
	if idx := findString(json, start); idx != -1 {
		// Find the value part and replace it
		valueStart := idx + len(start)
		for valueStart < len(json) && (json[valueStart] == ' ' || json[valueStart] == '\t') {
			valueStart++
		}
		
		if valueStart < len(json) {
			if json[valueStart] == '"' {
				// String value
				valueEnd := valueStart + 1
				for valueEnd < len(json) && json[valueEnd] != '"' {
					if json[valueEnd] == '\\' {
						valueEnd++ // Skip escaped character
					}
					valueEnd++
				}
				if valueEnd < len(json) {
					valueEnd++ // Include closing quote
					return json[:valueStart] + `"` + replacement + `"` + json[valueEnd:]
				}
			}
		}
	}
	return json
}

func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
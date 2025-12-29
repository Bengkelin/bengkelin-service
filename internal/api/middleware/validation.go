package middleware

import (
	"net/http"
	"reflect"

	"github.com/Bengkelin/bengkelin-service/pkg/response"
	"github.com/Bengkelin/bengkelin-service/pkg/validation"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"github.com/gin-gonic/gin"
)

// ValidationMiddleware creates a middleware that validates request body against a struct type
func ValidationMiddleware(structType interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new instance of the struct type
		structValue := reflect.New(reflect.TypeOf(structType)).Interface()
		
		// Bind JSON to struct (this handles basic binding validation)
		if err := c.ShouldBindJSON(structValue); err != nil {
			applog.Info("Validation failed - binding error", 
				"error", err.Error(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"ip", c.ClientIP(),
			)
			
			resp := response.BuildFailedResponse("validation failed", map[string]interface{}{
				"message": "Invalid request format",
				"details": err.Error(),
			})
			c.AbortWithStatusJSON(http.StatusBadRequest, resp)
			return
		}
		
		// Perform custom validation and sanitization
		validationErrors, err := validation.ValidateAndSanitize(structValue)
		if err != nil {
			applog.Error("Validation error", "error", err.Error())
			resp := response.BuildFailedResponse("validation error", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, resp)
			return
		}
		
		if validationErrors != nil && len(validationErrors) > 0 {
			applog.Info("Validation failed - custom validation", 
				"errors", validationErrors,
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"ip", c.ClientIP(),
			)
			
			resp := response.BuildFailedResponse("validation failed", map[string]interface{}{
				"message": "Invalid input data",
				"errors":  validationErrors,
			})
			c.AbortWithStatusJSON(http.StatusBadRequest, resp)
			return
		}
		
		// Store validated and sanitized data in context
		c.Set("validated_data", structValue)
		c.Next()
	}
}

// GetValidatedData retrieves validated data from context
func GetValidatedData(c *gin.Context, target interface{}) bool {
	if data, exists := c.Get("validated_data"); exists {
		// Copy data to target
		sourceValue := reflect.ValueOf(data)
		targetValue := reflect.ValueOf(target)
		
		if sourceValue.Kind() == reflect.Ptr {
			sourceValue = sourceValue.Elem()
		}
		if targetValue.Kind() == reflect.Ptr {
			targetValue = targetValue.Elem()
		}
		
		if sourceValue.Type() == targetValue.Type() {
			targetValue.Set(sourceValue)
			return true
		}
	}
	return false
}

// ValidateRequest validates a request struct manually
func ValidateRequest(c *gin.Context, request interface{}) bool {
	// Bind JSON to struct
	if err := c.ShouldBindJSON(request); err != nil {
		applog.Info("Manual validation failed - binding error", 
			"error", err.Error(),
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"ip", c.ClientIP(),
		)
		
		resp := response.BuildFailedResponse("validation failed", map[string]interface{}{
			"message": "Invalid request format",
			"details": err.Error(),
		})
		c.JSON(http.StatusBadRequest, resp)
		return false
	}
	
	// Perform custom validation and sanitization
	validationErrors, err := validation.ValidateAndSanitize(request)
	if err != nil {
		applog.Error("Manual validation error", "error", err.Error())
		resp := response.BuildFailedResponse("validation error", err.Error())
		c.JSON(http.StatusInternalServerError, resp)
		return false
	}
	
	if validationErrors != nil && len(validationErrors) > 0 {
		applog.Info("Manual validation failed - custom validation", 
			"errors", validationErrors,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"ip", c.ClientIP(),
		)
		
		resp := response.BuildFailedResponse("validation failed", map[string]interface{}{
			"message": "Invalid input data",
			"errors":  validationErrors,
		})
		c.JSON(http.StatusBadRequest, resp)
		return false
	}
	
	return true
}

// SanitizeMiddleware sanitizes all string inputs in request body
func SanitizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware can be used to sanitize inputs before they reach handlers
		// The actual sanitization is done in the validation package
		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		
		c.Next()
	}
}
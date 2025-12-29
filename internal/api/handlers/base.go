package handlers

import (
	"net/http"
	"strconv"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/constants"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/pkg/errors"
	"github.com/Bengkelin/bengkelin-service/pkg/response"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
	"github.com/gin-gonic/gin"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct{}

// HandleError handles errors consistently across all handlers
func (h *BaseHandler) HandleError(c *gin.Context, err error) {
	if appErr, ok := appErrors.IsAppError(err); ok {
		applog.Info("Application error occurred", 
			"code", appErr.Code,
			"message", appErr.Message,
			"details", appErr.Details,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"ip", c.ClientIP(),
		)
		
		resp := response.BuildFailedResponse(appErr.Message, map[string]interface{}{
			"code":    appErr.Code,
			"message": appErr.Message,
			"details": appErr.Details,
		})
		c.JSON(appErr.StatusCode, resp)
		return
	}
	
	// Handle generic errors
	applog.Error("Unexpected error occurred",
		"error", err.Error(),
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"ip", c.ClientIP(),
	)
	
	resp := response.BuildFailedResponse("Internal server error", map[string]interface{}{
		"code":    "INTERNAL_SERVER_ERROR",
		"message": "An unexpected error occurred",
	})
	c.JSON(http.StatusInternalServerError, resp)
}

// HandleSuccess handles successful responses consistently
func (h *BaseHandler) HandleSuccess(c *gin.Context, message string, data interface{}) {
	resp := response.BuildSuccessResponse(message, data)
	c.JSON(http.StatusOK, resp)
}

// HandleCreated handles created responses consistently
func (h *BaseHandler) HandleCreated(c *gin.Context, message string, data interface{}) {
	resp := response.BuildSuccessResponse(message, data)
	c.JSON(http.StatusCreated, resp)
}

// GetUserID extracts user ID from JWT context
func (h *BaseHandler) GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("id")
	if !exists {
		return "", appErrors.ErrUnauthorized
	}
	
	userIDStr, ok := userID.(string)
	if !ok {
		return "", appErrors.ErrUnauthorized
	}
	
	return userIDStr, nil
}

// GetMitraID extracts mitra ID from JWT context
func (h *BaseHandler) GetMitraID(c *gin.Context) (string, error) {
	mitraID, exists := c.Get("id")
	if !exists {
		return "", appErrors.ErrUnauthorized
	}
	
	mitraIDStr, ok := mitraID.(string)
	if !ok {
		return "", appErrors.ErrUnauthorized
	}
	
	return mitraIDStr, nil
}

// ParsePagination extracts pagination parameters from query
func (h *BaseHandler) ParsePagination(c *gin.Context) dto.PaginationRequest {
	page := constants.DefaultPage
	limit := constants.DefaultLimit
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= constants.MaxLimit {
			limit = l
		}
	}
	
	return dto.PaginationRequest{
		Page:  page,
		Limit: limit,
	}
}

// ParseLocation extracts location parameters from query
func (h *BaseHandler) ParseLocation(c *gin.Context) (float64, float64, error) {
	latStr := c.Query("latitude")
	lngStr := c.Query("longitude")
	
	if latStr == "" || lngStr == "" {
		return 0, 0, appErrors.ErrInvalidLocation.WithDetails("latitude and longitude are required")
	}
	
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return 0, 0, appErrors.ErrInvalidLocation.WithDetails("invalid latitude format")
	}
	
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return 0, 0, appErrors.ErrInvalidLocation.WithDetails("invalid longitude format")
	}
	
	// Validate coordinate ranges
	if lat < -90 || lat > 90 {
		return 0, 0, appErrors.ErrInvalidLocation.WithDetails("latitude must be between -90 and 90")
	}
	
	if lng < -180 || lng > 180 {
		return 0, 0, appErrors.ErrInvalidLocation.WithDetails("longitude must be between -180 and 180")
	}
	
	return lat, lng, nil
}

// ParseRadius extracts radius parameter from query
func (h *BaseHandler) ParseRadius(c *gin.Context) float64 {
	radiusStr := c.Query("radius")
	if radiusStr == "" {
		return constants.DefaultRadius
	}
	
	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 || radius > constants.MaxRadius {
		return constants.DefaultRadius
	}
	
	return radius
}

// LogRequest logs incoming requests for debugging
func (h *BaseHandler) LogRequest(c *gin.Context, action string) {
	applog.Debug("Handler request",
		"action", action,
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"query", c.Request.URL.RawQuery,
		"ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
	)
}

// ValidateOwnership validates if the user owns the resource
func (h *BaseHandler) ValidateOwnership(userID, resourceOwnerID string) error {
	if userID != resourceOwnerID {
		return appErrors.ErrNotOwner
	}
	return nil
}

// ValidateRequiredFields validates that required fields are not empty
func (h *BaseHandler) ValidateRequiredFields(fields map[string]interface{}) error {
	for fieldName, fieldValue := range fields {
		if fieldValue == nil {
			return appErrors.ErrMissingField.WithDetails(fieldName + " is required")
		}
		
		switch v := fieldValue.(type) {
		case string:
			if v == "" {
				return appErrors.ErrMissingField.WithDetails(fieldName + " is required")
			}
		case *string:
			if v == nil || *v == "" {
				return appErrors.ErrMissingField.WithDetails(fieldName + " is required")
			}
		}
	}
	return nil
}
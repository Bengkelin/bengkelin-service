package helpers

import (
	"github.com/Bengkelin/bengkelin-service/internal/response"
	"github.com/gin-gonic/gin"
)

// ErrorResponse sends an error response using the response package
func ErrorResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := response.BuildFailedResponse(message, data)
	c.JSON(statusCode, resp)
}

// SuccessResponse sends a success response using the response package
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := response.BuildSuccessResponse(message, data)
	c.JSON(statusCode, resp)
}
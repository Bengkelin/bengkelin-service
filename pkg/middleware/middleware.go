package middleware

import (
	"net/http"

	"github.com/Bengkelin/bengkelin-service/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateRequest[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request T

		if err := c.ShouldBindJSON(&request); err != nil {
			resp := response.BuildFailedResponse("validation failed", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, resp)
			return
		}

		if err := validate.Struct(request); err != nil {
			resp := response.BuildFailedResponse("validation failed", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, resp)
			return
		}

		c.Set("validatedRequest", request)
		c.Next()
	}
}

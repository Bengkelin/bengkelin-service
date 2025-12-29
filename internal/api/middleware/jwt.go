package middleware

import (
	"net/http"
	"strings"

	"github.com/Bengkelin/bengkelin-service/pkg/crypto"
	"github.com/Bengkelin/bengkelin-service/pkg/response"
	"github.com/gin-gonic/gin"
)

// Func to authorizing jwt token
func AuthJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response := response.BuildFailedResponse("no token provided", nil)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response := response.BuildFailedResponse("invalid token format", nil)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		jwtHelper := crypto.GetJWTCrypto()
		token, err := jwtHelper.ValidateToken(tokenString)

		if err != nil || token == nil {
			response := response.BuildFailedResponse("token invalid", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		// Extract user ID using the new method
		userID, err := jwtHelper.GetUserIDFromToken(tokenString)
		if err != nil {
			response := response.BuildFailedResponse("failed to extract user ID from token", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		c.Set("id", userID)
	}
}

// func to authorizing jwt token for mitra
func AuthJWTMitra() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response := response.BuildFailedResponse("no token provided", nil)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response := response.BuildFailedResponse("invalid token format", nil)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		jwtHelper := crypto.GetJWTCrypto()
		token, err := jwtHelper.ValidateTokenMitra(tokenString)

		if err != nil || token == nil {
			response := response.BuildFailedResponse("token invalid", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		// Extract mitra ID using the new method
		mitraID, err := jwtHelper.GetMitraIDFromToken(tokenString)
		if err != nil {
			response := response.BuildFailedResponse("failed to extract mitra ID from token", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		c.Set("id", mitraID)
	}
}

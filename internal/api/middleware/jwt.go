package middleware

import (
	"net/http"
	"strings"

	"github.com/Bengkelin/bengkelin-service/internal/crypto"
	"github.com/Bengkelin/bengkelin-service/internal/response"
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
// AuthJWTFlexible - Middleware that accepts both user and mitra JWT tokens
func AuthJWTFlexible() gin.HandlerFunc {
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
		
		// Use the combined validation function that handles both user and mitra tokens
		claims, err := crypto.ValidateJWT(tokenString)
		if err != nil {
			response := response.BuildFailedResponse("token invalid", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		// Set context based on token type
		if claims.UserID != "" {
			// User token
			c.Set("id", claims.UserID)
			c.Set("user_id", claims.UserID)
			c.Set("user_type", "user")
		} else if claims.MitraID != "" {
			// Mitra token
			c.Set("id", claims.MitraID)
			c.Set("mitra_id", claims.MitraID)
			c.Set("user_type", "mitra")
		} else {
			response := response.BuildFailedResponse("invalid token claims", "no user or mitra ID found")
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		c.Next()
	}
}

// AuthJWTOptional - Optional authentication that allows both authenticated and anonymous access
func AuthJWTOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		
		// If no auth header, continue as anonymous user
		if authHeader == "" {
			c.Set("user_type", "anonymous")
			c.Next()
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Set("user_type", "anonymous")
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		// Try to validate token
		claims, err := crypto.ValidateJWT(tokenString)
		if err != nil {
			// If token is invalid, continue as anonymous
			c.Set("user_type", "anonymous")
			c.Next()
			return
		}

		// Set context based on token type
		if claims.UserID != "" {
			c.Set("id", claims.UserID)
			c.Set("user_id", claims.UserID)
			c.Set("user_type", "user")
		} else if claims.MitraID != "" {
			c.Set("id", claims.MitraID)
			c.Set("mitra_id", claims.MitraID)
			c.Set("user_type", "mitra")
		} else {
			c.Set("user_type", "anonymous")
		}

		c.Next()
	}
}

// AuthJWTAdmin - Middleware that validates JWT and requires admin role
func AuthJWTAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			resp := response.BuildFailedResponse("no token provided", nil)
			c.AbortWithStatusJSON(http.StatusBadRequest, resp)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			resp := response.BuildFailedResponse("invalid token format", nil)
			c.AbortWithStatusJSON(http.StatusBadRequest, resp)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		jwtHelper := crypto.GetJWTCrypto()
		token, err := jwtHelper.ValidateToken(tokenString)

		if err != nil || token == nil {
			resp := response.BuildFailedResponse("token invalid", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, resp)
			return
		}

		claims, ok := token.Claims.(*crypto.JwtCustomClaim)
		if !ok || !token.Valid {
			resp := response.BuildFailedResponse("token invalid", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, resp)
			return
		}

		if claims.Role != "admin" {
			resp := response.BuildFailedResponse("admin access required", nil)
			c.AbortWithStatusJSON(http.StatusForbidden, resp)
			return
		}

		c.Set("id", claims.UserID)
		c.Set("user_type", "admin")
		c.Next()
	}
}

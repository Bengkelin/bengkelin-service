package middleware

import (
	"net/http"
	"strings"

	"github.com/Bengkelin/bengkelin-service/pkg/crypto"
	"github.com/Bengkelin/bengkelin-service/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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

		tokenString := strings.Split(authHeader, " ")[1]
		jwtHelper := crypto.GetJWTCrypto()
		token, err := jwtHelper.ValidateToken(tokenString)

		if err != nil || token == nil {
			response := response.BuildFailedResponse("token invalid", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		claim, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			response := response.BuildFailedResponse("token is expired, please try login", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		c.Set("id", claim["user_id"])
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

		tokenString := strings.Split(authHeader, " ")[1]
		jwtHelper := crypto.GetJWTCrypto()
		token, err := jwtHelper.ValidateTokenMitra(tokenString)

		if err != nil || token == nil {
			response := response.BuildFailedResponse("token invalid", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		claim, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			response := response.BuildFailedResponse("token is expired, please try login", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		c.Set("id", claim["mitra_id"])
	}
}

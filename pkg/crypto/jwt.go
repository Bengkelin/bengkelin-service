package crypto

import (
	"context"
	"fmt"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	"github.com/golang-jwt/jwt"
)

var jwtHelper *jwtCryptoHelper

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Contract for JWT Crypto Helper
type JWTCryptoHelper interface {
	GenerateTokenPair(userID string) (*TokenPair, error)
	GenerateTokenPairMitra(mitraID string) (*TokenPair, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	ValidateTokenMitra(tokenString string) (*jwt.Token, error)
	ValidateRefreshToken(tokenString string) (*jwt.Token, error)
	ValidateRefreshTokenMitra(tokenString string) (*jwt.Token, error)
	RefreshAccessToken(refreshToken string) (*TokenPair, error)
	RefreshAccessTokenMitra(refreshToken string) (*TokenPair, error)
	RevokeRefreshToken(refreshToken string) error
	RevokeAllUserTokens(userID string) error
	RevokeAllMitraTokens(mitraID string) error
	GetUserIDFromToken(tokenString string) (string, error)
	GetMitraIDFromToken(tokenString string) (string, error)
}

// Struct for jwt custom claim
type JwtCustomClaim struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role,omitempty"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.StandardClaims
}

type JwtCustomClaimMitra struct {
	MitraID   string `json:"mitra_id"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.StandardClaims
}

// Struct for JWTHelper
type jwtCryptoHelper struct {
	refreshTokenRepo repository.RefreshTokenRepositoryInterface
	userRepo         repository.UserRepositoryInterface
}

// Func to initialize new jwt crypto helper
func GetJWTCrypto() JWTCryptoHelper {
	if jwtHelper == nil {
		jwtHelper = &jwtCryptoHelper{
			refreshTokenRepo: repository.GetRefreshTokenRepository(),
			userRepo:         repository.GetUserRepository(),
		}
	}
	return jwtHelper
}

// GenerateTokenPair generates both access and refresh tokens for user
func (helper *jwtCryptoHelper) GenerateTokenPair(userID string) (*TokenPair, error) {
	serverConfiguration := config.GetConfig().Server

	// Look up user role from database
	role := "user"
	if helper.userRepo != nil {
		user, err := helper.userRepo.FindUserByID(context.Background(), userID)
		if err == nil && user.Role != "" {
			role = user.Role
		}
	}

	// Generate access token (short-lived)
	accessClaims := &JwtCustomClaim{
		UserID:    userID,
		Role:      role,
		TokenType: "access",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(serverConfiguration.ExpiresHour)).Unix(),
			Issuer:    serverConfiguration.Name,
			IssuedAt:  time.Now().Unix(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(serverConfiguration.Secret))
	if err != nil {
		return nil, err
	}

	// Generate refresh token (long-lived)
	refreshClaims := &JwtCustomClaim{
		UserID:    userID,
		TokenType: "refresh",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(serverConfiguration.RefreshExpiresHour)).Unix(),
			Issuer:    serverConfiguration.Name,
			IssuedAt:  time.Now().Unix(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(serverConfiguration.RefreshSecret))
	if err != nil {
		return nil, err
	}

	// Store refresh token in database
	refreshTokenModel := models.RefreshToken{
		ID:        helpers.GenerateUUID(),
		UserID:    &userID,
		Token:     refreshTokenString,
		ExpiresAt: time.Unix(refreshClaims.ExpiresAt, 0),
		IsRevoked: false,
	}

	_, err = helper.refreshTokenRepo.CreateRefreshToken(context.Background(), refreshTokenModel)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    serverConfiguration.ExpiresHour * 3600, // Convert hours to seconds
		TokenType:    "Bearer",
	}, nil
}

// GenerateTokenPairMitra generates both access and refresh tokens for mitra
func (helper *jwtCryptoHelper) GenerateTokenPairMitra(mitraID string) (*TokenPair, error) {
	serverConfiguration := config.GetConfig().Server

	// Generate access token (short-lived)
	accessClaims := &JwtCustomClaimMitra{
		MitraID:   mitraID,
		TokenType: "access",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(serverConfiguration.ExpiresHour)).Unix(),
			Issuer:    serverConfiguration.Name,
			IssuedAt:  time.Now().Unix(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(serverConfiguration.Secret2))
	if err != nil {
		return nil, err
	}

	// Generate refresh token (long-lived)
	refreshClaims := &JwtCustomClaimMitra{
		MitraID:   mitraID,
		TokenType: "refresh",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(serverConfiguration.RefreshExpiresHour)).Unix(),
			Issuer:    serverConfiguration.Name,
			IssuedAt:  time.Now().Unix(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(serverConfiguration.RefreshSecret2))
	if err != nil {
		return nil, err
	}

	// Store refresh token in database
	refreshTokenModel := models.RefreshToken{
		ID:        helpers.GenerateUUID(),
		MitraID:   &mitraID,
		Token:     refreshTokenString,
		ExpiresAt: time.Unix(refreshClaims.ExpiresAt, 0),
		IsRevoked: false,
	}

	_, err = helper.refreshTokenRepo.CreateRefreshToken(context.Background(), refreshTokenModel)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    serverConfiguration.ExpiresHour * 3600, // Convert hours to seconds
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates access token for users
func (helper *jwtCryptoHelper) ValidateToken(tokenString string) (*jwt.Token, error) {
	serverConfiguration := config.GetConfig().Server
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaim{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(serverConfiguration.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if token type is access
	if claims, ok := token.Claims.(*JwtCustomClaim); ok && token.Valid {
		if claims.TokenType != "access" {
			return nil, fmt.Errorf("invalid token type")
		}
	}

	return token, nil
}

// ValidateTokenMitra validates access token for mitras
func (helper *jwtCryptoHelper) ValidateTokenMitra(tokenString string) (*jwt.Token, error) {
	serverConfiguration := config.GetConfig().Server
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaimMitra{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(serverConfiguration.Secret2), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if token type is access
	if claims, ok := token.Claims.(*JwtCustomClaimMitra); ok && token.Valid {
		if claims.TokenType != "access" {
			return nil, fmt.Errorf("invalid token type")
		}
	}

	return token, nil
}

// ValidateRefreshToken validates refresh token for users
func (helper *jwtCryptoHelper) ValidateRefreshToken(tokenString string) (*jwt.Token, error) {
	serverConfiguration := config.GetConfig().Server
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaim{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(serverConfiguration.RefreshSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if token type is refresh
	if claims, ok := token.Claims.(*JwtCustomClaim); ok && token.Valid {
		if claims.TokenType != "refresh" {
			return nil, fmt.Errorf("invalid token type")
		}

		// Check if token exists in database and is not revoked
		_, err := helper.refreshTokenRepo.FindRefreshTokenByToken(context.Background(), tokenString)
		if err != nil {
			return nil, fmt.Errorf("refresh token not found or revoked")
		}
	}

	return token, nil
}

// ValidateRefreshTokenMitra validates refresh token for mitras
func (helper *jwtCryptoHelper) ValidateRefreshTokenMitra(tokenString string) (*jwt.Token, error) {
	serverConfiguration := config.GetConfig().Server
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaimMitra{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(serverConfiguration.RefreshSecret2), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if token type is refresh
	if claims, ok := token.Claims.(*JwtCustomClaimMitra); ok && token.Valid {
		if claims.TokenType != "refresh" {
			return nil, fmt.Errorf("invalid token type")
		}

		// Check if token exists in database and is not revoked
		_, err := helper.refreshTokenRepo.FindRefreshTokenByToken(context.Background(), tokenString)
		if err != nil {
			return nil, fmt.Errorf("refresh token not found or revoked")
		}
	}

	return token, nil
}

// RefreshAccessToken generates new access token using refresh token for users
func (helper *jwtCryptoHelper) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	// Validate refresh token
	token, err := helper.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JwtCustomClaim)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Revoke old refresh token
	err = helper.RevokeRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	return helper.GenerateTokenPair(claims.UserID)
}

// RefreshAccessTokenMitra generates new access token using refresh token for mitras
func (helper *jwtCryptoHelper) RefreshAccessTokenMitra(refreshToken string) (*TokenPair, error) {
	// Validate refresh token
	token, err := helper.ValidateRefreshTokenMitra(refreshToken)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JwtCustomClaimMitra)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Revoke old refresh token
	err = helper.RevokeRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	return helper.GenerateTokenPairMitra(claims.MitraID)
}

// RevokeRefreshToken revokes a specific refresh token
func (helper *jwtCryptoHelper) RevokeRefreshToken(refreshToken string) error {
	tokenModel, err := helper.refreshTokenRepo.FindRefreshTokenByToken(context.Background(), refreshToken)
	if err != nil {
		return err
	}

	return helper.refreshTokenRepo.RevokeRefreshToken(context.Background(), tokenModel.ID)
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (helper *jwtCryptoHelper) RevokeAllUserTokens(userID string) error {
	return helper.refreshTokenRepo.RevokeAllUserRefreshTokens(context.Background(), userID)
}

// RevokeAllMitraTokens revokes all refresh tokens for a mitra
func (helper *jwtCryptoHelper) RevokeAllMitraTokens(mitraID string) error {
	return helper.refreshTokenRepo.RevokeAllMitraRefreshTokens(context.Background(), mitraID)
}

// GetUserIDFromToken extracts user ID from access token
func (helper *jwtCryptoHelper) GetUserIDFromToken(tokenString string) (string, error) {
	token, err := helper.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*JwtCustomClaim)
	if !ok || !token.Valid {
		return "", fmt.Errorf("token is invalid")
	}

	return claims.UserID, nil
}

// GetMitraIDFromToken extracts mitra ID from access token
func (helper *jwtCryptoHelper) GetMitraIDFromToken(tokenString string) (string, error) {
	token, err := helper.ValidateTokenMitra(tokenString)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*JwtCustomClaimMitra)
	if !ok || !token.Valid {
		return "", fmt.Errorf("token is invalid")
	}

	return claims.MitraID, nil
}

// CombinedClaims represents claims that can be either user or mitra
type CombinedClaims struct {
	UserID  string `json:"user_id,omitempty"`
	MitraID string `json:"mitra_id,omitempty"`
	Role    string `json:"role,omitempty"`
}

// ValidateJWT validates JWT token for both users and mitras
func ValidateJWT(tokenString string) (*CombinedClaims, error) {
	helper := GetJWTCrypto()

	// Try to validate as user token first
	userToken, err := helper.ValidateToken(tokenString)
	if err == nil {
		if claims, ok := userToken.Claims.(*JwtCustomClaim); ok && userToken.Valid {
			return &CombinedClaims{
				UserID: claims.UserID,
				Role:   claims.Role,
			}, nil
		}
	}

	// Try to validate as mitra token
	mitraToken, err := helper.ValidateTokenMitra(tokenString)
	if err == nil {
		if claims, ok := mitraToken.Claims.(*JwtCustomClaimMitra); ok && mitraToken.Valid {
			return &CombinedClaims{
				MitraID: claims.MitraID,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid token")
}

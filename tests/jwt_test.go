package tests

import (
	"testing"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/config"
	"github.com/Bengkelin/bengkelin-service/internal/db"
	"github.com/Bengkelin/bengkelin-service/internal/crypto"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
	"github.com/stretchr/testify/assert"
)

func setupTestEnvironment() {
	// Setup logger
	applog.Setup("test")
	
	// Setup config with test values
	config.Config = &config.Configuration{
		Server: config.ServerConnection{
			Secret:               "test-secret-key-for-access-tokens",
			Secret2:              "test-secret-key-for-mitra-access-tokens",
			RefreshSecret:        "test-secret-key-for-refresh-tokens",
			RefreshSecret2:       "test-secret-key-for-mitra-refresh-tokens",
			Name:                 "bengkelin-test-api",
			ExpiresHour:          1,  // 1 hour for access tokens
			RefreshExpiresHour:   24, // 24 hours for refresh tokens
		},
		Database: config.DatabaseConfiguration{
			Driver:   "sqlite",
			Dbname:   ":memory:",
			Host:     "",
			Port:     "",
			Username: "",
			Password: "",
		},
	}
	
	// Setup in-memory database for testing
	db.SetupDB()
}

func TestJWTTokenPairGeneration(t *testing.T) {
	setupTestEnvironment()
	
	jwtHelper := crypto.GetJWTCrypto()
	userID := "test-user-123"
	
	// Test user token pair generation
	tokenPair, err := jwtHelper.GenerateTokenPair(userID)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, "Bearer", tokenPair.TokenType)
	assert.Equal(t, int64(3600), tokenPair.ExpiresIn) // 1 hour in seconds
}

func TestJWTTokenPairGenerationMitra(t *testing.T) {
	setupTestEnvironment()
	
	jwtHelper := crypto.GetJWTCrypto()
	mitraID := "test-mitra-456"
	
	// Test mitra token pair generation
	tokenPair, err := jwtHelper.GenerateTokenPairMitra(mitraID)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, "Bearer", tokenPair.TokenType)
	assert.Equal(t, int64(3600), tokenPair.ExpiresIn) // 1 hour in seconds
}

func TestAccessTokenValidation(t *testing.T) {
	setupTestEnvironment()
	
	jwtHelper := crypto.GetJWTCrypto()
	userID := "test-user-123"
	
	// Generate token pair
	tokenPair, err := jwtHelper.GenerateTokenPair(userID)
	assert.NoError(t, err)
	
	// Validate access token
	token, err := jwtHelper.ValidateToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.True(t, token.Valid)
	
	// Extract user ID from token
	extractedUserID, err := jwtHelper.GetUserIDFromToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
}

func TestRefreshTokenValidation(t *testing.T) {
	setupTestEnvironment()
	
	jwtHelper := crypto.GetJWTCrypto()
	userID := "test-user-123"
	
	// Generate token pair
	tokenPair, err := jwtHelper.GenerateTokenPair(userID)
	assert.NoError(t, err)
	
	// Validate refresh token
	token, err := jwtHelper.ValidateRefreshToken(tokenPair.RefreshToken)
	assert.NoError(t, err)
	assert.True(t, token.Valid)
}

func TestTokenRefresh(t *testing.T) {
	setupTestEnvironment()
	
	jwtHelper := crypto.GetJWTCrypto()
	userID := "test-user-123"
	
	// Generate initial token pair
	initialTokenPair, err := jwtHelper.GenerateTokenPair(userID)
	assert.NoError(t, err)
	
	// Wait a moment to ensure different issued times
	time.Sleep(1 * time.Second)
	
	// Refresh the token
	newTokenPair, err := jwtHelper.RefreshAccessToken(initialTokenPair.RefreshToken)
	assert.NoError(t, err)
	
	// New tokens should be different
	assert.NotEqual(t, initialTokenPair.AccessToken, newTokenPair.AccessToken)
	assert.NotEqual(t, initialTokenPair.RefreshToken, newTokenPair.RefreshToken)
	
	// Old refresh token should be invalid now
	_, err = jwtHelper.ValidateRefreshToken(initialTokenPair.RefreshToken)
	assert.Error(t, err)
	
	// New tokens should be valid
	_, err = jwtHelper.ValidateToken(newTokenPair.AccessToken)
	assert.NoError(t, err)
	
	_, err = jwtHelper.ValidateRefreshToken(newTokenPair.RefreshToken)
	assert.NoError(t, err)
}

func TestTokenRevocation(t *testing.T) {
	setupTestEnvironment()
	
	jwtHelper := crypto.GetJWTCrypto()
	userID := "test-user-123"
	
	// Generate token pair
	tokenPair, err := jwtHelper.GenerateTokenPair(userID)
	assert.NoError(t, err)
	
	// Revoke refresh token
	err = jwtHelper.RevokeRefreshToken(tokenPair.RefreshToken)
	assert.NoError(t, err)
	
	// Refresh token should be invalid now
	_, err = jwtHelper.ValidateRefreshToken(tokenPair.RefreshToken)
	assert.Error(t, err)
}

func TestRevokeAllUserTokens(t *testing.T) {
	setupTestEnvironment()
	
	jwtHelper := crypto.GetJWTCrypto()
	userID := "test-user-123"
	
	// Generate multiple token pairs
	tokenPair1, err := jwtHelper.GenerateTokenPair(userID)
	assert.NoError(t, err)
	
	tokenPair2, err := jwtHelper.GenerateTokenPair(userID)
	assert.NoError(t, err)
	
	// Both should be valid initially
	_, err = jwtHelper.ValidateRefreshToken(tokenPair1.RefreshToken)
	assert.NoError(t, err)
	
	_, err = jwtHelper.ValidateRefreshToken(tokenPair2.RefreshToken)
	assert.NoError(t, err)
	
	// Revoke all tokens for user
	err = jwtHelper.RevokeAllUserTokens(userID)
	assert.NoError(t, err)
	
	// Both should be invalid now
	_, err = jwtHelper.ValidateRefreshToken(tokenPair1.RefreshToken)
	assert.Error(t, err)
	
	_, err = jwtHelper.ValidateRefreshToken(tokenPair2.RefreshToken)
	assert.Error(t, err)
}

func TestInvalidTokenValidation(t *testing.T) {
	setupTestEnvironment()
	
	jwtHelper := crypto.GetJWTCrypto()
	
	// Test with invalid token
	_, err := jwtHelper.ValidateToken("invalid-token")
	assert.Error(t, err)
	
	// Test with empty token
	_, err = jwtHelper.ValidateToken("")
	assert.Error(t, err)
	
	// Test refresh token as access token (should fail)
	userID := "test-user-123"
	tokenPair, err := jwtHelper.GenerateTokenPair(userID)
	assert.NoError(t, err)
	
	_, err = jwtHelper.ValidateToken(tokenPair.RefreshToken)
	assert.Error(t, err)
}

func TestMitraTokenFunctionality(t *testing.T) {
	setupTestEnvironment()
	
	jwtHelper := crypto.GetJWTCrypto()
	mitraID := "test-mitra-456"
	
	// Generate mitra token pair
	tokenPair, err := jwtHelper.GenerateTokenPairMitra(mitraID)
	assert.NoError(t, err)
	
	// Validate mitra access token
	token, err := jwtHelper.ValidateTokenMitra(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.True(t, token.Valid)
	
	// Extract mitra ID from token
	extractedMitraID, err := jwtHelper.GetMitraIDFromToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, mitraID, extractedMitraID)
	
	// Test mitra token refresh
	newTokenPair, err := jwtHelper.RefreshAccessTokenMitra(tokenPair.RefreshToken)
	assert.NoError(t, err)
	assert.NotEqual(t, tokenPair.AccessToken, newTokenPair.AccessToken)
}
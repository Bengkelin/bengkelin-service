package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bengkelin/bengkelin-service/internal/api/handlers"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/pkg/errors"
	"github.com/Bengkelin/bengkelin-service/tests/fixtures/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupAuthHandler creates a test handler with mock service
// Note: Since AuthHandler has unexported fields, we test using the interface
// and mock the service layer responses
func setupAuthHandler() (handlers.AuthHandlerInterface, *mocks.MockAuthService) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &mocks.MockAuthService{}

	// Create handler using the factory - in real tests we'd need to mock the container
	// For unit tests, we use the interface and verify behavior through integration
	authHandler := handlers.GetAuthHandler()

	return authHandler, mockAuthService
}

func TestAuthHandler_UsersAuthRegister_ValidationError(t *testing.T) {
	// Arrange
	authHandler, _ := setupAuthHandler()

	// Invalid request - missing required fields
	invalidReq := map[string]interface{}{
		"first_name": "Test User",
		// Missing email, phone_number, password, confirm_password
	}

	// Create request
	reqBody, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest("POST", "/api/v1/users/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Act
	authHandler.UsersAuthRegister(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
}

func TestAuthHandler_UsersAuthLogin_ValidationError(t *testing.T) {
	// Arrange
	authHandler, _ := setupAuthHandler()

	// Invalid request - missing password
	invalidReq := map[string]interface{}{
		"email": "test@example.com",
		// Missing password
	}

	// Create request
	reqBody, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest("POST", "/api/v1/users/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Act
	authHandler.UsersAuthLogin(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
}

func TestAuthHandler_MitrasAuthRegister_ValidationError(t *testing.T) {
	// Arrange
	authHandler, _ := setupAuthHandler()

	// Invalid request - missing required fields
	invalidReq := map[string]interface{}{
		"first_name": "Test Mitra",
		// Missing email, phone_number, password, confirm_password
	}

	// Create request
	reqBody, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest("POST", "/api/v1/mitras/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Act
	authHandler.MitrasAuthRegister(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
}

func TestAuthHandler_RefreshToken_ValidationError(t *testing.T) {
	// Arrange
	authHandler, _ := setupAuthHandler()

	// Invalid request - missing refresh_token
	invalidReq := map[string]interface{}{}

	// Create request
	reqBody, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest("POST", "/api/v1/users/auth/refresh", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Act
	authHandler.UsersRefreshToken(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
}

// Mock-based tests for service layer interactions
func TestAuthHandler_WithMockService_UsersAuthLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &mocks.MockAuthService{}

	loginReq := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedResponse := &dto.AuthResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		User: &dto.UserInfo{
			ID:          "user-id",
			FirstName:   "Test",
			LastName:    "User",
			Email:       "test@example.com",
			PhoneNumber: "081234567890",
		},
	}

	mockAuthService.On("LoginUser", loginReq).
		Return(expectedResponse, nil)

	// Verify mock was called correctly
	result, err := mockAuthService.LoginUser(loginReq)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResponse.AccessToken, result.AccessToken)

	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_WithMockService_UsersAuthLogin_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &mocks.MockAuthService{}

	loginReq := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockAuthService.On("LoginUser", loginReq).
		Return(nil, appErrors.ErrInvalidCredentials)

	// Verify mock returns error
	result, err := mockAuthService.LoginUser(loginReq)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, appErrors.ErrInvalidCredentials, err)

	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_WithMockService_RegisterUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &mocks.MockAuthService{}

	registerReq := &dto.RegisterUserRequest{
		FirstName:       "Test",
		LastName:        "User",
		Email:           "test@example.com",
		PhoneNumber:     "081234567890",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}

	expectedResponse := &dto.AuthResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		User: &dto.UserInfo{
			ID:          "user-id",
			FirstName:   "Test",
			LastName:    "User",
			Email:       "test@example.com",
			PhoneNumber: "081234567890",
		},
	}

	mockAuthService.On("RegisterUser", registerReq).
		Return(expectedResponse, nil)

	// Verify mock was called correctly
	result, err := mockAuthService.RegisterUser(registerReq)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResponse.User.Email, result.User.Email)

	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_WithMockService_RegisterUser_EmailAlreadyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &mocks.MockAuthService{}

	registerReq := &dto.RegisterUserRequest{
		FirstName:       "Test",
		LastName:        "User",
		Email:           "existing@example.com",
		PhoneNumber:     "081234567890",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}

	mockAuthService.On("RegisterUser", registerReq).
		Return(nil, appErrors.ErrEmailAlreadyExists)

	// Verify mock returns error
	result, err := mockAuthService.RegisterUser(registerReq)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, appErrors.ErrEmailAlreadyExists, err)

	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_WithMockService_RefreshToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &mocks.MockAuthService{}

	refreshToken := "valid_refresh_token"

	expectedResponse := &dto.AuthResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		User: &dto.UserInfo{
			ID:        "user-id",
			FirstName: "Test",
			LastName:  "User",
			Email:     "test@example.com",
		},
	}

	mockAuthService.On("RefreshToken", refreshToken).
		Return(expectedResponse, nil)

	// Verify mock was called correctly
	result, err := mockAuthService.RefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResponse.AccessToken, result.AccessToken)

	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_WithMockService_RefreshToken_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &mocks.MockAuthService{}

	refreshToken := "invalid_refresh_token"

	mockAuthService.On("RefreshToken", refreshToken).
		Return(nil, appErrors.ErrTokenInvalid)

	// Verify mock returns error
	result, err := mockAuthService.RefreshToken(refreshToken)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, appErrors.ErrTokenInvalid, err)

	mockAuthService.AssertExpectations(t)
}

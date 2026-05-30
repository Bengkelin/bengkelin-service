package services_test

import (
	"testing"

	"github.com/Bengkelin/bengkelin-service/internal/dto"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/errors"
	"github.com/Bengkelin/bengkelin-service/tests/fixtures/mocks"
	"github.com/stretchr/testify/assert"
)

// TestAuthService_MockLoginUser_Success tests the mock service directly
func TestAuthService_MockLoginUser_Success(t *testing.T) {
	// Arrange
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

	mockAuthService.On("LoginUser", loginReq).Return(expectedResponse, nil)

	// Act
	result, err := mockAuthService.LoginUser(loginReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResponse.AccessToken, result.AccessToken)
	assert.Equal(t, expectedResponse.RefreshToken, result.RefreshToken)
	assert.Equal(t, expectedResponse.User.Email, result.User.Email)

	mockAuthService.AssertExpectations(t)
}

// TestAuthService_MockLoginUser_InvalidCredentials tests invalid credentials handling
func TestAuthService_MockLoginUser_InvalidCredentials(t *testing.T) {
	// Arrange
	mockAuthService := &mocks.MockAuthService{}

	loginReq := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockAuthService.On("LoginUser", loginReq).Return(nil, appErrors.ErrInvalidCredentials)

	// Act
	result, err := mockAuthService.LoginUser(loginReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, appErrors.ErrInvalidCredentials, err)

	mockAuthService.AssertExpectations(t)
}

// TestAuthService_MockRegisterUser_Success tests successful registration
func TestAuthService_MockRegisterUser_Success(t *testing.T) {
	// Arrange
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

	mockAuthService.On("RegisterUser", registerReq).Return(expectedResponse, nil)

	// Act
	result, err := mockAuthService.RegisterUser(registerReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResponse.User.Email, result.User.Email)
	assert.Equal(t, expectedResponse.User.FirstName, result.User.FirstName)

	mockAuthService.AssertExpectations(t)
}

// TestAuthService_MockRegisterUser_EmailAlreadyExists tests duplicate email handling
func TestAuthService_MockRegisterUser_EmailAlreadyExists(t *testing.T) {
	// Arrange
	mockAuthService := &mocks.MockAuthService{}

	registerReq := &dto.RegisterUserRequest{
		FirstName:       "Test",
		LastName:        "User",
		Email:           "existing@example.com",
		PhoneNumber:     "081234567890",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}

	mockAuthService.On("RegisterUser", registerReq).Return(nil, appErrors.ErrEmailAlreadyExists)

	// Act
	result, err := mockAuthService.RegisterUser(registerReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, appErrors.ErrEmailAlreadyExists, err)

	mockAuthService.AssertExpectations(t)
}

// TestAuthService_MockLoginMitra_Success tests mitra login
func TestAuthService_MockLoginMitra_Success(t *testing.T) {
	// Arrange
	mockAuthService := &mocks.MockAuthService{}

	loginReq := &dto.LoginRequest{
		Email:    "mitra@example.com",
		Password: "password123",
	}

	expectedResponse := &dto.AuthResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		Mitra: &dto.MitraInfo{
			ID:          "mitra-id",
			FirstName:   "Test",
			LastName:    "Mitra",
			Email:       "mitra@example.com",
			PhoneNumber: "081234567891",
		},
	}

	mockAuthService.On("LoginMitra", loginReq).Return(expectedResponse, nil)

	// Act
	result, err := mockAuthService.LoginMitra(loginReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResponse.Mitra.Email, result.Mitra.Email)
	assert.Equal(t, expectedResponse.Mitra.FirstName, result.Mitra.FirstName)

	mockAuthService.AssertExpectations(t)
}

// TestAuthService_MockRegisterMitra_Success tests successful mitra registration
func TestAuthService_MockRegisterMitra_Success(t *testing.T) {
	// Arrange
	mockAuthService := &mocks.MockAuthService{}

	registerReq := &dto.RegisterMitraRequest{
		FirstName:       "Test",
		LastName:        "Mitra",
		Email:           "mitra@example.com",
		PhoneNumber:     "081234567891",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}

	expectedResponse := &dto.AuthResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		Mitra: &dto.MitraInfo{
			ID:          "mitra-id",
			FirstName:   "Test",
			LastName:    "Mitra",
			Email:       "mitra@example.com",
			PhoneNumber: "081234567891",
		},
	}

	mockAuthService.On("RegisterMitra", registerReq).Return(expectedResponse, nil)

	// Act
	result, err := mockAuthService.RegisterMitra(registerReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResponse.Mitra.ID, result.Mitra.ID)

	mockAuthService.AssertExpectations(t)
}

// TestAuthService_MockRefreshToken_Success tests token refresh
func TestAuthService_MockRefreshToken_Success(t *testing.T) {
	// Arrange
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

	mockAuthService.On("RefreshToken", refreshToken).Return(expectedResponse, nil)

	// Act
	result, err := mockAuthService.RefreshToken(refreshToken)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResponse.AccessToken, result.AccessToken)
	assert.Equal(t, expectedResponse.RefreshToken, result.RefreshToken)

	mockAuthService.AssertExpectations(t)
}

// TestAuthService_MockRefreshToken_Invalid tests invalid token handling
func TestAuthService_MockRefreshToken_Invalid(t *testing.T) {
	// Arrange
	mockAuthService := &mocks.MockAuthService{}

	refreshToken := "invalid_refresh_token"

	mockAuthService.On("RefreshToken", refreshToken).Return(nil, appErrors.ErrTokenInvalid)

	// Act
	result, err := mockAuthService.RefreshToken(refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, appErrors.ErrTokenInvalid, err)

	mockAuthService.AssertExpectations(t)
}

// TestAuthService_MockLogout_Success tests logout
func TestAuthService_MockLogout_Success(t *testing.T) {
	// Arrange
	mockAuthService := &mocks.MockAuthService{}

	userID := "user-id"

	mockAuthService.On("Logout", userID).Return(nil)

	// Act
	err := mockAuthService.Logout(userID)

	// Assert
	assert.NoError(t, err)

	mockAuthService.AssertExpectations(t)
}

// TestAuthService_MockLogoutAll_Success tests logout from all devices
func TestAuthService_MockLogoutAll_Success(t *testing.T) {
	// Arrange
	mockAuthService := &mocks.MockAuthService{}

	userID := "user-id"

	mockAuthService.On("LogoutAll", userID).Return(nil)

	// Act
	err := mockAuthService.LogoutAll(userID)

	// Assert
	assert.NoError(t, err)

	mockAuthService.AssertExpectations(t)
}

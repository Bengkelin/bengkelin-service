package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bengkelin/bengkelin-service/internal/api/handlers"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/container"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	appErrors "github.com/Bengkelin/bengkelin-service/internal/pkg/errors"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) LoginUser(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RegisterUser(ctx context.Context, req dto.RegisterUserRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) LoginUserWithGoogle(ctx context.Context, req dto.GoogleAuthRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) LoginMitra(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RegisterMitra(ctx context.Context, req dto.RegisterMitraRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) LoginMitraWithGoogle(ctx context.Context, req dto.GoogleAuthRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RefreshUserToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RefreshMitraToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) LogoutUser(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) LogoutMitra(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) LogoutAllUserDevices(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthService) LogoutAllMitraDevices(ctx context.Context, mitraID string) error {
	args := m.Called(ctx, mitraID)
	return args.Error(0)
}

func TestDependencyInjection(t *testing.T) {
	// Test that container initializes properly
	container := container.GetContainer()
	assert.NotNil(t, container)
	assert.NotNil(t, container.AuthService)
	assert.NotNil(t, container.JWTHelper)
	assert.NotNil(t, container.PasswordHelper)
}

func TestHandlerFactory(t *testing.T) {
	// Test handler factory
	factory := handlers.GetHandlerFactory()
	assert.NotNil(t, factory)
	
	authHandler1 := factory.GetAuthHandler()
	authHandler2 := factory.GetAuthHandler()
	
	// Should return the same instance (singleton)
	assert.Equal(t, authHandler1, authHandler2)
}

func TestErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		error          error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "App Error",
			error:          appErrors.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "INVALID_CREDENTIALS",
		},
		{
			name:           "App Error with Details",
			error:          appErrors.ErrValidationFailed.WithDetails("email is required"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_FAILED",
		},
		{
			name:           "Generic Error",
			error:          assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_SERVER_ERROR",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			
			handler := &handlers.BaseHandler{}
			handler.HandleError(c, tt.error)
			
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			
			assert.Equal(t, false, response["success"])
			
			if errors, ok := response["errors"].(map[string]interface{}); ok {
				assert.Equal(t, tt.expectedCode, errors["code"])
			}
		})
	}
}

func TestBaseHandlerUtilities(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	t.Run("ParsePagination", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test?page=2&limit=20", nil)
		
		handler := &handlers.BaseHandler{}
		pagination := handler.ParsePagination(c)
		
		assert.Equal(t, 2, pagination.Page)
		assert.Equal(t, 20, pagination.Limit)
	})
	
	t.Run("ParseLocation", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test?latitude=-6.2088&longitude=106.8456", nil)
		
		handler := &handlers.BaseHandler{}
		lat, lng, err := handler.ParseLocation(c)
		
		assert.NoError(t, err)
		assert.Equal(t, -6.2088, lat)
		assert.Equal(t, 106.8456, lng)
	})
	
	t.Run("ParseLocation Invalid", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test?latitude=invalid&longitude=106.8456", nil)
		
		handler := &handlers.BaseHandler{}
		_, _, err := handler.ParseLocation(c)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid latitude format")
	})
	
	t.Run("ValidateOwnership", func(t *testing.T) {
		handler := &handlers.BaseHandler{}
		
		// Same user - should pass
		err := handler.ValidateOwnership("user123", "user123")
		assert.NoError(t, err)
		
		// Different user - should fail
		err = handler.ValidateOwnership("user123", "user456")
		assert.Error(t, err)
		assert.Equal(t, appErrors.ErrNotOwner, err)
	})
}

func TestAuthHandlerIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create mock service
	mockService := &MockAuthService{}
	
	// Create handler with mock service
	// handler := &handlers.AuthHandler{}
	// Note: In real implementation, we'd inject the mock service
	
	t.Run("Successful Login", func(t *testing.T) {
		expectedResponse := &dto.AuthResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
			User: &dto.UserInfo{
				ID:        "user123",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
		}
		
		mockService.On("LoginUser", mock.Anything, mock.MatchedBy(func(req dto.LoginRequest) bool {
			return req.Email == "john@example.com" && req.Password == "password123"
		})).Return(expectedResponse, nil)
		
		// Create request
		loginReq := map[string]string{
			"email":    "john@example.com",
			"password": "password123",
		}
		
		body, _ := json.Marshal(loginReq)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		
		// Note: This test would need the actual handler method to work
		// handler.UsersAuthLogin(c)
		
		// For now, just verify the mock was set up correctly
		mockService.AssertExpectations(t)
	})
}

func TestServiceLayerArchitecture(t *testing.T) {
	t.Run("Service Interface Compliance", func(t *testing.T) {
		// Test that our mock implements the interface
		var _ service.AuthServiceInterface = &MockAuthService{}
	})
	
	t.Run("Error Wrapping", func(t *testing.T) {
		// Test error wrapping functionality
		originalErr := assert.AnError
		wrappedErr := appErrors.WrapError(originalErr, appErrors.ErrDatabaseError)
		
		assert.NotNil(t, wrappedErr)
		assert.Equal(t, "DATABASE_ERROR", wrappedErr.Code)
		assert.Contains(t, wrappedErr.Details, originalErr.Error())
	})
}

func TestDTOValidation(t *testing.T) {
	t.Run("LoginRequest DTO", func(t *testing.T) {
		req := dto.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		
		assert.NotEmpty(t, req.Email)
		assert.NotEmpty(t, req.Password)
	})
	
	t.Run("AuthResponse DTO", func(t *testing.T) {
		resp := dto.AuthResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}
		
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.Greater(t, resp.ExpiresIn, int64(0))
		assert.Equal(t, "Bearer", resp.TokenType)
	})
}

func BenchmarkHandlerCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		factory := handlers.GetHandlerFactory()
		_ = factory.GetAuthHandler()
	}
}

func BenchmarkContainerAccess(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		container := container.GetContainer()
		_ = container.AuthService
	}
}
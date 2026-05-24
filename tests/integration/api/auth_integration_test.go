package api

import (
	"net/http"
	"testing"

	"github.com/Bengkelin/bengkelin-service/tests/helpers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthIntegrationTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *AuthIntegrationTestSuite) SetupSuite() {
	// Setup test database
	helpers.SetupTestDB()

	// Setup test router
	suite.router = helpers.SetupTestRouter()
}

func (suite *AuthIntegrationTestSuite) SetupTest() {
	// Clean database before each test
	helpers.CleanupTestDB()
}

func (suite *AuthIntegrationTestSuite) TearDownSuite() {
	// Cleanup after all tests
	helpers.CleanupTestDB()
}

func TestAuthIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationTestSuite))
}

func (suite *AuthIntegrationTestSuite) TestUserRegistrationFlow() {
	t := suite.T()

	// Test user registration
	userData := map[string]interface{}{
		"first_name":   "Integration",
		"last_name":    "Test User",
		"email":        "integration@example.com",
		"phone_number": "081234567890",
		"password":     "password123",
	}

	w := helpers.MakeRequest(t, suite.router, "POST", "/api/v1/users/auth/register", userData)
	response := helpers.AssertSuccessResponse(t, w, http.StatusCreated)

	// Verify response structure
	data := helpers.GetResponseData(t, response)
	helpers.AssertResponseContainsFields(t, data, []string{"access_token", "refresh_token", "user"})

	// Verify user data
	user := data["user"].(map[string]interface{})
	assert.Equal(t, userData["first_name"], user["first_name"])
	assert.Equal(t, userData["email"], user["email"])

	// Verify sensitive data is not exposed
	helpers.AssertResponseDoesNotContainFields(t, user, []string{"password"})
}

func (suite *AuthIntegrationTestSuite) TestUserRegistrationDuplicateEmail() {
	t := suite.T()

	// Register first user
	userData := map[string]interface{}{
		"first_name":   "First",
		"last_name":    "User",
		"email":        "duplicate@example.com",
		"phone_number": "081234567890",
		"password":     "password123",
	}

	w := helpers.MakeRequest(t, suite.router, "POST", "/api/v1/users/auth/register", userData)
	helpers.AssertSuccessResponse(t, w, http.StatusCreated)

	// Try to register with same email
	w = helpers.MakeRequest(t, suite.router, "POST", "/api/v1/users/auth/register", userData)
	helpers.AssertErrorResponse(t, w, http.StatusConflict)
}

func (suite *AuthIntegrationTestSuite) TestUserLoginFlow() {
	t := suite.T()

	// First register a user
	userData := map[string]interface{}{
		"first_name":   "Login",
		"last_name":    "Test",
		"email":        "login@example.com",
		"phone_number": "081234567890",
		"password":     "password123",
	}

	w := helpers.MakeRequest(t, suite.router, "POST", "/api/v1/users/auth/register", userData)
	helpers.AssertSuccessResponse(t, w, http.StatusCreated)

	// Then login
	loginData := map[string]interface{}{
		"email":    "login@example.com",
		"password": "password123",
	}

	w = helpers.MakeRequest(t, suite.router, "POST", "/api/v1/users/auth/login", loginData)
	response := helpers.AssertSuccessResponse(t, w, http.StatusOK)

	// Verify tokens are returned
	data := helpers.GetResponseData(t, response)
	helpers.AssertResponseContainsFields(t, data, []string{"access_token", "refresh_token"})
}

func (suite *AuthIntegrationTestSuite) TestUserLoginInvalidCredentials() {
	t := suite.T()

	loginData := map[string]interface{}{
		"email":    "nonexistent@example.com",
		"password": "wrongpassword",
	}

	w := helpers.MakeRequest(t, suite.router, "POST", "/api/v1/users/auth/login", loginData)
	helpers.AssertErrorResponse(t, w, http.StatusUnauthorized)
}

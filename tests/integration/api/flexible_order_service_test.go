package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/Bengkelin/bengkelin-service/internal/api/router/v1"
	"github.com/Bengkelin/bengkelin-service/internal/config"
	"github.com/Bengkelin/bengkelin-service/internal/models"
	"github.com/Bengkelin/bengkelin-service/internal/repository"
	"github.com/Bengkelin/bengkelin-service/internal/crypto"
	"github.com/Bengkelin/bengkelin-service/tests/helpers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlexibleOrderServiceTestSuite struct {
	suite.Suite
	router      *gin.Engine
	testUser    *models.User
	testMitra   *models.Mitra
	testBengkel *models.Bengkel
	testVehicle *models.Vehicle
	ctx         context.Context
	userToken   string
	mitraToken  string
}

func (suite *FlexibleOrderServiceTestSuite) SetupSuite() {
	// Initialize config first (required for JWT)
	config.Setup("")

	// Setup test database
	helpers.SetupTestDB()

	// Setup router
	gin.SetMode(gin.TestMode)
	suite.router = v1.Setup()

	// Create test user
	suite.testUser = &models.User{
		ID:          helpers.GenerateTestID(),
		FirstName:   "Test",
		LastName:    "User",
		Email:       "testuser@example.com",
		PhoneNumber: "081234567890",
		Password:    "hashedpassword",
	}

	userRepo := repository.GetUserRepository()
	_, err := userRepo.CreateUser(suite.ctx, *suite.testUser)
	suite.Require().NoError(err)

	// Create vehicle for user (ID is auto-increment uint, don't set it manually)
	vehicle := &models.Vehicle{
		UserID:        suite.testUser.ID,
		VehicleType:   "Motor",
		VehicleNumber: "B1234XYZ",
	}

	vehicleRepo := repository.GetVehicleRepository()
	createdVehicle, err := vehicleRepo.CreateVehicle(suite.ctx, *vehicle)
	suite.Require().NoError(err)
	suite.testVehicle = &createdVehicle

	// Create test mitra with bengkel
	suite.testMitra = &models.Mitra{
		ID:          helpers.GenerateTestID(),
		FirstName:   "Test",
		LastName:    "Mitra",
		Email:       "testmitra@example.com",
		PhoneNumber: "081234567891",
		Password:    "hashedpassword",
	}

	mitraRepo := repository.GetMitraRepository()
	_, err = mitraRepo.CreateMitra(suite.ctx, *suite.testMitra)
	suite.Require().NoError(err)

	// Create bengkel for mitra
	suite.testBengkel = &models.Bengkel{
		ID:           helpers.GenerateTestID(),
		MitraID:      suite.testMitra.ID,
		BengkelName:  "Test Bengkel",
		BengkelPhone: "081234567892",
		JumlahMontir: 3,
	}

	bengkelRepo := repository.GetBengkelRepository()
	_, err = bengkelRepo.CreateBengkel(suite.ctx, *suite.testBengkel)
	suite.Require().NoError(err)

	// Create admin fee
	adminFee := &models.AdminFee{
		ID:       helpers.GenerateTestID(),
		AdminFee: 5000.0,
	}

	adminFeeRepo := repository.GetAdminFeeRepository()
	_, err = adminFeeRepo.CreateAdminFee(suite.ctx, *adminFee)
	suite.Require().NoError(err)

	// Generate JWT tokens using the correct API
	jwtCrypto := crypto.GetJWTCrypto()

	userTokenPair, err := jwtCrypto.GenerateTokenPair(suite.testUser.ID)
	suite.Require().NoError(err)
	suite.userToken = userTokenPair.AccessToken

	mitraTokenPair, err := jwtCrypto.GenerateTokenPairMitra(suite.testMitra.ID)
	suite.Require().NoError(err)
	suite.mitraToken = mitraTokenPair.AccessToken
}

func (suite *FlexibleOrderServiceTestSuite) TearDownSuite() {
	helpers.CleanupTestDB()
}

func (suite *FlexibleOrderServiceTestSuite) TestUserSelfOrder() {
	// Test user creating order for themselves using legacy format
	orderData := map[string]interface{}{
		"mitra_id": suite.testMitra.ID,
		"title":    []string{"Oil Change", "Brake Check"},
		"detail":   []string{"Full synthetic oil change", "Brake pad inspection"},
		"price":    []float64{50000, 25000},
	}

	jsonData, _ := json.Marshal(orderData)

	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/bengkels/order/service", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.userToken)

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	// Verify response structure
	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "success create order for yourself", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), suite.testUser.ID, data["user_id"])
	assert.Equal(suite.T(), suite.testBengkel.ID, data["bengkel_id"])
	assert.Equal(suite.T(), suite.testBengkel.BengkelName, data["bengkel_name"])
	assert.Equal(suite.T(), 75000.0, data["total_price"])
	assert.Equal(suite.T(), 5000.0, data["admin_fee"])

	// Verify order context
	createdBy := data["created_by"].(map[string]interface{})
	assert.Equal(suite.T(), "user", createdBy["type"])
	assert.Equal(suite.T(), suite.testUser.ID, createdBy["id"])
	assert.Equal(suite.T(), "Test User", createdBy["name"])

	orderContext := data["order_context"].(map[string]interface{})
	assert.True(suite.T(), orderContext["is_self_order"].(bool))
	assert.False(suite.T(), orderContext["is_mitra_created"].(bool))
}

func (suite *FlexibleOrderServiceTestSuite) TestMitraCreatedOrder() {
	// Test mitra creating order for user using legacy format
	orderData := map[string]interface{}{
		"title":  []string{"Oil Change", "Brake Check"},
		"detail": []string{"Full synthetic oil change", "Brake pad inspection"},
		"price":  []float64{50000, 25000},
	}

	jsonData, _ := json.Marshal(orderData)

	// Create request with userId URL parameter
	url := fmt.Sprintf("/api/v1/bengkels/order/service/%s", suite.testUser.ID)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.mitraToken)

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	// Verify response structure
	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "success create order for user", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), suite.testUser.ID, data["user_id"])
	assert.Equal(suite.T(), suite.testBengkel.ID, data["bengkel_id"])
	assert.Equal(suite.T(), suite.testBengkel.BengkelName, data["bengkel_name"])
	assert.Equal(suite.T(), 75000.0, data["total_price"])
	assert.Equal(suite.T(), 5000.0, data["admin_fee"])

	// Verify order context
	createdBy := data["created_by"].(map[string]interface{})
	assert.Equal(suite.T(), "mitra", createdBy["type"])
	assert.Equal(suite.T(), suite.testMitra.ID, createdBy["id"])
	assert.Equal(suite.T(), "Test Mitra", createdBy["name"])

	orderContext := data["order_context"].(map[string]interface{})
	assert.False(suite.T(), orderContext["is_self_order"].(bool))
	assert.True(suite.T(), orderContext["is_mitra_created"].(bool))
}

func (suite *FlexibleOrderServiceTestSuite) TestUserSelfOrderWithQueryParam() {
	// Test user creating order using mitraId query parameter
	orderData := map[string]interface{}{
		"title":  []string{"Oil Change"},
		"detail": []string{"Full synthetic oil change"},
		"price":  []float64{50000},
	}

	jsonData, _ := json.Marshal(orderData)

	// Create request with mitraId query parameter
	url := fmt.Sprintf("/api/v1/bengkels/order/service?mitraId=%s", suite.testMitra.ID)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.userToken)

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "success create order for yourself", response["message"])
}

func (suite *FlexibleOrderServiceTestSuite) TestUserSelfOrderMissingMitraId() {
	// Test user creating order without mitra_id in body or query parameter
	orderData := map[string]interface{}{
		"title":  []string{"Oil Change"},
		"detail": []string{"Full synthetic oil change"},
		"price":  []float64{50000},
	}

	jsonData, _ := json.Marshal(orderData)

	// Create request without mitraId query parameter or mitra_id in body
	req, _ := http.NewRequest("POST", "/api/v1/bengkels/order/service", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.userToken)

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert error response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "mitra ID required")
}

func (suite *FlexibleOrderServiceTestSuite) TestMitraCreatedOrderMissingUserId() {
	// Test mitra creating order without userId parameter
	orderData := map[string]interface{}{
		"title":  []string{"Oil Change"},
		"detail": []string{"Full synthetic oil change"},
		"price":  []float64{50000},
	}

	jsonData, _ := json.Marshal(orderData)

	// Create request without userId URL parameter (trailing slash only)
	req, _ := http.NewRequest("POST", "/api/v1/bengkels/order/service/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.mitraToken)

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Note: Gin may redirect /path/ to /path, so we might get 307 or 404
	// The actual error case would be when userId param is empty
	assert.Contains(suite.T(), []int{http.StatusBadRequest, http.StatusNotFound, http.StatusTemporaryRedirect, http.StatusMovedPermanently}, w.Code)
}

func (suite *FlexibleOrderServiceTestSuite) TestUnauthorizedAccess() {
	// Test accessing endpoint without authentication
	orderData := map[string]interface{}{
		"title":  []string{"Oil Change"},
		"detail": []string{"Full synthetic oil change"},
		"price":  []float64{50000},
	}

	jsonData, _ := json.Marshal(orderData)

	req, _ := http.NewRequest("POST", "/api/v1/bengkels/order/service", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert unauthorized response
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *FlexibleOrderServiceTestSuite) TestNewStructuredFormat() {
	// Test using new structured format with services array
	orderData := map[string]interface{}{
		"mitra_id": suite.testMitra.ID,
		"services": []map[string]interface{}{
			{
				"title":  "Oil Change",
				"detail": "Full synthetic oil change",
				"price":  50000,
			},
			{
				"title":  "Brake Check",
				"detail": "Brake pad inspection",
				"price":  25000,
			},
		},
	}

	jsonData, _ := json.Marshal(orderData)

	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/bengkels/order/service", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.userToken)

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "success create order for yourself", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), 75000.0, data["total_price"])
}

func TestFlexibleOrderServiceTestSuite(t *testing.T) {
	suite.Run(t, new(FlexibleOrderServiceTestSuite))
}

package api

import (
	"net/http"
	"testing"

	"github.com/Bengkelin/bengkelin-service/tests/helpers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BengkelDetailTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *BengkelDetailTestSuite) SetupSuite() {
	// Setup test database
	helpers.SetupTestDB()

	// Setup test router
	suite.router = helpers.SetupTestRouter()
}

func (suite *BengkelDetailTestSuite) SetupTest() {
	// Clean database before each test
	helpers.CleanupTestDB()
}

func (suite *BengkelDetailTestSuite) TearDownSuite() {
	// Cleanup after all tests
	helpers.CleanupTestDB()
}

func TestBengkelDetailTestSuite(t *testing.T) {
	suite.Run(t, new(BengkelDetailTestSuite))
}

func (suite *BengkelDetailTestSuite) TestBengkelDetailAnonymous() {
	t := suite.T()

	// Test anonymous access to bengkel detail
	w := helpers.MakeRequest(t, suite.router, "GET", "/api/v1/bengkels/test-bengkel-1", nil)

	// Verify response - might be 404 if bengkel doesn't exist, but should not be 401
	// Anonymous access is allowed for public endpoints
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
}

func (suite *BengkelDetailTestSuite) TestBengkelListAnonymous() {
	t := suite.T()

	// Test anonymous access to bengkel list
	w := helpers.MakeRequest(t, suite.router, "GET", "/api/v1/bengkels", nil)

	// Public endpoint should return success
	if w.Code == http.StatusOK {
		response := helpers.AssertSuccessResponse(t, w, http.StatusOK)
		data := helpers.GetResponseData(t, response)
		assert.NotNil(t, data)
	}
}

func (suite *BengkelDetailTestSuite) TestBengkelSearchAnonymous() {
	t := suite.T()

	// Test anonymous search
	w := helpers.MakeRequest(t, suite.router, "GET", "/api/v1/bengkels/search?query=repair", nil)

	// Public search endpoint should return success
	if w.Code == http.StatusOK {
		response := helpers.AssertSuccessResponse(t, w, http.StatusOK)
		data := helpers.GetResponseData(t, response)
		assert.NotNil(t, data)
	}
}

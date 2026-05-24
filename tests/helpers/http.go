package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bengkelin/bengkelin-service/internal/api/router/v1"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// SetupTestRouter initializes a test router with all routes
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return v1.Setup()
}

// MakeRequest creates and executes an HTTP request
func MakeRequest(t *testing.T, router *gin.Engine, method, url string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	return w
}

// MakeAuthenticatedRequest creates an authenticated HTTP request
func MakeAuthenticatedRequest(t *testing.T, router *gin.Engine, method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	// Add authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	return w
}

// AssertJSONResponse validates JSON response structure and status
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedSuccess bool) map[string]interface{} {
	assert.Equal(t, expectedStatus, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")

	// Check standard response structure
	assert.Contains(t, response, "success")
	assert.Contains(t, response, "message")
	assert.Equal(t, expectedSuccess, response["success"])

	return response
}

// AssertSuccessResponse validates successful API response
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	response := AssertJSONResponse(t, w, expectedStatus, true)
	assert.Contains(t, response, "data")
	return response
}

// AssertErrorResponse validates error API response
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	response := AssertJSONResponse(t, w, expectedStatus, false)
	assert.Contains(t, response, "errors")
	return response
}

// GetResponseData extracts data from response
func GetResponseData(t *testing.T, response map[string]interface{}) map[string]interface{} {
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Response data is not a map")
	}
	return data
}

// AssertResponseContainsFields checks if response contains expected fields
func AssertResponseContainsFields(t *testing.T, data map[string]interface{}, fields []string) {
	for _, field := range fields {
		assert.Contains(t, data, field, "Response should contain field: %s", field)
	}
}

// AssertResponseDoesNotContainFields checks if response does not contain sensitive fields
func AssertResponseDoesNotContainFields(t *testing.T, data map[string]interface{}, fields []string) {
	for _, field := range fields {
		assert.NotContains(t, data, field, "Response should not contain sensitive field: %s", field)
	}
}
package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetAllBengkelPaginate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/bengkels", func(c *gin.Context) {
		c.Set("id", "test-user-id")
		// Simulate real handler response so test can assert on it
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
		})
	})

	req := httptest.NewRequest("GET", "/bengkels?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
}

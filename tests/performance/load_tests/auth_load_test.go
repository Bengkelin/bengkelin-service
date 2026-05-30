package load_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	v1 "github.com/Bengkelin/bengkelin-service/internal/api/router/v1"
	"github.com/Bengkelin/bengkelin-service/internal/config"
	"github.com/Bengkelin/bengkelin-service/tests/helpers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// LoadTestConfig defines configuration for load tests
type LoadTestConfig struct {
	ConcurrentUsers int
	RequestsPerUser int
	TestDuration    time.Duration
	TargetRPS       int
}

// LoadTestResult contains results of load test
type LoadTestResult struct {
	TotalRequests     int
	SuccessfulReqs    int
	FailedReqs        int
	AverageLatency    time.Duration
	MaxLatency        time.Duration
	MinLatency        time.Duration
	RequestsPerSecond float64
	ErrorRate         float64
}

func TestAuthEndpointsLoadTest(t *testing.T) {
	// Initialize config first (required for the application)
	config.Setup("")

	// Setup test environment
	helpers.SetupTestDB()
	defer helpers.CleanupTestDB()

	// Setup test router
	gin.SetMode(gin.TestMode)
	router := v1.Setup()

	config := LoadTestConfig{
		ConcurrentUsers: 50,
		RequestsPerUser: 20,
		TestDuration:    30 * time.Second,
		TargetRPS:       100,
	}

	t.Run("UserRegistrationLoad", func(t *testing.T) {
		result := runUserRegistrationLoadTest(t, router, config)

		// Assert performance requirements
		assert.Less(t, result.AverageLatency, 500*time.Millisecond, "Average latency should be under 500ms")
		assert.Less(t, result.ErrorRate, 0.05, "Error rate should be under 5%")
		assert.Greater(t, result.RequestsPerSecond, float64(config.TargetRPS)*0.8, "Should achieve at least 80% of target RPS")

		t.Logf("User Registration Load Test Results:")
		t.Logf("  Total Requests: %d", result.TotalRequests)
		t.Logf("  Successful: %d", result.SuccessfulReqs)
		t.Logf("  Failed: %d", result.FailedReqs)
		t.Logf("  Average Latency: %v", result.AverageLatency)
		t.Logf("  Max Latency: %v", result.MaxLatency)
		t.Logf("  Min Latency: %v", result.MinLatency)
		t.Logf("  Requests/Second: %.2f", result.RequestsPerSecond)
		t.Logf("  Error Rate: %.2f%%", result.ErrorRate*100)
	})

	t.Run("UserLoginLoad", func(t *testing.T) {
		// Pre-create users for login test
		createTestUsers(t, router, config.ConcurrentUsers)

		result := runUserLoginLoadTest(t, router, config)

		// Assert performance requirements
		assert.Less(t, result.AverageLatency, 300*time.Millisecond, "Login latency should be under 300ms")
		assert.Less(t, result.ErrorRate, 0.02, "Login error rate should be under 2%")

		t.Logf("User Login Load Test Results:")
		t.Logf("  Total Requests: %d", result.TotalRequests)
		t.Logf("  Successful: %d", result.SuccessfulReqs)
		t.Logf("  Failed: %d", result.FailedReqs)
		t.Logf("  Average Latency: %v", result.AverageLatency)
		t.Logf("  Requests/Second: %.2f", result.RequestsPerSecond)
		t.Logf("  Error Rate: %.2f%%", result.ErrorRate*100)
	})

	t.Run("TokenRefreshLoad", func(t *testing.T) {
		// Pre-create users and get refresh tokens
		refreshTokens := createTestUsersWithTokens(t, router, config.ConcurrentUsers)

		result := runTokenRefreshLoadTest(t, router, config, refreshTokens)

		// Assert performance requirements
		assert.Less(t, result.AverageLatency, 200*time.Millisecond, "Token refresh should be under 200ms")
		assert.Less(t, result.ErrorRate, 0.01, "Token refresh error rate should be under 1%")

		t.Logf("Token Refresh Load Test Results:")
		t.Logf("  Total Requests: %d", result.TotalRequests)
		t.Logf("  Successful: %d", result.SuccessfulReqs)
		t.Logf("  Average Latency: %v", result.AverageLatency)
		t.Logf("  Requests/Second: %.2f", result.RequestsPerSecond)
		t.Logf("  Error Rate: %.2f%%", result.ErrorRate*100)
	})
}

func runUserRegistrationLoadTest(t *testing.T, router *gin.Engine, config LoadTestConfig) LoadTestResult {
	var wg sync.WaitGroup
	var mu sync.Mutex

	results := make([]RequestResult, 0, config.ConcurrentUsers*config.RequestsPerUser)
	startTime := time.Now()

	// Launch concurrent users
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			for j := 0; j < config.RequestsPerUser; j++ {
				// Create unique user data with correct field names
				userData := map[string]interface{}{
					"first_name":   fmt.Sprintf("LoadTest User %d-%d", userID, j),
					"last_name":    "Test",
					"email":        fmt.Sprintf("loadtest%d-%d@example.com", userID, j),
					"phone_number": fmt.Sprintf("08123456%04d", userID*1000+j),
					"password":     "password123",
				}

				reqStart := time.Now()
				w := helpers.MakeRequest(t, router, "POST", "/api/v1/users/auth/register", userData)
				latency := time.Since(reqStart)

				success := w.Code == http.StatusCreated

				mu.Lock()
				results = append(results, RequestResult{
					Success: success,
					Latency: latency,
					Status:  w.Code,
				})
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	return calculateLoadTestResult(results, totalDuration)
}

func runUserLoginLoadTest(t *testing.T, router *gin.Engine, config LoadTestConfig) LoadTestResult {
	var wg sync.WaitGroup
	var mu sync.Mutex

	results := make([]RequestResult, 0, config.ConcurrentUsers*config.RequestsPerUser)
	startTime := time.Now()

	// Launch concurrent users
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			for j := 0; j < config.RequestsPerUser; j++ {
				// Use pre-created user credentials
				loginData := map[string]interface{}{
					"email":    fmt.Sprintf("testuser%d@example.com", userID),
					"password": "password123",
				}

				reqStart := time.Now()
				w := helpers.MakeRequest(t, router, "POST", "/api/v1/users/auth/login", loginData)
				latency := time.Since(reqStart)

				success := w.Code == http.StatusOK

				mu.Lock()
				results = append(results, RequestResult{
					Success: success,
					Latency: latency,
					Status:  w.Code,
				})
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	return calculateLoadTestResult(results, totalDuration)
}

func runTokenRefreshLoadTest(t *testing.T, router *gin.Engine, config LoadTestConfig, refreshTokens []string) LoadTestResult {
	var wg sync.WaitGroup
	var mu sync.Mutex

	results := make([]RequestResult, 0, config.ConcurrentUsers*config.RequestsPerUser)
	startTime := time.Now()

	// Launch concurrent users
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			refreshToken := refreshTokens[userID%len(refreshTokens)]

			for j := 0; j < config.RequestsPerUser; j++ {
				refreshData := map[string]interface{}{
					"refresh_token": refreshToken,
				}

				reqStart := time.Now()
				w := helpers.MakeRequest(t, router, "POST", "/api/v1/users/auth/refresh", refreshData)
				latency := time.Since(reqStart)

				success := w.Code == http.StatusOK

				// Update refresh token if successful
				if success {
					var response map[string]interface{}
					json.Unmarshal(w.Body.Bytes(), &response)
					if data, ok := response["data"].(map[string]interface{}); ok {
						if newToken, ok := data["refresh_token"].(string); ok {
							refreshToken = newToken
						}
					}
				}

				mu.Lock()
				results = append(results, RequestResult{
					Success: success,
					Latency: latency,
					Status:  w.Code,
				})
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	return calculateLoadTestResult(results, totalDuration)
}

// RequestResult represents the result of a single request
type RequestResult struct {
	Success bool
	Latency time.Duration
	Status  int
}

func calculateLoadTestResult(results []RequestResult, totalDuration time.Duration) LoadTestResult {
	if len(results) == 0 {
		return LoadTestResult{}
	}

	var totalLatency time.Duration
	var successCount, failCount int
	var maxLatency, minLatency time.Duration

	minLatency = results[0].Latency

	for _, result := range results {
		totalLatency += result.Latency

		if result.Success {
			successCount++
		} else {
			failCount++
		}

		if result.Latency > maxLatency {
			maxLatency = result.Latency
		}
		if result.Latency < minLatency {
			minLatency = result.Latency
		}
	}

	totalRequests := len(results)
	averageLatency := totalLatency / time.Duration(totalRequests)
	requestsPerSecond := float64(totalRequests) / totalDuration.Seconds()
	errorRate := float64(failCount) / float64(totalRequests)

	return LoadTestResult{
		TotalRequests:     totalRequests,
		SuccessfulReqs:    successCount,
		FailedReqs:        failCount,
		AverageLatency:    averageLatency,
		MaxLatency:        maxLatency,
		MinLatency:        minLatency,
		RequestsPerSecond: requestsPerSecond,
		ErrorRate:         errorRate,
	}
}

func createTestUsers(t *testing.T, router *gin.Engine, count int) {
	for i := 0; i < count; i++ {
		userData := map[string]interface{}{
			"first_name":   fmt.Sprintf("Test User %d", i),
			"last_name":    "Test",
			"email":        fmt.Sprintf("testuser%d@example.com", i),
			"phone_number": fmt.Sprintf("08123456%04d", i),
			"password":     "password123",
		}

		helpers.MakeRequest(t, router, "POST", "/api/v1/users/auth/register", userData)
	}
}

func createTestUsersWithTokens(t *testing.T, router *gin.Engine, count int) []string {
	refreshTokens := make([]string, count)

	for i := 0; i < count; i++ {
		userData := map[string]interface{}{
			"first_name":   fmt.Sprintf("Token Test User %d", i),
			"last_name":    "Test",
			"email":        fmt.Sprintf("tokenuser%d@example.com", i),
			"phone_number": fmt.Sprintf("08123457%04d", i),
			"password":     "password123",
		}

		w := helpers.MakeRequest(t, router, "POST", "/api/v1/users/auth/register", userData)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if data, ok := response["data"].(map[string]interface{}); ok {
			if token, ok := data["refresh_token"].(string); ok {
				refreshTokens[i] = token
			}
		}
	}

	return refreshTokens
}

// setupBenchmarkHelper sets up the test environment for benchmarks
func setupBenchmarkHelper() (*gin.Engine, func()) {
	config.Setup("")
	helpers.SetupTestDB()
	gin.SetMode(gin.TestMode)
	router := v1.Setup()
	cleanup := func() {
		helpers.CleanupTestDB()
	}
	return router, cleanup
}

// BenchmarkUserRegistration benchmarks user registration endpoint
func BenchmarkUserRegistration(b *testing.B) {
	router, cleanup := setupBenchmarkHelper()
	defer cleanup()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			userData := map[string]interface{}{
				"first_name":   fmt.Sprintf("Bench User %d", i),
				"last_name":    "Test",
				"email":        fmt.Sprintf("bench%d@example.com", i),
				"phone_number": fmt.Sprintf("08123458%04d", i),
				"password":     "password123",
			}

			reqBody, _ := json.Marshal(userData)
			req, _ := http.NewRequest("POST", "/api/v1/users/auth/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			i++
		}
	})
}

// BenchmarkUserLogin benchmarks user login endpoint
func BenchmarkUserLogin(b *testing.B) {
	router, cleanup := setupBenchmarkHelper()
	defer cleanup()

	// Create a test user
	userData := map[string]interface{}{
		"first_name":   "Bench",
		"last_name":    "Login User",
		"email":        "benchlogin@example.com",
		"phone_number": "081234567890",
		"password":     "password123",
	}

	// Use httptest directly to avoid testing.T requirement
	reqBody, _ := json.Marshal(userData)
	req, _ := http.NewRequest("POST", "/api/v1/users/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginData := map[string]interface{}{
		"email":    "benchlogin@example.com",
		"password": "password123",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reqBody, _ := json.Marshal(loginData)
			req, _ := http.NewRequest("POST", "/api/v1/users/auth/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

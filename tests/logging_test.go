package tests

import (
	"context"
	"runtime"
	"testing"
	"time"

	applog "github.com/Bengkelin/bengkelin-service/internal/log"
)

func TestLoggingEnhancement(t *testing.T) {
	// Initialize logger for testing
	applog.Setup("test")
	
	// Test basic logging functions
	t.Run("Basic Logging Functions", func(t *testing.T) {
		applog.Info("Test info message", "key", "value")
		applog.Debug("Test debug message", "key", "value")
		applog.Warn("Test warn message", "key", "value")
		applog.Error("Test error message", "key", "value")
	})
	
	// Test context-aware logging
	t.Run("Context-Aware Logging", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "request_id", "test-request-123")
		ctx = context.WithValue(ctx, "user_id", "test-user-456")
		
		applog.InfoCtx(ctx, "Test context-aware info", "operation", "test")
		applog.DebugCtx(ctx, "Test context-aware debug", "operation", "test")
		applog.WarnCtx(ctx, "Test context-aware warn", "operation", "test")
		applog.ErrorCtx(ctx, "Test context-aware error", "operation", "test")
	})
	
	// Test performance logging
	t.Run("Performance Logging", func(t *testing.T) {
		start := time.Now()
		time.Sleep(10 * time.Millisecond) // Simulate work
		
		applog.LogDuration("test_operation", start, "test_key", "test_value")
		
		ctx := context.WithValue(context.Background(), "request_id", "perf-test-123")
		applog.LogDurationCtx(ctx, "test_operation_ctx", start, "test_key", "test_value")
	})
	
	// Test HTTP request logging
	t.Run("HTTP Request Logging", func(t *testing.T) {
		applog.LogHTTPRequest(
			"GET", 
			"/api/v1/test", 
			"test-agent", 
			"127.0.0.1", 
			200, 
			50*time.Millisecond,
			"test_key", "test_value",
		)
	})
	
	// Test database operation logging
	t.Run("Database Operation Logging", func(t *testing.T) {
		// Test successful operation
		applog.LogDBOperation("SELECT", "users", 25*time.Millisecond, nil, "query", "test")
		
		// Test slow operation
		applog.LogDBOperation("SELECT", "users", 150*time.Millisecond, nil, "query", "slow_test")
	})
	
	// Test security event logging
	t.Run("Security Event Logging", func(t *testing.T) {
		applog.LogSecurityEvent(
			"suspicious_login_attempt",
			"test-user-123",
			"192.168.1.100",
			"suspicious-agent",
			"reason", "multiple_failed_attempts",
		)
		
		applog.LogAuthEvent(
			"login",
			"test-user-123",
			"test@example.com",
			"192.168.1.100",
			true,
			"method", "password",
		)
	})
	
	// Test business event logging
	t.Run("Business Event Logging", func(t *testing.T) {
		applog.LogBusinessEvent(
			"user_registration",
			"user_id", "test-user-123",
			"email", "test@example.com",
		)
	})
	
	// Test error logging with stack trace
	t.Run("Error Logging with Stack Trace", func(t *testing.T) {
		testErr := &TestError{message: "test error"}
		applog.LogError(testErr, "Test error occurred", "context", "test")
		
		ctx := context.WithValue(context.Background(), "request_id", "error-test-123")
		applog.LogErrorCtx(ctx, testErr, "Test context error occurred", "context", "test")
	})
}

// TestError is a simple error type for testing
type TestError struct {
	message string
}

func (e *TestError) Error() string {
	return e.message
}

func TestLoggerConfiguration(t *testing.T) {
	// Test different environment configurations
	environments := []string{"development", "staging", "test"}
	
	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			applog.Setup(env)
			applog.Info("Logger configured for environment", "env", env)
		})
	}
	
	// Test production environment separately with error handling
	t.Run("Environment_production", func(t *testing.T) {
		// Skip production test on Windows as it requires specific log directories
		if runtime.GOOS == "windows" {
			t.Skip("Skipping production environment test on Windows")
			return
		}
		
		applog.Setup("production")
		applog.Info("Logger configured for environment", "env", "production")
	})
}

func TestLogLevelChanges(t *testing.T) {
	applog.Setup("test")
	
	// Test different log levels
	levels := []applog.LogLevel{
		applog.DebugLevel,
		applog.InfoLevel,
		applog.WarnLevel,
		applog.ErrorLevel,
	}
	
	for _, level := range levels {
		t.Run("Level_"+string(level), func(t *testing.T) {
			applog.SetLevel(level)
			applog.Info("Log level changed", "level", level)
		})
	}
}
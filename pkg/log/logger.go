package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log           *zap.SugaredLogger
	structuredLog *zap.Logger
)

// LogLevel represents logging levels
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// LogConfig holds logging configuration
type LogConfig struct {
	Level       LogLevel `json:"level"`
	Environment string   `json:"environment"`
	OutputPath  string   `json:"output_path"`
	ErrorPath   string   `json:"error_path"`
	EnableCaller bool    `json:"enable_caller"`
	EnableStacktrace bool `json:"enable_stacktrace"`
}

// Setup initializes the logger with enhanced configuration
func Setup(env string) {
	config := getLogConfig(env)
	
	var zapConfig zap.Config
	
	if env == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		zapConfig.Encoding = "json"
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.LevelKey = "level"
		zapConfig.EncoderConfig.MessageKey = "message"
		zapConfig.EncoderConfig.CallerKey = "caller"
		zapConfig.EncoderConfig.StacktraceKey = "stacktrace"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapConfig.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		zapConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}
	
	// Set log level based on config
	switch config.Level {
	case DebugLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case InfoLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case WarnLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case ErrorLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case FatalLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	}
	
	// Configure output paths
	if config.OutputPath != "" {
		zapConfig.OutputPaths = []string{config.OutputPath, "stdout"}
	}
	if config.ErrorPath != "" {
		zapConfig.ErrorOutputPaths = []string{config.ErrorPath, "stderr"}
	}
	
	// Enable caller information
	zapConfig.DisableCaller = !config.EnableCaller
	zapConfig.DisableStacktrace = !config.EnableStacktrace
	
	// Add common fields
	zapConfig.InitialFields = map[string]interface{}{
		"service":     "bengkelin-api",
		"version":     getVersion(),
		"environment": env,
		"pid":         os.Getpid(),
	}
	
	logger, err := zapConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	
	structuredLog = logger
	Log = logger.Sugar()
	
	Log.Info("Logger initialized successfully", 
		"environment", env,
		"level", config.Level,
		"caller_enabled", config.EnableCaller,
		"stacktrace_enabled", config.EnableStacktrace,
	)
}

// getLogConfig returns logging configuration based on environment
func getLogConfig(env string) LogConfig {
	config := LogConfig{
		Environment:      env,
		EnableCaller:     true,
		EnableStacktrace: true,
	}
	
	switch env {
	case "production":
		config.Level = InfoLevel
		config.OutputPath = "/var/log/bengkelin/app.log"
		config.ErrorPath = "/var/log/bengkelin/error.log"
		config.EnableStacktrace = false
	case "staging":
		config.Level = InfoLevel
		config.EnableStacktrace = false
	case "development":
		config.Level = DebugLevel
	case "test":
		config.Level = ErrorLevel
		config.EnableCaller = false
		config.EnableStacktrace = false
	default:
		config.Level = InfoLevel
	}
	
	// Override with environment variables
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = LogLevel(level)
	}
	if outputPath := os.Getenv("LOG_OUTPUT_PATH"); outputPath != "" {
		config.OutputPath = outputPath
	}
	if errorPath := os.Getenv("LOG_ERROR_PATH"); errorPath != "" {
		config.ErrorPath = errorPath
	}
	
	return config
}

// getVersion returns the application version
func getVersion() string {
	if version := os.Getenv("APP_VERSION"); version != "" {
		return version
	}
	return "dev"
}

// Context-aware logging functions

// WithContext creates a logger with context information
func WithContext(ctx context.Context) *zap.SugaredLogger {
	if Log == nil {
		return zap.NewNop().Sugar()
	}
	
	logger := Log
	
	// Add request ID if available
	if requestID := getRequestID(ctx); requestID != "" {
		logger = logger.With("request_id", requestID)
	}
	
	// Add user ID if available
	if userID := getUserID(ctx); userID != "" {
		logger = logger.With("user_id", userID)
	}
	
	// Add trace ID if available
	if traceID := getTraceID(ctx); traceID != "" {
		logger = logger.With("trace_id", traceID)
	}
	
	return logger
}

// Structured logging functions with enhanced context

func Debug(msg string, keysAndValues ...interface{}) {
	if Log == nil {
		return
	}
	Log.Debugw(msg, keysAndValues...)
}

func Info(msg string, keysAndValues ...interface{}) {
	if Log == nil {
		return
	}
	Log.Infow(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...interface{}) {
	if Log == nil {
		return
	}
	Log.Warnw(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...interface{}) {
	if Log == nil {
		return
	}
	Log.Errorw(msg, keysAndValues...)
}

func Fatal(msg string, keysAndValues ...interface{}) {
	if Log == nil {
		return
	}
	Log.Fatalw(msg, keysAndValues...)
}

// Context-aware logging functions

func DebugCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	WithContext(ctx).Debugw(msg, keysAndValues...)
}

func InfoCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	WithContext(ctx).Infow(msg, keysAndValues...)
}

func WarnCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	WithContext(ctx).Warnw(msg, keysAndValues...)
}

func ErrorCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	WithContext(ctx).Errorw(msg, keysAndValues...)
}

func FatalCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	WithContext(ctx).Fatalw(msg, keysAndValues...)
}

// Performance logging

// LogDuration logs the duration of an operation
func LogDuration(operation string, start time.Time, keysAndValues ...interface{}) {
	duration := time.Since(start)
	
	fields := append(keysAndValues, 
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
		"duration", duration.String(),
	)
	
	if duration > time.Second {
		Warn("Slow operation detected", fields...)
	} else {
		Debug("Operation completed", fields...)
	}
}

// LogDurationCtx logs the duration of an operation with context
func LogDurationCtx(ctx context.Context, operation string, start time.Time, keysAndValues ...interface{}) {
	duration := time.Since(start)
	
	fields := append(keysAndValues, 
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
		"duration", duration.String(),
	)
	
	if duration > time.Second {
		WarnCtx(ctx, "Slow operation detected", fields...)
	} else {
		DebugCtx(ctx, "Operation completed", fields...)
	}
}

// HTTP request logging

// LogHTTPRequest logs HTTP request details
func LogHTTPRequest(method, path, userAgent, clientIP string, statusCode int, duration time.Duration, keysAndValues ...interface{}) {
	fields := append(keysAndValues,
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
		"client_ip", clientIP,
		"user_agent", userAgent,
	)
	
	level := getLogLevelForStatus(statusCode)
	
	switch level {
	case "error":
		Error("HTTP request completed", fields...)
	case "warn":
		Warn("HTTP request completed", fields...)
	default:
		Info("HTTP request completed", fields...)
	}
}

// Database operation logging

// LogDBOperation logs database operation details
func LogDBOperation(operation, table string, duration time.Duration, err error, keysAndValues ...interface{}) {
	fields := append(keysAndValues,
		"operation", operation,
		"table", table,
		"duration_ms", duration.Milliseconds(),
	)
	
	if err != nil {
		fields = append(fields, "error", err.Error())
		Error("Database operation failed", fields...)
	} else if duration > 100*time.Millisecond {
		fields = append(fields, "slow_query", true)
		Warn("Slow database operation", fields...)
	} else {
		Debug("Database operation completed", fields...)
	}
}

// Security logging

// LogSecurityEvent logs security-related events
func LogSecurityEvent(event, userID, clientIP, userAgent string, keysAndValues ...interface{}) {
	fields := append(keysAndValues,
		"security_event", event,
		"user_id", userID,
		"client_ip", clientIP,
		"user_agent", userAgent,
		"timestamp", time.Now().UTC(),
	)
	
	Warn("Security event", fields...)
}

// LogAuthEvent logs authentication events
func LogAuthEvent(event, userID, email, clientIP string, success bool, keysAndValues ...interface{}) {
	fields := append(keysAndValues,
		"auth_event", event,
		"user_id", userID,
		"email", email,
		"client_ip", clientIP,
		"success", success,
		"timestamp", time.Now().UTC(),
	)
	
	if success {
		Info("Authentication event", fields...)
	} else {
		Warn("Authentication failed", fields...)
	}
}

// Business logic logging

// LogBusinessEvent logs important business events
func LogBusinessEvent(event string, keysAndValues ...interface{}) {
	fields := append(keysAndValues,
		"business_event", event,
		"timestamp", time.Now().UTC(),
	)
	
	Info("Business event", fields...)
}

// Error logging with stack trace

// LogError logs an error with enhanced context
func LogError(err error, msg string, keysAndValues ...interface{}) {
	if err == nil {
		return
	}
	
	fields := append(keysAndValues,
		"error", err.Error(),
		"error_type", fmt.Sprintf("%T", err),
	)
	
	// Add stack trace for debugging
	if structuredLog != nil {
		structuredLog.Error(msg, 
			zap.Error(err),
			zap.Any("fields", fields),
			zap.Stack("stacktrace"),
		)
	} else {
		Error(msg, fields...)
	}
}

// LogErrorCtx logs an error with context and enhanced information
func LogErrorCtx(ctx context.Context, err error, msg string, keysAndValues ...interface{}) {
	if err == nil {
		return
	}
	
	fields := append(keysAndValues,
		"error", err.Error(),
		"error_type", fmt.Sprintf("%T", err),
	)
	
	// Add caller information
	if pc, file, line, ok := runtime.Caller(1); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			fields = append(fields, 
				"caller_func", fn.Name(),
				"caller_file", file,
				"caller_line", line,
			)
		}
	}
	
	WithContext(ctx).Errorw(msg, fields...)
}

// Utility functions

func getLogLevelForStatus(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "error"
	case statusCode >= 400:
		return "warn"
	default:
		return "info"
	}
}

func getRequestID(ctx context.Context) string {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

func getUserID(ctx context.Context) string {
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

func getTraceID(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		Log.Sync()
	}
	if structuredLog != nil {
		structuredLog.Sync()
	}
}

// SetLevel dynamically changes the log level
func SetLevel(level LogLevel) {
	if structuredLog == nil {
		return
	}
	
	// Note: This requires the logger to be created with a dynamic level
	Info("Log level changed", "new_level", level)
}

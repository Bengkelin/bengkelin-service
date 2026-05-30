package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Custom error types
var (
	// Authentication errors
	ErrInvalidCredentials = NewAppError("INVALID_CREDENTIALS", "Invalid email or password", http.StatusUnauthorized)
	ErrEmailAlreadyExists = NewAppError("EMAIL_EXISTS", "Email already registered", http.StatusConflict)
	ErrPasswordMismatch   = NewAppError("PASSWORD_MISMATCH", "Passwords do not match", http.StatusBadRequest)
	ErrTokenExpired       = NewAppError("TOKEN_EXPIRED", "Token has expired", http.StatusUnauthorized)
	ErrTokenInvalid       = NewAppError("TOKEN_INVALID", "Invalid token", http.StatusUnauthorized)
	ErrUnauthorized       = NewAppError("UNAUTHORIZED", "Unauthorized access", http.StatusUnauthorized)

	// Resource errors
	ErrUserNotFound    = NewAppError("USER_NOT_FOUND", "User not found", http.StatusNotFound)
	ErrMitraNotFound   = NewAppError("MITRA_NOT_FOUND", "Mitra not found", http.StatusNotFound)
	ErrBengkelNotFound = NewAppError("BENGKEL_NOT_FOUND", "Bengkel not found", http.StatusNotFound)
	ErrOrderNotFound   = NewAppError("ORDER_NOT_FOUND", "Order not found", http.StatusNotFound)
	ErrAddressNotFound = NewAppError("ADDRESS_NOT_FOUND", "Address not found", http.StatusNotFound)
	ErrVehicleNotFound = NewAppError("VEHICLE_NOT_FOUND", "Vehicle not found", http.StatusNotFound)

	// Permission errors
	ErrForbidden      = NewAppError("FORBIDDEN", "Access forbidden", http.StatusForbidden)
	ErrNotOwner       = NewAppError("NOT_OWNER", "You don't own this resource", http.StatusForbidden)
	ErrInvalidStatus  = NewAppError("INVALID_STATUS", "Invalid status transition", http.StatusBadRequest)

	// Validation errors
	ErrValidationFailed = NewAppError("VALIDATION_FAILED", "Validation failed", http.StatusBadRequest)
	ErrInvalidInput     = NewAppError("INVALID_INPUT", "Invalid input data", http.StatusBadRequest)
	ErrMissingField     = NewAppError("MISSING_FIELD", "Required field is missing", http.StatusBadRequest)

	// Business logic errors
	ErrBengkelNotOpen     = NewAppError("BENGKEL_NOT_OPEN", "Bengkel is currently closed", http.StatusBadRequest)
	ErrServiceNotAvailable = NewAppError("SERVICE_NOT_AVAILABLE", "Service is not available", http.StatusBadRequest)
	ErrInvalidLocation    = NewAppError("INVALID_LOCATION", "Invalid location coordinates", http.StatusBadRequest)

	// System errors
	ErrInternalServer = NewAppError("INTERNAL_SERVER_ERROR", "Internal server error", http.StatusInternalServerError)
	ErrDatabaseError  = NewAppError("DATABASE_ERROR", "Database operation failed", http.StatusInternalServerError)
	ErrExternalAPI    = NewAppError("EXTERNAL_API_ERROR", "External API error", http.StatusBadGateway)
)

// AppError represents an application-specific error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	Details    string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError creates a new application error
func NewAppError(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to an existing error
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		StatusCode: e.StatusCode,
		Details:    details,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// WrapError wraps a generic error into an AppError
func WrapError(err error, appErr *AppError) *AppError {
	if err == nil {
		return nil
	}
	
	return &AppError{
		Code:       appErr.Code,
		Message:    appErr.Message,
		StatusCode: appErr.StatusCode,
		Details:    err.Error(),
	}
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", ve[0].Message)
}

// ToAppError converts validation errors to AppError
func (ve ValidationErrors) ToAppError() *AppError {
	return &AppError{
		Code:       "VALIDATION_FAILED",
		Message:    "Validation failed",
		StatusCode: http.StatusBadRequest,
		Details:    ve.Error(),
	}
}
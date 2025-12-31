package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var (
	// Custom validator instance
	customValidator *validator.Validate
	
	// Common regex patterns
	phoneRegex     = regexp.MustCompile(`^(\+62|62|0)[0-9]{9,13}$`)
	alphaNumeric   = regexp.MustCompile(`^[a-zA-Z0-9\s]+$`)
	noScript       = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	sqlInjection   = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`)
)

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// Initialize custom validator
func init() {
	customValidator = validator.New()
	
	// Register custom validation tags
	registerCustomValidations()
	
	// Register custom tag name function
	customValidator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	// Replace gin's validator with our custom one
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		registerCustomValidations(v)
	}
}

// registerCustomValidations registers all custom validation rules
func registerCustomValidations(v ...*validator.Validate) {
	var validate *validator.Validate
	if len(v) > 0 {
		validate = v[0]
	} else {
		validate = customValidator
	}
	
	// Phone number validation
	validate.RegisterValidation("phone", validatePhone)
	
	// Strong password validation
	validate.RegisterValidation("strong_password", validateStrongPassword)
	
	// No XSS validation
	validate.RegisterValidation("no_xss", validateNoXSS)
	
	// No SQL injection validation
	validate.RegisterValidation("no_sql_injection", validateNoSQLInjection)
	
	// Alpha numeric with spaces
	validate.RegisterValidation("alpha_numeric_space", validateAlphaNumericSpace)
	
	// Latitude validation
	validate.RegisterValidation("latitude", validateLatitude)
	
	// Longitude validation
	validate.RegisterValidation("longitude", validateLongitude)
	
	// Rating validation (1-5)
	validate.RegisterValidation("rating", validateRating)
	
	// Indonesian vehicle number validation
	validate.RegisterValidation("vehicle_number", validateVehicleNumber)
	
	// Bank account number validation
	validate.RegisterValidation("bank_account", validateBankAccount)
	
	// Time format validation (HH:MM)
	validate.RegisterValidation("time_format", validateTimeFormat)
	
	// Day name validation
	validate.RegisterValidation("day_name", validateDayName)
	
	// URL validation (more strict than default)
	validate.RegisterValidation("strict_url", validateStrictURL)
	
	// File extension validation
	validate.RegisterValidation("file_ext", validateFileExtension)
}

// Custom validation functions

func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	return phoneRegex.MatchString(phone)
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	if len(password) < 8 {
		return false
	}
	
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasNumber && hasSpecial
}

func validateNoXSS(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return !noScript.MatchString(value) && !strings.Contains(strings.ToLower(value), "javascript:")
}

func validateNoSQLInjection(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return !sqlInjection.MatchString(value)
}

func validateAlphaNumericSpace(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return alphaNumeric.MatchString(value)
}

func validateLatitude(fl validator.FieldLevel) bool {
	lat := fl.Field().Float()
	return lat >= -90 && lat <= 90
}

func validateLongitude(fl validator.FieldLevel) bool {
	lng := fl.Field().Float()
	return lng >= -180 && lng <= 180
}

func validateRating(fl validator.FieldLevel) bool {
	rating := fl.Field().Int()
	return rating >= 1 && rating <= 5
}

func validateVehicleNumber(fl validator.FieldLevel) bool {
	vehicleNumber := fl.Field().String()
	// Indonesian vehicle number format: B 1234 ABC or B1234ABC
	vehicleRegex := regexp.MustCompile(`^[A-Z]{1,2}\s?\d{1,4}\s?[A-Z]{1,3}$`)
	return vehicleRegex.MatchString(strings.ToUpper(vehicleNumber))
}

func validateBankAccount(fl validator.FieldLevel) bool {
	account := fl.Field().String()
	// Bank account should be 10-20 digits
	bankRegex := regexp.MustCompile(`^\d{10,20}$`)
	return bankRegex.MatchString(account)
}

func validateTimeFormat(fl validator.FieldLevel) bool {
	timeStr := fl.Field().String()
	timeRegex := regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)
	return timeRegex.MatchString(timeStr)
}

func validateDayName(fl validator.FieldLevel) bool {
	day := strings.ToLower(fl.Field().String())
	validDays := []string{"senin", "selasa", "rabu", "kamis", "jumat", "sabtu", "minggu"}
	for _, validDay := range validDays {
		if day == validDay {
			return true
		}
	}
	return false
}

func validateStrictURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/.*)?$`)
	return urlRegex.MatchString(url)
}

func validateFileExtension(fl validator.FieldLevel) bool {
	filename := fl.Field().String()
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"}
	
	for _, ext := range allowedExts {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return true
		}
	}
	return false
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) ValidationErrors {
	err := customValidator.Struct(s)
	if err == nil {
		return nil
	}
	
	var validationErrors ValidationErrors
	
	for _, err := range err.(validator.ValidationErrors) {
		validationError := ValidationError{
			Field:   err.Field(),
			Tag:     err.Tag(),
			Value:   fmt.Sprintf("%v", err.Value()),
			Message: getErrorMessage(err),
		}
		validationErrors = append(validationErrors, validationError)
	}
	
	return validationErrors
}

// getErrorMessage returns user-friendly error messages
func getErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, fe.Param())
	case "phone":
		return fmt.Sprintf("%s must be a valid Indonesian phone number", field)
	case "strong_password":
		return fmt.Sprintf("%s must contain at least 8 characters with uppercase, lowercase, number, and special character", field)
	case "no_xss":
		return fmt.Sprintf("%s contains potentially dangerous content", field)
	case "no_sql_injection":
		return fmt.Sprintf("%s contains potentially dangerous SQL content", field)
	case "alpha_numeric_space":
		return fmt.Sprintf("%s must contain only letters, numbers, and spaces", field)
	case "latitude":
		return fmt.Sprintf("%s must be a valid latitude (-90 to 90)", field)
	case "longitude":
		return fmt.Sprintf("%s must be a valid longitude (-180 to 180)", field)
	case "rating":
		return fmt.Sprintf("%s must be between 1 and 5", field)
	case "vehicle_number":
		return fmt.Sprintf("%s must be a valid Indonesian vehicle number", field)
	case "bank_account":
		return fmt.Sprintf("%s must be a valid bank account number (10-20 digits)", field)
	case "time_format":
		return fmt.Sprintf("%s must be in HH:MM format", field)
	case "day_name":
		return fmt.Sprintf("%s must be a valid day name in Indonesian", field)
	case "strict_url":
		return fmt.Sprintf("%s must be a valid HTTP/HTTPS URL", field)
	case "file_ext":
		return fmt.Sprintf("%s must have a valid file extension", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// SanitizeInput sanitizes input to prevent XSS and other attacks
func SanitizeInput(input string) string {
	// Remove script tags
	input = noScript.ReplaceAllString(input, "")
	
	// Remove javascript: protocol (case-insensitive)
	jsRegex := regexp.MustCompile(`(?i)javascript:`)
	input = jsRegex.ReplaceAllString(input, "")
	
	// Remove common XSS patterns (case-insensitive)
	xssPatterns := []string{
		`(?i)<script`,
		`(?i)</script>`,
		`(?i)onclick=`,
		`(?i)onload=`,
		`(?i)onerror=`,
		`(?i)vbscript:`,
		`(?i)data:text/html`,
	}
	
	for _, pattern := range xssPatterns {
		regex := regexp.MustCompile(pattern)
		input = regex.ReplaceAllString(input, "")
	}
	
	// Trim whitespace
	input = strings.TrimSpace(input)
	
	return input
}

// ValidateAndSanitize validates and sanitizes input
func ValidateAndSanitize(s interface{}) (ValidationErrors, error) {
	// First sanitize string fields
	sanitizeStructFields(s)
	
	// Then validate
	return ValidateStruct(s), nil
}

// sanitizeStructFields sanitizes all string fields in a struct
func sanitizeStructFields(s interface{}) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return
	}
	
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String && field.CanSet() {
			sanitized := SanitizeInput(field.String())
			field.SetString(sanitized)
		}
	}
}
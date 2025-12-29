package tests

import (
	"testing"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/validator"
	"github.com/Bengkelin/bengkelin-service/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestLoginRequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		request     validator.LoginRequest
		expectError bool
		errorCount  int
	}{
		{
			name: "Valid login request",
			request: validator.LoginRequest{
				Email:    "test@example.com",
				Password: "ValidPass123!",
			},
			expectError: false,
		},
		{
			name: "Invalid email format",
			request: validator.LoginRequest{
				Email:    "invalid-email",
				Password: "ValidPass123!",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "Password too short",
			request: validator.LoginRequest{
				Email:    "test@example.com",
				Password: "short",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "XSS attempt in email",
			request: validator.LoginRequest{
				Email:    "test@example.com<script>alert('xss')</script>",
				Password: "ValidPass123!",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "SQL injection attempt in password",
			request: validator.LoginRequest{
				Email:    "test@example.com",
				Password: "password' OR '1'='1",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "Empty fields",
			request: validator.LoginRequest{
				Email:    "",
				Password: "",
			},
			expectError: true,
			errorCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validation.ValidateStruct(tt.request)
			
			if tt.expectError {
				assert.NotNil(t, errors)
				if tt.errorCount > 0 {
					assert.Len(t, errors, tt.errorCount)
				}
			} else {
				assert.Nil(t, errors)
			}
		})
	}
}

func TestRegisterRequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		request     validator.RegisterNewUserRequest
		expectError bool
	}{
		{
			name: "Valid register request",
			request: validator.RegisterNewUserRequest{
				FirstName:       "John",
				LastName:        "Doe",
				Email:           "john.doe@example.com",
				PhoneNumber:     "081234567890",
				Password:        "StrongPass123!",
				ConfirmPassword: "StrongPass123!",
			},
			expectError: false,
		},
		{
			name: "Invalid phone number",
			request: validator.RegisterNewUserRequest{
				FirstName:       "John",
				LastName:        "Doe",
				Email:           "john.doe@example.com",
				PhoneNumber:     "123", // Invalid Indonesian phone
				Password:        "StrongPass123!",
				ConfirmPassword: "StrongPass123!",
			},
			expectError: true,
		},
		{
			name: "Weak password",
			request: validator.RegisterNewUserRequest{
				FirstName:       "John",
				LastName:        "Doe",
				Email:           "john.doe@example.com",
				PhoneNumber:     "081234567890",
				Password:        "weakpass", // No uppercase, number, or special char
				ConfirmPassword: "weakpass",
			},
			expectError: true,
		},
		{
			name: "Name with special characters",
			request: validator.RegisterNewUserRequest{
				FirstName:       "John<script>",
				LastName:        "Doe",
				Email:           "john.doe@example.com",
				PhoneNumber:     "081234567890",
				Password:        "StrongPass123!",
				ConfirmPassword: "StrongPass123!",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validation.ValidateStruct(tt.request)
			
			if tt.expectError {
				assert.NotNil(t, errors)
			} else {
				assert.Nil(t, errors)
			}
		})
	}
}

func TestBengkelValidation(t *testing.T) {
	tests := []struct {
		name        string
		request     validator.BengkelRegisterRequest
		expectError bool
	}{
		{
			name: "Valid bengkel request",
			request: validator.BengkelRegisterRequest{
				BengkelName:  "Honda Service Center",
				BengkelPhone: "081234567890",
				JumlahMontir: 5,
				Hari:         []string{"senin", "selasa", "rabu"},
				JamBuka:      []string{"08:00", "09:00", "10:00"},
			},
			expectError: false,
		},
		{
			name: "Invalid day name",
			request: validator.BengkelRegisterRequest{
				BengkelName:  "Honda Service Center",
				BengkelPhone: "081234567890",
				JumlahMontir: 5,
				Hari:         []string{"monday", "tuesday"}, // English day names
				JamBuka:      []string{"08:00", "09:00"},
			},
			expectError: true,
		},
		{
			name: "Invalid time format",
			request: validator.BengkelRegisterRequest{
				BengkelName:  "Honda Service Center",
				BengkelPhone: "081234567890",
				JumlahMontir: 5,
				Hari:         []string{"senin", "selasa"},
				JamBuka:      []string{"8:00", "25:00"}, // Invalid time format
			},
			expectError: true,
		},
		{
			name: "Too many mechanics",
			request: validator.BengkelRegisterRequest{
				BengkelName:  "Honda Service Center",
				BengkelPhone: "081234567890",
				JumlahMontir: 100, // Exceeds max limit
				Hari:         []string{"senin"},
				JamBuka:      []string{"08:00"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validation.ValidateStruct(tt.request)
			
			if tt.expectError {
				assert.NotNil(t, errors)
			} else {
				assert.Nil(t, errors)
			}
		})
	}
}

func TestAddressValidation(t *testing.T) {
	tests := []struct {
		name        string
		request     validator.AddressUserRequest
		expectError bool
	}{
		{
			name: "Valid address request",
			request: validator.AddressUserRequest{
				Latitude:     -6.2088,
				Longitude:    106.8456,
				AddressLabel: "Home",
				FullAddress:  "Jl. Sudirman No. 1, Jakarta",
				Note:         "Near the mall",
			},
			expectError: false,
		},
		{
			name: "Invalid latitude",
			request: validator.AddressUserRequest{
				Latitude:     -95.0, // Invalid latitude
				Longitude:    106.8456,
				AddressLabel: "Home",
				FullAddress:  "Jl. Sudirman No. 1, Jakarta",
			},
			expectError: true,
		},
		{
			name: "Invalid longitude",
			request: validator.AddressUserRequest{
				Latitude:     -6.2088,
				Longitude:    200.0, // Invalid longitude
				AddressLabel: "Home",
				FullAddress:  "Jl. Sudirman No. 1, Jakarta",
			},
			expectError: true,
		},
		{
			name: "Address too short",
			request: validator.AddressUserRequest{
				Latitude:     -6.2088,
				Longitude:    106.8456,
				AddressLabel: "Home",
				FullAddress:  "Short", // Too short
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validation.ValidateStruct(tt.request)
			
			if tt.expectError {
				assert.NotNil(t, errors)
			} else {
				assert.Nil(t, errors)
			}
		})
	}
}

func TestVehicleValidation(t *testing.T) {
	tests := []struct {
		name        string
		request     validator.VehicleUserRequest
		expectError bool
	}{
		{
			name: "Valid vehicle request",
			request: validator.VehicleUserRequest{
				VehicleType:   "Motorcycle",
				VehicleColor:  "Red",
				VehicleNumber: "B1234ABC",
			},
			expectError: false,
		},
		{
			name: "Invalid vehicle number format",
			request: validator.VehicleUserRequest{
				VehicleType:   "Motorcycle",
				VehicleColor:  "Red",
				VehicleNumber: "INVALID123", // Invalid format
			},
			expectError: true,
		},
		{
			name: "Vehicle type with special characters",
			request: validator.VehicleUserRequest{
				VehicleType:   "Motor<script>",
				VehicleColor:  "Red",
				VehicleNumber: "B1234ABC",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validation.ValidateStruct(tt.request)
			
			if tt.expectError {
				assert.NotNil(t, errors)
			} else {
				assert.Nil(t, errors)
			}
		})
	}
}

func TestBankValidation(t *testing.T) {
	tests := []struct {
		name        string
		request     validator.BankMitraRequest
		expectError bool
	}{
		{
			name: "Valid bank request",
			request: validator.BankMitraRequest{
				BankName:   "Bank BCA",
				BankNumber: "1234567890123456",
			},
			expectError: false,
		},
		{
			name: "Invalid bank account number",
			request: validator.BankMitraRequest{
				BankName:   "Bank BCA",
				BankNumber: "123", // Too short
			},
			expectError: true,
		},
		{
			name: "Bank account with letters",
			request: validator.BankMitraRequest{
				BankName:   "Bank BCA",
				BankNumber: "123ABC456789", // Contains letters
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validation.ValidateStruct(tt.request)
			
			if tt.expectError {
				assert.NotNil(t, errors)
			} else {
				assert.Nil(t, errors)
			}
		})
	}
}

func TestTestimoniValidation(t *testing.T) {
	tests := []struct {
		name        string
		request     validator.BengkelTestimoniRequest
		expectError bool
	}{
		{
			name: "Valid testimoni request",
			request: validator.BengkelTestimoniRequest{
				Testimoni: "Great service, very professional and fast!",
				Rating:    5,
			},
			expectError: false,
		},
		{
			name: "Rating out of range",
			request: validator.BengkelTestimoniRequest{
				Testimoni: "Great service, very professional and fast!",
				Rating:    6, // Invalid rating
			},
			expectError: true,
		},
		{
			name: "Testimoni too short",
			request: validator.BengkelTestimoniRequest{
				Testimoni: "Good", // Too short
				Rating:    5,
			},
			expectError: true,
		},
		{
			name: "XSS attempt in testimoni",
			request: validator.BengkelTestimoniRequest{
				Testimoni: "Great service <script>alert('xss')</script>",
				Rating:    5,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validation.ValidateStruct(tt.request)
			
			if tt.expectError {
				assert.NotNil(t, errors)
			} else {
				assert.Nil(t, errors)
			}
		})
	}
}

func TestSanitization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove script tags",
			input:    "Hello <script>alert('xss')</script> World",
			expected: "Hello  World",
		},
		{
			name:     "Remove javascript protocol",
			input:    "javascript:alert('xss')",
			expected: "alert('xss')",
		},
		{
			name:     "Clean input",
			input:    "Normal text input",
			expected: "Normal text input",
		},
		{
			name:     "Trim whitespace",
			input:    "  spaced text  ",
			expected: "spaced text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomValidators(t *testing.T) {
	t.Run("Phone validation", func(t *testing.T) {
		validPhones := []string{
			"081234567890",
			"6281234567890",
			"+6281234567890",
			"08123456789",
		}
		
		invalidPhones := []string{
			"123456789",
			"08123",
			"abc123456789",
			"081234567890123456", // Too long
		}
		
		for _, phone := range validPhones {
			request := validator.RegisterNewUserRequest{
				FirstName:       "John",
				LastName:        "Doe",
				Email:           "test@example.com",
				PhoneNumber:     phone,
				Password:        "StrongPass123!",
				ConfirmPassword: "StrongPass123!",
			}
			errors := validation.ValidateStruct(request)
			assert.Nil(t, errors, "Phone %s should be valid", phone)
		}
		
		for _, phone := range invalidPhones {
			request := validator.RegisterNewUserRequest{
				FirstName:       "John",
				LastName:        "Doe",
				Email:           "test@example.com",
				PhoneNumber:     phone,
				Password:        "StrongPass123!",
				ConfirmPassword: "StrongPass123!",
			}
			errors := validation.ValidateStruct(request)
			assert.NotNil(t, errors, "Phone %s should be invalid", phone)
		}
	})
	
	t.Run("Strong password validation", func(t *testing.T) {
		validPasswords := []string{
			"StrongPass123!",
			"MyP@ssw0rd",
			"Secure123$",
		}
		
		invalidPasswords := []string{
			"password",      // No uppercase, number, special
			"PASSWORD",      // No lowercase, number, special
			"Password",      // No number, special
			"Password123",   // No special
			"Pass123!",      // Too short
		}
		
		for _, password := range validPasswords {
			request := validator.RegisterNewUserRequest{
				FirstName:       "John",
				LastName:        "Doe",
				Email:           "test@example.com",
				PhoneNumber:     "081234567890",
				Password:        password,
				ConfirmPassword: password,
			}
			errors := validation.ValidateStruct(request)
			assert.Nil(t, errors, "Password %s should be valid", password)
		}
		
		for _, password := range invalidPasswords {
			request := validator.RegisterNewUserRequest{
				FirstName:       "John",
				LastName:        "Doe",
				Email:           "test@example.com",
				PhoneNumber:     "081234567890",
				Password:        password,
				ConfirmPassword: password,
			}
			errors := validation.ValidateStruct(request)
			assert.NotNil(t, errors, "Password %s should be invalid", password)
		}
	})
}
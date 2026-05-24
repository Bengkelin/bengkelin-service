package dto

import (
	"time"
)

// UpdateUserRequest for updating user profile
type UpdateUserRequest struct {
	FirstName   string `json:"first_name" validate:"min=1,max=50,alpha_numeric_space,no_xss"`
	LastName    string `json:"last_name" validate:"min=1,max=50,alpha_numeric_space,no_xss"`
	PhoneNumber string `json:"phone_number" validate:"phone"`
}

// UpdateMitraRequest for updating mitra profile
type UpdateMitraRequest struct {
	FirstName   string `json:"first_name" validate:"min=1,max=50,alpha_numeric_space,no_xss"`
	LastName    string `json:"last_name" validate:"min=1,max=50,alpha_numeric_space,no_xss"`
	PhoneNumber string `json:"phone_number" validate:"phone"`
}

// MitraBankRequest for bank operations
type MitraBankRequest struct {
	BankName   string `json:"bank_name" validate:"required,min=1,max=100,no_xss"`
	BankNumber string `json:"bank_number" validate:"required,min=5,max=50,no_xss"`
}

// MitraBankResponse for bank data
type MitraBankResponse struct {
	BankName   string `json:"bank_name"`
	BankNumber string `json:"bank_number"`
}

// MitraProfileResponse for mitra profile data
type MitraProfileResponse struct {
	ID          string            `json:"mitra_id"`
	FirstName   string            `json:"first_name"`
	LastName    string            `json:"last_name"`
	Email       string            `json:"email"`
	PhoneNumber string            `json:"phone_number"`
	BankName    string            `json:"bank_name,omitempty"`
	BankNumber  string            `json:"bank_number,omitempty"`
	Bengkel     []BengkelResponse `json:"bengkel,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// UserProfileResponse for user profile data
type UserProfileResponse struct {
	ID          string            `json:"id"`
	FirstName   string            `json:"first_name"`
	LastName    string            `json:"last_name"`
	Email       string            `json:"email"`
	PhoneNumber string            `json:"phone_number"`
	AvatarURL   string            `json:"avatar_url,omitempty"`
	Addresses   []AddressResponse `json:"addresses,omitempty"`
	Vehicles    []VehicleResponse `json:"vehicles,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

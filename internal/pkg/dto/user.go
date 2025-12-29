package dto

import "time"

// UpdateUserRequest for updating user profile
type UpdateUserRequest struct {
	FirstName   string  `json:"first_name" validate:"min=1,max=50,alpha_numeric_space,no_xss"`
	LastName    string  `json:"last_name" validate:"min=1,max=50,alpha_numeric_space,no_xss"`
	PhoneNumber string  `json:"phone_number" validate:"phone"`
	Latitude    float64 `json:"latitude" validate:"latitude"`
	Longitude   float64 `json:"longitude" validate:"longitude"`
	FullAddress string  `json:"full_address" validate:"max=500,no_xss"`
	Note        string  `json:"note" validate:"max=200,no_xss"`
}

// UserProfileResponse for user profile data
type UserProfileResponse struct {
	ID          string              `json:"id"`
	FirstName   string              `json:"first_name"`
	LastName    string              `json:"last_name"`
	Email       string              `json:"email"`
	PhoneNumber string              `json:"phone_number"`
	AvatarURL   string              `json:"avatar_url,omitempty"`
	Addresses   []AddressResponse   `json:"addresses,omitempty"`
	Vehicles    []VehicleResponse   `json:"vehicles,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}
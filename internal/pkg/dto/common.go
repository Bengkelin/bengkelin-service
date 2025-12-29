package dto

import "time"

// Common DTOs used across services

// PaginationRequest for paginated requests
type PaginationRequest struct {
	Page  int `json:"page" validate:"min=1" form:"page"`
	Limit int `json:"limit" validate:"min=1,max=100" form:"limit"`
}

// PaginationResponse for paginated responses
type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// AddressRequest for address-related operations
type AddressRequest struct {
	Latitude     float64 `json:"latitude" validate:"required,latitude"`
	Longitude    float64 `json:"longitude" validate:"required,longitude"`
	AddressLabel string  `json:"address_label" validate:"required,min=1,max=100,no_xss"`
	FullAddress  string  `json:"full_address" validate:"required,min=10,max=500,no_xss"`
	Note         string  `json:"note" validate:"max=200,no_xss"`
}

// AddressResponse for address data
type AddressResponse struct {
	ID           string    `json:"id"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	AddressLabel string    `json:"address_label"`
	FullAddress  string    `json:"full_address"`
	Note         string    `json:"note,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// VehicleRequest for vehicle-related operations
type VehicleRequest struct {
	VehicleType   string `json:"vehicle_type" validate:"required,min=1,max=50,alpha_numeric_space,no_xss"`
	VehicleColor  string `json:"vehicle_color" validate:"required,min=1,max=30,alpha_numeric_space,no_xss"`
	VehicleNumber string `json:"vehicle_number" validate:"required,vehicle_number"`
}

// VehicleResponse for vehicle data
type VehicleResponse struct {
	ID            string    `json:"id"`
	VehicleType   string    `json:"vehicle_type"`
	VehicleColor  string    `json:"vehicle_color"`
	VehicleNumber string    `json:"vehicle_number"`
	Photos        []string  `json:"photos,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ErrorResponse for error responses
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse for success responses
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// FileUploadResponse for file upload operations
type FileUploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// LocationRequest for location-based operations
type LocationRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
	Radius    float64 `json:"radius" validate:"min=0.1,max=50"` // in kilometers
}

// TimeRange for time-based filtering
type TimeRange struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// SortRequest for sorting operations
type SortRequest struct {
	Field string `json:"field" validate:"alpha_numeric"`
	Order string `json:"order" validate:"oneof=asc desc"`
}

// FilterRequest for filtering operations
type FilterRequest struct {
	Field    string      `json:"field" validate:"alpha_numeric"`
	Operator string      `json:"operator" validate:"oneof=eq ne gt lt gte lte like in"`
	Value    interface{} `json:"value"`
}
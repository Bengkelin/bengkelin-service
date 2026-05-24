package dto

import "time"

// CreateBengkelRequest for creating a new bengkel
type CreateBengkelRequest struct {
	BengkelName  string   `json:"bengkel_name" validate:"required,min=2,max=100,no_xss"`
	BengkelPhone string   `json:"bengkel_phone" validate:"required,phone"`
	JumlahMontir uint     `json:"jumlah_montir" validate:"required,min=1,max=50"`
	Hari         []string `json:"hari" validate:"required,min=1,max=7,dive,day_name"`
	JamBuka      []string `json:"jam_buka" validate:"required,min=1,max=7,dive,time_format"`
}

// UpdateBengkelRequest for updating bengkel information
type UpdateBengkelRequest struct {
	BengkelName  string `json:"bengkel_name" validate:"min=2,max=100,no_xss"`
	BengkelPhone string `json:"bengkel_phone" validate:"phone"`
	JumlahMontir uint   `json:"jumlah_montir" validate:"min=1,max=50"`
}

// OperationalRequest for updating operational hours
type OperationalRequest struct {
	Hari    []string `json:"hari" validate:"required,min=1,max=7,dive,day_name"`
	JamBuka []string `json:"jam_buka" validate:"required,min=1,max=7,dive,time_format"`
}

// OperationalItemRequest for individual operational hours
type OperationalItemRequest struct {
	ID       uint   `json:"id,omitempty"`
	Hari     string `json:"hari" validate:"required,day_name"`
	JamBuka  string `json:"jam_buka" validate:"required,time_format"`
	JamTutup string `json:"jam_tutup,omitempty" validate:"omitempty,time_format"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// BengkelAddressRequest for creating bengkel address
type BengkelAddressRequest struct {
	Latitude     float64 `json:"latitude" validate:"required,latitude"`
	Longitude    float64 `json:"longitude" validate:"required,longitude"`
	AddressLabel string  `json:"address_label,omitempty" validate:"omitempty,max=100,no_xss"`
	FullAddress  string  `json:"full_address" validate:"required,min=5,max=500,no_xss"`
	Note         string  `json:"note,omitempty" validate:"omitempty,max=200,no_xss"`
}

// BengkelServiceItemRequest for individual service items
type BengkelServiceItemRequest struct {
	ID          uint    `json:"id,omitempty"`
	NamaService string  `json:"nama_service" validate:"required,min=2,max=100,no_xss"`
	Description string  `json:"description,omitempty" validate:"omitempty,max=500,no_xss"`
	Price       float64 `json:"price,omitempty" validate:"omitempty,min=0"`
	IsAvailable *bool   `json:"is_available,omitempty"`
}

// ServiceOptionRequest for updating service options
type ServiceOptionRequest struct {
	HomeService  bool `json:"home_service"`
	StoreService bool `json:"store_service"`
	IsOpen       bool `json:"is_open"`
}

// TestimonialRequest for adding testimonials
type TestimonialRequest struct {
	Testimoni string `json:"testimoni" validate:"required,min=10,max=1000,no_xss"`
	Rating    int    `json:"rating" validate:"required,rating"`
}

// SearchBengkelRequest for searching bengkels
type SearchBengkelRequest struct {
	PaginationRequest
	Query     string   `json:"query" validate:"max=100,no_xss" form:"query"`
	Latitude  float64  `json:"latitude" validate:"latitude" form:"latitude"`
	Longitude float64  `json:"longitude" validate:"longitude" form:"longitude"`
	Radius    float64  `json:"radius" validate:"min=0.1,max=50" form:"radius"` // in km
	Services  []string `json:"services" validate:"dive,alpha_numeric_space" form:"services"`
	IsOpen    *bool    `json:"is_open" form:"is_open"`
}

// NearestBengkelRequest for finding nearest bengkels
type NearestBengkelRequest struct {
	PaginationRequest
	Latitude  float64 `json:"latitude" validate:"required,latitude" form:"latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude" form:"longitude"`
	Radius    float64 `json:"radius" validate:"min=0.1,max=50" form:"radius"` // in km
}

// BengkelResponse for bengkel data
type BengkelResponse struct {
	ID           string                `json:"id"`
	MitraID      string                `json:"mitra_id"`
	BengkelName  string                `json:"bengkel_name"`
	BengkelPhone string                `json:"bengkel_phone"`
	JumlahMontir uint                  `json:"jumlah_montir"`
	AvatarURL    string                `json:"avatar_url,omitempty"`
	HomeService  bool                  `json:"home_service"`
	StoreService bool                  `json:"store_service"`
	IsOpen       bool                  `json:"is_open"`
	Rating       float64               `json:"rating"`
	TotalReviews int                   `json:"total_reviews"`
	Address      *AddressResponse      `json:"address,omitempty"`
	Services     []string              `json:"services,omitempty"`
	Photos       []string              `json:"photos,omitempty"`
	Operational  []OperationalResponse `json:"operational,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

// BengkelDetailResponse for detailed bengkel information
type BengkelDetailResponse struct {
	BengkelResponse
	Testimonials []TestimonialResponse `json:"testimonials,omitempty"`
	Distance     float64               `json:"distance,omitempty"` // in km
}

// PaginatedBengkelResponse for paginated bengkel list
type PaginatedBengkelResponse struct {
	PaginationResponse
	Data []BengkelResponse `json:"data"`
}

// OperationalResponse for operational hours
type OperationalResponse struct {
	ID        string `json:"id"`
	BengkelID string `json:"bengkel_id"`
	Hari      string `json:"hari"`
	JamBuka   string `json:"jam_buka"`
}

// TestimonialResponse for testimonial data
type TestimonialResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	BengkelID string    `json:"bengkel_id"`
	UserName  string    `json:"user_name"`
	Testimoni string    `json:"testimoni"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
}

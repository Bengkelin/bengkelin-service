package validator

type UserUpdateRequest struct {
	FirstName   string `json:"first_name" binding:"min=1,max=50" validate:"alpha_numeric_space,no_xss"`
	LastName    string `json:"last_name" binding:"min=1,max=50" validate:"alpha_numeric_space,no_xss"`
	PhoneNumber string `json:"phone_number" validate:"phone"`
}

type UserAddressCreateRequest struct {
	Latitude     float64 `json:"latitude" binding:"required" validate:"latitude"`
	Longitude    float64 `json:"longitude" binding:"required" validate:"longitude"`
	AddressLabel string  `json:"address_label" binding:"required,min=1,max=100" validate:"no_xss,no_sql_injection"`
	FullAddress  string  `json:"full_address" binding:"required,min=1,max=500" validate:"no_xss,no_sql_injection"`
	Note         string  `json:"note,omitempty" binding:"max=200" validate:"no_xss,no_sql_injection"`
	IsPrimary    *bool   `json:"is_primary,omitempty"`
}

type UserAddressUpdateRequest struct {
	Latitude     float64 `json:"latitude,omitempty" validate:"latitude"`
	Longitude    float64 `json:"longitude,omitempty" validate:"longitude"`
	AddressLabel string  `json:"address_label,omitempty" binding:"max=100" validate:"no_xss,no_sql_injection"`
	FullAddress  string  `json:"full_address,omitempty" binding:"max=500" validate:"no_xss,no_sql_injection"`
	Note         string  `json:"note,omitempty" binding:"max=200" validate:"no_xss,no_sql_injection"`
	IsPrimary    *bool   `json:"is_primary,omitempty"`
}

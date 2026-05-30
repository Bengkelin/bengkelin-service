package validator

// Struct that define the validator/binding of Login Request
type LoginRequest struct {
	Email    string `json:"email" form:"email" binding:"required,email" validate:"no_xss,no_sql_injection"`
	Password string `json:"password" form:"password" binding:"required,min=8,max=128" validate:"no_xss"`
}

type GoogleAuthRequest struct {
	Email     string `json:"email" binding:"required,email" validate:"no_xss"`
	FirstName string `json:"first_name" binding:"required,min=1,max=50" validate:"alpha_numeric_space,no_xss"`
}

// Struct that define the validator/binding of Register Request
type RegisterRequest struct {
	Name     string `json:"name" form:"name" binding:"required,min=1,max=100" validate:"alpha_numeric_space,no_xss"`
	Email    string `json:"email" form:"email" binding:"required,email" validate:"no_xss"`
	Password string `json:"password" form:"password" binding:"required,min=8,max=128" validate:"strong_password,no_xss"`
}

type RegisterNewUserRequest struct {
	FirstName       string `json:"first_name" binding:"required,min=1,max=50" validate:"alpha_numeric_space,no_xss"`
	LastName        string `json:"last_name" binding:"required,min=1,max=50" validate:"alpha_numeric_space,no_xss"`
	Email           string `json:"email" binding:"required,email" validate:"no_xss"`
	PhoneNumber     string `json:"phone_number" binding:"required" validate:"phone"`
	Password        string `json:"password" binding:"required,min=8,max=128" validate:"strong_password,no_xss"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=8,max=128" validate:"no_xss"`
}

type RegisterNewMitraRequest struct {
	FirstName       string `json:"first_name" binding:"required,min=1,max=50" validate:"alpha_numeric_space,no_xss"`
	LastName        string `json:"last_name" binding:"required,min=1,max=50" validate:"alpha_numeric_space,no_xss"`
	Email           string `json:"email" binding:"required,email" validate:"no_xss"`
	PhoneNumber     string `json:"phone_number" binding:"required" validate:"phone"`
	Password        string `json:"password" binding:"required,min=8,max=128" validate:"strong_password,no_xss"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=8,max=128" validate:"no_xss"`
}

type BankMitraRequest struct {
	BankName   string `json:"bank_name" binding:"required,min=2,max=50" validate:"alpha_numeric_space,no_xss"`
	BankNumber string `json:"bank_number" binding:"required" validate:"bank_account"`
}

type AddressUserRequest struct {
	Latitude     float64 `json:"latitude" binding:"required" validate:"latitude"`
	Longitude    float64 `json:"longitude" binding:"required" validate:"longitude"`
	AddressLabel string  `json:"address_label" binding:"required,min=1,max=100" validate:"no_xss,no_sql_injection"`
	FullAddress  string  `json:"full_address" binding:"required,min=10,max=500" validate:"no_xss,no_sql_injection"`
	Note         string  `json:"note" binding:"max=200" validate:"no_xss,no_sql_injection"`
}

type VehicleUserRequest struct {
	VehicleType   string `json:"vehicle_type" form:"vehicle_type" binding:"required,min=1,max=50" validate:"alpha_numeric_space,no_xss"`
	VehicleColor  string `json:"vehicle_color" form:"vehicle_color" binding:"required,min=1,max=30" validate:"alpha_numeric_space,no_xss"`
	VehicleNumber string `json:"vehicle_number" form:"vehicle_number" binding:"required" validate:"vehicle_number"`
}

// RefreshTokenRequest for token refresh validation
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,min=10" validate:"no_xss,no_sql_injection"`
}

// LogoutRequest for logout validation
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,min=10" validate:"no_xss,no_sql_injection"`
}

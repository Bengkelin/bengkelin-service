package validator

// Struct that define the validator/binding of Login Request
type LoginRequest struct {
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required"`
}

// Struct that define the validator/binding of Register Request
type RegisterRequest struct {
	Name     string `json:"name" form:"name" binding:"required,min=1"`
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required,min=8"`
}

type RegisterNewUserRequest struct {
	FirstName       string `json:"first_name" binding:"required"`
	LastName        string `json:"last_name" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	PhoneNumber     string `json:"phone_number" validation:"phone_number"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

type RegisterNewMitraRequest struct {
	FirstName       string `json:"first_name" binding:"required"`
	LastName        string `json:"last_name" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	PhoneNumber     string `json:"phone_number" validation:"phone_number"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

type BankMitraRequest struct {
	BankName   string `json:"bank_name" binding:"required"`
	BankNumber string `json:"bank_number" binding:"required"`
}

type AddressUserRequest struct {
	Latitude     float64 `json:"latitude" binding:"required"`
	Longitude    float64 `json:"longitude" binding:"required"`
	AddressLabel string  `json:"address_label" binding:"required"`
	FullAddress  string  `json:"full_address" binding:"required"`
	Note         string  `json:"note"`
}

type VehicleUserRequest struct {
	VehicleBrand  string `json:"vehicle_type" binding:"required"`
	VehicleColor  string `json:"vehicle_color" binding:"required"`
	VehicleNumber string `json:"vehicle_number" binding:"required"`
}

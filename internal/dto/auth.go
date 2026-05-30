package dto

// Request DTOs
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,no_xss"`
	Password string `json:"password" validate:"required,min=8,max=128,no_xss"`
}

type RegisterUserRequest struct {
	FirstName       string `json:"first_name" validate:"required,min=1,max=50,alpha_numeric_space,no_xss"`
	LastName        string `json:"last_name" validate:"required,min=1,max=50,alpha_numeric_space,no_xss"`
	Email           string `json:"email" validate:"required,email,no_xss"`
	PhoneNumber     string `json:"phone_number" validate:"required,phone"`
	Password        string `json:"password" validate:"required,min=8,max=128,strong_password,no_xss"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8,max=128,no_xss"`
}

type RegisterMitraRequest struct {
	FirstName       string `json:"first_name" validate:"required,min=1,max=50,alpha_numeric_space,no_xss"`
	LastName        string `json:"last_name" validate:"required,min=1,max=50,alpha_numeric_space,no_xss"`
	Email           string `json:"email" validate:"required,email,no_xss"`
	PhoneNumber     string `json:"phone_number" validate:"required,phone"`
	Password        string `json:"password" validate:"required,min=8,max=128,strong_password,no_xss"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8,max=128,no_xss"`
}

type GoogleAuthRequest struct {
	Email     string `json:"email" validate:"required,email,no_xss"`
	FirstName string `json:"first_name" validate:"required,min=1,max=50,alpha_numeric_space,no_xss"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=10,no_xss"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=10,no_xss"`
}

// Response DTOs
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	TokenType    string    `json:"token_type"`
	User         *UserInfo `json:"user,omitempty"`
	Mitra        *MitraInfo `json:"mitra,omitempty"`
}

type UserInfo struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type MitraInfo struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	BankName    string `json:"bank_name,omitempty"`
	BankNumber  string `json:"bank_number,omitempty"`
}
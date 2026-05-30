package validator

type MitraBankUpdateRequest struct {
	BankName   string `json:"bank_name" binding:"required,min=2,max=50" validate:"alpha_numeric_space,no_xss"`
	BankNumber string `json:"bank_number" binding:"required" validate:"bank_account"`
}

type MitraUpdateProfileRequest struct {
	FirstName   string `json:"first_name" binding:"omitempty,min=1,max=50" validate:"alpha_numeric_space,no_xss"`
	LastName    string `json:"last_name" binding:"omitempty,min=1,max=50" validate:"alpha_numeric_space,no_xss"`
	PhoneNumber string `json:"phone_number" binding:"omitempty" validate:"phone"`
}

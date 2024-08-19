package validator

type MitraBankUpdateRequest struct {
	BankName   string `binding:"required" json:"bank_name"`
	BankNumber string `binding:"required" json:"bank_number"`
}

type MitraUpdateProfileRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
}

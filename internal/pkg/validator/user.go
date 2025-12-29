package validator

type UserUpdateRequest struct {
	FirstName   string  `json:"first_name" binding:"min=1,max=50" validate:"alpha_numeric_space,no_xss"`
	LastName    string  `json:"last_name" binding:"min=1,max=50" validate:"alpha_numeric_space,no_xss"`
	PhoneNumber string  `json:"phone_number" validate:"phone"`
	Latitude    float64 `json:"latitude" validate:"latitude"`
	Longitude   float64 `json:"longitude" validate:"longitude"`
	FullAddress string  `json:"full_address" binding:"max=500" validate:"no_xss,no_sql_injection"`
	Note        string  `json:"note" binding:"max=200" validate:"no_xss,no_sql_injection"`
}

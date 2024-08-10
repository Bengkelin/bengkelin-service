package validator

type UserUpdateRequest struct {
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	PhoneNumber string  `json:"phone_number"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	FullAddress string  `json:"full_address"`
	Note        string  `json:"note"`
}

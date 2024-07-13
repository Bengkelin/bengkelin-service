package validator

type UserUpdateRequest struct {
	PhoneNumber string  `json:"phone_number"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	FullAddress string  `json:"full_address"`
	Note        string  `json:"note"`
}

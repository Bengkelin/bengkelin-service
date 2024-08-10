package validator

type PesananUpdateRequest struct {
	IsHomeService       bool   `json:"is_home_service"`
	HomeServiceSchedule string `json:"home_service_schedule"`
	PaymentMethod       string `json:"payment_method"`
}

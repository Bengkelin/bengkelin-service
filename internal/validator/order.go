package validator

type PesananUpdateRequest struct {
	IsHomeService       bool   `json:"is_home_service"`
	HomeServiceSchedule string `json:"home_service_schedule"`
	PaymentMethod       string `json:"payment_method"`
}

type PesananStatusUpdateRequest struct {
	Status uint   `json:"status" binding:"required,min=0,max=4"`
	Reason string `json:"reason,omitempty"` // Required when status = 4 (cancelled)
}

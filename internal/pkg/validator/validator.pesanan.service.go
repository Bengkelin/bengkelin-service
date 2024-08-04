package validator

type PesananServiceRequest struct {
	ServiceName []string  `json:"service_name" binding:"required"`
	Note        []string  `json:"note"`
	Price       []float64 `json:"price" binding:"required"`
	TotalPrice  float64   `json:"total_price"`
}

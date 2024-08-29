package validator

type PesananServiceRequest struct {
	Title  []string  `json:"title" binding:"required"`
	Detail []string  `json:"detail"`
	Price  []float64 `json:"price" binding:"required"`
}

package validator

// OrderServiceItem represents a single service item in an order
type OrderServiceItem struct {
	Title  string  `json:"title" binding:"required"`
	Detail string  `json:"detail"`
	Price  float64 `json:"price" binding:"required,min=0"`
}

// OrderServiceRequest represents the request for creating order services
type OrderServiceRequest struct {
	MitraId  string             `json:"mitra_id,omitempty"` // Optional for user self-orders
	Services []OrderServiceItem `json:"services" binding:"required,min=1,dive"`
}

// Legacy support - for backward compatibility (deprecated)
type OrderServiceRequestLegacy struct {
	MitraId string    `json:"mitra_id,omitempty"` // Optional for user self-orders
	Title   []string  `json:"title" binding:"required"`
	Detail  []string  `json:"detail"`
	Price   []float64 `json:"price" binding:"required"`
}

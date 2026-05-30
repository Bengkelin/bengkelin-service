package dto

import "time"

// CreateOrderRequest for creating a new order
type CreateOrderRequest struct {
	VehicleID   string   `json:"vehicle_id" validate:"required"`
	Services    []string `json:"services" validate:"required,min=1,dive,alpha_numeric_space"`
	AddressID   string   `json:"address_id" validate:"required"`
	Notes       string   `json:"notes" validate:"max=500,no_xss"`
	ServiceType string   `json:"service_type" validate:"required,oneof=home store"`
}

// OrderServiceItem for structured order service creation
type OrderServiceItem struct {
	Title  string  `json:"title"`
	Detail string  `json:"detail"`
	Price  float64 `json:"price"`
}

// CreateOrderWithServicesRequest for creating an order with structured services
type CreateOrderWithServicesRequest struct {
	MitraID         string              `json:"mitra_id"`
	Services        []OrderServiceItem  `json:"services"`
	IsHomeService   bool                `json:"is_home_service"`
	HomeServiceSchedule string          `json:"home_service_schedule,omitempty"`
	PaymentMethod   string              `json:"payment_method,omitempty"`
}

// CreateOrderResult contains the result of order creation
type CreateOrderResult struct {
	OrderID       string  `json:"order_id"`
	UserID        string  `json:"user_id"`
	BengkelID     string  `json:"bengkel_id"`
	BengkelName   string  `json:"bengkel_name"`
	TotalPrice    float64 `json:"total_price"`
	AdminFee      float64 `json:"admin_fee"`
	Status        int     `json:"status"`
	CreatedByName string  `json:"created_by_name"`
	IsSelfOrder   bool    `json:"is_self_order"`
}

// UpdateOrderStatusRequest for updating order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending confirmed in_progress completed cancelled"`
	Notes  string `json:"notes" validate:"max=500,no_xss"`
}

// OrderResponse for order data
type OrderResponse struct {
	ID          string                `json:"id"`
	UserID      string                `json:"user_id"`
	BengkelID   string                `json:"bengkel_id"`
	VehicleID   string                `json:"vehicle_id"`
	AddressID   string                `json:"address_id"`
	Status      string                `json:"status"`
	ServiceType string                `json:"service_type"`
	TotalPrice  float64               `json:"total_price"`
	Notes       string                `json:"notes,omitempty"`
	Services    []OrderServiceResponse `json:"services"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// OrderDetailResponse for detailed order information
type OrderDetailResponse struct {
	OrderResponse
	User    UserInfo        `json:"user"`
	Bengkel BengkelResponse `json:"bengkel"`
	Vehicle VehicleResponse `json:"vehicle"`
	Address AddressResponse `json:"address"`
}

// OrderServiceResponse for order service items
type OrderServiceResponse struct {
	ID          string  `json:"id"`
	OrderID     string  `json:"order_id"`
	ServiceName string  `json:"service_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

// PaginatedOrderResponse for paginated order list
type PaginatedOrderResponse struct {
	PaginationResponse
	Data []OrderResponse `json:"data"`
}
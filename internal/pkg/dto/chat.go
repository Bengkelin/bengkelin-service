package dto

import "time"

// ChatTokenResponse for chat token data
type ChatTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ChatHistoryRequest for saving chat history
type ChatHistoryRequest struct {
	UserID    string `json:"user_id" validate:"required"`
	BengkelID string `json:"bengkel_id" validate:"required"`
	Message   string `json:"message" validate:"required,max=1000,no_xss"`
	UserType  string `json:"user_type" validate:"required,oneof=user mitra"`
}

// ChatHistoryResponse for chat history data
type ChatHistoryResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	BengkelID string    `json:"bengkel_id"`
	Message   string    `json:"message"`
	UserType  string    `json:"user_type"`
	CreatedAt time.Time `json:"created_at"`
}
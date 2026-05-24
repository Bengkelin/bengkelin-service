package models

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken model for storing refresh tokens
type RefreshToken struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID    *string        `gorm:"type:varchar(36);index:idx_refresh_tokens_user_id" json:"user_id,omitempty"`
	MitraID   *string        `gorm:"type:varchar(36);index:idx_refresh_tokens_mitra_id" json:"mitra_id,omitempty"`
	Token     string         `gorm:"type:text;not null" json:"token"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	IsRevoked bool           `gorm:"default:false" json:"is_revoked"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	User  *User  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
	Mitra *Mitra `gorm:"foreignKey:MitraID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"mitra,omitempty"`
}

// TableName specifies the table name for RefreshToken model
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsExpired checks if the refresh token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid checks if the refresh token is valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked
}

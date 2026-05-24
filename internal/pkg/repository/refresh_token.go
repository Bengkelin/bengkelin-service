package repository

import (
	"context"
	"sync"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"gorm.io/gorm"
)

var (
	refreshTokenRepository *RefreshTokenRepository
	refreshTokenOnce       sync.Once
)

// RefreshTokenRepository struct
type RefreshTokenRepository struct {
	connection *gorm.DB
}

// RefreshTokenRepositoryInterface interface
type RefreshTokenRepositoryInterface interface {
	CreateRefreshToken(ctx context.Context, refreshToken models.RefreshToken) (models.RefreshToken, error)
	FindRefreshTokenByToken(ctx context.Context, token string) (models.RefreshToken, error)
	FindRefreshTokenByUserID(ctx context.Context, userID string) ([]models.RefreshToken, error)
	FindRefreshTokenByMitraID(ctx context.Context, mitraID string) ([]models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenID string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error
	RevokeAllMitraRefreshTokens(ctx context.Context, mitraID string) error
	DeleteExpiredTokens(ctx context.Context) error
	UpdateRefreshToken(ctx context.Context, refreshToken models.RefreshToken) (models.RefreshToken, error)
}

// GetRefreshTokenRepository function
func GetRefreshTokenRepository() RefreshTokenRepositoryInterface {
	refreshTokenOnce.Do(func() {
		refreshTokenRepository = &RefreshTokenRepository{
			connection: db.GetDB(),
		}
	})
	return refreshTokenRepository
}

// CreateRefreshToken creates a new refresh token
func (repo *RefreshTokenRepository) CreateRefreshToken(ctx context.Context, refreshToken models.RefreshToken) (models.RefreshToken, error) {
	err := repo.connection.WithContext(ctx).Create(&refreshToken).Error
	if err != nil {
		return models.RefreshToken{}, err
	}
	return refreshToken, nil
}

// FindRefreshTokenByToken finds a refresh token by token string
func (repo *RefreshTokenRepository) FindRefreshTokenByToken(ctx context.Context, token string) (models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := repo.connection.WithContext(ctx).Where("token = ? AND is_revoked = ? AND expires_at > ?", token, false, time.Now()).First(&refreshToken).Error
	if err != nil {
		return models.RefreshToken{}, err
	}
	return refreshToken, nil
}

// FindRefreshTokenByUserID finds all refresh tokens for a user
func (repo *RefreshTokenRepository) FindRefreshTokenByUserID(ctx context.Context, userID string) ([]models.RefreshToken, error) {
	var refreshTokens []models.RefreshToken
	err := repo.connection.WithContext(ctx).Where("user_id = ? AND is_revoked = ? AND expires_at > ?", userID, false, time.Now()).Find(&refreshTokens).Error
	if err != nil {
		return []models.RefreshToken{}, err
	}
	return refreshTokens, nil
}

// FindRefreshTokenByMitraID finds all refresh tokens for a mitra
func (repo *RefreshTokenRepository) FindRefreshTokenByMitraID(ctx context.Context, mitraID string) ([]models.RefreshToken, error) {
	var refreshTokens []models.RefreshToken
	err := repo.connection.WithContext(ctx).Where("mitra_id = ? AND is_revoked = ? AND expires_at > ?", mitraID, false, time.Now()).Find(&refreshTokens).Error
	if err != nil {
		return []models.RefreshToken{}, err
	}
	return refreshTokens, nil
}

// RevokeRefreshToken revokes a specific refresh token
func (repo *RefreshTokenRepository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	err := repo.connection.WithContext(ctx).Model(&models.RefreshToken{}).Where("id = ?", tokenID).Update("is_revoked", true).Error
	return err
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user
func (repo *RefreshTokenRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	err := repo.connection.WithContext(ctx).Model(&models.RefreshToken{}).Where("user_id = ?", userID).Update("is_revoked", true).Error
	return err
}

// RevokeAllMitraRefreshTokens revokes all refresh tokens for a mitra
func (repo *RefreshTokenRepository) RevokeAllMitraRefreshTokens(ctx context.Context, mitraID string) error {
	err := repo.connection.WithContext(ctx).Model(&models.RefreshToken{}).Where("mitra_id = ?", mitraID).Update("is_revoked", true).Error
	return err
}

// DeleteExpiredTokens deletes all expired refresh tokens
func (repo *RefreshTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	err := repo.connection.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.RefreshToken{}).Error
	return err
}

// UpdateRefreshToken updates a refresh token
func (repo *RefreshTokenRepository) UpdateRefreshToken(ctx context.Context, refreshToken models.RefreshToken) (models.RefreshToken, error) {
	err := repo.connection.WithContext(ctx).Save(&refreshToken).Error
	if err != nil {
		return models.RefreshToken{}, err
	}
	return refreshToken, nil
}

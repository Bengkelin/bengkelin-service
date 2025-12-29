package repository

import (
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"gorm.io/gorm"
)

var refreshTokenRepository *RefreshTokenRepository

// RefreshTokenRepository struct
type RefreshTokenRepository struct {
	connection *gorm.DB
}

// RefreshTokenRepositoryInterface interface
type RefreshTokenRepositoryInterface interface {
	CreateRefreshToken(refreshToken models.RefreshToken) (models.RefreshToken, error)
	FindRefreshTokenByToken(token string) (models.RefreshToken, error)
	FindRefreshTokenByUserID(userID string) ([]models.RefreshToken, error)
	FindRefreshTokenByMitraID(mitraID string) ([]models.RefreshToken, error)
	RevokeRefreshToken(tokenID string) error
	RevokeAllUserRefreshTokens(userID string) error
	RevokeAllMitraRefreshTokens(mitraID string) error
	DeleteExpiredTokens() error
	UpdateRefreshToken(refreshToken models.RefreshToken) (models.RefreshToken, error)
}

// GetRefreshTokenRepository function
func GetRefreshTokenRepository() RefreshTokenRepositoryInterface {
	if refreshTokenRepository == nil {
		refreshTokenRepository = &RefreshTokenRepository{
			connection: db.GetDB(),
		}
	}
	return refreshTokenRepository
}

// CreateRefreshToken creates a new refresh token
func (repo *RefreshTokenRepository) CreateRefreshToken(refreshToken models.RefreshToken) (models.RefreshToken, error) {
	err := repo.connection.Create(&refreshToken).Error
	if err != nil {
		return models.RefreshToken{}, err
	}
	return refreshToken, nil
}

// FindRefreshTokenByToken finds a refresh token by token string
func (repo *RefreshTokenRepository) FindRefreshTokenByToken(token string) (models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := repo.connection.Where("token = ? AND is_revoked = ? AND expires_at > ?", token, false, time.Now()).First(&refreshToken).Error
	if err != nil {
		return models.RefreshToken{}, err
	}
	return refreshToken, nil
}

// FindRefreshTokenByUserID finds all refresh tokens for a user
func (repo *RefreshTokenRepository) FindRefreshTokenByUserID(userID string) ([]models.RefreshToken, error) {
	var refreshTokens []models.RefreshToken
	err := repo.connection.Where("user_id = ? AND is_revoked = ? AND expires_at > ?", userID, false, time.Now()).Find(&refreshTokens).Error
	if err != nil {
		return []models.RefreshToken{}, err
	}
	return refreshTokens, nil
}

// FindRefreshTokenByMitraID finds all refresh tokens for a mitra
func (repo *RefreshTokenRepository) FindRefreshTokenByMitraID(mitraID string) ([]models.RefreshToken, error) {
	var refreshTokens []models.RefreshToken
	err := repo.connection.Where("mitra_id = ? AND is_revoked = ? AND expires_at > ?", mitraID, false, time.Now()).Find(&refreshTokens).Error
	if err != nil {
		return []models.RefreshToken{}, err
	}
	return refreshTokens, nil
}

// RevokeRefreshToken revokes a specific refresh token
func (repo *RefreshTokenRepository) RevokeRefreshToken(tokenID string) error {
	err := repo.connection.Model(&models.RefreshToken{}).Where("id = ?", tokenID).Update("is_revoked", true).Error
	return err
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user
func (repo *RefreshTokenRepository) RevokeAllUserRefreshTokens(userID string) error {
	err := repo.connection.Model(&models.RefreshToken{}).Where("user_id = ?", userID).Update("is_revoked", true).Error
	return err
}

// RevokeAllMitraRefreshTokens revokes all refresh tokens for a mitra
func (repo *RefreshTokenRepository) RevokeAllMitraRefreshTokens(mitraID string) error {
	err := repo.connection.Model(&models.RefreshToken{}).Where("mitra_id = ?", mitraID).Update("is_revoked", true).Error
	return err
}

// DeleteExpiredTokens deletes all expired refresh tokens
func (repo *RefreshTokenRepository) DeleteExpiredTokens() error {
	err := repo.connection.Where("expires_at < ?", time.Now()).Delete(&models.RefreshToken{}).Error
	return err
}

// UpdateRefreshToken updates a refresh token
func (repo *RefreshTokenRepository) UpdateRefreshToken(refreshToken models.RefreshToken) (models.RefreshToken, error) {
	err := repo.connection.Save(&refreshToken).Error
	if err != nil {
		return models.RefreshToken{}, err
	}
	return refreshToken, nil
}
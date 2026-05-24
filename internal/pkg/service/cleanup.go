package service

import (
	"context"
	"sync"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
)

var (
	cleanupService *CleanupService
	cleanupOnce    sync.Once
)

// CleanupService handles periodic cleanup tasks
type CleanupService struct {
	refreshTokenRepo repository.RefreshTokenRepositoryInterface
}

// CleanupServiceInterface defines the cleanup service contract
type CleanupServiceInterface interface {
	StartPeriodicCleanup()
	CleanupExpiredTokens() error
}

// NewCleanupService creates a new cleanup service with injected dependencies
func NewCleanupService(refreshTokenRepo repository.RefreshTokenRepositoryInterface) CleanupServiceInterface {
	return &CleanupService{
		refreshTokenRepo: refreshTokenRepo,
	}
}

// GetCleanupService returns singleton instance of cleanup service
func GetCleanupService() CleanupServiceInterface {
	cleanupOnce.Do(func() {
		cleanupService = &CleanupService{
			refreshTokenRepo: repository.GetRefreshTokenRepository(),
		}
	})
	return cleanupService
}

// StartPeriodicCleanup starts a goroutine that periodically cleans up expired tokens
func (s *CleanupService) StartPeriodicCleanup() {
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Run cleanup every 24 hours
		defer ticker.Stop()

		applog.Info("Started periodic cleanup service for expired refresh tokens")

		for {
			select {
			case <-ticker.C:
				err := s.CleanupExpiredTokens()
				if err != nil {
					applog.Error("Failed to cleanup expired tokens", "error", err.Error())
				} else {
					applog.Info("Successfully cleaned up expired refresh tokens")
				}
			}
		}
	}()
}

// CleanupExpiredTokens removes all expired refresh tokens from database
func (s *CleanupService) CleanupExpiredTokens() error {
	applog.Debug("Starting cleanup of expired refresh tokens")
	err := s.refreshTokenRepo.DeleteExpiredTokens(context.Background())
	if err != nil {
		applog.Error("Error during token cleanup", "error", err.Error())
		return err
	}
	applog.Debug("Completed cleanup of expired refresh tokens")
	return nil
}

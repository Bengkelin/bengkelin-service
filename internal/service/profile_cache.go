package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/events"
	"github.com/Bengkelin/bengkelin-service/internal/models"
	redisClient "github.com/Bengkelin/bengkelin-service/internal/redis"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
)

const (
	// Cache TTL for profile data
	MitraProfileCacheTTL  = 5 * time.Minute
	UserProfileCacheTTL   = 5 * time.Minute
	BengkelDetailCacheTTL = 2 * time.Minute

	// Cache key prefixes
	MitraProfileCachePrefix  = "profile:mitra:"
	UserProfileCachePrefix   = "profile:user:"
	BengkelDetailCachePrefix = "bengkel:detail:"
	// Repository-level cache key prefix (used in repository/bengkel.go)
	BengkelRepositoryCachePrefix = "bengkel:"

	// Redis pub/sub channel for cache invalidation
	CacheInvalidationChannel = "cache:invalidation"
)

// ProfileCacheService handles caching for profile endpoints
type ProfileCacheService struct {
	redis         *redisClient.RedisCache
	pubsubEnabled bool
	subscriptions map[string]struct{}
	mu            sync.RWMutex
	instanceID    string
}

// NewProfileCacheService creates a new profile cache service
func NewProfileCacheService() *ProfileCacheService {
	service := &ProfileCacheService{
		redis:         redisClient.GetRedisClient(),
		pubsubEnabled: os.Getenv("REDIS_PUBSUB_ENABLED") != "false", // Default to true
		subscriptions: make(map[string]struct{}),
		instanceID:    generateInstanceID(),
	}

	// Start the pub/sub subscriber for distributed cache invalidation
	if service.pubsubEnabled && service.redis != nil {
		go service.startInvalidationSubscriber()
	}

	return service
}

// generateInstanceID creates a unique identifier for this service instance
func generateInstanceID() string {
	return fmt.Sprintf("%s-%d", time.Now().Format("20060102150405"), time.Now().UnixNano())
}

// startInvalidationSubscriber subscribes to Redis pub/sub for distributed cache invalidation
func (s *ProfileCacheService) startInvalidationSubscriber() {
	ctx := context.Background()

	client := s.redis.GetClient()
	if client == nil {
		applog.Warn("Redis client not available, skipping pub/sub subscriber")
		return
	}

	// Subscribe to cache invalidation channel
	pubsub := client.Subscribe(ctx, CacheInvalidationChannel)
	defer pubsub.Close()

	// Wait for confirmation that subscription is created
	_, err := pubsub.Receive(ctx)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to subscribe to cache invalidation channel")
		return
	}

	applog.Info("Started cache invalidation subscriber", "instance_id", s.instanceID, "channel", CacheInvalidationChannel)

	// Listen for messages
	ch := pubsub.Channel()
	for msg := range ch {
		s.handleInvalidationMessage(ctx, msg.Payload)
	}
}

// handleInvalidationMessage processes incoming cache invalidation messages
func (s *ProfileCacheService) handleInvalidationMessage(ctx context.Context, payload string) {
	var event events.CacheInvalidationEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		applog.WarnCtx(ctx, "Failed to unmarshal cache invalidation event", "error", err, "payload", payload)
		return
	}

	// Skip events from this same instance (already handled locally)
	if event.Source == s.instanceID {
		return
	}

	applog.DebugCtx(ctx, "Received cache invalidation event",
		"entity_type", event.EntityType,
		"entity_id", event.EntityID,
		"action", event.Action,
		"source", event.Source,
	)

	// Handle the invalidation based on entity type
	switch event.EntityType {
	case "mitra":
		_ = s.InvalidateMitraProfile(ctx, event.EntityID)
	case "user":
		_ = s.InvalidateUserProfile(ctx, event.EntityID)
	case "bengkel":
		_ = s.InvalidateBengkelDetail(ctx, event.EntityID)
	default:
		applog.WarnCtx(ctx, "Unknown entity type in cache invalidation event", "entity_type", event.EntityType)
	}
}

// PublishInvalidationEvent publishes a cache invalidation event to Redis pub/sub
func (s *ProfileCacheService) PublishInvalidationEvent(ctx context.Context, entityType, entityID, action string) error {
	if !s.pubsubEnabled || s.redis == nil {
		return nil
	}

	event := events.CacheInvalidationEvent{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Timestamp:  time.Now().Unix(),
		Source:     s.instanceID,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal cache invalidation event: %w", err)
	}

	client := s.redis.GetClient()
	if client == nil {
		return fmt.Errorf("redis client not available")
	}

	result := client.Publish(ctx, CacheInvalidationChannel, payload)
	if err := result.Err(); err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to publish cache invalidation event",
			"entity_type", entityType,
			"entity_id", entityID,
		)
		return err
	}

	applog.DebugCtx(ctx, "Published cache invalidation event",
		"entity_type", entityType,
		"entity_id", entityID,
		"action", action,
		"subscribers", result.Val(),
	)

	return nil
}

// GetMitraProfile retrieves mitra profile from cache or returns nil if not found
func (s *ProfileCacheService) GetMitraProfile(ctx context.Context, mitraID string) (*models.Mitra, error) {
	if s.redis == nil {
		return nil, nil
	}

	cacheKey := fmt.Sprintf("%s%s", MitraProfileCachePrefix, mitraID)

	var mitra models.Mitra
	err := s.redis.GetWithContext(ctx, cacheKey, &mitra)
	if err != nil {
		// Cache miss or error
		return nil, err
	}

	applog.DebugCtx(ctx, "Profile cache hit", "mitra_id", mitraID, "cache_key", cacheKey)
	return &mitra, nil
}

// SetMitraProfile caches mitra profile
func (s *ProfileCacheService) SetMitraProfile(ctx context.Context, mitra *models.Mitra) error {
	if s.redis == nil || mitra == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("%s%s", MitraProfileCachePrefix, mitra.ID)

	err := s.redis.SetWithContext(ctx, cacheKey, mitra, MitraProfileCacheTTL)
	if err != nil {
		applog.WarnCtx(ctx, "Failed to cache mitra profile", "mitra_id", mitra.ID, "error", err)
		return err
	}

	applog.DebugCtx(ctx, "Profile cached", "mitra_id", mitra.ID, "cache_key", cacheKey, "ttl", MitraProfileCacheTTL)
	return nil
}

// InvalidateMitraProfile removes mitra profile from cache and publishes invalidation event
func (s *ProfileCacheService) InvalidateMitraProfile(ctx context.Context, mitraID string) error {
	if s.redis == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("%s%s", MitraProfileCachePrefix, mitraID)
	err := s.redis.DeleteWithContext(ctx, cacheKey)
	if err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate mitra profile cache", "mitra_id", mitraID, "error", err)
		return err
	}

	applog.DebugCtx(ctx, "Profile cache invalidated", "mitra_id", mitraID, "cache_key", cacheKey)

	// Publish event for distributed cache invalidation
	_ = s.PublishInvalidationEvent(ctx, "mitra", mitraID, "updated")

	return nil
}

// GetUserProfile retrieves user profile from cache
func (s *ProfileCacheService) GetUserProfile(ctx context.Context, userID string) (*models.User, error) {
	if s.redis == nil {
		return nil, nil
	}

	cacheKey := fmt.Sprintf("%s%s", UserProfileCachePrefix, userID)

	var user models.User
	err := s.redis.GetWithContext(ctx, cacheKey, &user)
	if err != nil {
		return nil, err
	}

	applog.DebugCtx(ctx, "User profile cache hit", "user_id", userID, "cache_key", cacheKey)
	return &user, nil
}

// SetUserProfile caches user profile
func (s *ProfileCacheService) SetUserProfile(ctx context.Context, user *models.User) error {
	if s.redis == nil || user == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("%s%s", UserProfileCachePrefix, user.ID)

	err := s.redis.SetWithContext(ctx, cacheKey, user, UserProfileCacheTTL)
	if err != nil {
		applog.WarnCtx(ctx, "Failed to cache user profile", "user_id", user.ID, "error", err)
		return err
	}

	applog.DebugCtx(ctx, "User profile cached", "user_id", user.ID, "cache_key", cacheKey, "ttl", UserProfileCacheTTL)
	return nil
}

// InvalidateUserProfile removes user profile from cache and publishes invalidation event
func (s *ProfileCacheService) InvalidateUserProfile(ctx context.Context, userID string) error {
	if s.redis == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("%s%s", UserProfileCachePrefix, userID)
	err := s.redis.DeleteWithContext(ctx, cacheKey)
	if err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate user profile cache", "user_id", userID, "error", err)
		return err
	}

	applog.DebugCtx(ctx, "User profile cache invalidated", "user_id", userID, "cache_key", cacheKey)

	// Publish event for distributed cache invalidation
	_ = s.PublishInvalidationEvent(ctx, "user", userID, "updated")

	return nil
}

// GetBengkelDetail retrieves bengkel detail from cache
func (s *ProfileCacheService) GetBengkelDetail(ctx context.Context, bengkelID string) (*models.Bengkel, error) {
	if s.redis == nil {
		return nil, nil
	}

	cacheKey := fmt.Sprintf("%s%s", BengkelDetailCachePrefix, bengkelID)

	var bengkel models.Bengkel
	err := s.redis.GetWithContext(ctx, cacheKey, &bengkel)
	if err != nil {
		return nil, err
	}

	applog.DebugCtx(ctx, "Bengkel detail cache hit", "bengkel_id", bengkelID, "cache_key", cacheKey)
	return &bengkel, nil
}

// SetBengkelDetail caches bengkel detail
func (s *ProfileCacheService) SetBengkelDetail(ctx context.Context, bengkel *models.Bengkel) error {
	if s.redis == nil || bengkel == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("%s%s", BengkelDetailCachePrefix, bengkel.ID)

	err := s.redis.SetWithContext(ctx, cacheKey, bengkel, BengkelDetailCacheTTL)
	if err != nil {
		applog.WarnCtx(ctx, "Failed to cache bengkel detail", "bengkel_id", bengkel.ID, "error", err)
		return err
	}

	applog.DebugCtx(ctx, "Bengkel detail cached", "bengkel_id", bengkel.ID, "cache_key", cacheKey, "ttl", BengkelDetailCacheTTL)
	return nil
}

// InvalidateBengkelDetail removes bengkel detail from cache and publishes invalidation event
func (s *ProfileCacheService) InvalidateBengkelDetail(ctx context.Context, bengkelID string) error {
	if s.redis == nil {
		return nil
	}

	// Delete ProfileCacheService key (bengkel:detail:)
	cacheKey := fmt.Sprintf("%s%s", BengkelDetailCachePrefix, bengkelID)
	err := s.redis.DeleteWithContext(ctx, cacheKey)
	if err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate bengkel detail cache", "bengkel_id", bengkelID, "error", err)
	}

	// Also delete repository-level cache key (bengkel:) to ensure consistency
	repoCacheKey := fmt.Sprintf("%s%s", BengkelRepositoryCachePrefix, bengkelID)
	err = s.redis.DeleteWithContext(ctx, repoCacheKey)
	if err != nil {
		applog.WarnCtx(ctx, "Failed to invalidate bengkel repository cache", "bengkel_id", bengkelID, "error", err)
	}

	applog.DebugCtx(ctx, "Bengkel detail cache invalidated", "bengkel_id", bengkelID,
		"cache_key", cacheKey, "repo_cache_key", repoCacheKey)

	// Publish event for distributed cache invalidation
	_ = s.PublishInvalidationEvent(ctx, "bengkel", bengkelID, "updated")

	return nil
}

// InvalidateMitraRelatedCaches invalidates all caches related to a mitra (profile + bengkel details)
func (s *ProfileCacheService) InvalidateMitraRelatedCaches(ctx context.Context, mitraID string, bengkelIDs []string) error {
	// Invalidate mitra profile
	_ = s.InvalidateMitraProfile(ctx, mitraID)

	// Invalidate all bengkels owned by this mitra
	for _, bengkelID := range bengkelIDs {
		_ = s.InvalidateBengkelDetail(ctx, bengkelID)
	}

	applog.InfoCtx(ctx, "Invalidated mitra related caches",
		"mitra_id", mitraID,
		"bengkel_count", len(bengkelIDs),
	)

	return nil
}

// IsAvailable checks if cache service is available
func (s *ProfileCacheService) IsAvailable() bool {
	return s.redis != nil && s.redis.IsConnected()
}

// CacheStats provides cache statistics (for monitoring)
type CacheStats struct {
	Available     bool   `json:"available"`
	PubSubEnabled bool   `json:"pubsub_enabled"`
	InstanceID    string `json:"instance_id"`
}

// GetStats returns cache statistics
func (s *ProfileCacheService) GetStats() CacheStats {
	return CacheStats{
		Available:     s.IsAvailable(),
		PubSubEnabled: s.pubsubEnabled,
		InstanceID:    s.instanceID,
	}
}

// Global instance
var profileCacheService *ProfileCacheService
var profileCacheOnce sync.Once

// GetProfileCacheService returns the global profile cache service instance
func GetProfileCacheService() *ProfileCacheService {
	profileCacheOnce.Do(func() {
		profileCacheService = NewProfileCacheService()
	})
	return profileCacheService
}

// ResetProfileCacheService resets the global instance (useful for testing)
func ResetProfileCacheService() {
	profileCacheService = nil
	profileCacheOnce = sync.Once{}
}

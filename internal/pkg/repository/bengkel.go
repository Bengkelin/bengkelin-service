package repository

import (
	"context"
	"sync"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/redis"
)

var (
	bengkelRepository *BengkelRepository
	bengkelOnce       sync.Once
)

type BengkelWithDistance struct {
	models.Bengkel
	Distance float64 `json:"distance"`
}

type BengkelRepositoryInterface interface {
	CreateBengkel(ctx context.Context, bengkel models.Bengkel) (models.Bengkel, error)
	UpdateBengkelById(ctx context.Context, bengkelId string, bengkel *models.Bengkel) error
	GetBengkelById(ctx context.Context, bengkelId string) (*models.Bengkel, error)
	GetBengkelByIdFresh(ctx context.Context, bengkelId string) (*models.Bengkel, error) // Bypass cache
	FindBengkelById(ctx context.Context, bengkelId string, page, size int) (*models.Bengkel, []models.BengkelTestimonial, int, error)
	GetAllBengkel(ctx context.Context) ([]models.Bengkel, error)
	GetAllBengkelPaginate(ctx context.Context, page int, limit int) ([]models.Bengkel, int, error)
	GetNearestBengkelPaginate(ctx context.Context, lat, lng float64, page, limit int) ([]BengkelWithDistance, int, error)
	GetBengkelSearch(ctx context.Context, query string, page int, limit int) ([]models.Bengkel, int, error)
	GetBengkelByFilterService(ctx context.Context, service string, page int, limit int) ([]models.Bengkel, int, error)
	GetBengkelSearchV2(ctx context.Context, service, query string, page int, limit int) ([]models.Bengkel, int, error)
	SearchBengkelPublic(ctx context.Context, criteria map[string]interface{}) ([]models.Bengkel, int, error)
}

type BengkelRepository struct{}

func GetBengkelRepository() BengkelRepositoryInterface {
	bengkelOnce.Do(func() {
		bengkelRepository = &BengkelRepository{}
	})
	return bengkelRepository
}

// CreateBengkel implements BengkelRepositoryInterface.
func (repo *BengkelRepository) CreateBengkel(ctx context.Context, bengkel models.Bengkel) (models.Bengkel, error) {
	err := Create(ctx, &bengkel)
	// If error when transaction to database i.e duplicate email
	if err != nil {
		return models.Bengkel{}, err
	}
	return bengkel, nil
}

// UpdateBengkelById implements BengkelRepositoryInterface.
func (*BengkelRepository) UpdateBengkelById(ctx context.Context, bengkelId string, bengkel *models.Bengkel) error {
	err := db.GetDB().WithContext(ctx).Model(&models.Bengkel{}).Where("id = ?", bengkelId).Updates(bengkel).Error

	if err != nil {
		return err
	}

	// Clear cache after successful update
	cache := redis.GetRedisClient()
	cacheKey := "bengkel:" + bengkelId
	cache.DeleteWithContext(ctx, cacheKey) // Ignore error as cache deletion is not critical

	return nil
}

// FindBengkelById implements BengkelRepositoryInterface.
func (*BengkelRepository) FindBengkelById(ctx context.Context, bengkelId string, page, size int) (*models.Bengkel, []models.BengkelTestimonial, int, error) {
	var bengkel models.Bengkel
	where := models.Bengkel{}
	where.ID = bengkelId
	_, err := First(ctx, where, &bengkel, []string{"Photos", "Services", "Addresses", "Operasionals"})
	if err != nil {
		return nil, nil, 0, err
	}

	var count int64

	err = db.GetDB().WithContext(ctx).Model(&models.BengkelTestimonial{}).Where("bengkel_id = ?", bengkelId).Count(&count).Error
	if err != nil {
		return nil, nil, 0, err
	}

	offset := (page - 1) * size

	var testimonies []models.BengkelTestimonial
	err = db.GetDB().WithContext(ctx).Where("bengkel_id = ?", bengkelId).Preload("User").Offset(offset).Limit(size).Find(&testimonies).Error
	if err != nil {
		return nil, nil, 0, err
	}

	return &bengkel, testimonies, int(count), nil
}

// GetBengkelById implements BengkelRepositoryInterface.
func (*BengkelRepository) GetBengkelById(ctx context.Context, bengkelId string) (*models.Bengkel, error) {
	cache := redis.GetRedisClient()
	cacheKey := "bengkel:" + bengkelId

	// Try cache first
	var bengkel models.Bengkel
	if err := cache.GetWithContext(ctx, cacheKey, &bengkel); err == nil {
		return &bengkel, nil
	}

	// Fallback to database with preloaded relationships
	where := models.Bengkel{}
	where.ID = bengkelId
	_, err := First(ctx, where, &bengkel, []string{"Photos", "Services", "Addresses", "Operasionals"})
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	cache.SetWithContext(ctx, cacheKey, bengkel, 5*time.Minute)
	return &bengkel, nil
}

// GetAllBengkel implements BengkelRepositoryInterface.
func (*BengkelRepository) GetAllBengkel(ctx context.Context) ([]models.Bengkel, error) {
	var bengkels []models.Bengkel
	err := Find(ctx, models.Bengkel{}, &bengkels, []string{"Photos", "Addresses", "Operasionals"})
	if err != nil {
		return nil, err
	}
	return bengkels, nil
}

// GetAllBengkelPaginate implements BengkelRepositoryInterface.
func (*BengkelRepository) GetAllBengkelPaginate(ctx context.Context, page int, limit int) ([]models.Bengkel, int, error) {
	var bengkels []models.Bengkel
	var count int64
	err := db.GetDB().WithContext(ctx).Model(&models.Bengkel{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.GetDB().WithContext(ctx).Preload("Photos").Preload("Services").Preload("Addresses").Preload("Operasionals").Offset((page - 1) * limit).Limit(limit).Find(&bengkels).Error
	if err != nil {
		return nil, 0, err
	}
	return bengkels, int(count), nil
}

// GetNearestBengkelPaginate fetches bengkels sorted by distance from a given point.
// Uses Haversine formula in SQL to calculate and filter by distance, avoiding N+1 queries.
// Only preloads Addresses (needed for distance) and Photos (needed for response).
func (*BengkelRepository) GetNearestBengkelPaginate(ctx context.Context, lat, lng float64, page, limit int) ([]BengkelWithDistance, int, error) {
	maxDistanceKm := 50.0 // 50km radius

	haversineExpr := `(6371 * acos(cos(radians(?)) * cos(radians(ba.latitude)) * cos(radians(ba.longitude) - radians(?)) + sin(radians(?)) * sin(radians(ba.latitude))))`

	// Step 1: Get bengkel IDs within range, with distance, using a single JOIN query
	type bengkelIDWithDistance struct {
		BengkelID string
		Distance  float64
	}
	var idResults []bengkelIDWithDistance

	err := db.GetDB().WithContext(ctx).
		Table("bengkels AS b").
		Select("b.id AS bengkel_id, "+haversineExpr+" AS distance", lat, lng, lat).
		Joins("INNER JOIN bengkel_addresses AS ba ON ba.bengkel_id = b.id").
		Where(haversineExpr+" < ?", lat, lng, lat, maxDistanceKm).
		Group("b.id, ba.latitude, ba.longitude").
		Order("distance ASC").
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(&idResults).Error
	if err != nil {
		return nil, 0, err
	}

	if len(idResults) == 0 {
		return []BengkelWithDistance{}, 0, nil
	}

	// Step 2: Count total within range (for pagination)
	var totalCount int64
	err = db.GetDB().WithContext(ctx).
		Table("bengkels AS b").
		Select("b.id").
		Joins("INNER JOIN bengkel_addresses AS ba ON ba.bengkel_id = b.id").
		Where(haversineExpr+" < ?", lat, lng, lat, maxDistanceKm).
		Group("b.id").
		Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Step 3: Load full bengkel objects with necessary associations in batch
	ids := make([]string, len(idResults))
	distanceMap := make(map[string]float64, len(idResults))
	for i, r := range idResults {
		ids[i] = r.BengkelID
		distanceMap[r.BengkelID] = r.Distance
	}

	var bengkels []models.Bengkel
	err = db.GetDB().WithContext(ctx).
		Preload("Photos").
		Preload("Services").
		Preload("Addresses").
		Preload("Operasionals").
		Where("id IN ?", ids).
		Find(&bengkels).Error
	if err != nil {
		return nil, 0, err
	}

	// Step 4: Combine results, preserving distance order
	bengkelMap := make(map[string]models.Bengkel, len(bengkels))
	for _, b := range bengkels {
		bengkelMap[b.ID] = b
	}

	results := make([]BengkelWithDistance, 0, len(idResults))
	for _, r := range idResults {
		if b, ok := bengkelMap[r.BengkelID]; ok {
			results = append(results, BengkelWithDistance{
				Bengkel:  b,
				Distance: r.Distance,
			})
		}
	}

	return results, int(totalCount), nil
}

// GetBengkelSearch implements BengkelRepositoryInterface.
func (*BengkelRepository) GetBengkelSearch(ctx context.Context, query string, page int, limit int) ([]models.Bengkel, int, error) {
	var bengkels []models.Bengkel

	var count int64
	// Create a subquery to filter bengkels based on the services
	subQuery := db.GetDB().WithContext(ctx).Model(&models.Bengkel{}).
		Joins("LEFT JOIN bengkel_services as bs ON bs.bengkel_id = bengkels.id").
		Joins("LEFT JOIN bengkel_addresses as ba ON ba.bengkel_id = bengkels.id").
		Where("bengkel_name LIKE ? OR bs.nama_service LIKE ? OR ba.full_address LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Group("bengkels.id")

	// Count the total number of filtered bengkels
	err := db.GetDB().WithContext(ctx).Model(&models.Bengkel{}).
		Where("id IN (?)", subQuery.Select("bengkels.id")).
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Fetch the filtered bengkels with associations
	err = db.GetDB().WithContext(ctx).Model(&models.Bengkel{}).
		Where("id IN (?)", subQuery.Select("bengkels.id")).
		Preload("Photos").
		Preload("Services").
		Preload("Addresses").
		Preload("Operasionals").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&bengkels).Error
	if err != nil {
		return nil, 0, err
	}
	return bengkels, int(count), nil
}

// GetBengkelByFilterService implements BengkelRepositoryInterface.
func (*BengkelRepository) GetBengkelByFilterService(ctx context.Context, service string, page int, limit int) ([]models.Bengkel, int, error) {
	var bengkels []models.Bengkel

	var count int64

	query := db.GetDB().WithContext(ctx).Model(&models.Bengkel{})
	if service == "home_service" {
		err := query.Where("home_service = true").Count(&count).Error
		if err != nil {
			return nil, 0, err
		}
		err = query.
			Where("home_service = true").
			Preload("Photos").
			Preload("Services").
			Preload("Addresses").
			Preload("Operasionals").
			Offset((page - 1) * limit).
			Limit(limit).
			Find(&bengkels).Error
		if err != nil {
			return nil, 0, err
		}
		return bengkels, int(count), nil
	}

	err := query.Where("store_service = true").Count(&count).Error

	if err != nil {
		return nil, 0, err
	}
	err = query.
		Where("store_service = true").
		Preload("Photos").
		Preload("Services").
		Preload("Addresses").
		Preload("Operasionals").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&bengkels).Error
	if err != nil {
		return nil, 0, err
	}
	return bengkels, int(count), nil
}

// GetBengkelSearchV2 implements BengkelRepositoryInterface.
func (*BengkelRepository) GetBengkelSearchV2(ctx context.Context, service, query string, page int, limit int) ([]models.Bengkel, int, error) {
	var bengkels []models.Bengkel
	var count int64

	subQuery := db.GetDB().WithContext(ctx).Model(&models.Bengkel{})

	if service == "home_service" {
		subQuery = subQuery.Where("home_service = true")
	} else if service == "store_service" {
		subQuery = subQuery.Where("store_service = true")
	}

	if query != "" {
		subQuery = db.GetDB().WithContext(ctx).Model(&models.Bengkel{}).
			Joins("LEFT JOIN bengkel_services as bs ON bs.bengkel_id = bengkels.id").
			Joins("LEFT JOIN bengkel_addresses as ba ON ba.bengkel_id = bengkels.id").
			Where("bengkel_name LIKE ? OR bs.nama_service LIKE ? OR ba.full_address LIKE ?",
				"%"+query+"%", "%"+query+"%", "%"+query+"%").
			Group("bengkels.id")
	}

	// Count the total number of filtered bengkels
	err := db.GetDB().WithContext(ctx).Model(&models.Bengkel{}).
		Where("id IN (?)", subQuery.Select("bengkels.id")).
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Fetch the filtered bengkels with associations
	err = db.GetDB().WithContext(ctx).Model(&models.Bengkel{}).
		Where("id IN (?)", subQuery.Select("bengkels.id")).
		Preload("Photos").
		Preload("Services").
		Preload("Addresses").
		Preload("Operasionals").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&bengkels).Error
	if err != nil {
		return nil, 0, err
	}

	return bengkels, int(count), nil
}

// SearchBengkelPublic implements BengkelRepositoryInterface.
// Public search method with flexible criteria for unauthenticated users
func (repo *BengkelRepository) SearchBengkelPublic(ctx context.Context, criteria map[string]interface{}) ([]models.Bengkel, int, error) {
	var bengkels []models.Bengkel
	var count int64

	// Extract pagination parameters
	page, ok := criteria["page"].(int)
	if !ok || page < 1 {
		page = 1
	}

	limit, ok := criteria["limit"].(int)
	if !ok || limit < 1 {
		limit = 10
	}

	// Build the query
	query := db.GetDB().WithContext(ctx).Model(&models.Bengkel{})

	// Apply search filters
	if searchQuery, exists := criteria["query"].(string); exists && searchQuery != "" {
		query = query.Where("bengkel_name ILIKE ?",
			"%"+searchQuery+"%")
	}

	// Filter by service if specified
	if service, exists := criteria["service"].(string); exists && service != "" {
		query = query.Joins("JOIN bengkel_services ON bengkels.id = bengkel_services.bengkel_id").
			Where("bengkel_services.nama_service ILIKE ? AND bengkel_services.is_available = true",
				"%"+service+"%")
	}

	// Filter by location if specified
	if city, exists := criteria["city"].(string); exists && city != "" {
		query = query.Joins("JOIN bengkel_addresses ON bengkels.id = bengkel_addresses.bengkel_id").
			Where("bengkel_addresses.city ILIKE ?", "%"+city+"%")
	}

	if province, exists := criteria["province"].(string); exists && province != "" {
		if city, cityExists := criteria["city"].(string); !cityExists || city == "" {
			// Only join if we haven't already joined for city
			query = query.Joins("JOIN bengkel_addresses ON bengkels.id = bengkel_addresses.bengkel_id")
		}
		query = query.Where("bengkel_addresses.province ILIKE ?", "%"+province+"%")
	}

	// Only show active bengkels
	query = query.Where("bengkels.deleted_at IS NULL")

	// Get total count
	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results with preloaded relationships
	err = query.
		Preload("Photos").
		Preload("Services", "is_available = true"). // Only load available services
		Preload("Addresses").
		Preload("Operasionals").
		Distinct("bengkels.*"). // Ensure distinct results when joining
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&bengkels).Error
	if err != nil {
		return nil, 0, err
	}

	return bengkels, int(count), nil
}

// GetBengkelByIdFresh bypasses cache and gets fresh data from database
func (*BengkelRepository) GetBengkelByIdFresh(ctx context.Context, bengkelId string) (*models.Bengkel, error) {
	// Always query database, bypass cache
	var bengkel models.Bengkel
	where := models.Bengkel{}
	where.ID = bengkelId
	_, err := First(ctx, where, &bengkel, []string{"Photos", "Services", "Addresses", "Operasionals"})
	if err != nil {
		return nil, err
	}

	// Update cache with fresh data
	cache := redis.GetRedisClient()
	cacheKey := "bengkel:" + bengkelId
	cache.SetWithContext(ctx, cacheKey, bengkel, 5*time.Minute)

	return &bengkel, nil
}

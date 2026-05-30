package repository

import (
	"context"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/models"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
	"gorm.io/gorm"
)

// AuditRepository handles audit log database operations
type AuditRepository struct {
	db *gorm.DB
}

// AuditRepositoryInterface defines audit repository methods
type AuditRepositoryInterface interface {
	CreateAuditLog(ctx context.Context, audit *models.AuditLog) error
	GetAuditLogs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.AuditLog, error)
	GetAuditLogByID(ctx context.Context, id string) (*models.AuditLog, error)
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *gorm.DB) AuditRepositoryInterface {
	return &AuditRepository{
		db: db,
	}
}

// CreateAuditLog creates a new audit log entry
func (r *AuditRepository) CreateAuditLog(ctx context.Context, audit *models.AuditLog) error {
	audit.CreatedAt = time.Now()
	
	if err := r.db.WithContext(ctx).Create(audit).Error; err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to create audit log", "audit_id", audit.ID)
		return err
	}
	
	applog.InfoCtx(ctx, "Audit log created successfully", "audit_id", audit.ID)
	return nil
}

// GetAuditLogs retrieves audit logs with filters
func (r *AuditRepository) GetAuditLogs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.AuditLog, error) {
	var audits []models.AuditLog
	
	query := r.db.WithContext(ctx)
	
	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}
	
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&audits).Error; err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to get audit logs")
		return nil, err
	}
	
	return audits, nil
}

// GetAuditLogByID retrieves an audit log by ID
func (r *AuditRepository) GetAuditLogByID(ctx context.Context, id string) (*models.AuditLog, error) {
	var audit models.AuditLog
	
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&audit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		applog.LogErrorCtx(ctx, err, "Failed to get audit log by ID", "audit_id", id)
		return nil, err
	}
	
	return &audit, nil
}
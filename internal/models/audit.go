package models

import (
	"encoding/json"
	"time"
)

// AuditLog represents system audit logs for tracking user activities and system events
type AuditLog struct {
	ID          string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID      *string         `json:"user_id,omitempty" gorm:"type:varchar(36);index"`
	MitraID     *string         `json:"mitra_id,omitempty" gorm:"type:varchar(36);index"`
	SessionID   *string         `json:"session_id,omitempty" gorm:"type:varchar(255);index"`
	Action      string          `json:"action" gorm:"type:varchar(100);not null;index"`
	Resource    string          `json:"resource" gorm:"type:varchar(100);not null;index"`
	ResourceID  *string         `json:"resource_id,omitempty" gorm:"type:varchar(36);index"`
	Method      string          `json:"method" gorm:"type:varchar(10);not null"`
	Endpoint    string          `json:"endpoint" gorm:"type:varchar(255);not null"`
	IPAddress   string          `json:"ip_address" gorm:"type:varchar(45);not null;index"`
	UserAgent   string          `json:"user_agent" gorm:"type:text"`
	StatusCode  int             `json:"status_code" gorm:"not null;index"`
	Duration    int64           `json:"duration" gorm:"not null"` // in milliseconds
	RequestSize int64           `json:"request_size" gorm:"default:0"`
	ResponseSize int64          `json:"response_size" gorm:"default:0"`
	Metadata    json.RawMessage `json:"metadata,omitempty" gorm:"type:json"`
	ErrorMsg    *string         `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt   time.Time       `json:"created_at" gorm:"autoCreateTime;index"`
	
	// Relationships
	User  *User  `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	Mitra *Mitra `json:"mitra,omitempty" gorm:"foreignKey:MitraID;references:ID"`
}

// SecurityEvent represents security-related events for monitoring
type SecurityEvent struct {
	ID          string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	EventType   string          `json:"event_type" gorm:"type:varchar(50);not null;index"`
	Severity    string          `json:"severity" gorm:"type:enum('low','medium','high','critical');not null;index"`
	UserID      *string         `json:"user_id,omitempty" gorm:"type:varchar(36);index"`
	MitraID     *string         `json:"mitra_id,omitempty" gorm:"type:varchar(36);index"`
	IPAddress   string          `json:"ip_address" gorm:"type:varchar(45);not null;index"`
	UserAgent   string          `json:"user_agent" gorm:"type:text"`
	Description string          `json:"description" gorm:"type:text;not null"`
	Metadata    json.RawMessage `json:"metadata,omitempty" gorm:"type:json"`
	Resolved    bool            `json:"resolved" gorm:"default:false;index"`
	ResolvedAt  *time.Time      `json:"resolved_at,omitempty"`
	ResolvedBy  *string         `json:"resolved_by,omitempty" gorm:"type:varchar(36)"`
	CreatedAt   time.Time       `json:"created_at" gorm:"autoCreateTime;index"`
	
	// Relationships
	User  *User  `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	Mitra *Mitra `json:"mitra,omitempty" gorm:"foreignKey:MitraID;references:ID"`
}

// BusinessMetric represents business metrics for analytics
type BusinessMetric struct {
	ID         string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	MetricType string          `json:"metric_type" gorm:"type:varchar(50);not null;index"`
	MetricName string          `json:"metric_name" gorm:"type:varchar(100);not null;index"`
	Value      float64         `json:"value" gorm:"not null"`
	Unit       string          `json:"unit" gorm:"type:varchar(20)"`
	Dimensions json.RawMessage `json:"dimensions,omitempty" gorm:"type:json"`
	Timestamp  time.Time       `json:"timestamp" gorm:"not null;index"`
	CreatedAt  time.Time       `json:"created_at" gorm:"autoCreateTime"`
}

// UserActivity represents user activity tracking
type UserActivity struct {
	ID         string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID     *string         `json:"user_id,omitempty" gorm:"type:varchar(36);index"`
	MitraID    *string         `json:"mitra_id,omitempty" gorm:"type:varchar(36);index"`
	Activity   string          `json:"activity" gorm:"type:varchar(100);not null;index"`
	Category   string          `json:"category" gorm:"type:varchar(50);not null;index"`
	Details    json.RawMessage `json:"details,omitempty" gorm:"type:json"`
	IPAddress  string          `json:"ip_address" gorm:"type:varchar(45);not null"`
	UserAgent  string          `json:"user_agent" gorm:"type:text"`
	SessionID  *string         `json:"session_id,omitempty" gorm:"type:varchar(255);index"`
	CreatedAt  time.Time       `json:"created_at" gorm:"autoCreateTime;index"`
	
	// Relationships
	User  *User  `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	Mitra *Mitra `json:"mitra,omitempty" gorm:"foreignKey:MitraID;references:ID"`
}

// SystemEvent represents system-level events
type SystemEvent struct {
	ID          string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	EventType   string          `json:"event_type" gorm:"type:varchar(50);not null;index"`
	Component   string          `json:"component" gorm:"type:varchar(50);not null;index"`
	Level       string          `json:"level" gorm:"type:enum('debug','info','warn','error','fatal');not null;index"`
	Message     string          `json:"message" gorm:"type:text;not null"`
	Metadata    json.RawMessage `json:"metadata,omitempty" gorm:"type:json"`
	StackTrace  *string         `json:"stack_trace,omitempty" gorm:"type:text"`
	CreatedAt   time.Time       `json:"created_at" gorm:"autoCreateTime;index"`
}

// PerformanceMetric represents performance monitoring data
type PerformanceMetric struct {
	ID            string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	MetricType    string          `json:"metric_type" gorm:"type:varchar(50);not null;index"`
	Endpoint      string          `json:"endpoint" gorm:"type:varchar(255);not null;index"`
	Method        string          `json:"method" gorm:"type:varchar(10);not null"`
	ResponseTime  int64           `json:"response_time" gorm:"not null"` // in milliseconds
	MemoryUsage   int64           `json:"memory_usage" gorm:"default:0"` // in bytes
	CPUUsage      float64         `json:"cpu_usage" gorm:"default:0"`    // percentage
	DatabaseTime  int64           `json:"database_time" gorm:"default:0"` // in milliseconds
	RedisTime     int64           `json:"redis_time" gorm:"default:0"`    // in milliseconds
	ExternalTime  int64           `json:"external_time" gorm:"default:0"` // in milliseconds
	StatusCode    int             `json:"status_code" gorm:"not null;index"`
	RequestSize   int64           `json:"request_size" gorm:"default:0"`
	ResponseSize  int64           `json:"response_size" gorm:"default:0"`
	Metadata      json.RawMessage `json:"metadata,omitempty" gorm:"type:json"`
	Timestamp     time.Time       `json:"timestamp" gorm:"not null;index"`
	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime"`
}

// Table names
func (AuditLog) TableName() string {
	return "audit_logs"
}

func (SecurityEvent) TableName() string {
	return "security_events"
}

func (BusinessMetric) TableName() string {
	return "business_metrics"
}

func (UserActivity) TableName() string {
	return "user_activities"
}

func (SystemEvent) TableName() string {
	return "system_events"
}

func (PerformanceMetric) TableName() string {
	return "performance_metrics"
}

// Audit event types
const (
	// Authentication events
	AuditEventLogin          = "login"
	AuditEventLogout         = "logout"
	AuditEventRegister       = "register"
	AuditEventPasswordReset  = "password_reset"
	AuditEventTokenRefresh   = "token_refresh"
	
	// User management events
	AuditEventProfileUpdate  = "profile_update"
	AuditEventAvatarUpdate   = "avatar_update"
	AuditEventAddressCreate  = "address_create"
	AuditEventAddressUpdate  = "address_update"
	AuditEventAddressDelete  = "address_delete"
	AuditEventVehicleCreate  = "vehicle_create"
	AuditEventVehicleDelete  = "vehicle_delete"
	
	// Bengkel management events
	AuditEventBengkelCreate  = "bengkel_create"
	AuditEventBengkelUpdate  = "bengkel_update"
	AuditEventServiceCreate  = "service_create"
	AuditEventServiceUpdate  = "service_update"
	
	// Order events
	AuditEventOrderCreate    = "order_create"
	AuditEventOrderUpdate    = "order_update"
	AuditEventOrderCancel    = "order_cancel"
	AuditEventOrderComplete  = "order_complete"
	
	// Chat events
	AuditEventChatRoomCreate = "chat_room_create"
	AuditEventMessageSend    = "message_send"
	AuditEventMessageEdit    = "message_edit"
	AuditEventMessageDelete  = "message_delete"
	
	// System events
	AuditEventSystemStart    = "system_start"
	AuditEventSystemStop     = "system_stop"
	AuditEventDatabaseError  = "database_error"
	AuditEventRedisError     = "redis_error"
)

// Security event types
const (
	SecurityEventFailedLogin        = "failed_login"
	SecurityEventSuspiciousActivity = "suspicious_activity"
	SecurityEventRateLimitExceeded  = "rate_limit_exceeded"
	SecurityEventUnauthorizedAccess = "unauthorized_access"
	SecurityEventDataBreach         = "data_breach"
	SecurityEventMaliciousRequest   = "malicious_request"
	SecurityEventAccountLockout     = "account_lockout"
	SecurityEventPasswordBreach     = "password_breach"
)

// Business metric types
const (
	MetricTypeUser         = "user"
	MetricTypeMitra        = "mitra"
	MetricTypeBengkel      = "bengkel"
	MetricTypeOrder        = "order"
	MetricTypeRevenue      = "revenue"
	MetricTypePerformance  = "performance"
	MetricTypeEngagement   = "engagement"
	MetricTypeConversion   = "conversion"
)

// Activity categories
const (
	ActivityCategoryAuth     = "authentication"
	ActivityCategoryProfile  = "profile"
	ActivityCategoryBengkel  = "bengkel"
	ActivityCategoryOrder    = "order"
	ActivityCategoryChat     = "chat"
	ActivityCategorySearch   = "search"
	ActivityCategoryPayment  = "payment"
	ActivityCategorySystem   = "system"
)

// System event levels
const (
	SystemLevelDebug = "debug"
	SystemLevelInfo  = "info"
	SystemLevelWarn  = "warn"
	SystemLevelError = "error"
	SystemLevelFatal = "fatal"
)

// Security severity levels
const (
	SecuritySeverityLow      = "low"
	SecuritySeverityMedium   = "medium"
	SecuritySeverityHigh     = "high"
	SecuritySeverityCritical = "critical"
)
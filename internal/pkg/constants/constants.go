package constants

// User types
const (
	UserTypeUser  = "user"
	UserTypeMitra = "mitra"
)

// Order statuses
const (
	OrderStatusPending    = "pending"
	OrderStatusConfirmed  = "confirmed"
	OrderStatusInProgress = "in_progress"
	OrderStatusCompleted  = "completed"
	OrderStatusCancelled  = "cancelled"
)

// Service types
const (
	ServiceTypeHome  = "home"
	ServiceTypeStore = "store"
)

// Token types
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// File upload constants
const (
	MaxFileSize        = 10 * 1024 * 1024 // 10MB
	AllowedImageTypes  = "jpg,jpeg,png,gif"
	AllowedDocTypes    = "pdf,doc,docx"
	UploadPathAvatars  = "public/avatars"
	UploadPathVehicles = "public/vehicles"
	UploadPathBengkels = "public/bengkels"
)

// Pagination constants
const (
	DefaultPage      = 1
	DefaultLimit     = 10
	MaxLimit         = 100
	DefaultRadius    = 5.0  // 5km
	MaxRadius        = 50.0 // 50km
)

// Rating constants
const (
	MinRating = 1
	MaxRating = 5
)

// Indonesian days
var IndonesianDays = []string{
	"senin", "selasa", "rabu", "kamis", "jumat", "sabtu", "minggu",
}

// Time format
const (
	TimeFormat24Hour = "15:04"
	DateFormat       = "2006-01-02"
	DateTimeFormat   = "2006-01-02 15:04:05"
)

// Cache keys
const (
	CacheKeyUserProfile    = "user:profile:%s"
	CacheKeyMitraProfile   = "mitra:profile:%s"
	CacheKeyBengkelProfile = "bengkel:profile:%s"
	CacheKeyBengkelList    = "bengkel:list:%s"
	CacheKeyNearestBengkel = "bengkel:nearest:%f:%f:%f"
)

// Cache TTL (in seconds)
const (
	CacheTTLShort  = 300   // 5 minutes
	CacheTTLMedium = 1800  // 30 minutes
	CacheTTLLong   = 3600  // 1 hour
	CacheTTLDay    = 86400 // 24 hours
)

// HTTP headers
const (
	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"
	HeaderUserAgent     = "User-Agent"
	HeaderXForwardedFor = "X-Forwarded-For"
	HeaderXRealIP       = "X-Real-IP"
)

// Content types
const (
	ContentTypeJSON = "application/json"
	ContentTypeForm = "application/x-www-form-urlencoded"
	ContentTypeFile = "multipart/form-data"
)

// Database constraints
const (
	MaxNameLength        = 100
	MaxEmailLength       = 255
	MaxPhoneLength       = 20
	MaxAddressLength     = 500
	MaxNoteLength        = 200
	MaxTestimoniLength   = 1000
	MaxPasswordLength    = 128
	MinPasswordLength    = 8
	MaxMontirCount       = 50
	MinMontirCount       = 1
)

// Business rules
const (
	MaxAddressesPerUser  = 5
	MaxVehiclesPerUser   = 10
	MaxBengkelsPerMitra  = 1
	MaxServicesPerBengkel = 20
	MaxPhotosPerBengkel   = 10
	MaxOrdersPerDay       = 50
)

// Error codes
const (
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeAuthentication = "AUTH_ERROR"
	ErrCodeAuthorization  = "AUTHZ_ERROR"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeExternal       = "EXTERNAL_ERROR"
)

// Log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

// Environment types
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
	EnvTest        = "test"
)
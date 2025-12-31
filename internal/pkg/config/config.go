package config

import (
	"log"

	"github.com/spf13/viper"
)

var Config *Configuration

// Struct of Configuration instance.
// It include Database and Server configuration
type Configuration struct {
	App           AppConfiguration
	Database      DatabaseConfiguration
	Database_Test DatabaseTestConfiguration
	Cloudinary    StorageCloudinary
	Server        ServerConnection
	GoogleAuth    GoogleAuthConfiguration
	Midtrans      MidtransConfiguration
	SMTP          SMTPConfiguration
	Supabase      SupabaseConfiguration
	Agore         AgoreConfiguration
	RateLimit     RateLimitConfiguration
	Redis         RedisConfiguration
}

// Struct of App Configuration instance
type AppConfiguration struct {
	Name        string `mapstructure:"APP_NAME"`
	Version     string `mapstructure:"APP_VERSION"`
	Environment string `mapstructure:"APP_ENVIRONMENT"`
}

// Struct of Database Configuration instance.
type DatabaseConfiguration struct {
	Driver       string `mapstructure:"DATABASE_DRIVER"`
	Dbname       string `mapstructure:"DATABASE_NAME"`
	Username     string `mapstructure:"DATABASE_USERNAME"`
	Password     string `mapstructure:"DATABASE_PASSWORD"`
	Host         string `mapstructure:"DATABASE_HOST"`
	Port         string `mapstructure:"DATABASE_PORT"`
	MaxLifetime  int    `mapstructure:"DATABASE_MAX_LIFETIME"`
	MaxOpenConns int    `mapstructure:"DATABASE_MAX_OPEN_CONNS"`
	MaxIdleConns int    `mapstructure:"DATABASE_MAX_IDLE_CONNS"`
}

type SupabaseConfiguration struct {
	SupabaseUrl string `mapstructure:"SUPABASE_URL"`
	SupabaseKey string `mapstructure:"SUPABASE_KEY"`
}

type MidtransConfiguration struct {
	MidtransServerKey string `mapstructure:"MIDTRANS_SERVER_KEY"`
}

type GoogleAuthConfiguration struct {
	ClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	ClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
}

// Struct of Database for Testing Configuration instance
type DatabaseTestConfiguration struct {
	Driver       string `mapstructure:"DATABASE_TEST_DRIVER"`
	Dbname       string `mapstructure:"DATABASE_TEST_NAME"`
	Username     string `mapstructure:"DATABASE_TEST_USERNAME"`
	Password     string `mapstructure:"DATABASE_TEST_PASSWORD"`
	Host         string `mapstructure:"DATABASE_TEST_HOST"`
	Port         string `mapstructure:"DATABASE_TEST_PORT"`
	MaxLifetime  int    `mapstructure:"DATABASE_TEST_MAX_LIFETIME"`
	MaxOpenConns int    `mapstructure:"DATABASE_TEST_MAX_OPEN_CONNS"`
	MaxIdleConns int    `mapstructure:"DATABASE_TEST_MAX_IDLE_CONNS"`
}

type SMTPConfiguration struct {
	Host     string `mapstructure:"SMTP_HOST"`
	Port     string `mapstructure:"SMTP_PORT"`
	Username string `mapstructure:"SMTP_USERNAME"`
	Password string `mapstructure:"SMTP_PASSWORD"`
}

// Struct of Cloudinary Storage Configuration instance
type StorageCloudinary struct {
	CloudName    string `mapstructure:"CLOUDINARY_CLOUD_NAME"`
	ApiKey       string `mapstructure:"CLOUDINARY_API_KEY"`
	ApiSecret    string `mapstructure:"CLOUDINARY_API_SECRET"`
	UploadFolder string `mapstructure:"CLOUDINARY_UPLOAD_FOLDER"`
}

// Struct of Server Configuration instance.
type ServerConnection struct {
	Host                 string `mapstructure:"SERVER_HOST"`
	Port                 string `mapstructure:"SERVER_PORT"`
	Secret               string `mapstructure:"SERVER_SECRET"`
	Secret2              string `mapstructure:"SERVER_SECRET2"`
	ApiSecret            string `mapstructure:"API_SECRET"`
	DevMode              string `mapstructure:"SERVER_DEV_MODE"`
	Mode                 string `mapstructure:"SERVER_MODE"`
	Name                 string `mapstructure:"SERVER_NAME"`
	ExpiresHour          int64  `mapstructure:"SERVER_EXPIRES_HOUR"`
	RefreshSecret        string `mapstructure:"SERVER_REFRESH_SECRET"`
	RefreshSecret2       string `mapstructure:"SERVER_REFRESH_SECRET2"`
	RefreshExpiresHour   int64  `mapstructure:"SERVER_REFRESH_EXPIRES_HOUR"`
}

// Struct for Rate Limiting Configuration
type RateLimitConfiguration struct {
	GeneralRPS    float64 `mapstructure:"RATE_LIMIT_GENERAL_RPS"`
	GeneralBurst  int     `mapstructure:"RATE_LIMIT_GENERAL_BURST"`
	AuthRPS       float64 `mapstructure:"RATE_LIMIT_AUTH_RPS"`
	AuthBurst     int     `mapstructure:"RATE_LIMIT_AUTH_BURST"`
	StrictRPS     float64 `mapstructure:"RATE_LIMIT_STRICT_RPS"`
	StrictBurst   int     `mapstructure:"RATE_LIMIT_STRICT_BURST"`
	Enabled       bool    `mapstructure:"RATE_LIMIT_ENABLED"`
}

// Struct for Redis Configuration
type RedisConfiguration struct {
	URL      string `mapstructure:"REDIS_URL"`      // Full Redis URL (redis://user:pass@host:port/db)
	Host     string `mapstructure:"REDIS_HOST"`     // Individual host (fallback)
	Port     string `mapstructure:"REDIS_PORT"`     // Individual port (fallback)
	Password string `mapstructure:"REDIS_PASSWORD"` // Individual password (fallback)
	DB       int    `mapstructure:"REDIS_DB"`       // Individual DB (fallback)
	Enabled  bool   `mapstructure:"REDIS_ENABLED"`
}

type AgoreConfiguration struct {
	AppID          string `mapstructure:"AGORA_APP_ID"`
	AppCertificate string `mapstructure:"AGORA_APP_CERTIFICATE"`
	ExpiryTime     string `mapstructure:"AGORA_EXPIRY_TIME"`
}

// Setup the configuration
func Setup(configPath string) {
	var (
		appConfiguration          AppConfiguration
		databaseConfiguration     DatabaseConfiguration
		databaseTestConfiguration DatabaseTestConfiguration
		cloudinaryConfiguration   StorageCloudinary
		serverConfiguration       ServerConnection
		googleAuthConfiguration   GoogleAuthConfiguration
		midtransConfiguration     MidtransConfiguration
		smtpConfiguration         SMTPConfiguration
		supabaseConfiguration     SupabaseConfiguration
		agoreConfiguration        AgoreConfiguration
		rateLimitConfiguration    RateLimitConfiguration
		redisConfiguration        RedisConfiguration
	)

	viper.SetConfigFile(configPath)
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	unmarshalConfiguration(&appConfiguration)
	unmarshalConfiguration(&databaseConfiguration)
	unmarshalConfiguration(&databaseTestConfiguration)
	unmarshalConfiguration(&cloudinaryConfiguration)
	unmarshalConfiguration(&serverConfiguration)
	unmarshalConfiguration(&googleAuthConfiguration)
	unmarshalConfiguration(&midtransConfiguration)
	unmarshalConfiguration(&smtpConfiguration)
	unmarshalConfiguration(&supabaseConfiguration)
	unmarshalConfiguration(&agoreConfiguration)
	unmarshalConfiguration(&rateLimitConfiguration)
	unmarshalConfiguration(&redisConfiguration)

	// Set default values for app configuration if not set
	if appConfiguration.Name == "" {
		appConfiguration.Name = "Bengkelin API"
	}
	if appConfiguration.Version == "" {
		appConfiguration.Version = "1.0.0"
	}
	if appConfiguration.Environment == "" {
		appConfiguration.Environment = "development"
	}

	// Set default values for Redis configuration if not set
	if redisConfiguration.Host == "" {
		redisConfiguration.Host = "localhost"
	}
	if redisConfiguration.Port == "" {
		redisConfiguration.Port = "6379"
	}
	if redisConfiguration.DB == 0 {
		redisConfiguration.DB = 0
	}
	// Redis is enabled by default
	redisConfiguration.Enabled = true

	// Set default values for rate limiting if not configured
	if rateLimitConfiguration.GeneralRPS == 0 {
		rateLimitConfiguration.GeneralRPS = 100.0 / 60.0 // 100 requests per minute
	}
	if rateLimitConfiguration.GeneralBurst == 0 {
		rateLimitConfiguration.GeneralBurst = 10
	}
	if rateLimitConfiguration.AuthRPS == 0 {
		rateLimitConfiguration.AuthRPS = 10.0 / 60.0 // 10 requests per minute
	}
	if rateLimitConfiguration.AuthBurst == 0 {
		rateLimitConfiguration.AuthBurst = 5
	}
	if rateLimitConfiguration.StrictRPS == 0 {
		rateLimitConfiguration.StrictRPS = 5.0 / 60.0 // 5 requests per minute
	}
	if rateLimitConfiguration.StrictBurst == 0 {
		rateLimitConfiguration.StrictBurst = 2
	}
	// Rate limiting is enabled by default
	rateLimitConfiguration.Enabled = true

	configuration := Configuration{
		App:           appConfiguration,
		Database:      databaseConfiguration,
		Database_Test: databaseTestConfiguration,
		Cloudinary:    cloudinaryConfiguration,
		Server:        serverConfiguration,
		GoogleAuth:    googleAuthConfiguration,
		Midtrans:      midtransConfiguration,
		SMTP:          smtpConfiguration,
		Supabase:      supabaseConfiguration,
		Agore:         agoreConfiguration,
		RateLimit:     rateLimitConfiguration,
		Redis:         redisConfiguration,
	}

	Config = &configuration
}

// Helper to unmarshal
func unmarshalConfiguration(configuration interface{}) {
	err := viper.Unmarshal(configuration)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
}

// GetConfig return the configuration instance
func GetConfig() *Configuration {
	return Config
}

package config

import (
	"log"

	"github.com/spf13/viper"
)

var Config *Configuration

// Struct of Configuration instance.
// It include Database and Server configuration
type Configuration struct {
	Database      DatabaseConfiguration
	Database_Test DatabaseTestConfiguration
	Cloudinary    StorageCloudinary
	Server        ServerConnection
	GoogleAuth    GoogleAuthConfiguration
	Midtrans      MidtransConfiguration
	SMTP          SMTPConfiguration
	Supabase      SupabaseConfiguration
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
	Port        string `mapstructure:"SERVER_PORT"`
	Secret      string `mapstructure:"SERVER_SECRET"`
	Secret2     string `mapstructure:"SERVER_SECRET2"`
	Mode        string `mapstructure:"SERVER_MODE"`
	Name        string `mapstructure:"SERVER_NAME"`
	ExpiresHour int64  `mapstructure:"SERVER_EXPIRES_HOUR"`
}

// Setup the configuration
func Setup(configPath string) {
	var (
		databaseConfiguration     DatabaseConfiguration
		databaseTestConfiguration DatabaseTestConfiguration
		cloudinaryConfiguration   StorageCloudinary
		serverConfiguration       ServerConnection
		googleAuthConfiguration   GoogleAuthConfiguration
		midtransConfiguration     MidtransConfiguration
		smtpConfiguration         SMTPConfiguration
		supabaseConfiguration     SupabaseConfiguration
	)

	viper.SetConfigFile(configPath)
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	unmarshalConfiguration(&databaseConfiguration)
	unmarshalConfiguration(&databaseTestConfiguration)
	unmarshalConfiguration(&cloudinaryConfiguration)
	unmarshalConfiguration(&serverConfiguration)
	unmarshalConfiguration(&googleAuthConfiguration)
	unmarshalConfiguration(&midtransConfiguration)
	unmarshalConfiguration(&smtpConfiguration)
	unmarshalConfiguration(&supabaseConfiguration)

	configuration := Configuration{
		Database:      databaseConfiguration,
		Database_Test: databaseTestConfiguration,
		Cloudinary:    cloudinaryConfiguration,
		Server:        serverConfiguration,
		GoogleAuth:    googleAuthConfiguration,
		Midtrans:      midtransConfiguration,
		SMTP:          smtpConfiguration,
		Supabase:      supabaseConfiguration,
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

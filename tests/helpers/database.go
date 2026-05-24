package helpers

import (
	"fmt"
	"os"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var TestDB *gorm.DB

// SetupTestDB initializes a test database connection
func SetupTestDB() *gorm.DB {
	if TestDB != nil {
		return TestDB
	}

	// Get test database configuration
	testConfig := getTestDBConfig()
	
	// Create database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		testConfig.Host,
		testConfig.User,
		testConfig.Password,
		testConfig.Name,
		testConfig.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for tests
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	// Auto migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.UserAddress{},
		&models.Vehicle{},
		&models.Mitra{},
		&models.Bengkel{},
		&models.BengkelAddress{},
		&models.BengkelService{},
		&models.BengkelOperational{},
		&models.BengkelPhoto{},
		&models.BengkelTestimonial{},
		&models.Order{},
		&models.OrderService{},
		&models.AdminFee{},
		&models.ChatRoom{},
		&models.ChatMessage{},
		&models.RefreshToken{},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to migrate test database: %v", err))
	}

	TestDB = db
	return TestDB
}

// CleanupTestDB cleans up test database after tests
func CleanupTestDB() {
	if TestDB == nil {
		return
	}

	// Clean up all tables in reverse order to handle foreign keys
	tables := []interface{}{
		&models.OrderService{},
		&models.Order{},
		&models.BengkelTestimonial{},
		&models.BengkelPhoto{},
		&models.BengkelService{},
		&models.BengkelOperational{},
		&models.BengkelAddress{},
		&models.Bengkel{},
		&models.UserAddress{},
		&models.Vehicle{},
		&models.User{},
		&models.Mitra{},
		&models.ChatMessage{},
		&models.ChatRoom{},
		&models.AdminFee{},
		&models.RefreshToken{},
	}

	for _, table := range tables {
		if err := TestDB.Unscoped().Where("1 = 1").Delete(table).Error; err != nil {
			// Ignore errors during cleanup
		}
	}
}

// DatabaseConfig represents database configuration for tests
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// getTestDBConfig returns test database configuration
func getTestDBConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_DB_PORT", "5432"),
		User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "password"),
		Name:     getEnvOrDefault("TEST_DB_NAME", "bengkelin_test"),
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GenerateTestID generates a test UUID
func GenerateTestID() string {
	return helpers.GenerateUUID()
}
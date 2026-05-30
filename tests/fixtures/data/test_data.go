package data

import (
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/models"
)

// TestUsers provides sample user data for testing
var TestUsers = []models.User{
	{
		ID:          "test-user-1",
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john.doe@example.com",
		PhoneNumber: "081234567890",
		Password:    "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		AvatarUrl:   "https://example.com/avatar1.jpg",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-1 * time.Hour),
	},
	{
		ID:          "test-user-2",
		FirstName:   "Jane",
		LastName:    "Smith",
		Email:       "jane.smith@example.com",
		PhoneNumber: "081234567891",
		Password:    "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		AvatarUrl:   "https://example.com/avatar2.jpg",
		CreatedAt:   time.Now().Add(-48 * time.Hour),
		UpdatedAt:   time.Now().Add(-2 * time.Hour),
	},
}

// TestMitras provides sample mitra data for testing
var TestMitras = []models.Mitra{
	{
		ID:          "test-mitra-1",
		FirstName:   "Workshop",
		LastName:    "Owner 1",
		Email:       "owner1@workshop.com",
		PhoneNumber: "081234567892",
		Password:    "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		BankName:    "Bank BCA",
		BankNumber:  "1234567890",
		CreatedAt:   time.Now().Add(-72 * time.Hour),
		UpdatedAt:   time.Now().Add(-3 * time.Hour),
	},
	{
		ID:          "test-mitra-2",
		FirstName:   "Workshop",
		LastName:    "Owner 2",
		Email:       "owner2@workshop.com",
		PhoneNumber: "081234567893",
		Password:    "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		BankName:    "Bank Mandiri",
		BankNumber:  "0987654321",
		CreatedAt:   time.Now().Add(-96 * time.Hour),
		UpdatedAt:   time.Now().Add(-4 * time.Hour),
	},
}

// TestBengkels provides sample bengkel data for testing
var TestBengkels = []models.Bengkel{
	{
		ID:           "test-bengkel-1",
		MitraID:      "test-mitra-1",
		BengkelName:  "Auto Repair Shop 1",
		BengkelPhone: "081234567894",
		AvatarUrl:    "https://example.com/bengkel1.jpg",
		CreatedAt:    time.Now().Add(-48 * time.Hour),
		UpdatedAt:    time.Now().Add(-1 * time.Hour),
	},
	{
		ID:           "test-bengkel-2",
		MitraID:      "test-mitra-2",
		BengkelName:  "Quick Fix Garage",
		BengkelPhone: "081234567895",
		AvatarUrl:    "https://example.com/bengkel2.jpg",
		CreatedAt:    time.Now().Add(-72 * time.Hour),
		UpdatedAt:    time.Now().Add(-2 * time.Hour),
	},
}

// TestVehicles provides sample vehicle data for testing
var TestVehicles = []models.Vehicle{
	{
		ID:            1,
		UserID:        "test-user-1",
		VehicleType:   "Motor",
		VehicleColor:  "Red",
		VehicleNumber: "B 1234 ABC",
	},
	{
		ID:            2,
		UserID:        "test-user-2",
		VehicleType:   "Mobil",
		VehicleColor:  "Blue",
		VehicleNumber: "B 5678 DEF",
	},
}

// TestUserAddresses provides sample user address data for testing
var TestUserAddresses = []models.UserAddress{
	{
		ID:           1,
		UserID:       "test-user-1",
		AddressLabel: "Home",
		FullAddress:  "Jl. Sudirman No. 123, Jakarta, DKI Jakarta 12345",
		Latitude:     -6.2088,
		Longitude:    106.8456,
		Note:         "Near the mall",
		IsPrimary:    boolPtr(true),
	},
	{
		ID:           2,
		UserID:       "test-user-2",
		AddressLabel: "Office",
		FullAddress:  "Jl. Thamrin No. 456, Jakarta, DKI Jakarta 67890",
		Latitude:     -6.1751,
		Longitude:    106.8650,
		Note:         "Near the bus stop",
		IsPrimary:    boolPtr(false),
	},
}

// TestOrders provides sample order data for testing
var TestOrders = []models.Order{
	{
		ID:            "test-order-1",
		UserID:        "test-user-1",
		BengkelID:     "test-bengkel-1",
		VehicleID:     1,
		Status:        models.OrderStatusPending,
		TotalPrice:    150000.00,
		AdminFee:      5000.00,
		IsHomeService: boolPtr(false),
		CreatedAt:     time.Now().Add(-2 * time.Hour),
		UpdatedAt:     time.Now(),
	},
	{
		ID:             "test-order-2",
		UserID:         "test-user-2",
		BengkelID:      "test-bengkel-2",
		VehicleID:      2,
		Status:         models.OrderStatusConfirmed,
		TotalPrice:     250000.00,
		AdminFee:       5000.00,
		IsHomeService:  boolPtr(true),
		HomeServiceFee: 25000.00,
		CreatedAt:      time.Now().Add(-24 * time.Hour),
		UpdatedAt:      time.Now().Add(-23 * time.Hour),
	},
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

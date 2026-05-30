package service

import (
	"context"

	"github.com/Bengkelin/bengkelin-service/internal/dto"
	"github.com/Bengkelin/bengkelin-service/internal/models"
	"github.com/Bengkelin/bengkelin-service/internal/repository"
	"github.com/Bengkelin/bengkelin-service/internal/crypto"
)

// AuthServiceInterface defines authentication service contract
type AuthServiceInterface interface {
	// User authentication
	LoginUser(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	RegisterUser(ctx context.Context, req dto.RegisterUserRequest) (*dto.AuthResponse, error)
	LoginUserWithGoogle(ctx context.Context, req dto.GoogleAuthRequest) (*dto.AuthResponse, error)

	// Mitra authentication
	LoginMitra(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	RegisterMitra(ctx context.Context, req dto.RegisterMitraRequest) (*dto.AuthResponse, error)
	LoginMitraWithGoogle(ctx context.Context, req dto.GoogleAuthRequest) (*dto.AuthResponse, error)

	// Token management
	RefreshUserToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	RefreshMitraToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	LogoutUser(ctx context.Context, refreshToken string) error
	LogoutMitra(ctx context.Context, refreshToken string) error
	LogoutAllUserDevices(ctx context.Context, userID string) error
	LogoutAllMitraDevices(ctx context.Context, mitraID string) error
}

// UserServiceInterface defines user service contract
type UserServiceInterface interface {
	GetUserProfile(ctx context.Context, userID string) (*dto.UserProfileResponse, error)
	UpdateUserProfile(ctx context.Context, userID string, req dto.UpdateUserRequest) (*dto.UserProfileResponse, error)
	UpdateUserAvatar(ctx context.Context, userID string, avatarURL string) error

	// Address management
	AddUserAddress(ctx context.Context, userID string, req dto.AddressRequest) (*dto.AddressResponse, error)
	UpdateAddress(ctx context.Context, userID string, addressID uint, req dto.AddressRequest) (*dto.AddressResponse, error)
	CreateOrUpdateAddress(ctx context.Context, userID string, req dto.AddressRequest) (*dto.AddressResponse, error)
	GetUserAddress(ctx context.Context, userID, addressID string) (*dto.AddressResponse, error)
	DeleteUserAddress(ctx context.Context, userID, addressID string) error

	// Vehicle management
	AddUserVehicle(ctx context.Context, userID string, req dto.VehicleRequest) (*dto.VehicleResponse, error)
	GetUserVehicle(ctx context.Context, userID, vehicleID string) (*dto.VehicleResponse, error)
	GetAllUserVehicles(ctx context.Context, userID string) ([]dto.VehicleResponse, error)
	DeleteUserVehicle(ctx context.Context, userID, vehicleID string) error
}

// BengkelServiceInterface defines bengkel service contract
type BengkelServiceInterface interface {
	CreateBengkel(ctx context.Context, mitraID string, req dto.CreateBengkelRequest) (*dto.BengkelResponse, error)
	GetBengkelProfile(ctx context.Context, mitraID string) (*dto.BengkelResponse, error)
	UpdateBengkelProfile(ctx context.Context, mitraID string, req dto.UpdateBengkelRequest) (*dto.BengkelResponse, error)
	UpdateBengkelMontir(ctx context.Context, mitraID string, jumlahMontir uint) error
	UpdateBengkelOperational(ctx context.Context, mitraID string, operationals []dto.OperationalItemRequest) error
	CreateBengkelAddress(ctx context.Context, mitraID string, req dto.BengkelAddressRequest) error
	CreateBengkelServices(ctx context.Context, mitraID string, services []dto.BengkelServiceItemRequest) error
	UpdateBengkelServices(ctx context.Context, mitraID string, services []dto.BengkelServiceItemRequest) error
	CreateBengkelPhotos(ctx context.Context, mitraID string, photoURLs []string) error
	GetBengkelPhotoURL(ctx context.Context, photoID string) (string, error)
	DeleteBengkelPhoto(ctx context.Context, mitraID string, photoID string) error
	UpdateBengkelStatusOptions(ctx context.Context, mitraID string, homeService, storeService, isOpen bool) error
	UpdateBengkelAvatar(ctx context.Context, mitraID string, avatarURL string) error
	SearchBengkels(ctx context.Context, req dto.SearchBengkelRequest) (*dto.PaginatedBengkelResponse, error)
	GetNearestBengkels(ctx context.Context, req dto.NearestBengkelRequest) (*dto.PaginatedBengkelResponse, error)
	CreateTestimonial(ctx context.Context, userID, bengkelID, orderID string, testimoni string, rating int) error

	// Query methods for handlers
	GetMitraWithBengkel(ctx context.Context, mitraID string) (*models.Mitra, error)
	GetAllBengkels(ctx context.Context) ([]dto.BengkelResponse, error)
	GetAllBengkelsPaginate(ctx context.Context, page, limit int) ([]dto.BengkelResponse, int, error)
	SearchBengkelsV2(ctx context.Context, serviceType, query string, page, limit int) ([]dto.BengkelResponse, int, error)
	GetBengkelDetailWithTestimonials(ctx context.Context, bengkelID string, page, limit int) (*models.Bengkel, []models.BengkelTestimonial, int, error)
	GetBengkelFullDetail(ctx context.Context, bengkelID string, forceFresh bool, page, limit int) (*models.Bengkel, []models.BengkelTestimonial, int, error)
	GetBengkelOperationalTimeSlots(ctx context.Context, bengkelID, day string) (map[int]string, error)
	SearchBengkelsPublic(ctx context.Context, criteria map[string]interface{}) ([]models.Bengkel, int, error)
}

// OrderServiceInterface defines order service contract
type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, mitraID, userID string, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	CreateOrderWithServices(ctx context.Context, userID, mitraID string, req dto.CreateOrderWithServicesRequest) (*dto.CreateOrderResult, error)
	GetOrderDetails(ctx context.Context, orderID string, userType string) (*dto.OrderDetailResponse, error)
	UpdateOrderStatus(ctx context.Context, mitraID, orderID string, status string) error
	UpdateOrderStatusWithValidation(ctx context.Context, mitraID, orderID string, newStatus uint, reason string) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID string, req dto.PaginationRequest) (*dto.PaginatedOrderResponse, error)
	GetMitraOrders(ctx context.Context, mitraID string, req dto.PaginationRequest) (*dto.PaginatedOrderResponse, error)

	// Query methods for handlers
	GetOrderForUser(ctx context.Context, orderID, userID string) (*models.Order, error)
	GetOrderForMitra(ctx context.Context, orderID, mitraID string) (*models.Order, error)
	GetUserOrdersPaginate(ctx context.Context, userID string, page, limit int) ([]models.Order, int, error)
	GetMitraOrdersPaginate(ctx context.Context, mitraID string, page, limit int) ([]models.Order, int, error)
	UpdateOrderDetails(ctx context.Context, orderID, userID string, isHomeService *bool, homeServiceSchedule, paymentMethod string) error
	ValidateUserExists(ctx context.Context, userID string) error
	ValidateMitraExists(ctx context.Context, mitraID string) error
}

// ChatServiceInterface defines chat service contract
type ChatServiceInterface interface {
	CreateUserAppToken(ctx context.Context, userID string) (*dto.ChatTokenResponse, error)
	CreateMitraAppToken(ctx context.Context, mitraID string) (*dto.ChatTokenResponse, error)
	CreateUserChatToken(ctx context.Context, userID string) (*dto.ChatTokenResponse, error)
	CreateMitraChatToken(ctx context.Context, mitraID string) (*dto.ChatTokenResponse, error)
	SaveChatHistory(ctx context.Context, req dto.ChatHistoryRequest) error
	GetChatHistory(ctx context.Context, userID string, userType string) ([]dto.ChatHistoryResponse, error)
	GenerateRtmParams(ctx context.Context, userID string) (string, uint32, error)
	ValidateUserExists(ctx context.Context, userID string) error
	ValidateMitraExists(ctx context.Context, mitraID string) error
	GetChatHistoryPaginate(ctx context.Context, page, limit int, senderID, receiverID string) ([]models.ChatHistory, int, error)
}

// MitraServiceInterface defines mitra service contract
type MitraServiceInterface interface {
	GetMitraProfile(ctx context.Context, mitraID string) (*models.Mitra, error)
	UpdateMitraProfile(ctx context.Context, mitraID string, req dto.UpdateMitraRequest) (*models.Mitra, error)
	CreateMitraBank(ctx context.Context, mitraID string, req dto.MitraBankRequest) (*dto.MitraBankResponse, error)
	UpdateMitraBank(ctx context.Context, mitraID string, req dto.MitraBankRequest) (*dto.MitraBankResponse, error)
}

// AdminFeeServiceInterface defines admin fee service contract
type AdminFeeServiceInterface interface {
	CreateAdminFee(ctx context.Context, adminFee float64) (*models.AdminFee, error)
	GetLatestAdminFee(ctx context.Context) (*models.AdminFee, error)
}

// ServiceContainer holds all service instances
type ServiceContainer struct {
	AuthService     AuthServiceInterface
	UserService     UserServiceInterface
	BengkelService  BengkelServiceInterface
	OrderService    OrderServiceInterface
	ChatService     ChatServiceInterface
	MitraService    MitraServiceInterface
	AdminFeeService AdminFeeServiceInterface
}

// ServiceDependencies holds all dependencies needed by services
type ServiceDependencies struct {
	// Core repositories
	UserRepo         repository.UserRepositoryInterface
	MitraRepo        repository.MitraRepositoryInterface
	BengkelRepo      repository.BengkelRepositoryInterface
	AddressRepo      repository.AddressRepositoryInterface
	VehicleRepo      repository.VehicleRepositoryInterface
	RefreshTokenRepo repository.RefreshTokenRepositoryInterface

	// Bengkel sub-repositories
	BengkelAddressRepo      repository.BengkelAddressRepositoryInterface
	BengkelOperationalRepo  repository.BengkelOperationalRepositoryInterface
	BengkelServiceRepo      repository.BengkelServiceRepositoryInterface
	BengkelPhotoRepo        repository.BengkelPhotoRepositoryInterface
	BengkelTestimonialRepo  repository.BengkelTestimonialRepositoryInterface

	// Order repositories
	OrderRepo       repository.OrderRepositoryInterface
	OrderServiceRepo repository.OrderServiceRepositoryInterface

	// Chat repositories
	ChatHistoryRepo repository.ChatHistoryRepositoryInterface

	// Admin repositories
	AdminFeeRepo repository.AdminFeeRepositoryInterface

	// Helpers
	JWTHelper      crypto.JWTCryptoHelper
	PasswordHelper crypto.PasswordCryptoHelper
}

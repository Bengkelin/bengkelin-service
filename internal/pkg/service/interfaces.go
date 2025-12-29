package service

import (
	"context"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/dto"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/pkg/crypto"
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
	GetUserAddress(ctx context.Context, userID, addressID string) (*dto.AddressResponse, error)
	DeleteUserAddress(ctx context.Context, userID, addressID string) error
	
	// Vehicle management
	AddUserVehicle(ctx context.Context, userID string, req dto.VehicleRequest) (*dto.VehicleResponse, error)
	GetUserVehicle(ctx context.Context, userID, vehicleID string) (*dto.VehicleResponse, error)
	DeleteUserVehicle(ctx context.Context, userID, vehicleID string) error
}

// BengkelServiceInterface defines bengkel service contract
type BengkelServiceInterface interface {
	CreateBengkel(ctx context.Context, mitraID string, req dto.CreateBengkelRequest) (*dto.BengkelResponse, error)
	GetBengkelProfile(ctx context.Context, mitraID string) (*dto.BengkelResponse, error)
	UpdateBengkelProfile(ctx context.Context, mitraID string, req dto.UpdateBengkelRequest) (*dto.BengkelResponse, error)
	SearchBengkels(ctx context.Context, req dto.SearchBengkelRequest) (*dto.PaginatedBengkelResponse, error)
	GetNearestBengkels(ctx context.Context, req dto.NearestBengkelRequest) (*dto.PaginatedBengkelResponse, error)
}

// OrderServiceInterface defines order service contract
type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, mitraID, userID string, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetOrderDetails(ctx context.Context, orderID string, userType string) (*dto.OrderDetailResponse, error)
	UpdateOrderStatus(ctx context.Context, mitraID, orderID string, status string) error
	GetUserOrders(ctx context.Context, userID string, req dto.PaginationRequest) (*dto.PaginatedOrderResponse, error)
	GetMitraOrders(ctx context.Context, mitraID string, req dto.PaginationRequest) (*dto.PaginatedOrderResponse, error)
}

// ChatServiceInterface defines chat service contract
type ChatServiceInterface interface {
	CreateUserAppToken(ctx context.Context, userID string) (*dto.ChatTokenResponse, error)
	CreateMitraAppToken(ctx context.Context, mitraID string) (*dto.ChatTokenResponse, error)
	CreateUserChatToken(ctx context.Context, userID string) (*dto.ChatTokenResponse, error)
	CreateMitraChatToken(ctx context.Context, mitraID string) (*dto.ChatTokenResponse, error)
	SaveChatHistory(ctx context.Context, req dto.ChatHistoryRequest) error
	GetChatHistory(ctx context.Context, userID string, userType string) ([]dto.ChatHistoryResponse, error)
}

// ServiceContainer holds all service instances
type ServiceContainer struct {
	AuthService    AuthServiceInterface
	UserService    UserServiceInterface
	BengkelService BengkelServiceInterface
	OrderService   OrderServiceInterface
	ChatService    ChatServiceInterface
}

// ServiceDependencies holds all dependencies needed by services
type ServiceDependencies struct {
	UserRepo         repository.UserRepositoryInterface
	MitraRepo        repository.MitraRepositoryInterface
	BengkelRepo      repository.BengkelRepositoryInterface
	AddressRepo      repository.AddressRepositoryInterface
	VehicleRepo      repository.VehicleRepositoryInterface
	RefreshTokenRepo repository.RefreshTokenRepositoryInterface
	JWTHelper        crypto.JWTCryptoHelper
	PasswordHelper   crypto.PasswordCryptoHelper
}
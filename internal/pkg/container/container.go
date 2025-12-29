package container

import (
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/service"
	"github.com/Bengkelin/bengkelin-service/pkg/crypto"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
)

// Container holds all application dependencies
type Container struct {
	// Services
	AuthService    service.AuthServiceInterface
	UserService    service.UserServiceInterface
	BengkelService service.BengkelServiceInterface
	OrderService   service.OrderServiceInterface
	ChatService    service.ChatServiceInterface
	
	// Repositories
	UserRepo         repository.UserRepositoryInterface
	MitraRepo        repository.MitraRepositoryInterface
	BengkelRepo      repository.BengkelRepositoryInterface
	AddressRepo      repository.AddressRepositoryInterface
	VehicleRepo      repository.VehicleRepositoryInterface
	RefreshTokenRepo repository.RefreshTokenRepositoryInterface
	
	// Helpers
	JWTHelper      crypto.JWTCryptoHelper
	PasswordHelper crypto.PasswordCryptoHelper
}

var (
	container *Container
	once      sync.Once
)

// GetContainer returns the singleton container instance
func GetContainer() *Container {
	once.Do(func() {
		container = initializeContainer()
	})
	return container
}

// initializeContainer initializes all dependencies
func initializeContainer() *Container {
	applog.Info("Initializing dependency container")
	
	// Initialize helpers
	jwtHelper := crypto.GetJWTCrypto()
	passwordHelper := crypto.GetPasswordCryptoHelper()
	
	// Initialize repositories
	userRepo := repository.GetUserRepository()
	mitraRepo := repository.GetMitraRepository()
	bengkelRepo := repository.GetBengkelRepository()
	addressRepo := repository.GetAddressRepository()
	vehicleRepo := repository.GetVehicleRepository()
	refreshTokenRepo := repository.GetRefreshTokenRepository()
	
	// TODO: Initialize missing repositories
	// orderRepo := repository.GetOrderRepository()
	// chatRepo := repository.GetChatRepository()
	
	// Create service dependencies
	serviceDeps := service.ServiceDependencies{
		UserRepo:         userRepo,
		MitraRepo:        mitraRepo,
		BengkelRepo:      bengkelRepo,
		AddressRepo:      addressRepo,
		VehicleRepo:      vehicleRepo,
		RefreshTokenRepo: refreshTokenRepo,
		JWTHelper:        jwtHelper,
		PasswordHelper:   passwordHelper,
		// OrderRepo:        orderRepo,
		// ChatRepo:         chatRepo,
	}
	
	// Initialize services
	authService := service.NewAuthService(serviceDeps)
	// userService := service.NewUserService(serviceDeps)
	// bengkelService := service.NewBengkelService(serviceDeps)
	// orderService := service.NewOrderService(serviceDeps)
	// chatService := service.NewChatService(serviceDeps)
	
	applog.Info("Dependency container initialized successfully")
	
	return &Container{
		// Services
		AuthService: authService,
		// UserService:    userService,
		// BengkelService: bengkelService,
		// OrderService:   orderService,
		// ChatService:    chatService,
		
		// Repositories
		UserRepo:         userRepo,
		MitraRepo:        mitraRepo,
		BengkelRepo:      bengkelRepo,
		AddressRepo:      addressRepo,
		VehicleRepo:      vehicleRepo,
		RefreshTokenRepo: refreshTokenRepo,
		// OrderRepo:        orderRepo,
		// ChatRepo:         chatRepo,
		
		// Helpers
		JWTHelper:      jwtHelper,
		PasswordHelper: passwordHelper,
	}
}

// Reset resets the container (useful for testing)
func Reset() {
	container = nil
	once = sync.Once{}
}
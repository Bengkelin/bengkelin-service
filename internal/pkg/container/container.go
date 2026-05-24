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
	MitraService   service.MitraServiceInterface
	AdminFeeService service.AdminFeeServiceInterface
	CleanupService  service.CleanupServiceInterface

	// Repositories
	UserRepo                repository.UserRepositoryInterface
	MitraRepo               repository.MitraRepositoryInterface
	BengkelRepo             repository.BengkelRepositoryInterface
	AddressRepo             repository.AddressRepositoryInterface
	VehicleRepo             repository.VehicleRepositoryInterface
	RefreshTokenRepo        repository.RefreshTokenRepositoryInterface
	BengkelAddressRepo      repository.BengkelAddressRepositoryInterface
	BengkelOperationalRepo  repository.BengkelOperationalRepositoryInterface
	BengkelServiceRepo      repository.BengkelServiceRepositoryInterface
	BengkelPhotoRepo        repository.BengkelPhotoRepositoryInterface
	BengkelTestimonialRepo  repository.BengkelTestimonialRepositoryInterface
	OrderRepo               repository.OrderRepositoryInterface
	OrderServiceRepo        repository.OrderServiceRepositoryInterface
	ChatHistoryRepo         repository.ChatHistoryRepositoryInterface
	AdminFeeRepo            repository.AdminFeeRepositoryInterface

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
	bengkelAddressRepo := repository.GetBengkelAddressRepository()
	bengkelOperationalRepo := repository.GetBengkelOperationalRepository()
	bengkelServiceRepo := repository.GetBengkelServiceRepository()
	bengkelPhotoRepo := repository.GetBengkelPhotoRepository()
	bengkelTestimonialRepo := repository.GetBengkelTestimonialRepository()
	orderRepo := repository.GetOrderRepository()
	orderServiceRepo := repository.GetOrderServiceRepository()
	chatHistoryRepo := repository.GetChatHistoryRepository()
	adminFeeRepo := repository.GetAdminFeeRepository()

	// Create service dependencies
	serviceDeps := service.ServiceDependencies{
		UserRepo:                userRepo,
		MitraRepo:               mitraRepo,
		BengkelRepo:             bengkelRepo,
		AddressRepo:             addressRepo,
		VehicleRepo:             vehicleRepo,
		RefreshTokenRepo:        refreshTokenRepo,
		BengkelAddressRepo:      bengkelAddressRepo,
		BengkelOperationalRepo:  bengkelOperationalRepo,
		BengkelServiceRepo:      bengkelServiceRepo,
		BengkelPhotoRepo:        bengkelPhotoRepo,
		BengkelTestimonialRepo:  bengkelTestimonialRepo,
		OrderRepo:               orderRepo,
		OrderServiceRepo:        orderServiceRepo,
		ChatHistoryRepo:         chatHistoryRepo,
		AdminFeeRepo:            adminFeeRepo,
		JWTHelper:               jwtHelper,
		PasswordHelper:          passwordHelper,
	}

	// Initialize services
	authService := service.NewAuthService(serviceDeps)
	userService := service.NewUserService(serviceDeps)
	bengkelService := service.NewBengkelService(serviceDeps)
	orderService := service.NewOrderService(serviceDeps)
	chatService := service.NewChatService(serviceDeps)
	mitraService := service.NewMitraService(serviceDeps)
	adminFeeService := service.NewAdminFeeService(serviceDeps)
	cleanupService := service.NewCleanupService(refreshTokenRepo)

	applog.Info("Dependency container initialized successfully")

	return &Container{
		// Services
		AuthService:     authService,
		UserService:     userService,
		BengkelService:  bengkelService,
		OrderService:    orderService,
		ChatService:     chatService,
		MitraService:    mitraService,
		AdminFeeService: adminFeeService,
		CleanupService:  cleanupService,

		// Repositories
		UserRepo:               userRepo,
		MitraRepo:              mitraRepo,
		BengkelRepo:            bengkelRepo,
		AddressRepo:            addressRepo,
		VehicleRepo:            vehicleRepo,
		RefreshTokenRepo:       refreshTokenRepo,
		BengkelAddressRepo:     bengkelAddressRepo,
		BengkelOperationalRepo: bengkelOperationalRepo,
		BengkelServiceRepo:     bengkelServiceRepo,
		BengkelPhotoRepo:       bengkelPhotoRepo,
		BengkelTestimonialRepo: bengkelTestimonialRepo,
		OrderRepo:              orderRepo,
		OrderServiceRepo:       orderServiceRepo,
		ChatHistoryRepo:        chatHistoryRepo,
		AdminFeeRepo:           adminFeeRepo,

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

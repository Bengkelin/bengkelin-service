package mocks

import (
	"context"

	"github.com/Bengkelin/bengkelin-service/internal/models"
	"github.com/Bengkelin/bengkelin-service/internal/repository"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepositoryInterface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserRepository) FindUserByID(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetDetailUser(ctx context.Context, userId string) (*models.User, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserById(ctx context.Context, userId string, user *models.User) error {
	args := m.Called(ctx, userId, user)
	return args.Error(0)
}

// MockMitraRepository is a mock implementation of MitraRepositoryInterface
type MockMitraRepository struct {
	mock.Mock
}

func (m *MockMitraRepository) CreateMitra(ctx context.Context, mitra models.Mitra) (models.Mitra, error) {
	args := m.Called(ctx, mitra)
	return args.Get(0).(models.Mitra), args.Error(1)
}

func (m *MockMitraRepository) FindMitraByID(ctx context.Context, mitraID string) (*models.Mitra, error) {
	args := m.Called(ctx, mitraID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Mitra), args.Error(1)
}

func (m *MockMitraRepository) GetMitraByID(ctx context.Context, mitraID string) (*models.Mitra, error) {
	args := m.Called(ctx, mitraID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Mitra), args.Error(1)
}

func (m *MockMitraRepository) FindMitraByEmail(ctx context.Context, email string) (*models.Mitra, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Mitra), args.Error(1)
}

func (m *MockMitraRepository) UpdateMitra(ctx context.Context, mitraID string, mitra *models.Mitra) error {
	args := m.Called(ctx, mitraID, mitra)
	return args.Error(0)
}

// MockBengkelRepository is a mock implementation of BengkelRepositoryInterface
type MockBengkelRepository struct {
	mock.Mock
}

func (m *MockBengkelRepository) CreateBengkel(ctx context.Context, bengkel models.Bengkel) (models.Bengkel, error) {
	args := m.Called(ctx, bengkel)
	return args.Get(0).(models.Bengkel), args.Error(1)
}

func (m *MockBengkelRepository) GetBengkelById(ctx context.Context, bengkelId string) (*models.Bengkel, error) {
	args := m.Called(ctx, bengkelId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bengkel), args.Error(1)
}

func (m *MockBengkelRepository) GetBengkelByIdFresh(ctx context.Context, bengkelId string) (*models.Bengkel, error) {
	args := m.Called(ctx, bengkelId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bengkel), args.Error(1)
}

func (m *MockBengkelRepository) FindBengkelById(ctx context.Context, bengkelId string, page, size int) (*models.Bengkel, []models.BengkelTestimonial, int, error) {
	args := m.Called(ctx, bengkelId, page, size)
	var bengkel *models.Bengkel
	var testimonials []models.BengkelTestimonial
	if args.Get(0) != nil {
		bengkel = args.Get(0).(*models.Bengkel)
	}
	if args.Get(1) != nil {
		testimonials = args.Get(1).([]models.BengkelTestimonial)
	}
	return bengkel, testimonials, args.Int(2), args.Error(3)
}

func (m *MockBengkelRepository) UpdateBengkelById(ctx context.Context, bengkelId string, bengkel *models.Bengkel) error {
	args := m.Called(ctx, bengkelId, bengkel)
	return args.Error(0)
}

func (m *MockBengkelRepository) GetAllBengkel(ctx context.Context) ([]models.Bengkel, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Bengkel), args.Error(1)
}

func (m *MockBengkelRepository) GetAllBengkelPaginate(ctx context.Context, page int, limit int) ([]models.Bengkel, int, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]models.Bengkel), args.Int(1), args.Error(2)
}

func (m *MockBengkelRepository) GetNearestBengkelPaginate(ctx context.Context, lat, lng float64, page, limit int) ([]repository.BengkelWithDistance, int, error) {
	args := m.Called(ctx, lat, lng, page, limit)
	return args.Get(0).([]repository.BengkelWithDistance), args.Int(1), args.Error(2)
}

func (m *MockBengkelRepository) GetBengkelSearch(ctx context.Context, query string, page int, limit int) ([]models.Bengkel, int, error) {
	args := m.Called(ctx, query, page, limit)
	return args.Get(0).([]models.Bengkel), args.Int(1), args.Error(2)
}

func (m *MockBengkelRepository) GetBengkelByFilterService(ctx context.Context, service string, page int, limit int) ([]models.Bengkel, int, error) {
	args := m.Called(ctx, service, page, limit)
	return args.Get(0).([]models.Bengkel), args.Int(1), args.Error(2)
}

func (m *MockBengkelRepository) GetBengkelSearchV2(ctx context.Context, service, query string, page int, limit int) ([]models.Bengkel, int, error) {
	args := m.Called(ctx, service, query, page, limit)
	return args.Get(0).([]models.Bengkel), args.Int(1), args.Error(2)
}

func (m *MockBengkelRepository) SearchBengkelPublic(ctx context.Context, criteria map[string]interface{}) ([]models.Bengkel, int, error) {
	args := m.Called(ctx, criteria)
	return args.Get(0).([]models.Bengkel), args.Int(1), args.Error(2)
}

// MockAddressRepository is a mock implementation of AddressRepositoryInterface
type MockAddressRepository struct {
	mock.Mock
}

func (m *MockAddressRepository) CreateAddress(ctx context.Context, address models.UserAddress) (models.UserAddress, error) {
	args := m.Called(ctx, address)
	return args.Get(0).(models.UserAddress), args.Error(1)
}

func (m *MockAddressRepository) UpdateAddressById(ctx context.Context, addressId uint, userId string, address *models.UserAddress) error {
	args := m.Called(ctx, addressId, userId, address)
	return args.Error(0)
}

func (m *MockAddressRepository) GetAddressById(ctx context.Context, userId string, addressId uint) (*models.UserAddress, error) {
	args := m.Called(ctx, userId, addressId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserAddress), args.Error(1)
}

func (m *MockAddressRepository) DeleteAddressById(ctx context.Context, addressId uint, userId string) error {
	args := m.Called(ctx, addressId, userId)
	return args.Error(0)
}

// MockVehicleRepository is a mock implementation of VehicleRepositoryInterface
type MockVehicleRepository struct {
	mock.Mock
}

func (m *MockVehicleRepository) CreateVehicle(ctx context.Context, vehicle models.Vehicle) (models.Vehicle, error) {
	args := m.Called(ctx, vehicle)
	return args.Get(0).(models.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) UpdateVehicleById(ctx context.Context, vehicleId uint, userId string, vehicle *models.Vehicle) error {
	args := m.Called(ctx, vehicleId, userId, vehicle)
	return args.Error(0)
}

func (m *MockVehicleRepository) GetVehicleById(ctx context.Context, userId string, vehicleId uint) (*models.Vehicle, error) {
	args := m.Called(ctx, userId, vehicleId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) GetAllVehiclesByUserId(ctx context.Context, userId string) ([]models.Vehicle, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]models.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) DeleteVehicleById(ctx context.Context, vehicleId uint, userId string) error {
	args := m.Called(ctx, vehicleId, userId)
	return args.Error(0)
}

// MockOrderRepository is a mock implementation of OrderRepositoryInterface
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, order models.Order) (models.Order, error) {
	args := m.Called(ctx, order)
	return args.Get(0).(models.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateOrderById(ctx context.Context, orderId string, order *models.Order) error {
	args := m.Called(ctx, orderId, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetOrderById(ctx context.Context, orderId string) (*models.Order, error) {
	args := m.Called(ctx, orderId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetDetailOrderById(ctx context.Context, orderId, userId string) (*models.Order, error) {
	args := m.Called(ctx, orderId, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetAllOrderUserPaginate(ctx context.Context, userId string, page, limit int) ([]models.Order, int, error) {
	args := m.Called(ctx, userId, page, limit)
	return args.Get(0).([]models.Order), args.Int(1), args.Error(2)
}

func (m *MockOrderRepository) GetAllOrderMitraPaginate(ctx context.Context, bengkelId string, page, limit int) ([]models.Order, int, error) {
	args := m.Called(ctx, bengkelId, page, limit)
	return args.Get(0).([]models.Order), args.Int(1), args.Error(2)
}

// MockOrderServiceRepository is a mock implementation of OrderServiceRepositoryInterface
type MockOrderServiceRepository struct {
	mock.Mock
}

func (m *MockOrderServiceRepository) CreateOrderService(ctx context.Context, orderService models.OrderService) (models.OrderService, error) {
	args := m.Called(ctx, orderService)
	return args.Get(0).(models.OrderService), args.Error(1)
}

func (m *MockOrderServiceRepository) UpdateOrderServiceById(ctx context.Context, orderServiceId string, orderService *models.OrderService) error {
	args := m.Called(ctx, orderServiceId, orderService)
	return args.Error(0)
}

func (m *MockOrderServiceRepository) GetAllOrderService(ctx context.Context) ([]models.OrderService, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.OrderService), args.Error(1)
}

func (m *MockOrderServiceRepository) GetAllOrderServicePaginate(ctx context.Context, page int, limit int, userId string) ([]models.OrderService, int, error) {
	args := m.Called(ctx, page, limit, userId)
	return args.Get(0).([]models.OrderService), args.Int(1), args.Error(2)
}

// MockChatHistoryRepository is a mock implementation of ChatHistoryRepositoryInterface
type MockChatHistoryRepository struct {
	mock.Mock
}

func (m *MockChatHistoryRepository) CreateChatHistory(ctx context.Context, chatHistory models.ChatHistory) (models.ChatHistory, error) {
	args := m.Called(ctx, chatHistory)
	return args.Get(0).(models.ChatHistory), args.Error(1)
}

func (m *MockChatHistoryRepository) UpdateChatHistoryById(ctx context.Context, chatHistoryId string, chatHistory *models.ChatHistory) error {
	args := m.Called(ctx, chatHistoryId, chatHistory)
	return args.Error(0)
}

func (m *MockChatHistoryRepository) GetAllChatHistory(ctx context.Context) ([]models.ChatHistory, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.ChatHistory), args.Error(1)
}

func (m *MockChatHistoryRepository) GetAllChatHistoryPaginate(ctx context.Context, page int, limit int, senderId, receiverId string) ([]models.ChatHistory, int, error) {
	args := m.Called(ctx, page, limit, senderId, receiverId)
	return args.Get(0).([]models.ChatHistory), args.Int(1), args.Error(2)
}

// MockBengkelTestimonialRepository is a mock implementation of BengkelTestimonialRepositoryInterface
type MockBengkelTestimonialRepository struct {
	mock.Mock
}

func (m *MockBengkelTestimonialRepository) CreateBengkelTestimonial(ctx context.Context, bengkelTestimonial models.BengkelTestimonial) (models.BengkelTestimonial, error) {
	args := m.Called(ctx, bengkelTestimonial)
	return args.Get(0).(models.BengkelTestimonial), args.Error(1)
}

func (m *MockBengkelTestimonialRepository) UpdateBengkelTestimonialById(ctx context.Context, bengkelTestimonialId string, bengkelTestimonial *models.BengkelTestimonial) error {
	args := m.Called(ctx, bengkelTestimonialId, bengkelTestimonial)
	return args.Error(0)
}

func (m *MockBengkelTestimonialRepository) GetBengkelTestimonialById(ctx context.Context, bengkelTestimonialId string) (*models.BengkelTestimonial, error) {
	args := m.Called(ctx, bengkelTestimonialId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BengkelTestimonial), args.Error(1)
}

func (m *MockBengkelTestimonialRepository) GetAllBengkelTestimonialPaginate(ctx context.Context, page int, limit int) ([]models.BengkelTestimonial, int, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]models.BengkelTestimonial), args.Int(1), args.Error(2)
}

// MockBengkelOperationalRepository is a mock implementation of BengkelOperationalRepositoryInterface
type MockBengkelOperationalRepository struct {
	mock.Mock
}

func (m *MockBengkelOperationalRepository) CreateBengkelOperational(ctx context.Context, bengkelOperational models.BengkelOperational) (models.BengkelOperational, error) {
	args := m.Called(ctx, bengkelOperational)
	return args.Get(0).(models.BengkelOperational), args.Error(1)
}

func (m *MockBengkelOperationalRepository) UpdateBengkelOperationalById(ctx context.Context, bengkelOperationalId, bengkelOperationalHari string, bengkelOperational *models.BengkelOperational) error {
	args := m.Called(ctx, bengkelOperationalId, bengkelOperationalHari, bengkelOperational)
	return args.Error(0)
}

func (m *MockBengkelOperationalRepository) GetBengkelOperationalById(ctx context.Context, bengkelId string) (*models.BengkelOperational, error) {
	args := m.Called(ctx, bengkelId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BengkelOperational), args.Error(1)
}

func (m *MockBengkelOperationalRepository) GetBengkelOperationalByIdAndDay(ctx context.Context, bengkelId, day string) (*models.BengkelOperational, error) {
	args := m.Called(ctx, bengkelId, day)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BengkelOperational), args.Error(1)
}

// MockBengkelAddressRepository is a mock implementation of BengkelAddressRepositoryInterface
type MockBengkelAddressRepository struct {
	mock.Mock
}

func (m *MockBengkelAddressRepository) CreateBengkelAddress(ctx context.Context, bengkelAddress models.BengkelAddress) (models.BengkelAddress, error) {
	args := m.Called(ctx, bengkelAddress)
	return args.Get(0).(models.BengkelAddress), args.Error(1)
}

func (m *MockBengkelAddressRepository) UpdateBengkelAddressById(ctx context.Context, bengkelAddressId string, bengkelAddress *models.BengkelAddress) error {
	args := m.Called(ctx, bengkelAddressId, bengkelAddress)
	return args.Error(0)
}

func (m *MockBengkelAddressRepository) GetBengkelAddressById(ctx context.Context, bengkelAddressId string) (*models.BengkelAddress, error) {
	args := m.Called(ctx, bengkelAddressId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BengkelAddress), args.Error(1)
}

// MockBengkelPhotoRepository is a mock implementation of BengkelPhotoRepositoryInterface
type MockBengkelPhotoRepository struct {
	mock.Mock
}

func (m *MockBengkelPhotoRepository) CreateBengkelPhoto(ctx context.Context, bengkelPhoto models.BengkelPhoto) (models.BengkelPhoto, error) {
	args := m.Called(ctx, bengkelPhoto)
	return args.Get(0).(models.BengkelPhoto), args.Error(1)
}

func (m *MockBengkelPhotoRepository) UpdateBengkelPhotoById(ctx context.Context, bengkelPhotoId string, bengkelPhoto *models.BengkelPhoto) error {
	args := m.Called(ctx, bengkelPhotoId, bengkelPhoto)
	return args.Error(0)
}

func (m *MockBengkelPhotoRepository) GetBengkelPhotoById(ctx context.Context, bengkelId string) (*models.BengkelPhoto, error) {
	args := m.Called(ctx, bengkelId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BengkelPhoto), args.Error(1)
}

func (m *MockBengkelPhotoRepository) DeleteBengkelPhotoById(ctx context.Context, photoId string) error {
	args := m.Called(ctx, photoId)
	return args.Error(0)
}

func (m *MockBengkelPhotoRepository) GetBengkelPhotoByPK(ctx context.Context, photoId string) (*models.BengkelPhoto, error) {
	args := m.Called(ctx, photoId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BengkelPhoto), args.Error(1)
}

// MockBengkelServiceRepository is a mock implementation of BengkelServiceRepositoryInterface
type MockBengkelServiceRepository struct {
	mock.Mock
}

func (m *MockBengkelServiceRepository) CreateBengkelService(ctx context.Context, bengkelService models.BengkelService) (models.BengkelService, error) {
	args := m.Called(ctx, bengkelService)
	return args.Get(0).(models.BengkelService), args.Error(1)
}

func (m *MockBengkelServiceRepository) UpdateBengkelServiceById(ctx context.Context, bengkelServiceId string, bengkelService *models.BengkelService) error {
	args := m.Called(ctx, bengkelServiceId, bengkelService)
	return args.Error(0)
}

func (m *MockBengkelServiceRepository) UpdateBengkelService(ctx context.Context, serviceId uint, bengkelService *models.BengkelService) error {
	args := m.Called(ctx, serviceId, bengkelService)
	return args.Error(0)
}

func (m *MockBengkelServiceRepository) GetBengkelServiceById(ctx context.Context, bengkelServiceId string) (*models.BengkelService, error) {
	args := m.Called(ctx, bengkelServiceId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BengkelService), args.Error(1)
}

func (m *MockBengkelServiceRepository) GetBengkelServiceByServiceId(ctx context.Context, serviceId uint) (*models.BengkelService, error) {
	args := m.Called(ctx, serviceId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BengkelService), args.Error(1)
}

// MockAdminFeeRepository is a mock implementation of AdminFeeRepositoryInterface
type MockAdminFeeRepository struct {
	mock.Mock
}

func (m *MockAdminFeeRepository) CreateAdminFee(ctx context.Context, adminFee models.AdminFee) (models.AdminFee, error) {
	args := m.Called(ctx, adminFee)
	return args.Get(0).(models.AdminFee), args.Error(1)
}

func (m *MockAdminFeeRepository) UpdateAdminFeeById(ctx context.Context, adminFeeId string, adminFee *models.AdminFee) error {
	args := m.Called(ctx, adminFeeId, adminFee)
	return args.Error(0)
}

func (m *MockAdminFeeRepository) GetAdminFeeById(ctx context.Context, adminFeeId string) (*models.AdminFee, error) {
	args := m.Called(ctx, adminFeeId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AdminFee), args.Error(1)
}

func (m *MockAdminFeeRepository) GetOneAdminFeeLatest(ctx context.Context) (*models.AdminFee, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AdminFee), args.Error(1)
}

// MockRefreshTokenRepository is a mock implementation of RefreshTokenRepositoryInterface
type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) CreateRefreshToken(ctx context.Context, refreshToken models.RefreshToken) (models.RefreshToken, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) FindRefreshTokenByToken(ctx context.Context, token string) (models.RefreshToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) FindRefreshTokenByUserID(ctx context.Context, userID string) ([]models.RefreshToken, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) FindRefreshTokenByMitraID(ctx context.Context, mitraID string) ([]models.RefreshToken, error) {
	args := m.Called(ctx, mitraID)
	return args.Get(0).([]models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllMitraRefreshTokens(ctx context.Context, mitraID string) error {
	args := m.Called(ctx, mitraID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) UpdateRefreshToken(ctx context.Context, refreshToken models.RefreshToken) (models.RefreshToken, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(models.RefreshToken), args.Error(1)
}

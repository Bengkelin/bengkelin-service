package mocks

import (
	"github.com/Bengkelin/bengkelin-service/internal/dto"
	"github.com/Bengkelin/bengkelin-service/internal/models"
	"github.com/Bengkelin/bengkelin-service/internal/validator"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of AuthServiceInterface
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) RegisterUser(req *dto.RegisterUserRequest) (*dto.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) LoginUser(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RegisterMitra(req *dto.RegisterMitraRequest) (*dto.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) LoginMitra(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(refreshToken string) (*dto.AuthResponse, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Logout(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthService) LogoutAll(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// MockUserService is a mock implementation of UserServiceInterface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetProfile(userID string) (*models.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateProfile(userID string, req *dto.UpdateUserRequest) (*models.User, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) CreateVehicle(userID string, req *dto.VehicleRequest) (*models.Vehicle, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vehicle), args.Error(1)
}

func (m *MockUserService) GetVehicles(userID string) ([]*models.Vehicle, error) {
	args := m.Called(userID)
	return args.Get(0).([]*models.Vehicle), args.Error(1)
}

// MockMitraService is a mock implementation of MitraServiceInterface
type MockMitraService struct {
	mock.Mock
}

func (m *MockMitraService) GetProfile(mitraID string) (*models.Mitra, error) {
	args := m.Called(mitraID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Mitra), args.Error(1)
}

func (m *MockMitraService) UpdateProfile(mitraID string, req *validator.MitraUpdateProfileRequest) (*models.Mitra, error) {
	args := m.Called(mitraID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Mitra), args.Error(1)
}

func (m *MockMitraService) CreateBank(mitraID string, req *validator.MitraBankUpdateRequest) (*models.Mitra, error) {
	args := m.Called(mitraID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Mitra), args.Error(1)
}

func (m *MockMitraService) UpdateBank(mitraID string, req *validator.MitraBankUpdateRequest) (*models.Mitra, error) {
	args := m.Called(mitraID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Mitra), args.Error(1)
}

// MockBengkelService is a mock implementation of BengkelServiceInterface
type MockBengkelService struct {
	mock.Mock
}

func (m *MockBengkelService) GetBengkelByID(bengkelID string) (*models.Bengkel, error) {
	args := m.Called(bengkelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bengkel), args.Error(1)
}

func (m *MockBengkelService) GetAllBengkels(page, limit int) ([]*models.Bengkel, int, error) {
	args := m.Called(page, limit)
	return args.Get(0).([]*models.Bengkel), args.Int(1), args.Error(2)
}

func (m *MockBengkelService) SearchBengkels(query string, page, limit int) ([]*models.Bengkel, int, error) {
	args := m.Called(query, page, limit)
	return args.Get(0).([]*models.Bengkel), args.Int(1), args.Error(2)
}

func (m *MockBengkelService) CreateBengkel(mitraID string, req *dto.CreateBengkelRequest) (*models.Bengkel, error) {
	args := m.Called(mitraID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bengkel), args.Error(1)
}

func (m *MockBengkelService) UpdateBengkel(bengkelID string, req *dto.UpdateBengkelRequest) (*models.Bengkel, error) {
	args := m.Called(bengkelID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bengkel), args.Error(1)
}

// MockOrderService is a mock implementation of OrderServiceInterface
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(userID string, req *dto.CreateOrderRequest) (*models.Order, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderByID(orderID string) (*models.Order, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderService) GetUserOrders(userID string, page, limit int) ([]*models.Order, int, error) {
	args := m.Called(userID, page, limit)
	return args.Get(0).([]*models.Order), args.Int(1), args.Error(2)
}

func (m *MockOrderService) GetMitraOrders(mitraID string, page, limit int) ([]*models.Order, int, error) {
	args := m.Called(mitraID, page, limit)
	return args.Get(0).([]*models.Order), args.Int(1), args.Error(2)
}

func (m *MockOrderService) UpdateOrderStatus(orderID string, status models.OrderStatus) (*models.Order, error) {
	args := m.Called(orderID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

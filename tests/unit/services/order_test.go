package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/service"
	"github.com/Bengkelin/bengkelin-service/tests/fixtures/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupOrderService() (*mocks.MockUserRepository, *mocks.MockMitraRepository, *mocks.MockOrderRepository, *mocks.MockOrderServiceRepository, *mocks.MockAdminFeeRepository, service.OrderServiceInterface) {
	userRepo := new(mocks.MockUserRepository)
	mitraRepo := new(mocks.MockMitraRepository)
	orderRepo := new(mocks.MockOrderRepository)
	orderSvcRepo := new(mocks.MockOrderServiceRepository)
	adminFeeRepo := new(mocks.MockAdminFeeRepository)
	svc := service.NewOrderService(service.ServiceDependencies{
		UserRepo:         userRepo,
		MitraRepo:        mitraRepo,
		OrderRepo:        orderRepo,
		OrderServiceRepo: orderSvcRepo,
		AdminFeeRepo:     adminFeeRepo,
	})
	return userRepo, mitraRepo, orderRepo, orderSvcRepo, adminFeeRepo, svc
}

// --- CreateOrder (requires db.WithTransaction — skipped) ---

func TestCreateOrder_Success(t *testing.T) {
	t.Skip("Requires db.WithTransaction — needs integration test with real DB")
}

func TestCreateOrder_UserNotFound(t *testing.T) {
	t.Skip("Requires db.WithTransaction — needs integration test with real DB")
}

// --- CreateOrderWithServices (requires db.WithTransaction — skipped) ---

func TestCreateOrderWithServices_Success(t *testing.T) {
	t.Skip("Requires db.WithTransaction — needs integration test with real DB")
}

// --- UpdateOrderStatusWithValidation ---

func TestUpdateOrderStatus_Success(t *testing.T) {
	_, mitraRepo, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	order := &models.Order{
		ID:        "order-1",
		BengkelID: "bengkel-1",
		Status:    models.OrderStatusPending,
	}
	orderRepo.On("GetOrderById", ctx, "order-1").Return(order, nil)
	orderRepo.On("UpdateOrderById", ctx, "order-1", mock.Anything).Return(nil)

	result, err := svc.UpdateOrderStatusWithValidation(ctx, "mitra-1", "order-1", uint(models.OrderStatusConfirmed), "")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "order-1", result.ID)
	mitraRepo.AssertExpectations(t)
	orderRepo.AssertExpectations(t)
}

func TestUpdateOrderStatus_InvalidStatus(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	result, err := svc.UpdateOrderStatusWithValidation(ctx, "mitra-1", "order-1", 99, "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "status not valid")
}

func TestUpdateOrderStatus_InvalidTransition(t *testing.T) {
	_, mitraRepo, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	// Order is confirmed (1), trying to go to pending (0) — invalid
	order := &models.Order{
		ID:        "order-1",
		BengkelID: "bengkel-1",
		Status:    models.OrderStatusConfirmed,
	}
	orderRepo.On("GetOrderById", ctx, "order-1").Return(order, nil)

	result, err := svc.UpdateOrderStatusWithValidation(ctx, "mitra-1", "order-1", uint(models.OrderStatusPending), "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot transition")
}

func TestUpdateOrderStatus_CancelWithoutReason(t *testing.T) {
	_, mitraRepo, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	order := &models.Order{
		ID:        "order-1",
		BengkelID: "bengkel-1",
		Status:    models.OrderStatusPending,
	}
	orderRepo.On("GetOrderById", ctx, "order-1").Return(order, nil)

	result, err := svc.UpdateOrderStatusWithValidation(ctx, "mitra-1", "order-1", uint(models.OrderStatusCancelled), "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cancellation reason is required")
}

func TestUpdateOrderStatus_OrderNotFound(t *testing.T) {
	_, mitraRepo, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)
	orderRepo.On("GetOrderById", ctx, "order-999").Return(nil, errors.New("not found"))

	result, err := svc.UpdateOrderStatusWithValidation(ctx, "mitra-1", "order-999", uint(models.OrderStatusConfirmed), "")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUpdateOrderStatus_MitraNoBengkel(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID:       "mitra-1",
		Bengkel:  []models.Bengkel{},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	result, err := svc.UpdateOrderStatusWithValidation(ctx, "mitra-1", "order-1", uint(models.OrderStatusConfirmed), "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "mitra must have bengkel")
}

func TestUpdateOrderStatus_OrderNotOwned(t *testing.T) {
	_, mitraRepo, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	order := &models.Order{
		ID:        "order-1",
		BengkelID: "bengkel-OTHER", // Different bengkel
		Status:    models.OrderStatusPending,
	}
	orderRepo.On("GetOrderById", ctx, "order-1").Return(order, nil)

	result, err := svc.UpdateOrderStatusWithValidation(ctx, "mitra-1", "order-1", uint(models.OrderStatusConfirmed), "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "order does not belong")
}

func TestUpdateOrderStatus_MitraNotFound(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitraRepo.On("FindMitraByID", ctx, "mitra-999").Return(nil, errors.New("not found"))

	result, err := svc.UpdateOrderStatusWithValidation(ctx, "mitra-999", "order-1", uint(models.OrderStatusConfirmed), "")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUpdateOrderStatus_CancelWithReason(t *testing.T) {
	_, mitraRepo, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	order := &models.Order{
		ID:        "order-1",
		BengkelID: "bengkel-1",
		Status:    models.OrderStatusPending,
	}
	orderRepo.On("GetOrderById", ctx, "order-1").Return(order, nil)
	orderRepo.On("UpdateOrderById", ctx, "order-1", mock.Anything).Return(nil)

	result, err := svc.UpdateOrderStatusWithValidation(ctx, "mitra-1", "order-1", uint(models.OrderStatusCancelled), "no_show")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "order-1", result.ID)
}

// --- GetOrderForUser ---

func TestGetOrderForUser_Success(t *testing.T) {
	userRepo, _, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	user := &models.User{ID: "user-1"}
	userRepo.On("FindUserByID", ctx, "user-1").Return(user, nil)

	order := &models.Order{
		ID:        "order-1",
		UserID:    "user-1",
		BengkelID: "bengkel-1",
	}
	orderRepo.On("GetDetailOrderById", ctx, "order-1", "user-1").Return(order, nil)

	result, err := svc.GetOrderForUser(ctx, "order-1", "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "order-1", result.ID)
	userRepo.AssertExpectations(t)
	orderRepo.AssertExpectations(t)
}

func TestGetOrderForUser_UserNotFound(t *testing.T) {
	userRepo, _, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	userRepo.On("FindUserByID", ctx, "user-999").Return(nil, errors.New("not found"))

	result, err := svc.GetOrderForUser(ctx, "order-1", "user-999")

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- GetOrderForMitra ---

func TestGetOrderForMitra_Success(t *testing.T) {
	_, mitraRepo, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	order := &models.Order{
		ID:        "order-1",
		BengkelID: "bengkel-1",
	}
	orderRepo.On("GetOrderById", ctx, "order-1").Return(order, nil)

	result, err := svc.GetOrderForMitra(ctx, "order-1", "mitra-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "order-1", result.ID)
	mitraRepo.AssertExpectations(t)
	orderRepo.AssertExpectations(t)
}

func TestGetOrderForMitra_MitraNotFound(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitraRepo.On("FindMitraByID", ctx, "mitra-999").Return(nil, errors.New("not found"))

	result, err := svc.GetOrderForMitra(ctx, "order-1", "mitra-999")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetOrderForMitra_OrderNotOwned(t *testing.T) {
	_, mitraRepo, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	order := &models.Order{
		ID:        "order-1",
		BengkelID: "bengkel-OTHER",
	}
	orderRepo.On("GetOrderById", ctx, "order-1").Return(order, nil)

	result, err := svc.GetOrderForMitra(ctx, "order-1", "mitra-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "order does not belong")
}

func TestGetOrderForMitra_MitraNoBengkel(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID:      "mitra-1",
		Bengkel: []models.Bengkel{},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	result, err := svc.GetOrderForMitra(ctx, "order-1", "mitra-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "mitra has no bengkel")
}

// --- GetUserOrdersPaginate ---

func TestGetUserOrdersPaginate_Success(t *testing.T) {
	userRepo, _, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	user := &models.User{ID: "user-1"}
	userRepo.On("FindUserByID", ctx, "user-1").Return(user, nil)

	orders := []models.Order{
		{ID: "order-1", UserID: "user-1", TotalPrice: 100000, CreatedAt: time.Now()},
		{ID: "order-2", UserID: "user-1", TotalPrice: 200000, CreatedAt: time.Now()},
	}
	orderRepo.On("GetAllOrderUserPaginate", ctx, "user-1", 1, 10).Return(orders, 2, nil)

	result, count, err := svc.GetUserOrdersPaginate(ctx, "user-1", 1, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, count)
	assert.Equal(t, "order-1", result[0].ID)
	userRepo.AssertExpectations(t)
	orderRepo.AssertExpectations(t)
}

func TestGetUserOrdersPaginate_UserNotFound(t *testing.T) {
	userRepo, _, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	userRepo.On("FindUserByID", ctx, "user-999").Return(nil, errors.New("not found"))

	result, count, err := svc.GetUserOrdersPaginate(ctx, "user-999", 1, 10)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, count)
}

// --- GetMitraOrdersPaginate ---

func TestGetMitraOrdersPaginate_Success(t *testing.T) {
	_, mitraRepo, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID: "mitra-1",
		Bengkel: []models.Bengkel{
			{ID: "bengkel-1"},
		},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	orders := []models.Order{
		{ID: "order-1", BengkelID: "bengkel-1", TotalPrice: 150000, CreatedAt: time.Now()},
	}
	orderRepo.On("GetAllOrderMitraPaginate", ctx, "bengkel-1", 1, 10).Return(orders, 1, nil)

	result, count, err := svc.GetMitraOrdersPaginate(ctx, "mitra-1", 1, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, count)
	mitraRepo.AssertExpectations(t)
	orderRepo.AssertExpectations(t)
}

func TestGetMitraOrdersPaginate_MitraNotFound(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitraRepo.On("FindMitraByID", ctx, "mitra-999").Return(nil, errors.New("not found"))

	result, count, err := svc.GetMitraOrdersPaginate(ctx, "mitra-999", 1, 10)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, count)
}

func TestGetMitraOrdersPaginate_MitraNoBengkel(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{
		ID:      "mitra-1",
		Bengkel: []models.Bengkel{},
	}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	result, count, err := svc.GetMitraOrdersPaginate(ctx, "mitra-1", 1, 10)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, count)
	assert.Contains(t, err.Error(), "mitra has no bengkel")
}

// --- UpdateOrderDetails ---

func TestUpdateOrderDetails_Success(t *testing.T) {
	_, _, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	order := &models.Order{
		ID:     "order-1",
		UserID: "user-1",
	}
	orderRepo.On("GetDetailOrderById", ctx, "order-1", "user-1").Return(order, nil)
	orderRepo.On("UpdateOrderById", ctx, "order-1", mock.Anything).Return(nil)

	isHomeService := true
	err := svc.UpdateOrderDetails(ctx, "order-1", "user-1", &isHomeService, "2026-05-25 10:00", "cash")

	assert.NoError(t, err)
	orderRepo.AssertExpectations(t)
}

func TestUpdateOrderDetails_OrderNotFound(t *testing.T) {
	_, _, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	orderRepo.On("GetDetailOrderById", ctx, "order-999", "user-1").Return(nil, errors.New("not found"))

	err := svc.UpdateOrderDetails(ctx, "order-999", "user-1", nil, "", "")

	assert.Error(t, err)
}

// --- ValidateUserExists ---

func TestValidateUserExists_Success(t *testing.T) {
	userRepo, _, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	user := &models.User{ID: "user-1"}
	userRepo.On("FindUserByID", ctx, "user-1").Return(user, nil)

	err := svc.ValidateUserExists(ctx, "user-1")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestValidateUserExists_NotFound(t *testing.T) {
	userRepo, _, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	userRepo.On("FindUserByID", ctx, "user-999").Return(nil, errors.New("not found"))

	err := svc.ValidateUserExists(ctx, "user-999")

	assert.Error(t, err)
}

// --- ValidateMitraExists ---

func TestValidateMitraExists_Success(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitra := &models.Mitra{ID: "mitra-1"}
	mitraRepo.On("FindMitraByID", ctx, "mitra-1").Return(mitra, nil)

	err := svc.ValidateMitraExists(ctx, "mitra-1")

	assert.NoError(t, err)
	mitraRepo.AssertExpectations(t)
}

func TestValidateMitraExists_NotFound(t *testing.T) {
	_, mitraRepo, _, _, _, svc := setupOrderService()
	ctx := context.Background()

	mitraRepo.On("FindMitraByID", ctx, "mitra-999").Return(nil, errors.New("not found"))

	err := svc.ValidateMitraExists(ctx, "mitra-999")

	assert.Error(t, err)
}

// --- GetOrderDetails ---

func TestGetOrderDetails_Success(t *testing.T) {
	_, _, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	order := &models.Order{
		ID:         "order-1",
		UserID:     "user-1",
		BengkelID:  "bengkel-1",
		TotalPrice: 150000,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	orderRepo.On("GetOrderById", ctx, "order-1").Return(order, nil)

	result, err := svc.GetOrderDetails(ctx, "order-1", "user")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "order-1", result.ID)
	assert.Equal(t, float64(150000), result.TotalPrice)
	orderRepo.AssertExpectations(t)
}

func TestGetOrderDetails_NotFound(t *testing.T) {
	_, _, orderRepo, _, _, svc := setupOrderService()
	ctx := context.Background()

	orderRepo.On("GetOrderById", ctx, "order-999").Return(nil, errors.New("not found"))

	result, err := svc.GetOrderDetails(ctx, "order-999", "user")

	assert.Error(t, err)
	assert.Nil(t, result)
}

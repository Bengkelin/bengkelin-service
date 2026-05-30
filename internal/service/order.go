package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/db"
	"github.com/Bengkelin/bengkelin-service/internal/dto"
	"github.com/Bengkelin/bengkelin-service/internal/events"
	"github.com/Bengkelin/bengkelin-service/internal/models"
	"github.com/Bengkelin/bengkelin-service/internal/repository"
	"github.com/Bengkelin/bengkelin-service/internal/helpers"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
	"gorm.io/gorm"
)

type OrderServiceImpl struct {
	userRepo         repository.UserRepositoryInterface
	mitraRepo        repository.MitraRepositoryInterface
	orderRepo        repository.OrderRepositoryInterface
	orderServiceRepo repository.OrderServiceRepositoryInterface
	adminFeeRepo     repository.AdminFeeRepositoryInterface
}

func NewOrderService(deps ServiceDependencies) OrderServiceInterface {
	return &OrderServiceImpl{
		userRepo:         deps.UserRepo,
		mitraRepo:        deps.MitraRepo,
		orderRepo:        deps.OrderRepo,
		orderServiceRepo: deps.OrderServiceRepo,
		adminFeeRepo:     deps.AdminFeeRepo,
	}
}

// CreateOrder creates an order with services in a single transaction
func (s *OrderServiceImpl) CreateOrder(ctx context.Context, mitraID, userID string, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// Validate user exists
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Validate mitra exists and has bengkel
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return nil, fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return nil, fmt.Errorf("mitra has no bengkel")
	}

	// Validate user has vehicles
	if len(user.Vehicles) == 0 {
		return nil, fmt.Errorf("user must have at least one vehicle")
	}

	// Get admin fee
	adminFeeData, err := s.adminFeeRepo.GetOneAdminFeeLatest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin fee: %w", err)
	}

	// Build order and services
	orderID := helpers.GenerateUUID()
	isHomeService := req.ServiceType == "home"
	var orderServices []models.OrderService
	var totalPrice float64

	for _, svcName := range req.Services {
		orderServices = append(orderServices, models.OrderService{
			OrderID: orderID,
			Title:   svcName,
		})
	}

	order := models.Order{
		ID:            orderID,
		UserID:        userID,
		BengkelID:     mitra.Bengkel[0].ID,
		Status:        0,
		VehicleID:     user.Vehicles[0].ID,
		IsHomeService: &isHomeService,
		AdminFee:      adminFeeData.AdminFee,
		Note:          req.Notes,
	}

	// Execute in transaction
	err = db.WithTransaction(func(tx *gorm.DB) error {
		if _, err := s.orderRepo.CreateOrder(ctx, order); err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		for _, os := range orderServices {
			if _, err := s.orderServiceRepo.CreateOrderService(ctx, os); err != nil {
				return fmt.Errorf("failed to create order service: %w", err)
			}
		}

		order.TotalPrice = totalPrice
		if err := s.orderRepo.UpdateOrderById(ctx, orderID, &order); err != nil {
			return fmt.Errorf("failed to update order total: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish event
	eventBus := events.GetEventBus()
	eventBus.Publish(events.Event{
		Type:    events.OrderCreated,
		Payload: &order,
	})

	applog.InfoCtx(ctx, "Order created successfully", "order_id", orderID, "user_id", userID, "mitra_id", mitraID)

	return &dto.OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		BengkelID:   order.BengkelID,
		VehicleID:   fmt.Sprintf("%d", user.Vehicles[0].ID),
		Status:      "pending",
		ServiceType: req.ServiceType,
		TotalPrice:  totalPrice,
		Notes:       req.Notes,
	}, nil
}

// GetOrderDetails gets order details
func (s *OrderServiceImpl) GetOrderDetails(ctx context.Context, orderID string, userType string) (*dto.OrderDetailResponse, error) {
	order, err := s.orderRepo.GetOrderById(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	return &dto.OrderDetailResponse{
		OrderResponse: dto.OrderResponse{
			ID:         order.ID,
			UserID:     order.UserID,
			BengkelID:  order.BengkelID,
			TotalPrice: order.TotalPrice,
			CreatedAt:  order.CreatedAt,
			UpdatedAt:  order.UpdatedAt,
		},
	}, nil
}

// UpdateOrderStatus updates order status
func (s *OrderServiceImpl) UpdateOrderStatus(ctx context.Context, mitraID, orderID string, status string) error {
	return nil
}

// UpdateOrderStatusWithValidation validates mitra ownership, status transitions, and updates the order
func (s *OrderServiceImpl) UpdateOrderStatusWithValidation(ctx context.Context, mitraID, orderID string, newStatus uint, reason string) (*models.Order, error) {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return nil, fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return nil, fmt.Errorf("mitra must have bengkel to update orders")
	}

	if !models.OrderStatus(newStatus).IsValid() {
		return nil, fmt.Errorf("status not valid (0-4)")
	}

	if newStatus == uint(models.OrderStatusCancelled) && reason == "" {
		return nil, fmt.Errorf("cancellation reason is required when status is 4")
	}

	order, err := s.orderRepo.GetOrderById(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	if order.BengkelID != mitra.Bengkel[0].ID {
		return nil, fmt.Errorf("order does not belong to your bengkel")
	}

	targetStatus := models.OrderStatus(newStatus)
	if !order.Status.CanTransitionTo(targetStatus) {
		return nil, fmt.Errorf("cannot transition from %s to %s", order.Status.String(), targetStatus.String())
	}

	now := time.Now()
	updateData := &models.Order{
		Status: targetStatus,
	}

	switch targetStatus {
	case models.OrderStatusConfirmed:
		updateData.ConfirmedAt = &now
	case models.OrderStatusInProgress:
		updateData.FinishedAt = &now
	case models.OrderStatusCancelled:
		updateData.CancelledAt = &now
		updateData.CancelledBy = mitraID
		updateData.CancelledReason = models.CancelledByMitra
		if reason != "" {
			switch reason {
			case "no_show":
				updateData.CancelledReason = models.CancelledNoShow
			case "service_unavailable":
				updateData.CancelledReason = models.CancelledServiceUnavailable
			case "payment_failed":
				updateData.CancelledReason = models.CancelledPaymentFailed
			default:
				updateData.CancelledReason = models.CancelledByMitra
			}
		}
	}

	if err := s.orderRepo.UpdateOrderById(ctx, orderID, updateData); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	updatedOrder, err := s.orderRepo.GetOrderById(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated order: %w", err)
	}

	return updatedOrder, nil
}

// GetUserOrders gets paginated orders for a user
func (s *OrderServiceImpl) GetUserOrders(ctx context.Context, userID string, req dto.PaginationRequest) (*dto.PaginatedOrderResponse, error) {
	orders, count, err := s.orderRepo.GetAllOrderUserPaginate(ctx, userID, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	var orderResponses []dto.OrderResponse
	for _, o := range orders {
		orderResponses = append(orderResponses, dto.OrderResponse{
			ID:         o.ID,
			UserID:     o.UserID,
			BengkelID:  o.BengkelID,
			TotalPrice: o.TotalPrice,
			CreatedAt:  o.CreatedAt,
			UpdatedAt:  o.UpdatedAt,
		})
	}

	return &dto.PaginatedOrderResponse{
		PaginationResponse: dto.PaginationResponse{
			Page:  req.Page,
			Limit: req.Limit,
			Total: int64(count),
		},
		Data: orderResponses,
	}, nil
}

// GetMitraOrders gets paginated orders for a mitra
func (s *OrderServiceImpl) GetMitraOrders(ctx context.Context, mitraID string, req dto.PaginationRequest) (*dto.PaginatedOrderResponse, error) {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return nil, fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return nil, fmt.Errorf("mitra has no bengkel")
	}

	orders, count, err := s.orderRepo.GetAllOrderMitraPaginate(ctx, mitra.Bengkel[0].ID, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	var orderResponses []dto.OrderResponse
	for _, o := range orders {
		orderResponses = append(orderResponses, dto.OrderResponse{
			ID:         o.ID,
			UserID:     o.UserID,
			BengkelID:  o.BengkelID,
			TotalPrice: o.TotalPrice,
			CreatedAt:  o.CreatedAt,
			UpdatedAt:  o.UpdatedAt,
		})
	}

	return &dto.PaginatedOrderResponse{
		PaginationResponse: dto.PaginationResponse{
			Page:  req.Page,
			Limit: req.Limit,
			Total: int64(count),
		},
		Data: orderResponses,
	}, nil
}

// CreateOrderWithServices creates an order with structured services in a transaction
func (s *OrderServiceImpl) CreateOrderWithServices(ctx context.Context, userID, mitraID string, req dto.CreateOrderWithServicesRequest) (*dto.CreateOrderResult, error) {
	// Validate user exists
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Validate mitra exists and has bengkel
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return nil, fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return nil, fmt.Errorf("mitra has no bengkel")
	}

	// Validate user has vehicles
	if len(user.Vehicles) == 0 {
		return nil, fmt.Errorf("user must have at least one vehicle")
	}

	// Get admin fee
	adminFeeData, err := s.adminFeeRepo.GetOneAdminFeeLatest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin fee: %w", err)
	}

	// Build order and services
	orderID := helpers.GenerateUUID()
	var orderServices []models.OrderService
	var totalPrice float64

	for _, svc := range req.Services {
		orderServices = append(orderServices, models.OrderService{
			OrderID: orderID,
			Title:   svc.Title,
			Detail:  svc.Detail,
			Price:   svc.Price,
		})
		totalPrice += svc.Price
	}

	order := models.Order{
		ID:            orderID,
		UserID:        userID,
		BengkelID:     mitra.Bengkel[0].ID,
		Status:        0,
		VehicleID:     user.Vehicles[0].ID,
		IsHomeService: &req.IsHomeService,
		AdminFee:      adminFeeData.AdminFee,
	}

	// Execute in transaction
	err = db.WithTransaction(func(tx *gorm.DB) error {
		if _, err := s.orderRepo.CreateOrder(ctx, order); err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		for _, os := range orderServices {
			if _, err := s.orderServiceRepo.CreateOrderService(ctx, os); err != nil {
				return fmt.Errorf("failed to create order service: %w", err)
			}
		}

		order.TotalPrice = totalPrice
		if err := s.orderRepo.UpdateOrderById(ctx, orderID, &order); err != nil {
			return fmt.Errorf("failed to update order total: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish event
	eventBus := events.GetEventBus()
	eventBus.Publish(events.Event{
		Type:    events.OrderCreated,
		Payload: &order,
	})

	// Build creator name
	creatorName := user.FirstName + " " + user.LastName

	applog.InfoCtx(ctx, "Order created successfully", "order_id", orderID, "user_id", userID, "mitra_id", mitraID)

	return &dto.CreateOrderResult{
		OrderID:       order.ID,
		UserID:        userID,
		BengkelID:     mitra.Bengkel[0].ID,
		BengkelName:   mitra.Bengkel[0].BengkelName,
		TotalPrice:    totalPrice,
		AdminFee:      adminFeeData.AdminFee,
		Status:        int(order.Status),
		CreatedByName: creatorName,
		IsSelfOrder:   true,
	}, nil
}

// GetOrderForUser gets an order with user ownership validation
func (s *OrderServiceImpl) GetOrderForUser(ctx context.Context, orderID, userID string) (*models.Order, error) {
	if _, err := s.userRepo.FindUserByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return s.orderRepo.GetDetailOrderById(ctx, orderID, userID)
}

// GetOrderForMitra gets an order with mitra ownership validation
func (s *OrderServiceImpl) GetOrderForMitra(ctx context.Context, orderID, mitraID string) (*models.Order, error) {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return nil, fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return nil, fmt.Errorf("mitra has no bengkel")
	}

	order, err := s.orderRepo.GetOrderById(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	if order.BengkelID != mitra.Bengkel[0].ID {
		return nil, fmt.Errorf("order does not belong to your bengkel")
	}

	return order, nil
}

// GetUserOrdersPaginate gets paginated orders for a user
func (s *OrderServiceImpl) GetUserOrdersPaginate(ctx context.Context, userID string, page, limit int) ([]models.Order, int, error) {
	if _, err := s.userRepo.FindUserByID(ctx, userID); err != nil {
		return nil, 0, fmt.Errorf("user not found: %w", err)
	}
	return s.orderRepo.GetAllOrderUserPaginate(ctx, userID, page, limit)
}

// GetMitraOrdersPaginate gets paginated orders for a mitra
func (s *OrderServiceImpl) GetMitraOrdersPaginate(ctx context.Context, mitraID string, page, limit int) ([]models.Order, int, error) {
	mitra, err := s.mitraRepo.FindMitraByID(ctx, mitraID)
	if err != nil {
		return nil, 0, fmt.Errorf("mitra not found: %w", err)
	}
	if len(mitra.Bengkel) == 0 {
		return nil, 0, fmt.Errorf("mitra has no bengkel")
	}
	return s.orderRepo.GetAllOrderMitraPaginate(ctx, mitra.Bengkel[0].ID, page, limit)
}

// UpdateOrderDetails updates order details with user validation
func (s *OrderServiceImpl) UpdateOrderDetails(ctx context.Context, orderID, userID string, isHomeService *bool, homeServiceSchedule, paymentMethod string) error {
	order, err := s.orderRepo.GetDetailOrderById(ctx, orderID, userID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	return s.orderRepo.UpdateOrderById(ctx, order.ID, &models.Order{
		IsHomeService:       isHomeService,
		HomeServiceSchedule: homeServiceSchedule,
		PaymentMethod:       paymentMethod,
	})
}

// ValidateUserExists validates that a user exists
func (s *OrderServiceImpl) ValidateUserExists(ctx context.Context, userID string) error {
	if _, err := s.userRepo.FindUserByID(ctx, userID); err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	return nil
}

// ValidateMitraExists validates that a mitra exists
func (s *OrderServiceImpl) ValidateMitraExists(ctx context.Context, mitraID string) error {
	if _, err := s.mitraRepo.FindMitraByID(ctx, mitraID); err != nil {
		return fmt.Errorf("mitra not found: %w", err)
	}
	return nil
}

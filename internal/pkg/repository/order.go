package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	orderRepository *OrderRepository
	orderOnce       sync.Once
)

type OrderRepositoryInterface interface {
	CreateOrder(ctx context.Context, order models.Order) (models.Order, error)
	UpdateOrderById(ctx context.Context, orderId string, order *models.Order) error
	GetOrderById(ctx context.Context, orderId string) (*models.Order, error)
	GetDetailOrderById(ctx context.Context, OrderId, userId string) (*models.Order, error)
	GetAllOrderUserPaginate(ctx context.Context, userId string, page, limit int) ([]models.Order, int, error)
	GetAllOrderMitraPaginate(ctx context.Context, bengkelId string, page, limit int) ([]models.Order, int, error)
}

type OrderRepository struct{}

func GetOrderRepository() OrderRepositoryInterface {
	orderOnce.Do(func() {
		orderRepository = &OrderRepository{}
	})
	return orderRepository
}

// CreateOrder implements OrderRepositoryInterface.
func (repo *OrderRepository) CreateOrder(ctx context.Context, Order models.Order) (models.Order, error) {
	err := db.GetDB().WithContext(ctx).Create(&Order).Error
	if err != nil {
		return models.Order{}, err
	}
	return Order, nil
}

// UpdateOrderById implements OrderRepositoryInterface.
func (*OrderRepository) UpdateOrderById(ctx context.Context, OrderId string, Order *models.Order) error {
	err := db.GetDB().WithContext(ctx).Model(&models.Order{}).Where("id = ?", OrderId).Updates(Order).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAllOrderUserPaginate implements OrderRepositoryInterface.
func (*OrderRepository) GetAllOrderUserPaginate(ctx context.Context, userId string, page, limit int) ([]models.Order, int, error) {
	var Orders []models.Order
	var count int64

	err := db.GetDB().WithContext(ctx).Model(&models.Order{}).Where("user_id = ?", userId).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.GetDB().WithContext(ctx).Preload("OrderServices").Preload("User").Preload("Bengkel").Preload("Bengkel.Addresses").Preload("Vehicle").Where("user_id = ?", userId).Offset((page - 1) * limit).Limit(limit).Find(&Orders).Error

	if err != nil {
		return nil, 0, err
	}

	return Orders, int(count), nil
}

// GetAllOrderMitraPaginate implements OrderRepositoryInterface.
func (*OrderRepository) GetAllOrderMitraPaginate(ctx context.Context, bengkelId string, page, limit int) ([]models.Order, int, error) {
	var Orders []models.Order
	var count int64

	err := db.GetDB().WithContext(ctx).Model(&models.Order{}).Where("bengkel_id = ?", bengkelId).Count(&count).Error

	if err != nil {
		return nil, 0, err
	}

	err = db.GetDB().WithContext(ctx).Preload("OrderServices").Preload("User").Preload("Bengkel").Preload("Bengkel.Addresses").Preload("Vehicle").Where("bengkel_id = ?", bengkelId).Offset((page - 1) * limit).Limit(limit).Find(&Orders).Error

	if err != nil {
		return nil, 0, err
	}

	return Orders, int(count), nil
}

// GetOrderById implements OrderRepositoryInterface.
func (*OrderRepository) GetOrderById(ctx context.Context, OrderId string) (*models.Order, error) {
	var Order models.Order
	where := models.Order{}
	where.ID = OrderId
	_, err := First(ctx, where, &Order, []string{"OrderServices", "User", "Bengkel", "Bengkel.Addresses", "Vehicle"})
	if err != nil {
		return nil, err
	}
	return &Order, nil
}

// GetDetailOrderById implements OrderRepositoryInterface.
func (*OrderRepository) GetDetailOrderById(ctx context.Context, OrderId, userId string) (*models.Order, error) {
	var Order models.Order
	err := db.GetDB().WithContext(ctx).
		Model(&models.Order{}).
		Preload("OrderServices").
		Preload("User").
		Preload("Bengkel").
		Preload("Bengkel.Addresses").
		Preload("Vehicle").
		Where("id = ? AND user_id = ?", OrderId, userId).
		Take(&Order).Error
	if err != nil {
		return nil, err
	}
	return &Order, nil
}

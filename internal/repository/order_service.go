package repository

import (
	"context"
	"sync"

	"github.com/Bengkelin/bengkelin-service/internal/db"
	"github.com/Bengkelin/bengkelin-service/internal/models"
)

var (
	orderServiceRepository *OrderServiceRepository
	orderServiceOnce       sync.Once
)

type OrderServiceRepositoryInterface interface {
	CreateOrderService(ctx context.Context, orderService models.OrderService) (models.OrderService, error)
	UpdateOrderServiceById(ctx context.Context, orderServiceId string, orderService *models.OrderService) error
	GetAllOrderService(ctx context.Context) ([]models.OrderService, error)
	GetAllOrderServicePaginate(ctx context.Context, page int, limit int, userId string) ([]models.OrderService, int, error)
}

type OrderServiceRepository struct{}

func GetOrderServiceRepository() OrderServiceRepositoryInterface {
	orderServiceOnce.Do(func() {
		orderServiceRepository = &OrderServiceRepository{}
	})
	return orderServiceRepository
}

// CreateOrderService implements OrderServiceRepositoryInterface.
func (repo *OrderServiceRepository) CreateOrderService(ctx context.Context, orderService models.OrderService) (models.OrderService, error) {
	err := db.GetDB().WithContext(ctx).Create(&orderService).Error
	if err != nil {
		return models.OrderService{}, err
	}
	return orderService, nil
}

// UpdateOrderServiceById implements OrderServiceRepositoryInterface.
func (*OrderServiceRepository) UpdateOrderServiceById(ctx context.Context, orderServiceId string, orderService *models.OrderService) error {
	err := db.GetDB().WithContext(ctx).Model(&models.OrderService{}).Where("id = ?", orderServiceId).Updates(orderService).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAllOrderService implements OrderServiceRepositoryInterface.
func (*OrderServiceRepository) GetAllOrderService(ctx context.Context) ([]models.OrderService, error) {
	var orderServices []models.OrderService
	err := db.GetDB().WithContext(ctx).Find(&orderServices).Error
	if err != nil {
		return nil, err
	}
	return orderServices, nil
}

// GetAllOrderServicePaginate implements OrderServiceRepositoryInterface.
func (*OrderServiceRepository) GetAllOrderServicePaginate(ctx context.Context, page int, limit int, userId string) ([]models.OrderService, int, error) {
	var orderServices []models.OrderService
	var total int64
	err := db.GetDB().WithContext(ctx).Model(&models.OrderService{}).Where("user_id = ?", userId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.GetDB().WithContext(ctx).Where("user_id = ?", userId).Offset((page - 1) * limit).Limit(limit).Find(&orderServices).Error
	if err != nil {
		return nil, 0, err
	}

	return orderServices, int(total), nil
}

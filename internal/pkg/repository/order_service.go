package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	orderServiceRepository *OrderServiceRepository
)

type OrderServiceRepositoryInterface interface {
	CreateOrderService(orderService models.OrderService) (models.OrderService, error)
	UpdateOrderServiceById(orderServiceId string, orderService *models.OrderService) error
	GetAllOrderService() ([]models.OrderService, error)
	GetAllOrderServicePaginate(page int, limit int, userId string) ([]models.OrderService, int, error)
}

type OrderServiceRepository struct{}

func GetOrderServiceRepository() OrderServiceRepositoryInterface {
	if orderServiceRepository == nil {
		orderServiceRepository = &OrderServiceRepository{}
	}
	return orderServiceRepository
}

// CreateOrderService implements OrderServiceRepositoryInterface.
func (repo *OrderServiceRepository) CreateOrderService(orderService models.OrderService) (models.OrderService, error) {
	err := db.GetDB().Create(&orderService).Error
	if err != nil {
		return models.OrderService{}, err
	}
	return orderService, nil
}

// UpdateOrderServiceById implements OrderServiceRepositoryInterface.
func (*OrderServiceRepository) UpdateOrderServiceById(orderServiceId string, orderService *models.OrderService) error {
	err := db.GetDB().Model(&models.OrderService{}).Where("id = ?", orderServiceId).Updates(orderService).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAllOrderService implements OrderServiceRepositoryInterface.
func (*OrderServiceRepository) GetAllOrderService() ([]models.OrderService, error) {
	var orderServices []models.OrderService
	err := db.GetDB().Find(&orderServices).Error
	if err != nil {
		return nil, err
	}
	return orderServices, nil
}

// GetAllOrderServicePaginate implements OrderServiceRepositoryInterface.
func (*OrderServiceRepository) GetAllOrderServicePaginate(page int, limit int, userId string) ([]models.OrderService, int, error) {
	var orderServices []models.OrderService
	var total int64
	err := db.GetDB().Model(&models.OrderService{}).Where("user_id = ?", userId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.GetDB().Where("user_id = ?", userId).Offset((page - 1) * limit).Limit(limit).Find(&orderServices).Error
	if err != nil {
		return nil, 0, err
	}

	return orderServices, int(total), nil
}
	

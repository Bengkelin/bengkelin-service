package repository

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/db"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

var (
	orderRepository *OrderRepository
)

type OrderRepositoryInterface interface {
	CreateOrder(order models.Order) (models.Order, error)
	UpdateOrderById(orderId string, order *models.Order) error
	GetOrderById(orderId string) (*models.Order, error)
	GetDetailOrderById(OrderId, userId string) (*models.Order, error)
	GetAllOrderUserPaginate(userId string, page, limit int) ([]models.Order, int, error)
	GetAllOrderMitraPaginate(bengkelId string, page, limit int) ([]models.Order, int, error)
}

type OrderRepository struct{}

func GetOrderRepository() OrderRepositoryInterface {
	if orderRepository == nil {
		orderRepository = &OrderRepository{}
	}
	return orderRepository
}

// CreateOrder implements OrderRepositoryInterface.
func (repo *OrderRepository) CreateOrder(Order models.Order) (models.Order, error) {
	err := db.GetDB().Create(&Order).Error
	if err != nil {
		return models.Order{}, err
	}
	return Order, nil
}

// UpdateOrderById implements OrderRepositoryInterface.
func (*OrderRepository) UpdateOrderById(OrderId string, Order *models.Order) error {
	err := db.GetDB().Model(&models.Order{}).Where("id = ?", OrderId).Updates(Order).Error

	if err != nil {
		return err
	}
	return nil
}

// GetAllOrderUserPaginate implements OrderRepositoryInterface.
func (*OrderRepository) GetAllOrderUserPaginate(userId string, page, limit int) ([]models.Order, int, error) {
	var Orders []models.Order
	var count int64

	err := db.GetDB().Model(&models.Order{}).Where("user_id = ?", userId).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.GetDB().Preload("OrderService").Preload("User").Preload("Bengkel").Preload("Bengkel.Addresses").Preload("Vehicle").Where("user_id = ?", userId).Offset((page - 1) * limit).Limit(limit).Find(&Orders).Error

	if err != nil {
		return nil, 0, err
	}

	return Orders, int(count), nil
}

// GetAllOrderMitraPaginate implements OrderRepositoryInterface.
func (*OrderRepository) GetAllOrderMitraPaginate(bengkelId string, page, limit int) ([]models.Order, int, error) {
	var Orders []models.Order
	var count int64

	err := db.GetDB().Model(&models.Order{}).Where("bengkel_id = ?", bengkelId).Count(&count).Error

	if err != nil {
		return nil, 0, err
	}

	err = db.GetDB().Preload("OrderService").Preload("User").Preload("Bengkel").Preload("Bengkel.Addresses").Preload("Vehicle").Where("bengkel_id = ?", bengkelId).Offset((page - 1) * limit).Limit(limit).Find(&Orders).Error

	if err != nil {
		return nil, 0, err
	}

	return Orders, int(count), nil
}

// GetOrderById implements OrderRepositoryInterface.
func (*OrderRepository) GetOrderById(OrderId string) (*models.Order, error) {
	var Order models.Order
	where := models.Order{}
	where.ID = OrderId
	_, err := First(where, &Order, []string{"OrderService", "User", "Bengkel", "Bengkel.Addresses", "Vehicle"})
	if err != nil {
		return nil, err
	}
	return &Order, nil
}

// GetDetailOrderById implements OrderRepositoryInterface.
func (*OrderRepository) GetDetailOrderById(OrderId, userId string) (*models.Order, error) {
	var Order models.Order
	err := db.GetDB().
		Model(&models.Order{}).
		Preload("OrderService").
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

//

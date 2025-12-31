package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/rabbitmq"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
)

// OrderProcessorService handles order processing operations
type OrderProcessorService struct {
	rabbitMQ           *rabbitmq.RabbitMQ
	notificationService NotificationServiceInterface
}

// OrderProcessorServiceInterface defines the order processor service contract
type OrderProcessorServiceInterface interface {
	// Order processing
	ProcessNewOrder(ctx context.Context, orderID, userID, bengkelID string) error
	ProcessOrderStatusUpdate(ctx context.Context, orderID, status, updatedBy string) error
	ProcessOrderCancellation(ctx context.Context, orderID, reason, cancelledBy string) error
	ProcessOrderCompletion(ctx context.Context, orderID string) error
	
	// Payment processing
	ProcessPaymentInitiation(ctx context.Context, orderID, paymentMethod string, amount float64) error
	ProcessPaymentSuccess(ctx context.Context, orderID, transactionID string) error
	ProcessPaymentFailure(ctx context.Context, orderID, reason string) error
	
	// Inventory management
	ProcessInventoryReservation(ctx context.Context, orderID string, items []InventoryItem) error
	ProcessInventoryRelease(ctx context.Context, orderID string, items []InventoryItem) error
	
	// Integration events
	ProcessMidtransWebhook(ctx context.Context, webhookData interface{}) error
	ProcessAgoraTokenRefresh(ctx context.Context, userID, channelName string) error
}

// Order processing payloads
type OrderProcessingEvent struct {
	OrderID     string                 `json:"order_id"`
	UserID      string                 `json:"user_id"`
	BengkelID   string                 `json:"bengkel_id"`
	Status      string                 `json:"status"`
	EventType   string                 `json:"event_type"`
	Data        map[string]interface{} `json:"data"`
	ProcessedBy string                 `json:"processed_by"`
	Timestamp   time.Time              `json:"timestamp"`
}

type PaymentProcessingEvent struct {
	OrderID       string                 `json:"order_id"`
	PaymentMethod string                 `json:"payment_method"`
	Amount        float64                `json:"amount"`
	TransactionID string                 `json:"transaction_id,omitempty"`
	Status        string                 `json:"status"`
	EventType     string                 `json:"event_type"`
	Data          map[string]interface{} `json:"data"`
	Timestamp     time.Time              `json:"timestamp"`
}

type InventoryItem struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
	Type     string `json:"type"` // service, part, etc.
}

type InventoryEvent struct {
	OrderID   string          `json:"order_id"`
	Items     []InventoryItem `json:"items"`
	EventType string          `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
}

type IntegrationEvent struct {
	Service   string                 `json:"service"`
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewOrderProcessorService creates a new order processor service
func NewOrderProcessorService(notificationService NotificationServiceInterface) OrderProcessorServiceInterface {
	return &OrderProcessorService{
		rabbitMQ:            rabbitmq.GetInstance(),
		notificationService: notificationService,
	}
}

// Order processing methods
func (s *OrderProcessorService) ProcessNewOrder(ctx context.Context, orderID, userID, bengkelID string) error {
	event := OrderProcessingEvent{
		OrderID:     orderID,
		UserID:      userID,
		BengkelID:   bengkelID,
		Status:      "pending",
		EventType:   "order_created",
		ProcessedBy: "system",
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"created_at": time.Now(),
		},
	}
	
	message := rabbitmq.Message{
		Type:    "order.created",
		Payload: event,
		Headers: map[string]interface{}{
			"order_id":   orderID,
			"user_id":    userID,
			"bengkel_id": bengkelID,
			"priority":   "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeOrders, "order.created", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue new order processing", 
			"order_id", orderID, "user_id", userID)
		return fmt.Errorf("failed to queue new order processing: %w", err)
	}
	
	applog.InfoCtx(ctx, "New order processing queued successfully", 
		"order_id", orderID, "user_id", userID, "bengkel_id", bengkelID)
	return nil
}

func (s *OrderProcessorService) ProcessOrderStatusUpdate(ctx context.Context, orderID, status, updatedBy string) error {
	event := OrderProcessingEvent{
		OrderID:     orderID,
		Status:      status,
		EventType:   "status_updated",
		ProcessedBy: updatedBy,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"previous_status": "", // This would be fetched from database
			"updated_at":      time.Now(),
		},
	}
	
	message := rabbitmq.Message{
		Type:    "order.status_updated",
		Payload: event,
		Headers: map[string]interface{}{
			"order_id":   orderID,
			"status":     status,
			"updated_by": updatedBy,
			"priority":   "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeOrders, "order.status", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue order status update", 
			"order_id", orderID, "status", status)
		return fmt.Errorf("failed to queue order status update: %w", err)
	}
	
	applog.InfoCtx(ctx, "Order status update queued successfully", 
		"order_id", orderID, "status", status, "updated_by", updatedBy)
	return nil
}

func (s *OrderProcessorService) ProcessOrderCancellation(ctx context.Context, orderID, reason, cancelledBy string) error {
	event := OrderProcessingEvent{
		OrderID:     orderID,
		Status:      "cancelled",
		EventType:   "order_cancelled",
		ProcessedBy: cancelledBy,
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"cancellation_reason": reason,
			"cancelled_at":        time.Now(),
		},
	}
	
	message := rabbitmq.Message{
		Type:    "order.cancelled",
		Payload: event,
		Headers: map[string]interface{}{
			"order_id":     orderID,
			"reason":       reason,
			"cancelled_by": cancelledBy,
			"priority":     "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeOrders, "order.cancelled", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue order cancellation", 
			"order_id", orderID, "reason", reason)
		return fmt.Errorf("failed to queue order cancellation: %w", err)
	}
	
	applog.InfoCtx(ctx, "Order cancellation queued successfully", 
		"order_id", orderID, "reason", reason, "cancelled_by", cancelledBy)
	return nil
}

func (s *OrderProcessorService) ProcessOrderCompletion(ctx context.Context, orderID string) error {
	event := OrderProcessingEvent{
		OrderID:     orderID,
		Status:      "completed",
		EventType:   "order_completed",
		ProcessedBy: "system",
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"completed_at": time.Now(),
		},
	}
	
	message := rabbitmq.Message{
		Type:    "order.completed",
		Payload: event,
		Headers: map[string]interface{}{
			"order_id": orderID,
			"priority": "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeOrders, "order.completed", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue order completion", "order_id", orderID)
		return fmt.Errorf("failed to queue order completion: %w", err)
	}
	
	applog.InfoCtx(ctx, "Order completion queued successfully", "order_id", orderID)
	return nil
}

// Payment processing methods
func (s *OrderProcessorService) ProcessPaymentInitiation(ctx context.Context, orderID, paymentMethod string, amount float64) error {
	event := PaymentProcessingEvent{
		OrderID:       orderID,
		PaymentMethod: paymentMethod,
		Amount:        amount,
		Status:        "initiated",
		EventType:     "payment_initiated",
		Timestamp:     time.Now(),
		Data: map[string]interface{}{
			"initiated_at": time.Now(),
		},
	}
	
	message := rabbitmq.Message{
		Type:    "payment.initiated",
		Payload: event,
		Headers: map[string]interface{}{
			"order_id":       orderID,
			"payment_method": paymentMethod,
			"amount":         amount,
			"priority":       "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeOrders, "payment.initiated", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue payment initiation", 
			"order_id", orderID, "payment_method", paymentMethod)
		return fmt.Errorf("failed to queue payment initiation: %w", err)
	}
	
	applog.InfoCtx(ctx, "Payment initiation queued successfully", 
		"order_id", orderID, "payment_method", paymentMethod, "amount", amount)
	return nil
}

func (s *OrderProcessorService) ProcessPaymentSuccess(ctx context.Context, orderID, transactionID string) error {
	event := PaymentProcessingEvent{
		OrderID:       orderID,
		TransactionID: transactionID,
		Status:        "success",
		EventType:     "payment_success",
		Timestamp:     time.Now(),
		Data: map[string]interface{}{
			"completed_at": time.Now(),
		},
	}
	
	message := rabbitmq.Message{
		Type:    "payment.success",
		Payload: event,
		Headers: map[string]interface{}{
			"order_id":       orderID,
			"transaction_id": transactionID,
			"priority":       "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeOrders, "payment.success", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue payment success", 
			"order_id", orderID, "transaction_id", transactionID)
		return fmt.Errorf("failed to queue payment success: %w", err)
	}
	
	applog.InfoCtx(ctx, "Payment success queued successfully", 
		"order_id", orderID, "transaction_id", transactionID)
	return nil
}

func (s *OrderProcessorService) ProcessPaymentFailure(ctx context.Context, orderID, reason string) error {
	event := PaymentProcessingEvent{
		OrderID:   orderID,
		Status:    "failed",
		EventType: "payment_failed",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"failure_reason": reason,
			"failed_at":      time.Now(),
		},
	}
	
	message := rabbitmq.Message{
		Type:    "payment.failed",
		Payload: event,
		Headers: map[string]interface{}{
			"order_id": orderID,
			"reason":   reason,
			"priority": "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeOrders, "payment.failed", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue payment failure", 
			"order_id", orderID, "reason", reason)
		return fmt.Errorf("failed to queue payment failure: %w", err)
	}
	
	applog.InfoCtx(ctx, "Payment failure queued successfully", 
		"order_id", orderID, "reason", reason)
	return nil
}

// Inventory management methods
func (s *OrderProcessorService) ProcessInventoryReservation(ctx context.Context, orderID string, items []InventoryItem) error {
	event := InventoryEvent{
		OrderID:   orderID,
		Items:     items,
		EventType: "inventory_reserved",
		Timestamp: time.Now(),
	}
	
	message := rabbitmq.Message{
		Type:    "inventory.reserved",
		Payload: event,
		Headers: map[string]interface{}{
			"order_id":   orderID,
			"item_count": len(items),
			"priority":   "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeOrders, "inventory.reserved", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue inventory reservation", 
			"order_id", orderID, "item_count", len(items))
		return fmt.Errorf("failed to queue inventory reservation: %w", err)
	}
	
	applog.InfoCtx(ctx, "Inventory reservation queued successfully", 
		"order_id", orderID, "item_count", len(items))
	return nil
}

func (s *OrderProcessorService) ProcessInventoryRelease(ctx context.Context, orderID string, items []InventoryItem) error {
	event := InventoryEvent{
		OrderID:   orderID,
		Items:     items,
		EventType: "inventory_released",
		Timestamp: time.Now(),
	}
	
	message := rabbitmq.Message{
		Type:    "inventory.released",
		Payload: event,
		Headers: map[string]interface{}{
			"order_id":   orderID,
			"item_count": len(items),
			"priority":   "normal",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeOrders, "inventory.released", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue inventory release", 
			"order_id", orderID, "item_count", len(items))
		return fmt.Errorf("failed to queue inventory release: %w", err)
	}
	
	applog.InfoCtx(ctx, "Inventory release queued successfully", 
		"order_id", orderID, "item_count", len(items))
	return nil
}

// Integration event methods
func (s *OrderProcessorService) ProcessMidtransWebhook(ctx context.Context, webhookData interface{}) error {
	event := IntegrationEvent{
		Service:   "midtrans",
		EventType: "webhook_received",
		Data: map[string]interface{}{
			"webhook_data": webhookData,
		},
		Timestamp: time.Now(),
	}
	
	message := rabbitmq.Message{
		Type:    "integration.midtrans_webhook",
		Payload: event,
		Headers: map[string]interface{}{
			"service":  "midtrans",
			"priority": "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeIntegrations, "integration.midtrans", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue Midtrans webhook processing")
		return fmt.Errorf("failed to queue Midtrans webhook processing: %w", err)
	}
	
	applog.InfoCtx(ctx, "Midtrans webhook processing queued successfully")
	return nil
}

func (s *OrderProcessorService) ProcessAgoraTokenRefresh(ctx context.Context, userID, channelName string) error {
	event := IntegrationEvent{
		Service:   "agora",
		EventType: "token_refresh",
		Data: map[string]interface{}{
			"user_id":      userID,
			"channel_name": channelName,
		},
		Timestamp: time.Now(),
	}
	
	message := rabbitmq.Message{
		Type:    "integration.agora_token_refresh",
		Payload: event,
		Headers: map[string]interface{}{
			"service":      "agora",
			"user_id":      userID,
			"channel_name": channelName,
			"priority":     "normal",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeIntegrations, "integration.agora", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue Agora token refresh", 
			"user_id", userID, "channel_name", channelName)
		return fmt.Errorf("failed to queue Agora token refresh: %w", err)
	}
	
	applog.InfoCtx(ctx, "Agora token refresh queued successfully", 
		"user_id", userID, "channel_name", channelName)
	return nil
}

// StartOrderProcessingConsumers starts all order processing consumers
func StartOrderProcessingConsumers() error {
	rabbitMQ := rabbitmq.GetInstance()
	
	// Order processing consumer
	orderConsumer := rabbitmq.Consumer{
		Queue:       rabbitmq.QueueOrderProcessing,
		Handler:     handleOrderProcessing,
		Concurrency: 5,
		AutoAck:     false,
	}
	
	err := rabbitMQ.Consume(orderConsumer)
	if err != nil {
		return fmt.Errorf("failed to start order processing consumer: %w", err)
	}
	
	// Payment processing consumer
	paymentConsumer := rabbitmq.Consumer{
		Queue:       rabbitmq.QueuePaymentProcessing,
		Handler:     handlePaymentProcessing,
		Concurrency: 3,
		AutoAck:     false,
	}
	
	err = rabbitMQ.Consume(paymentConsumer)
	if err != nil {
		return fmt.Errorf("failed to start payment processing consumer: %w", err)
	}
	
	// Integration events consumer
	integrationConsumer := rabbitmq.Consumer{
		Queue:       rabbitmq.QueueIntegrationEvents,
		Handler:     handleIntegrationEvent,
		Concurrency: 2,
		AutoAck:     false,
	}
	
	err = rabbitMQ.Consume(integrationConsumer)
	if err != nil {
		return fmt.Errorf("failed to start integration events consumer: %w", err)
	}
	
	applog.Info("All order processing consumers started successfully")
	return nil
}

// Message handlers
func handleOrderProcessing(message rabbitmq.Message) error {
	applog.Info("Processing order event", "message_id", message.ID, "type", message.Type)
	
	// TODO: Implement actual order processing logic
	// This would include:
	// - Updating order status in database
	// - Sending notifications to users/mitras
	// - Triggering inventory updates
	// - Integrating with external services
	
	// Simulate processing time
	time.Sleep(500 * time.Millisecond)
	
	applog.Info("Order event processed successfully", "message_id", message.ID)
	return nil
}

func handlePaymentProcessing(message rabbitmq.Message) error {
	applog.Info("Processing payment event", "message_id", message.ID, "type", message.Type)
	
	// TODO: Implement actual payment processing logic
	// This would include:
	// - Updating payment status in database
	// - Integrating with Midtrans API
	// - Handling payment confirmations
	// - Processing refunds
	
	// Simulate processing time
	time.Sleep(1 * time.Second)
	
	applog.Info("Payment event processed successfully", "message_id", message.ID)
	return nil
}

func handleIntegrationEvent(message rabbitmq.Message) error {
	applog.Info("Processing integration event", "message_id", message.ID, "type", message.Type)
	
	// TODO: Implement actual integration event processing
	// This would include:
	// - Processing webhooks from external services
	// - Refreshing tokens for third-party APIs
	// - Syncing data with external systems
	
	// Simulate processing time
	time.Sleep(200 * time.Millisecond)
	
	applog.Info("Integration event processed successfully", "message_id", message.ID)
	return nil
}
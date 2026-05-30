package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/rabbitmq"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
)

// NotificationService handles all notification-related operations
type NotificationService struct {
	rabbitMQ *rabbitmq.RabbitMQ
}

// NotificationServiceInterface defines the notification service contract
type NotificationServiceInterface interface {
	// Email notifications
	SendWelcomeEmail(ctx context.Context, userID, email, firstName string) error
	SendOrderConfirmationEmail(ctx context.Context, userID, email, orderID string) error
	SendPasswordResetEmail(ctx context.Context, userID, email, resetToken string) error
	
	// SMS notifications
	SendOrderStatusSMS(ctx context.Context, phoneNumber, orderID, status string) error
	SendVerificationSMS(ctx context.Context, phoneNumber, code string) error
	
	// Push notifications
	SendOrderUpdatePush(ctx context.Context, userID, title, message string) error
	SendChatMessagePush(ctx context.Context, userID, senderName, message string) error
	
	// Bulk notifications
	SendBulkEmail(ctx context.Context, recipients []EmailRecipient, template string, data interface{}) error
	SendBulkPush(ctx context.Context, userIDs []string, title, message string) error
}

// Email notification payloads
type EmailNotification struct {
	UserID    string                 `json:"user_id"`
	Email     string                 `json:"email"`
	Template  string                 `json:"template"`
	Subject   string                 `json:"subject"`
	Data      map[string]interface{} `json:"data"`
	Priority  string                 `json:"priority"` // high, normal, low
}

type SMSNotification struct {
	UserID      string `json:"user_id,omitempty"`
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
	Template    string `json:"template,omitempty"`
	Priority    string `json:"priority"` // high, normal, low
}

type PushNotification struct {
	UserID   string                 `json:"user_id"`
	Title    string                 `json:"title"`
	Message  string                 `json:"message"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Priority string                 `json:"priority"` // high, normal, low
}

type EmailRecipient struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// NewNotificationService creates a new notification service
func NewNotificationService() NotificationServiceInterface {
	return &NotificationService{
		rabbitMQ: rabbitmq.GetInstance(),
	}
}

// Email notification methods
func (s *NotificationService) SendWelcomeEmail(ctx context.Context, userID, email, firstName string) error {
	notification := EmailNotification{
		UserID:   userID,
		Email:    email,
		Template: "welcome",
		Subject:  "Welcome to Bengkelin!",
		Data: map[string]interface{}{
			"first_name": firstName,
			"app_name":   "Bengkelin",
		},
		Priority: "normal",
	}
	
	message := rabbitmq.Message{
		Type:    "email.welcome",
		Payload: notification,
		Headers: map[string]interface{}{
			"user_id":  userID,
			"priority": "normal",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeNotifications, "email.welcome", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue welcome email", "user_id", userID, "email", email)
		return fmt.Errorf("failed to queue welcome email: %w", err)
	}
	
	applog.InfoCtx(ctx, "Welcome email queued successfully", "user_id", userID, "email", email)
	return nil
}

func (s *NotificationService) SendOrderConfirmationEmail(ctx context.Context, userID, email, orderID string) error {
	notification := EmailNotification{
		UserID:   userID,
		Email:    email,
		Template: "order_confirmation",
		Subject:  "Order Confirmation - " + orderID,
		Data: map[string]interface{}{
			"order_id": orderID,
		},
		Priority: "high",
	}
	
	message := rabbitmq.Message{
		Type:    "email.order_confirmation",
		Payload: notification,
		Headers: map[string]interface{}{
			"user_id":  userID,
			"order_id": orderID,
			"priority": "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeNotifications, "email.order", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue order confirmation email", 
			"user_id", userID, "email", email, "order_id", orderID)
		return fmt.Errorf("failed to queue order confirmation email: %w", err)
	}
	
	applog.InfoCtx(ctx, "Order confirmation email queued successfully", 
		"user_id", userID, "email", email, "order_id", orderID)
	return nil
}

func (s *NotificationService) SendPasswordResetEmail(ctx context.Context, userID, email, resetToken string) error {
	notification := EmailNotification{
		UserID:   userID,
		Email:    email,
		Template: "password_reset",
		Subject:  "Password Reset Request",
		Data: map[string]interface{}{
			"reset_token": resetToken,
			"reset_url":   fmt.Sprintf("https://bengkelin.com/reset-password?token=%s", resetToken),
		},
		Priority: "high",
	}
	
	message := rabbitmq.Message{
		Type:    "email.password_reset",
		Payload: notification,
		Headers: map[string]interface{}{
			"user_id":  userID,
			"priority": "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeNotifications, "email.auth", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue password reset email", "user_id", userID, "email", email)
		return fmt.Errorf("failed to queue password reset email: %w", err)
	}
	
	applog.InfoCtx(ctx, "Password reset email queued successfully", "user_id", userID, "email", email)
	return nil
}

// SMS notification methods
func (s *NotificationService) SendOrderStatusSMS(ctx context.Context, phoneNumber, orderID, status string) error {
	notification := SMSNotification{
		PhoneNumber: phoneNumber,
		Template:    "order_status",
		Message:     fmt.Sprintf("Your order %s status has been updated to: %s", orderID, status),
		Priority:    "high",
	}
	
	message := rabbitmq.Message{
		Type:    "sms.order_status",
		Payload: notification,
		Headers: map[string]interface{}{
			"phone_number": phoneNumber,
			"order_id":     orderID,
			"priority":     "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeNotifications, "sms.order", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue order status SMS", 
			"phone_number", phoneNumber, "order_id", orderID)
		return fmt.Errorf("failed to queue order status SMS: %w", err)
	}
	
	applog.InfoCtx(ctx, "Order status SMS queued successfully", 
		"phone_number", phoneNumber, "order_id", orderID)
	return nil
}

func (s *NotificationService) SendVerificationSMS(ctx context.Context, phoneNumber, code string) error {
	notification := SMSNotification{
		PhoneNumber: phoneNumber,
		Template:    "verification",
		Message:     fmt.Sprintf("Your Bengkelin verification code is: %s", code),
		Priority:    "high",
	}
	
	message := rabbitmq.Message{
		Type:    "sms.verification",
		Payload: notification,
		Headers: map[string]interface{}{
			"phone_number": phoneNumber,
			"priority":     "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeNotifications, "sms.auth", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue verification SMS", "phone_number", phoneNumber)
		return fmt.Errorf("failed to queue verification SMS: %w", err)
	}
	
	applog.InfoCtx(ctx, "Verification SMS queued successfully", "phone_number", phoneNumber)
	return nil
}

// Push notification methods
func (s *NotificationService) SendOrderUpdatePush(ctx context.Context, userID, title, message string) error {
	notification := PushNotification{
		UserID:   userID,
		Title:    title,
		Message:  message,
		Priority: "high",
		Data: map[string]interface{}{
			"type": "order_update",
		},
	}
	
	rabbitMessage := rabbitmq.Message{
		Type:    "push.order_update",
		Payload: notification,
		Headers: map[string]interface{}{
			"user_id":  userID,
			"priority": "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeNotifications, "push.order", rabbitMessage)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue order update push notification", "user_id", userID)
		return fmt.Errorf("failed to queue order update push notification: %w", err)
	}
	
	applog.InfoCtx(ctx, "Order update push notification queued successfully", "user_id", userID)
	return nil
}

func (s *NotificationService) SendChatMessagePush(ctx context.Context, userID, senderName, message string) error {
	notification := PushNotification{
		UserID:   userID,
		Title:    fmt.Sprintf("New message from %s", senderName),
		Message:  message,
		Priority: "normal",
		Data: map[string]interface{}{
			"type":        "chat_message",
			"sender_name": senderName,
		},
	}
	
	rabbitMessage := rabbitmq.Message{
		Type:    "push.chat_message",
		Payload: notification,
		Headers: map[string]interface{}{
			"user_id":     userID,
			"sender_name": senderName,
			"priority":    "normal",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeNotifications, "push.chat", rabbitMessage)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue chat message push notification", "user_id", userID)
		return fmt.Errorf("failed to queue chat message push notification: %w", err)
	}
	
	applog.InfoCtx(ctx, "Chat message push notification queued successfully", "user_id", userID)
	return nil
}

// Bulk notification methods
func (s *NotificationService) SendBulkEmail(ctx context.Context, recipients []EmailRecipient, template string, data interface{}) error {
	bulkNotification := struct {
		Recipients []EmailRecipient `json:"recipients"`
		Template   string           `json:"template"`
		Data       interface{}      `json:"data"`
		Priority   string           `json:"priority"`
	}{
		Recipients: recipients,
		Template:   template,
		Data:       data,
		Priority:   "normal",
	}
	
	message := rabbitmq.Message{
		Type:    "email.bulk",
		Payload: bulkNotification,
		Headers: map[string]interface{}{
			"recipient_count": len(recipients),
			"template":        template,
			"priority":        "normal",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeNotifications, "email.bulk", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue bulk email", 
			"recipient_count", len(recipients), "template", template)
		return fmt.Errorf("failed to queue bulk email: %w", err)
	}
	
	applog.InfoCtx(ctx, "Bulk email queued successfully", 
		"recipient_count", len(recipients), "template", template)
	return nil
}

func (s *NotificationService) SendBulkPush(ctx context.Context, userIDs []string, title, message string) error {
	bulkNotification := struct {
		UserIDs  []string `json:"user_ids"`
		Title    string   `json:"title"`
		Message  string   `json:"message"`
		Priority string   `json:"priority"`
	}{
		UserIDs:  userIDs,
		Title:    title,
		Message:  message,
		Priority: "normal",
	}
	
	rabbitMessage := rabbitmq.Message{
		Type:    "push.bulk",
		Payload: bulkNotification,
		Headers: map[string]interface{}{
			"user_count": len(userIDs),
			"priority":   "normal",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeNotifications, "push.bulk", rabbitMessage)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue bulk push notification", "user_count", len(userIDs))
		return fmt.Errorf("failed to queue bulk push notification: %w", err)
	}
	
	applog.InfoCtx(ctx, "Bulk push notification queued successfully", "user_count", len(userIDs))
	return nil
}

// StartNotificationConsumers starts all notification consumers
func StartNotificationConsumers() error {
	rabbitMQ := rabbitmq.GetInstance()
	
	// Email consumer
	emailConsumer := rabbitmq.Consumer{
		Queue:       rabbitmq.QueueEmailNotifications,
		Handler:     handleEmailNotification,
		Concurrency: 5,
		AutoAck:     false,
	}
	
	err := rabbitMQ.Consume(emailConsumer)
	if err != nil {
		return fmt.Errorf("failed to start email consumer: %w", err)
	}
	
	// SMS consumer
	smsConsumer := rabbitmq.Consumer{
		Queue:       rabbitmq.QueueSMSNotifications,
		Handler:     handleSMSNotification,
		Concurrency: 3,
		AutoAck:     false,
	}
	
	err = rabbitMQ.Consume(smsConsumer)
	if err != nil {
		return fmt.Errorf("failed to start SMS consumer: %w", err)
	}
	
	// Push notification consumer
	pushConsumer := rabbitmq.Consumer{
		Queue:       rabbitmq.QueuePushNotifications,
		Handler:     handlePushNotification,
		Concurrency: 10,
		AutoAck:     false,
	}
	
	err = rabbitMQ.Consume(pushConsumer)
	if err != nil {
		return fmt.Errorf("failed to start push notification consumer: %w", err)
	}
	
	applog.Info("All notification consumers started successfully")
	return nil
}

// Message handlers
func handleEmailNotification(message rabbitmq.Message) error {
	applog.Info("Processing email notification", "message_id", message.ID, "type", message.Type)
	
	// TODO: Implement actual email sending logic
	// This would integrate with SMTP service, SendGrid, AWS SES, etc.
	
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)
	
	applog.Info("Email notification processed successfully", "message_id", message.ID)
	return nil
}

func handleSMSNotification(message rabbitmq.Message) error {
	applog.Info("Processing SMS notification", "message_id", message.ID, "type", message.Type)
	
	// TODO: Implement actual SMS sending logic
	// This would integrate with Twilio, AWS SNS, etc.
	
	// Simulate processing time
	time.Sleep(200 * time.Millisecond)
	
	applog.Info("SMS notification processed successfully", "message_id", message.ID)
	return nil
}

func handlePushNotification(message rabbitmq.Message) error {
	applog.Info("Processing push notification", "message_id", message.ID, "type", message.Type)
	
	// TODO: Implement actual push notification logic
	// This would integrate with FCM, APNS, etc.
	
	// Simulate processing time
	time.Sleep(50 * time.Millisecond)
	
	applog.Info("Push notification processed successfully", "message_id", message.ID)
	return nil
}
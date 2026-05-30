package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/rabbitmq"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
)

// AuditService handles audit logging operations
type AuditService struct {
	rabbitMQ *rabbitmq.RabbitMQ
}

// AuditServiceInterface defines audit service methods
type AuditServiceInterface interface {
	LogUserAction(ctx context.Context, userID, action, resource string, metadata map[string]interface{}) error
	LogSystemEvent(ctx context.Context, event, description string, metadata map[string]interface{}) error
	LogSecurityEvent(ctx context.Context, userID, event, ipAddress string, metadata map[string]interface{}) error
}

// AuditEvent represents an audit event
type AuditEvent struct {
	UserID      string                 `json:"user_id,omitempty"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource,omitempty"`
	Description string                 `json:"description,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"` // user_action, system_event, security_event
}

// NewAuditService creates a new audit service
func NewAuditService() AuditServiceInterface {
	return &AuditService{
		rabbitMQ: rabbitmq.GetInstance(),
	}
}

// LogUserAction logs a user action
func (s *AuditService) LogUserAction(ctx context.Context, userID, action, resource string, metadata map[string]interface{}) error {
	event := AuditEvent{
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Metadata:  metadata,
		Timestamp: time.Now(),
		EventType: "user_action",
	}
	
	message := rabbitmq.Message{
		Type:    "audit.user_action",
		Payload: event,
		Headers: map[string]interface{}{
			"user_id":    userID,
			"action":     action,
			"event_type": "user_action",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeAudit, "", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue user action audit", 
			"user_id", userID, "action", action, "resource", resource)
		return fmt.Errorf("failed to queue user action audit: %w", err)
	}
	
	applog.InfoCtx(ctx, "User action audit queued successfully", 
		"user_id", userID, "action", action, "resource", resource)
	return nil
}

// LogSystemEvent logs a system event
func (s *AuditService) LogSystemEvent(ctx context.Context, event, description string, metadata map[string]interface{}) error {
	auditEvent := AuditEvent{
		Action:      event,
		Description: description,
		Metadata:    metadata,
		Timestamp:   time.Now(),
		EventType:   "system_event",
	}
	
	message := rabbitmq.Message{
		Type:    "audit.system_event",
		Payload: auditEvent,
		Headers: map[string]interface{}{
			"event":      event,
			"event_type": "system_event",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeAudit, "", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue system event audit", 
			"event", event, "description", description)
		return fmt.Errorf("failed to queue system event audit: %w", err)
	}
	
	applog.InfoCtx(ctx, "System event audit queued successfully", 
		"event", event, "description", description)
	return nil
}

// LogSecurityEvent logs a security event
func (s *AuditService) LogSecurityEvent(ctx context.Context, userID, event, ipAddress string, metadata map[string]interface{}) error {
	auditEvent := AuditEvent{
		UserID:    userID,
		Action:    event,
		IPAddress: ipAddress,
		Metadata:  metadata,
		Timestamp: time.Now(),
		EventType: "security_event",
	}
	
	message := rabbitmq.Message{
		Type:    "audit.security_event",
		Payload: auditEvent,
		Headers: map[string]interface{}{
			"user_id":    userID,
			"event":      event,
			"ip_address": ipAddress,
			"event_type": "security_event",
			"priority":   "high",
		},
	}
	
	err := s.rabbitMQ.Publish(rabbitmq.ExchangeAudit, "", message)
	if err != nil {
		applog.LogErrorCtx(ctx, err, "Failed to queue security event audit", 
			"user_id", userID, "event", event, "ip_address", ipAddress)
		return fmt.Errorf("failed to queue security event audit: %w", err)
	}
	
	applog.InfoCtx(ctx, "Security event audit queued successfully", 
		"user_id", userID, "event", event, "ip_address", ipAddress)
	return nil
}

// StartAuditConsumers starts audit message consumers
func StartAuditConsumers() error {
	rabbitMQ := rabbitmq.GetInstance()
	
	// Audit consumer
	auditConsumer := rabbitmq.Consumer{
		Queue:       rabbitmq.QueueAuditLogs,
		Handler:     handleAuditEvent,
		Concurrency: 3,
		AutoAck:     false,
	}
	
	err := rabbitMQ.Consume(auditConsumer)
	if err != nil {
		return fmt.Errorf("failed to start audit consumer: %w", err)
	}
	
	applog.Info("Audit consumers started successfully")
	return nil
}

// handleAuditEvent processes audit events
func handleAuditEvent(message rabbitmq.Message) error {
	applog.Info("Processing audit event", "message_id", message.ID, "type", message.Type)
	
	// TODO: Implement actual audit log storage
	// This would typically save to database, send to log aggregation system, etc.
	
	// Simulate processing time
	time.Sleep(10 * time.Millisecond)
	
	applog.Info("Audit event processed successfully", "message_id", message.ID)
	return nil
}
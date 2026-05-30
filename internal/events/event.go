package events

import (
	"context"
	"encoding/json"
	"log"
)

type EventType string

const (
	OrderCreated   EventType = "order.created"
	OrderConfirmed EventType = "order.confirmed"
	OrderCompleted EventType = "order.completed"

	// Cache invalidation events
	CacheInvalidated      EventType = "cache.invalidated"
	MitraProfileUpdated   EventType = "mitra.profile.updated"
	UserProfileUpdated    EventType = "user.profile.updated"
	BengkelUpdated        EventType = "bengkel.updated"
	BengkelServiceUpdated EventType = "bengkel.service.updated"
	BengkelAddressUpdated EventType = "bengkel.address.updated"
)

type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

type EventPublisher interface {
	Publish(event Event) error
}

type EventSubscriber interface {
	Subscribe(eventType EventType, handler func(payload []byte)) error
}

// Simple in-memory event bus (replace with Redis Pub/Sub or RabbitMQ in production)
type EventBus struct {
	handlers map[EventType][]func(payload []byte)
}

var eventBus *EventBus

func GetEventBus() *EventBus {
	if eventBus == nil {
		eventBus = &EventBus{
			handlers: make(map[EventType][]func(payload []byte)),
		}
	}
	return eventBus
}

func (eb *EventBus) Publish(event Event) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	if handlers, ok := eb.handlers[event.Type]; ok {
		for _, handler := range handlers {
			go handler(payload) // Async execution
		}
	}
	log.Printf("Event published: %s", event.Type)
	return nil
}

func (eb *EventBus) Subscribe(eventType EventType, handler func(payload []byte)) {
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// CacheInvalidationEvent represents a cache invalidation event
// Used for both local in-memory events and distributed Redis pub/sub
type CacheInvalidationEvent struct {
	EntityType string `json:"entity_type"` // e.g., "mitra", "user", "bengkel"
	EntityID   string `json:"entity_id"`   // ID of the entity that was updated
	Action     string `json:"action"`      // e.g., "updated", "deleted"
	Timestamp  int64  `json:"timestamp"`   // Unix timestamp for event ordering
	Source     string `json:"source"`      // Source service/instance identifier
}

// CacheInvalidationHandler is the interface for handling cache invalidation events
type CacheInvalidationHandler interface {
	HandleCacheInvalidation(ctx context.Context, event CacheInvalidationEvent) error
}

// CacheInvalidationPublisher publishes cache invalidation events for distributed systems
type CacheInvalidationPublisher interface {
	PublishCacheInvalidation(ctx context.Context, event CacheInvalidationEvent) error
}

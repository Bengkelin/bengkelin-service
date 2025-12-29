package events

import (
	"encoding/json"
	"log"
)

type EventType string

const (
	OrderCreated   EventType = "order.created"
	OrderConfirmed EventType = "order.confirmed"
	OrderCompleted EventType = "order.completed"
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

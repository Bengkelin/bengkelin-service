package rabbitmq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRabbitMQ is a mock implementation for testing
type MockRabbitMQ struct {
	mock.Mock
}

func (m *MockRabbitMQ) Publish(exchange, routingKey string, message Message) error {
	args := m.Called(exchange, routingKey, message)
	return args.Error(0)
}

func (m *MockRabbitMQ) Consume(consumer Consumer) error {
	args := m.Called(consumer)
	return args.Error(0)
}

func (m *MockRabbitMQ) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRabbitMQ) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestMessage_Creation(t *testing.T) {
	message := Message{
		ID:        "test-123",
		Type:      "test.message",
		Payload:   map[string]interface{}{"key": "value"},
		Headers:   map[string]interface{}{"priority": "high"},
		Timestamp: time.Now(),
		Retry:     0,
	}

	assert.Equal(t, "test-123", message.ID)
	assert.Equal(t, "test.message", message.Type)
	assert.NotNil(t, message.Payload)
	assert.NotNil(t, message.Headers)
	assert.Equal(t, 0, message.Retry)
}

func TestConsumer_Creation(t *testing.T) {
	handler := func(msg Message) error {
		return nil
	}

	consumer := Consumer{
		Queue:       QueueEmailNotifications,
		Handler:     handler,
		Concurrency: 5,
		AutoAck:     false,
	}

	assert.Equal(t, QueueEmailNotifications, consumer.Queue)
	assert.NotNil(t, consumer.Handler)
	assert.Equal(t, 5, consumer.Concurrency)
	assert.False(t, consumer.AutoAck)
}

func TestMockRabbitMQ_Publish(t *testing.T) {
	mockRabbitMQ := new(MockRabbitMQ)
	
	message := Message{
		ID:      "test-123",
		Type:    "test.message",
		Payload: "test payload",
	}

	// Setup expectation
	mockRabbitMQ.On("Publish", ExchangeNotifications, "test.route", message).Return(nil)

	// Execute
	err := mockRabbitMQ.Publish(ExchangeNotifications, "test.route", message)

	// Assert
	assert.NoError(t, err)
	mockRabbitMQ.AssertExpectations(t)
}

func TestMockRabbitMQ_IsHealthy(t *testing.T) {
	mockRabbitMQ := new(MockRabbitMQ)

	// Setup expectation
	mockRabbitMQ.On("IsHealthy").Return(true)

	// Execute
	healthy := mockRabbitMQ.IsHealthy()

	// Assert
	assert.True(t, healthy)
	mockRabbitMQ.AssertExpectations(t)
}

// Integration test helper - only runs if RabbitMQ is available
func TestRabbitMQ_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require a real RabbitMQ instance
	// For now, we'll just test the constants and structure
	assert.NotEmpty(t, QueueEmailNotifications)
	assert.NotEmpty(t, ExchangeNotifications)
	assert.Equal(t, "email_notifications", QueueEmailNotifications)
	assert.Equal(t, "notifications", ExchangeNotifications)
}

// Benchmark test for message creation
func BenchmarkMessage_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Message{
			ID:        "test-123",
			Type:      "test.message",
			Payload:   map[string]interface{}{"key": "value"},
			Headers:   map[string]interface{}{"priority": "high"},
			Timestamp: time.Now(),
			Retry:     0,
		}
	}
}
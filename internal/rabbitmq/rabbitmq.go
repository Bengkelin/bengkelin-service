package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/config"
	applog "github.com/Bengkelin/bengkelin-service/internal/log"
	"github.com/streadway/amqp"
)

// RabbitMQ connection and channel management
type RabbitMQ struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	url          string
	exchanges    map[string]bool
	queues       map[string]bool
	mu           sync.RWMutex
	reconnecting bool
	ctx          context.Context
	cancel       context.CancelFunc
}

// Message represents a message to be published
type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   interface{}            `json:"payload"`
	Headers   map[string]interface{} `json:"headers,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Retry     int                    `json:"retry,omitempty"`
}

// Consumer represents a message consumer
type Consumer struct {
	Queue       string
	Handler     func(Message) error
	Concurrency int
	AutoAck     bool
}

// Exchange types
const (
	ExchangeTypeDirect  = "direct"
	ExchangeTypeFanout  = "fanout"
	ExchangeTypeTopic   = "topic"
	ExchangeTypeHeaders = "headers"
)

// Queue names
const (
	QueueEmailNotifications = "email_notifications"
	QueueSMSNotifications   = "sms_notifications"
	QueuePushNotifications  = "push_notifications"
	QueueOrderProcessing    = "order_processing"
	QueuePaymentProcessing  = "payment_processing"
	QueueFileProcessing     = "file_processing"
	QueueAuditLogs          = "audit_logs"
	QueueIntegrationEvents  = "integration_events"
	QueueChatEvents         = "chat_events"
	QueueDeadLetter         = "dead_letter"
)

// Exchange names
const (
	ExchangeNotifications = "notifications"
	ExchangeOrders        = "orders"
	ExchangeFiles         = "files"
	ExchangeAudit         = "audit"
	ExchangeIntegrations  = "integrations"
	ExchangeChat          = "chat"
	ExchangeDeadLetter    = "dead_letter"
)

var (
	instance *RabbitMQ
	once     sync.Once
)

// Setup initializes RabbitMQ connection
func Setup() error {
	var err error
	once.Do(func() {
		conf := config.GetConfig()
		
		ctx, cancel := context.WithCancel(context.Background())
		
		instance = &RabbitMQ{
			url:       conf.RabbitMQ.URL,
			exchanges: make(map[string]bool),
			queues:    make(map[string]bool),
			ctx:       ctx,
			cancel:    cancel,
		}
		
		err = instance.connect()
		if err != nil {
			applog.LogError(err, "Failed to setup RabbitMQ")
			return
		}
		
		// Setup default exchanges and queues
		err = instance.setupInfrastructure()
		if err != nil {
			applog.LogError(err, "Failed to setup RabbitMQ infrastructure")
			return
		}
		
		// Start connection monitoring
		go instance.monitorConnection()
		
		applog.Info("RabbitMQ setup completed successfully")
	})
	
	return err
}

// GetInstance returns the RabbitMQ instance
func GetInstance() *RabbitMQ {
	if instance == nil {
		panic("RabbitMQ not initialized. Call Setup() first.")
	}
	return instance
}

// connect establishes connection to RabbitMQ
func (r *RabbitMQ) connect() error {
	var err error
	
	r.conn, err = amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	
	r.channel, err = r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	
	// Set QoS for fair dispatch
	err = r.channel.Qos(10, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}
	
	applog.Info("Connected to RabbitMQ successfully")
	return nil
}

// setupInfrastructure creates exchanges and queues
func (r *RabbitMQ) setupInfrastructure() error {
	// Declare exchanges
	exchanges := map[string]string{
		ExchangeNotifications: ExchangeTypeTopic,
		ExchangeOrders:        ExchangeTypeTopic,
		ExchangeFiles:         ExchangeTypeDirect,
		ExchangeAudit:         ExchangeTypeFanout,
		ExchangeIntegrations:  ExchangeTypeTopic,
		ExchangeChat:          ExchangeTypeTopic,
		ExchangeDeadLetter:    ExchangeTypeDirect,
	}
	
	for exchange, exchangeType := range exchanges {
		err := r.DeclareExchange(exchange, exchangeType)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}
	
	// Declare queues with their routing keys
	queueBindings := map[string][]struct {
		exchange   string
		routingKey string
	}{
		QueueEmailNotifications: {
			{ExchangeNotifications, "email.*"},
		},
		QueueSMSNotifications: {
			{ExchangeNotifications, "sms.*"},
		},
		QueuePushNotifications: {
			{ExchangeNotifications, "push.*"},
		},
		QueueOrderProcessing: {
			{ExchangeOrders, "order.*"},
		},
		QueuePaymentProcessing: {
			{ExchangeOrders, "payment.*"},
		},
		QueueFileProcessing: {
			{ExchangeFiles, "file.process"},
		},
		QueueAuditLogs: {
			{ExchangeAudit, ""},
		},
		QueueIntegrationEvents: {
			{ExchangeIntegrations, "integration.*"},
		},
		QueueChatEvents: {
			{ExchangeChat, "chat.*"},
		},
		QueueDeadLetter: {
			{ExchangeDeadLetter, "dead_letter"},
		},
	}
	
	for queue, bindings := range queueBindings {
		// Declare queue with dead letter exchange
		args := amqp.Table{
			"x-dead-letter-exchange":    ExchangeDeadLetter,
			"x-dead-letter-routing-key": "dead_letter",
			"x-message-ttl":             300000, // 5 minutes TTL
		}
		
		err := r.DeclareQueue(queue, true, false, false, false, args)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queue, err)
		}
		
		// Bind queue to exchanges
		for _, binding := range bindings {
			err := r.BindQueue(queue, binding.routingKey, binding.exchange)
			if err != nil {
				return fmt.Errorf("failed to bind queue %s to exchange %s: %w", queue, binding.exchange, err)
			}
		}
	}
	
	applog.Info("RabbitMQ infrastructure setup completed")
	return nil
}

// DeclareExchange declares an exchange
func (r *RabbitMQ) DeclareExchange(name, exchangeType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.exchanges[name] {
		return nil // Already declared
	}
	
	err := r.channel.ExchangeDeclare(
		name,         // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	
	if err != nil {
		return err
	}
	
	r.exchanges[name] = true
	applog.Debug("Exchange declared", "name", name, "type", exchangeType)
	return nil
}

// DeclareQueue declares a queue
func (r *RabbitMQ) DeclareQueue(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.queues[name] {
		return nil // Already declared
	}
	
	_, err := r.channel.QueueDeclare(
		name,       // name
		durable,    // durable
		autoDelete, // delete when unused
		exclusive,  // exclusive
		noWait,     // no-wait
		args,       // arguments
	)
	
	if err != nil {
		return err
	}
	
	r.queues[name] = true
	applog.Debug("Queue declared", "name", name)
	return nil
}

// BindQueue binds a queue to an exchange
func (r *RabbitMQ) BindQueue(queueName, routingKey, exchangeName string) error {
	return r.channel.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
}

// Publish publishes a message to an exchange
func (r *RabbitMQ) Publish(exchange, routingKey string, message Message) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.channel == nil {
		return fmt.Errorf("RabbitMQ channel is not available")
	}
	
	// Set timestamp if not provided
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	
	// Generate ID if not provided
	if message.ID == "" {
		message.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	headers := amqp.Table{}
	if message.Headers != nil {
		for k, v := range message.Headers {
			headers[k] = v
		}
	}
	headers["message_id"] = message.ID
	headers["message_type"] = message.Type
	headers["timestamp"] = message.Timestamp.Unix()
	
	err = r.channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Headers:      headers,
			DeliveryMode: amqp.Persistent, // Make message persistent
			Timestamp:    message.Timestamp,
			MessageId:    message.ID,
		},
	)
	
	if err != nil {
		applog.LogError(err, "Failed to publish message", 
			"exchange", exchange, 
			"routing_key", routingKey, 
			"message_id", message.ID)
		return fmt.Errorf("failed to publish message: %w", err)
	}
	
	applog.Debug("Message published successfully", 
		"exchange", exchange, 
		"routing_key", routingKey, 
		"message_id", message.ID,
		"message_type", message.Type)
	
	return nil
}

// Consume starts consuming messages from a queue
func (r *RabbitMQ) Consume(consumer Consumer) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.channel == nil {
		return fmt.Errorf("RabbitMQ channel is not available")
	}
	
	msgs, err := r.channel.Consume(
		consumer.Queue, // queue
		"",             // consumer
		consumer.AutoAck, // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}
	
	// Start worker goroutines
	for i := 0; i < consumer.Concurrency; i++ {
		go r.worker(msgs, consumer.Handler, consumer.AutoAck)
	}
	
	applog.Info("Consumer started", 
		"queue", consumer.Queue, 
		"concurrency", consumer.Concurrency,
		"auto_ack", consumer.AutoAck)
	
	return nil
}

// worker processes messages
func (r *RabbitMQ) worker(msgs <-chan amqp.Delivery, handler func(Message) error, autoAck bool) {
	for d := range msgs {
		var message Message
		err := json.Unmarshal(d.Body, &message)
		if err != nil {
			applog.LogError(err, "Failed to unmarshal message", "body", string(d.Body))
			if !autoAck {
				d.Nack(false, false) // Don't requeue malformed messages
			}
			continue
		}
		
		// Process message
		err = handler(message)
		if err != nil {
			applog.LogError(err, "Failed to process message", 
				"message_id", message.ID, 
				"message_type", message.Type)
			
			if !autoAck {
				// Check retry count
				if message.Retry < 3 {
					message.Retry++
					// Republish with delay (simple retry mechanism)
					go func() {
						time.Sleep(time.Duration(message.Retry) * time.Second)
						retryBody, _ := json.Marshal(message)
						r.channel.Publish(
							d.Exchange,
							d.RoutingKey,
							false,
							false,
							amqp.Publishing{
								ContentType:  "application/json",
								Body:         retryBody,
								Headers:      d.Headers,
								DeliveryMode: amqp.Persistent,
							},
						)
					}()
				}
				d.Nack(false, false) // Don't requeue, let retry mechanism handle it
			}
			continue
		}
		
		if !autoAck {
			d.Ack(false)
		}
		
		applog.Debug("Message processed successfully", 
			"message_id", message.ID, 
			"message_type", message.Type)
	}
}

// monitorConnection monitors the connection and reconnects if needed
func (r *RabbitMQ) monitorConnection() {
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.conn.NotifyClose(make(chan *amqp.Error)):
			if !r.reconnecting {
				applog.Warn("RabbitMQ connection lost, attempting to reconnect...")
				r.reconnect()
			}
		}
	}
}

// reconnect attempts to reconnect to RabbitMQ
func (r *RabbitMQ) reconnect() {
	r.mu.Lock()
	r.reconnecting = true
	r.mu.Unlock()
	
	defer func() {
		r.mu.Lock()
		r.reconnecting = false
		r.mu.Unlock()
	}()
	
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
			err := r.connect()
			if err != nil {
				applog.LogError(err, "Failed to reconnect to RabbitMQ, retrying in 5 seconds...")
				time.Sleep(5 * time.Second)
				continue
			}
			
			// Reset infrastructure tracking
			r.exchanges = make(map[string]bool)
			r.queues = make(map[string]bool)
			
			// Recreate infrastructure
			err = r.setupInfrastructure()
			if err != nil {
				applog.LogError(err, "Failed to recreate RabbitMQ infrastructure, retrying in 5 seconds...")
				time.Sleep(5 * time.Second)
				continue
			}
			
			applog.Info("Successfully reconnected to RabbitMQ")
			return
		}
	}
}

// Close closes the RabbitMQ connection
func (r *RabbitMQ) Close() error {
	r.cancel()
	
	if r.channel != nil {
		r.channel.Close()
	}
	
	if r.conn != nil {
		return r.conn.Close()
	}
	
	return nil
}

// Health check
func (r *RabbitMQ) IsHealthy() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return r.conn != nil && !r.conn.IsClosed() && r.channel != nil
}
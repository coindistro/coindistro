package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coindistro/backend/internal/uuid"
)

// Message represents a queue message.
type Message struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	Timestamp  time.Time              `json:"timestamp"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
}

// Queue defines the interface for message queue operations.
type Queue interface {
	// Publish publishes a message to the queue.
	Publish(ctx context.Context, topic string, message Message) error
	// Subscribe subscribes to a topic and returns a channel of messages.
	Subscribe(ctx context.Context, topic string) (<-chan Message, error)
	// Ack acknowledges a message as processed.
	Ack(ctx context.Context, topic string, messageID string) error
	// Nack negatively acknowledges a message (requeue or DLQ).
	Nack(ctx context.Context, topic string, messageID string, requeue bool) error
	// Close gracefully shuts down the queue connection.
	Close() error
}

// Handler processes a queue message.
type Handler interface {
	Handle(ctx context.Context, msg Message) error
}

// HandlerFunc is an adapter for using functions as queue handlers.
type HandlerFunc func(ctx context.Context, msg Message) error

// Handle implements the Handler interface.
func (f HandlerFunc) Handle(ctx context.Context, msg Message) error {
	return f(ctx, msg)
}

// ConsumerConfig holds configuration for a queue consumer.
type ConsumerConfig struct {
	Topic       string
	Handler     Handler
	Concurrency int
	MaxRetries  int
	DLQTopic    string
}

// NewMessage creates a new queue message.
func NewMessage(msgType string, payload map[string]interface{}) Message {
	return Message{
		ID:         uuid.NewString(),
		Type:       msgType,
		Payload:    payload,
		Timestamp:  time.Now().UTC(),
		MaxRetries: 3,
	}
}

// MarshalJSON implements json.Marshaler for Message.
func (m Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	return json.Marshal(&struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Timestamp: m.Timestamp.Format(time.RFC3339Nano),
		Alias:     (*Alias)(&m),
	})
}

// Predefined queue topics for the Coindistro platform.
const (
	// Email topics
	TopicEmailSend          = "email.send"
	TopicEmailVerification  = "email.verification"
	TopicEmailPasswordReset = "email.password_reset"

	// Notification topics
	TopicNotificationPush = "notification.push"
	TopicNotificationSMS  = "notification.sms"

	// Trading topics
	TopicSignalBroadcast = "signal.broadcast"
	TopicBotExecution    = "bot.execution"
	TopicTradeSettlement = "trade.settlement"

	// Payment topics
	TopicPaymentProcess    = "payment.process"
	TopicPaymentSettlement = "payment.settlement"
	TopicPaymentRefund     = "payment.refund"

	// Blockchain topics
	TopicBlockchainSync    = "blockchain.sync"
	TopicBlockchainConfirm = "blockchain.confirm"

	// System topics
	TopicSystemCleanup = "system.cleanup"
	TopicSystemReport  = "system.report"
	TopicSystemAlert   = "system.alert"
)

// InMemoryQueue is an in-memory queue implementation for development/testing.
type InMemoryQueue struct {
	topics    map[string][]Message
	consumers map[string][]Handler
}

// NewInMemoryQueue creates a new in-memory queue.
func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{
		topics:    make(map[string][]Message),
		consumers: make(map[string][]Handler),
	}
}

// Publish publishes a message to the queue.
func (q *InMemoryQueue) Publish(ctx context.Context, topic string, message Message) error {
	q.topics[topic] = append(q.topics[topic], message)

	// Process consumers synchronously for in-memory
	if handlers, ok := q.consumers[topic]; ok {
		for _, handler := range handlers {
			if err := handler.Handle(ctx, message); err != nil {
				return err
			}
		}
	}
	return nil
}

// Subscribe subscribes to a topic.
func (q *InMemoryQueue) Subscribe(ctx context.Context, topic string) (<-chan Message, error) {
	ch := make(chan Message, 100)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// Ack acknowledges a message.
func (q *InMemoryQueue) Ack(ctx context.Context, topic string, messageID string) error {
	return nil
}

// Nack negatively acknowledges a message.
func (q *InMemoryQueue) Nack(ctx context.Context, topic string, messageID string, requeue bool) error {
	return nil
}

// Close shuts down the queue.
func (q *InMemoryQueue) Close() error {
	q.topics = nil
	q.consumers = nil
	return nil
}

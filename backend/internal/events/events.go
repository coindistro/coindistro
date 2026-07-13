package events

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/uuid"
)

// Event represents a domain event.
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Handler processes an event.
type Handler interface {
	Handle(ctx context.Context, event Event) error
}

// HandlerFunc is an adapter for using functions as event handlers.
type HandlerFunc func(ctx context.Context, event Event) error

// Handle implements the Handler interface.
func (f HandlerFunc) Handle(ctx context.Context, event Event) error {
	return f(ctx, event)
}

// Bus defines the event bus interface.
type Bus interface {
	// Publish publishes an event to all subscribers.
	Publish(ctx context.Context, event Event) error
	// Subscribe registers a handler for a specific event type.
	Subscribe(eventType string, handler Handler)
	// SubscribeAll registers a handler for all events.
	SubscribeAll(handler Handler)
	// Close gracefully shuts down the event bus.
	Close() error
}

// InMemoryBus is an in-memory event bus implementation.
type InMemoryBus struct {
	mu          sync.RWMutex
	subscribers map[string][]Handler
	allHandlers []Handler
	logger      *zap.Logger
	closed      bool
}

// NewInMemoryBus creates a new in-memory event bus.
func NewInMemoryBus(logger *zap.Logger) *InMemoryBus {
	return &InMemoryBus{
		subscribers: make(map[string][]Handler),
		allHandlers: make([]Handler, 0),
		logger:      logger,
	}
}

// Publish publishes an event to all matching subscribers.
func (b *InMemoryBus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return nil
	}

	// Set event ID if not set
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Publish to type-specific subscribers
	if handlers, ok := b.subscribers[event.Type]; ok {
		for _, handler := range handlers {
			if err := handler.Handle(ctx, event); err != nil {
				b.logger.Error("event handler failed",
					zap.String("event_type", event.Type),
					zap.String("event_id", event.ID),
					zap.Error(err),
				)
			}
		}
	}

	// Publish to all-event subscribers
	for _, handler := range b.allHandlers {
		if err := handler.Handle(ctx, event); err != nil {
			b.logger.Error("global event handler failed",
				zap.String("event_type", event.Type),
				zap.String("event_id", event.ID),
				zap.Error(err),
			)
		}
	}

	return nil
}

// Subscribe registers a handler for a specific event type.
func (b *InMemoryBus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
	b.logger.Info("event subscriber registered", zap.String("event_type", eventType))
}

// SubscribeAll registers a handler for all events.
func (b *InMemoryBus) SubscribeAll(handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.allHandlers = append(b.allHandlers, handler)
	b.logger.Info("global event subscriber registered")
}

// Close gracefully shuts down the event bus.
func (b *InMemoryBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
	b.subscribers = nil
	b.allHandlers = nil
	b.logger.Info("event bus closed")
	return nil
}

// NewEvent creates a new event with the given type and data.
func NewEvent(eventType string, source string, data map[string]interface{}) Event {
	return Event{
		ID:        uuid.NewString(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now().UTC(),
		Data:      data,
		Metadata:  make(map[string]interface{}),
	}
}

// WithMetadata adds metadata to an event.
func WithMetadata(event Event, key string, value interface{}) Event {
	if event.Metadata == nil {
		event.Metadata = make(map[string]interface{})
	}
	event.Metadata[key] = value
	return event
}

// MarshalJSON implements json.Marshaler for Event.
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(&struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Timestamp: e.Timestamp.Format(time.RFC3339Nano),
		Alias:     (*Alias)(&e),
	})
}

// Predefined event types for the Coindistro platform.
const (
	// User events
	EventUserRegistered = "user.registered"
	EventUserVerified   = "user.verified"
	EventUserUpdated    = "user.updated"
	EventUserDeleted    = "user.deleted"
	EventUserLoggedIn   = "user.logged_in"
	EventUserLoggedOut  = "user.logged_out"

	// KYC events
	EventKYCSubmitted = "kyc.submitted"
	EventKYCApproved  = "kyc.approved"
	EventKYCRejected  = "kyc.rejected"

	// Merchant events
	EventMerchantCreated   = "merchant.created"
	EventMerchantApproved  = "merchant.approved"
	EventMerchantSuspended = "merchant.suspended"

	// Wallet events
	EventWalletCreated       = "wallet.created"
	EventDepositCompleted    = "wallet.deposit_completed"
	EventWithdrawalRequested = "wallet.withdrawal_requested"
	EventWithdrawalCompleted = "wallet.withdrawal_completed"

	// Trading events
	EventSignalPublished = "signal.published"
	EventBotStarted      = "bot.started"
	EventBotStopped      = "bot.stopped"
	EventTradeExecuted   = "trade.executed"

	// Payment events
	EventPaymentCompleted = "payment.completed"
	EventPaymentFailed    = "payment.failed"
	EventPaymentRefunded  = "payment.refunded"

	// Notification events
	EventNotificationRequested = "notification.requested"

	// Referral events
	EventReferralCreated    = "referral.created"
	EventInvitationAccepted = "invitation.accepted"
	EventInvitationConsumed = "invitation.consumed"

	// Genesis events
	EventGenesisGranted = "genesis.granted"
	EventFounderGranted = "founder.granted"

	// Password events
	EventPasswordChanged        = "password.changed"
	EventPasswordResetRequested = "password.reset_requested"
	EventPasswordResetCompleted = "password.reset_completed"

	// Email events
	EventEmailVerified = "email.verified"

	// System events
	EventSystemHealthCheck = "system.health_check"
	EventSystemError       = "system.error"
)

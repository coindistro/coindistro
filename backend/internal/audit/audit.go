package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coindistro/backend/internal/uuid"
)

// Action represents an auditable action.
type Action string

const (
	// Authentication actions
	ActionLogin          Action = "login"
	ActionLogout         Action = "logout"
	ActionLoginFailed    Action = "login_failed"
	ActionPasswordChange Action = "password_change"
	ActionEmailChange    Action = "email_change"
	ActionPasswordReset  Action = "password_reset"
	ActionTokenRefresh   Action = "token_refresh"

	// User management actions
	ActionUserCreated      Action = "user_created"
	ActionUserUpdated      Action = "user_updated"
	ActionUserDeleted      Action = "user_deleted"
	ActionRoleUpdate       Action = "role_update"
	ActionPermissionUpdate Action = "permission_update"

	// KYC actions
	ActionKYCSubmitted   Action = "kyc_submitted"
	ActionKYCApproved    Action = "kyc_approved"
	ActionKYCRejected    Action = "kyc_rejected"
	ActionKYCNeedsReview Action = "kyc_needs_review"

	// Merchant actions
	ActionMerchantCreated   Action = "merchant_created"
	ActionMerchantApproved  Action = "merchant_approved"
	ActionMerchantSuspended Action = "merchant_suspended"
	ActionMerchantUpdated   Action = "merchant_updated"

	// Wallet actions
	ActionWalletCreated      Action = "wallet_created"
	ActionDeposit            Action = "deposit"
	ActionWithdrawalRequest  Action = "withdrawal_request"
	ActionWithdrawalApproved Action = "withdrawal_approved"
	ActionWithdrawalRejected Action = "withdrawal_rejected"
	ActionWalletFrozen       Action = "wallet_frozen"
	ActionWalletUnfrozen     Action = "wallet_unfrozen"

	// Trading actions
	ActionSignalPublished Action = "signal_published"
	ActionSignalDeleted   Action = "signal_deleted"
	ActionBotStarted      Action = "bot_started"
	ActionBotStopped      Action = "bot_stopped"
	ActionTradeExecuted   Action = "trade_executed"

	// Admin actions
	ActionAdminAction        Action = "admin_action"
	ActionSettingsChanged    Action = "settings_changed"
	ActionFeatureFlagChanged Action = "feature_flag_changed"

	// Notification actions
	ActionNotificationSent Action = "notification_sent"

	// Earn actions
	ActionEarnProductCreated    Action = "earn_product_created"
	ActionEarnProductUpdated    Action = "earn_product_updated"
	ActionEarnJoin              Action = "earn_join"
	ActionEarnAddFunds          Action = "earn_add_funds"
	ActionEarnWithdraw          Action = "earn_withdraw"
	ActionEarnExit              Action = "earn_exit"
	ActionEarnLaunchpoolCreated Action = "earn_launchpool_created"
	ActionEarnLearnComplete     Action = "earn_learn_complete"
)

// EntityType represents the type of entity being acted upon.
type EntityType string

const (
	EntityUser              EntityType = "user"
	EntityKYC               EntityType = "kyc"
	EntityMerchant          EntityType = "merchant"
	EntityWallet            EntityType = "wallet"
	EntityTransaction       EntityType = "transaction"
	EntitySignal            EntityType = "signal"
	EntityBot               EntityType = "bot"
	EntityCourse            EntityType = "course"
	EntitySettings          EntityType = "settings"
	EntityFeatureFlag       EntityType = "feature_flag"
	EntityAPIKey            EntityType = "api_key"
	EntityRole              EntityType = "role"
	EntityPermission        EntityType = "permission"
	EntityEarnProduct       EntityType = "earn_product"
	EntityEarnParticipation EntityType = "earn_participation"
	EntityEarnLaunchpool    EntityType = "earn_launchpool"
	EntityEarnLearn         EntityType = "earn_learn"
)

// Event represents a single audit log entry.
type Event struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	ActorID    string                 `json:"actor_id"`
	UserID     string                 `json:"user_id,omitempty"`
	Action     Action                 `json:"action"`
	EntityType EntityType             `json:"entity_type,omitempty"`
	EntityID   string                 `json:"entity_id,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Changes    map[string]Change      `json:"changes,omitempty"`
	Outcome    string                 `json:"outcome"` // success, failure, pending
	Error      string                 `json:"error,omitempty"`
}

// Change represents a field-level change.
type Change struct {
	From interface{} `json:"from,omitempty"`
	To   interface{} `json:"to,omitempty"`
}

// Store defines the interface for audit log storage.
type Store interface {
	// Record persists an audit event.
	Record(ctx context.Context, event Event) error
	// Query retrieves audit events based on filters.
	Query(ctx context.Context, filter Filter) ([]Event, error)
	// GetByID retrieves a single audit event by ID.
	GetByID(ctx context.Context, id string) (*Event, error)
}

// Filter defines audit log query filters.
type Filter struct {
	ActorID    string     `json:"actor_id,omitempty"`
	UserID     string     `json:"user_id,omitempty"`
	Action     Action     `json:"action,omitempty"`
	EntityType EntityType `json:"entity_type,omitempty"`
	EntityID   string     `json:"entity_id,omitempty"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	Outcome    string     `json:"outcome,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Offset     int        `json:"offset,omitempty"`
}

// Logger handles audit event creation and storage.
type Logger struct {
	store Store
}

// NewLogger creates a new audit logger.
func NewLogger(store Store) *Logger {
	return &Logger{store: store}
}

// Record creates and stores an audit event.
func (l *Logger) Record(ctx context.Context, event Event) error {
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if event.Outcome == "" {
		event.Outcome = "success"
	}
	return l.store.Record(ctx, event)
}

// Query retrieves audit events.
func (l *Logger) Query(ctx context.Context, filter Filter) ([]Event, error) {
	return l.store.Query(ctx, filter)
}

// GetByID retrieves a single audit event.
func (l *Logger) GetByID(ctx context.Context, id string) (*Event, error) {
	return l.store.GetByID(ctx, id)
}

// NewEvent creates a new audit event builder.
func NewEvent(actorID string, action Action) EventBuilder {
	return EventBuilder{
		event: Event{
			ID:        uuid.NewString(),
			Timestamp: time.Now().UTC(),
			ActorID:   actorID,
			Action:    action,
			Outcome:   "success",
		},
	}
}

// EventBuilder provides a fluent interface for building audit events.
type EventBuilder struct {
	event Event
}

// WithUserID sets the affected user ID.
func (b EventBuilder) WithUserID(userID string) EventBuilder {
	b.event.UserID = userID
	return b
}

// WithEntity sets the entity type and ID.
func (b EventBuilder) WithEntity(entityType EntityType, entityID string) EventBuilder {
	b.event.EntityType = entityType
	b.event.EntityID = entityID
	return b
}

// WithIP sets the IP address.
func (b EventBuilder) WithIP(ip string) EventBuilder {
	b.event.IPAddress = ip
	return b
}

// WithUserAgent sets the user agent.
func (b EventBuilder) WithUserAgent(ua string) EventBuilder {
	b.event.UserAgent = ua
	return b
}

// WithMetadata sets arbitrary metadata.
func (b EventBuilder) WithMetadata(meta map[string]interface{}) EventBuilder {
	b.event.Metadata = meta
	return b
}

// WithChange records a field-level change.
func (b EventBuilder) WithChange(field string, from, to interface{}) EventBuilder {
	if b.event.Changes == nil {
		b.event.Changes = make(map[string]Change)
	}
	b.event.Changes[field] = Change{From: from, To: to}
	return b
}

// WithOutcome sets the outcome.
func (b EventBuilder) WithOutcome(outcome string) EventBuilder {
	b.event.Outcome = outcome
	return b
}

// WithError sets the error message.
func (b EventBuilder) WithError(err string) EventBuilder {
	b.event.Error = err
	b.event.Outcome = "failure"
	return b
}

// Build returns the constructed event.
func (b EventBuilder) Build() Event {
	return b.event
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

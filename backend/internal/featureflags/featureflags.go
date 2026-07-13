package featureflags

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Flag represents a feature flag.
type Flag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
	Environment string `json:"environment,omitempty"`
}

// Manager manages feature flags.
type Manager struct {
	mu          sync.RWMutex
	flags       map[string]*Flag
	logger      *zap.Logger
	environment string
}

// New creates a new feature flag manager.
func New(logger *zap.Logger, environment string) *Manager {
	m := &Manager{
		flags:       make(map[string]*Flag),
		logger:      logger,
		environment: environment,
	}
	m.registerDefaultFlags()
	return m
}

// IsEnabled checks if a feature flag is enabled.
func (m *Manager) IsEnabled(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	flag, ok := m.flags[name]
	if !ok {
		return false
	}
	return flag.Enabled
}

// Enable enables a feature flag.
func (m *Manager) Enable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	flag, ok := m.flags[name]
	if !ok {
		return fmt.Errorf("feature flag not found: %s", name)
	}

	flag.Enabled = true
	m.logger.Info("feature flag enabled", zap.String("flag", name))
	return nil
}

// Disable disables a feature flag.
func (m *Manager) Disable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	flag, ok := m.flags[name]
	if !ok {
		return fmt.Errorf("feature flag not found: %s", name)
	}

	flag.Enabled = false
	m.logger.Info("feature flag disabled", zap.String("flag", name))
	return nil
}

// Set sets a feature flag's enabled state.
func (m *Manager) Set(name string, enabled bool) error {
	if enabled {
		return m.Enable(name)
	}
	return m.Disable(name)
}

// GetFlag returns a feature flag by name.
func (m *Manager) GetFlag(name string) (*Flag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	flag, ok := m.flags[name]
	if !ok {
		return nil, fmt.Errorf("feature flag not found: %s", name)
	}
	return flag, nil
}

// GetAllFlags returns all feature flags.
func (m *Manager) GetAllFlags() []Flag {
	m.mu.RLock()
	defer m.mu.RUnlock()

	flags := make([]Flag, 0, len(m.flags))
	for _, flag := range m.flags {
		flags = append(flags, *flag)
	}
	return flags
}

// RegisterFlag registers a new feature flag.
func (m *Manager) RegisterFlag(flag Flag) {
	m.mu.Lock()
	defer m.mu.Unlock()

	flag.Environment = m.environment
	m.flags[flag.Name] = &flag
	m.logger.Info("feature flag registered",
		zap.String("flag", flag.Name),
		zap.Bool("enabled", flag.Enabled),
	)
}

// RegisterFlags registers multiple feature flags at once.
func (m *Manager) RegisterFlags(flags []Flag) {
	for _, flag := range flags {
		m.RegisterFlag(flag)
	}
}

// LoadFromConfig loads feature flags from a configuration map.
func (m *Manager) LoadFromConfig(ctx context.Context, config map[string]bool) {
	for name, enabled := range config {
		if flag, err := m.GetFlag(name); err == nil {
			flag.Enabled = enabled
		} else {
			m.RegisterFlag(Flag{
				Name:    name,
				Enabled: enabled,
			})
		}
	}
}

func (m *Manager) registerDefaultFlags() {
	defaultFlags := []Flag{
		{Name: "exchange.enabled", Description: "Enable exchange trading features", Enabled: true},
		{Name: "academy.enabled", Description: "Enable academy/education features", Enabled: true},
		{Name: "signals.enabled", Description: "Enable trading signals", Enabled: true},
		{Name: "merchant.enabled", Description: "Enable merchant services", Enabled: true},
		{Name: "bots.enabled", Description: "Enable trading bots", Enabled: true},
		{Name: "payments.enabled", Description: "Enable payment processing", Enabled: true},
		{Name: "bank.enabled", Description: "Enable banking features", Enabled: false},
		{Name: "gift_cards.enabled", Description: "Enable gift card services", Enabled: true},
		{Name: "invest.enabled", Description: "Enable investment/staking features", Enabled: true},
		{Name: "p2p.enabled", Description: "Enable peer-to-peer trading", Enabled: true},
		{Name: "fiat_ramp.enabled", Description: "Enable fiat on/off ramp", Enabled: true},
		{Name: "notifications.enabled", Description: "Enable notification system", Enabled: true},
		{Name: "audit_logging.enabled", Description: "Enable audit logging", Enabled: true},
		{Name: "analytics.enabled", Description: "Enable analytics and reporting", Enabled: true},
		{Name: "maintenance_mode", Description: "Put platform in maintenance mode", Enabled: false},
	}

	for _, flag := range defaultFlags {
		flag.Environment = m.environment
		m.flags[flag.Name] = &flag
	}
}

// Predefined feature flag names for the Coindistro platform.
const (
	FlagExchange      = "exchange.enabled"
	FlagAcademy       = "academy.enabled"
	FlagSignals       = "signals.enabled"
	FlagMerchant      = "merchant.enabled"
	FlagBots          = "bots.enabled"
	FlagPayments      = "payments.enabled"
	FlagBank          = "bank.enabled"
	FlagGiftCards     = "gift_cards.enabled"
	FlagInvest        = "invest.enabled"
	FlagP2P           = "p2p.enabled"
	FlagFiatRamp      = "fiat_ramp.enabled"
	FlagNotifications = "notifications.enabled"
	FlagAuditLogging  = "audit_logging.enabled"
	FlagAnalytics     = "analytics.enabled"
	FlagMaintenance   = "maintenance_mode"

	// Identity / Registration flags
	FlagRegistration      = "registration.enabled"
	FlagRequiresReferral  = "registration.requires_referral"
	FlagInviteOnly        = "registration.invite_only"
	FlagEmailVerification = "registration.email_verification"
	FlagAutoVerify        = "registration.auto_verify"
	FlagSocialLogin       = "registration.allow_social_login"
	FlagGenesis           = "genesis.enabled"
)

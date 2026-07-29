package seed

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/config"
	"github.com/coindistro/backend/internal/database"
	"github.com/coindistro/backend/internal/identity/store"
)

// DemoAccounts is the canonical set of development role accounts.
func DemoAccounts() []SeedUser {
	genesisSlot := 1
	return []SeedUser{
		{
			Label:            "Super Admin",
			Email:            "admin@coindistro.com",
			Password:         "Admin123!",
			DisplayName:      "Super Admin",
			Username:         "superadmin",
			Roles:            []string{"super_admin", "user"},
			Status:           "active",
			EmailVerified:    true,
			PhoneVerified:    true,
			IsGenesis:        true,
			GenesisNumber:    &genesisSlot,
			IsFounder:        true,
			KYCStatus:        "approved",
			ProvisionWallets: true,
			Country:          "NGA",
			Timezone:         "Africa/Lagos",
		},
		{
			Label:            "Platform Admin",
			Email:            "platform@coindistro.com",
			Password:         "Platform123!",
			DisplayName:      "Platform Admin",
			Username:         "platformadmin",
			Roles:            []string{"admin", "user"},
			Status:           "active",
			EmailVerified:    true,
			PhoneVerified:    true,
			KYCStatus:        "approved",
			ProvisionWallets: true,
			Country:          "NGA",
			Timezone:         "Africa/Lagos",
		},
		{
			Label:            "Moderator",
			Email:            "moderator@coindistro.com",
			Password:         "Moderator123!",
			DisplayName:      "Moderator",
			Username:         "moderator",
			Roles:            []string{"moderator", "user"},
			Status:           "active",
			EmailVerified:    true,
			ProvisionWallets: true,
			Country:          "NGA",
			Timezone:         "Africa/Lagos",
		},
		{
			Label:            "Merchant",
			Email:            "merchant@coindistro.com",
			Password:         "Merchant123!",
			DisplayName:      "Merchant Demo",
			Username:         "merchantdemo",
			Roles:            []string{"merchant", "user"},
			Status:           "active",
			EmailVerified:    true,
			KYCStatus:        "approved",
			ProvisionWallets: true,
			Country:          "NGA",
			Timezone:         "Africa/Lagos",
			Merchant: &MerchantSeed{
				BusinessName:       "CoinDistro Store",
				Status:             "approved",
				BusinessVerified:   true,
				CanReceivePayments: true,
			},
		},
		{
			Label:            "Test User",
			Email:            "user@coindistro.com",
			Password:         "User123!",
			DisplayName:      "Test User",
			Username:         "testuser",
			Roles:            []string{"user"},
			ReferralCode:     "TESTUSER",
			Status:           "active",
			EmailVerified:    true,
			PhoneVerified:    true,
			KYCStatus:        "approved",
			ProvisionWallets: true,
			Country:          "NGA",
			Timezone:         "Africa/Lagos",
		},
		{
			Label:            "Genesis Member",
			Email:            "genesis@coindistro.com",
			Password:         "Genesis123!",
			DisplayName:      "Genesis Member",
			Username:         "genesismember",
			Roles:            []string{"user"},
			ReferralCode:     "GENESIS",
			Status:           "active",
			EmailVerified:    true,
			IsGenesis:        true,
			GenesisNumber:    &genesisSlot,
			KYCStatus:        "approved",
			ProvisionWallets: true,
			Country:          "NGA",
			Timezone:         "Africa/Lagos",
		},
	}
}

// ShouldSeedDemoUsers decides whether demo accounts should run.
// - Production: only when COINDISTRO_ALLOW_PRODUCTION_DEMO_USERS=true
// - Non-production: default true unless COINDISTRO_SEED_DEMO_USERS=false
func ShouldSeedDemoUsers(cfg *config.Config) bool {
	if cfg != nil && cfg.App.IsProduction() {
		return envTruthy("COINDISTRO_ALLOW_PRODUCTION_DEMO_USERS")
	}
	// Explicit opt-out
	v := strings.ToLower(strings.TrimSpace(os.Getenv("COINDISTRO_SEED_DEMO_USERS")))
	if v == "false" || v == "0" || v == "no" || v == "off" {
		return false
	}
	// Explicit opt-in or default on for development
	if v == "" || v == "true" || v == "1" || v == "yes" || v == "on" {
		return true
	}
	return false
}

// SeedDemoAccounts ensures all canonical development accounts exist.
func SeedDemoAccounts(ctx context.Context, db *database.Database, logger *zap.Logger) error {
	if logger == nil {
		logger = zap.NewNop()
	}
	if db == nil || db.Pool == nil {
		return fmt.Errorf("database is required")
	}

	identityStore := store.New(db.Pool)
	accounts := DemoAccounts()
	created := 0
	for _, acc := range accounts {
		res, err := EnsureUser(ctx, db.Pool, identityStore, logger, acc)
		if err != nil {
			return fmt.Errorf("%s: %w", acc.Label, err)
		}
		if res.Created {
			logger.Info(fmt.Sprintf("✓ %s created", acc.Label), zap.String("email", acc.Email))
			created++
		} else {
			logger.Info(fmt.Sprintf("✓ %s exists", acc.Label), zap.String("email", acc.Email))
		}
	}
	logger.Info(fmt.Sprintf("Seeded %d development accounts successfully.", len(accounts)),
		zap.Int("created", created),
		zap.Int("total", len(accounts)),
	)
	return nil
}

// PoolForTests exposes ensure helpers for tests without a full Database wrapper.
func EnsureUserWithPool(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger, in SeedUser) (*EnsureResult, error) {
	return EnsureUser(ctx, pool, store.New(pool), logger, in)
}

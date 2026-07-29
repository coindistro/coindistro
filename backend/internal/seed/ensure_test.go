package seed

import (
	"os"
	"testing"
	"time"

	"github.com/coindistro/backend/internal/auth"
	"github.com/coindistro/backend/internal/config"
	"github.com/coindistro/backend/internal/identity/store"
)

func TestShouldSeedDemoUsers_ProductionBlocked(t *testing.T) {
	t.Setenv("COINDISTRO_ALLOW_PRODUCTION_DEMO_USERS", "")
	t.Setenv("COINDISTRO_SEED_DEMO_USERS", "true")
	cfg := &config.Config{}
	cfg.App.Environment = "production"
	if ShouldSeedDemoUsers(cfg) {
		t.Fatal("production must not seed without allow flag")
	}
}

func TestShouldSeedDemoUsers_ProductionAllowed(t *testing.T) {
	t.Setenv("COINDISTRO_ALLOW_PRODUCTION_DEMO_USERS", "true")
	cfg := &config.Config{}
	cfg.App.Environment = "production"
	if !ShouldSeedDemoUsers(cfg) {
		t.Fatal("expected allow flag to enable production seed")
	}
}

func TestShouldSeedDemoUsers_DevDefault(t *testing.T) {
	t.Setenv("COINDISTRO_SEED_DEMO_USERS", "")
	t.Setenv("COINDISTRO_ALLOW_PRODUCTION_DEMO_USERS", "")
	cfg := &config.Config{}
	cfg.App.Environment = "development"
	if !ShouldSeedDemoUsers(cfg) {
		t.Fatal("development should seed by default")
	}
}

func TestShouldSeedDemoUsers_DevOptOut(t *testing.T) {
	t.Setenv("COINDISTRO_SEED_DEMO_USERS", "false")
	cfg := &config.Config{}
	cfg.App.Environment = "development"
	if ShouldSeedDemoUsers(cfg) {
		t.Fatal("expected opt-out")
	}
}

func TestDemoAccounts_RolesAndPasswords(t *testing.T) {
	accounts := DemoAccounts()
	if len(accounts) != 6 {
		t.Fatalf("want 6 accounts, got %d", len(accounts))
	}

	byEmail := map[string]SeedUser{}
	for _, a := range accounts {
		byEmail[a.Email] = a
		hash, err := auth.HashPassword(a.Password)
		if err != nil {
			t.Fatalf("hash %s: %v", a.Email, err)
		}
		if err := auth.VerifyPassword(a.Password, hash); err != nil {
			t.Fatalf("verify %s: %v", a.Email, err)
		}
	}

	if !containsRole(byEmail["admin@coindistro.com"].Roles, "super_admin") {
		t.Fatal("super admin role missing")
	}
	if !containsRole(byEmail["platform@coindistro.com"].Roles, "admin") {
		t.Fatal("admin role missing")
	}
	if !containsRole(byEmail["moderator@coindistro.com"].Roles, "moderator") {
		t.Fatal("moderator role missing")
	}
	if !containsRole(byEmail["merchant@coindistro.com"].Roles, "merchant") {
		t.Fatal("merchant role missing")
	}
	if byEmail["merchant@coindistro.com"].Merchant == nil {
		t.Fatal("merchant profile required")
	}
	if byEmail["merchant@coindistro.com"].Merchant.BusinessName != "CoinDistro Store" {
		t.Fatal("business name mismatch")
	}
	if byEmail["user@coindistro.com"].ReferralCode != "TESTUSER" {
		t.Fatal("user referral code")
	}
	if !byEmail["genesis@coindistro.com"].IsGenesis {
		t.Fatal("genesis flag")
	}
	if byEmail["genesis@coindistro.com"].GenesisNumber == nil || *byEmail["genesis@coindistro.com"].GenesisNumber != 1 {
		t.Fatal("genesis slot")
	}
	if !byEmail["user@coindistro.com"].ProvisionWallets {
		t.Fatal("wallets should be provisioned for standard user")
	}
	if len(DefaultWalletCurrencies) != 5 {
		t.Fatal("expected 5 default wallets")
	}
}

func TestUnionRolesIdempotent(t *testing.T) {
	a := []string{"user", "admin"}
	b := []string{"admin", "user"}
	if !sameStringSet(unionRoles(a, b), []string{"user", "admin"}) {
		t.Fatal("union roles")
	}
}

func TestForceSeedPasswordsEnv(t *testing.T) {
	_ = os.Unsetenv("FORCE_SEED_PASSWORDS")
	if envTruthy("FORCE_SEED_PASSWORDS") {
		t.Fatal("default false")
	}
	t.Setenv("FORCE_SEED_PASSWORDS", "true")
	if !envTruthy("FORCE_SEED_PASSWORDS") {
		t.Fatal("expected true")
	}
}

func TestBuildUserModel_SetsVerificationAndGenesis(t *testing.T) {
	n := 1
	now := time.Now().UTC()
	in := SeedUser{
		DisplayName:   "X",
		Username:      "xuser",
		Roles:         []string{"user"},
		ReferralCode:  "ABC",
		EmailVerified: true,
		PhoneVerified: true,
		IsGenesis:     true,
		GenesisNumber: &n,
		Country:       "NGA",
	}
	u := buildUserModel(in, "x@coindistro.com", "hash", now)
	if u.EmailVerifiedAt == nil || u.PhoneVerifiedAt == nil {
		t.Fatal("verification timestamps")
	}
	if !u.IsGenesis || u.GenesisNumber == nil || *u.GenesisNumber != 1 {
		t.Fatal("genesis fields")
	}
	if u.ReferralCode != "ABC" {
		t.Fatal("referral code")
	}
}

func TestSameStringSet(t *testing.T) {
	if !sameStringSet([]string{"a", "b"}, []string{"b", "a"}) {
		t.Fatal("order independent")
	}
	if sameStringSet([]string{"a"}, []string{"a", "b"}) {
		t.Fatal("different lengths")
	}
}

// Document that EnsureUser relies on store.ErrUserNotFound.
var _ = store.ErrUserNotFound

func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

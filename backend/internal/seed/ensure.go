package seed

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/auth"
	"github.com/coindistro/backend/internal/identity/models"
	"github.com/coindistro/backend/internal/identity/store"
	uuidlib "github.com/coindistro/backend/internal/uuid"
)

// MerchantSeed describes optional merchant profile fields.
type MerchantSeed struct {
	BusinessName       string
	Status             string // approved maps to active + is_verified
	BusinessVerified   bool
	CanReceivePayments bool
}

// SeedUser is the declarative input for EnsureUser.
type SeedUser struct {
	Label            string // log label e.g. "Super Admin"
	Email            string
	Password         string
	DisplayName      string
	Username         string
	Roles            []string
	ReferralCode     string // optional fixed code
	Status           string
	EmailVerified    bool
	PhoneVerified    bool
	IsGenesis        bool
	GenesisNumber    *int
	IsFounder        bool
	KYCStatus        string // approved | pending | ""
	Merchant         *MerchantSeed
	ProvisionWallets bool
	Country          string
	Timezone         string
}

// EnsureResult describes what EnsureUser did.
type EnsureResult struct {
	User    *models.User
	Created bool
	Updated bool
	Label   string
}

// DefaultWalletCurrencies provisioned for seed users when ProvisionWallets is true.
var DefaultWalletCurrencies = []string{"USD", "NGN", "USDT", "BTC", "ETH"}

// EnsureUser creates or updates a seeded user idempotently.
// Passwords are never overwritten unless FORCE_SEED_PASSWORDS=true.
func EnsureUser(ctx context.Context, pool *pgxpool.Pool, identityStore *store.Store, logger *zap.Logger, in SeedUser) (*EnsureResult, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	if pool == nil || identityStore == nil {
		return nil, fmt.Errorf("pool and identity store are required")
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || in.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}
	if in.Status == "" {
		in.Status = "active"
	}
	if in.Timezone == "" {
		in.Timezone = "UTC"
	}
	if len(in.Roles) == 0 {
		in.Roles = []string{"user"}
	}

	existing, err := identityStore.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, store.ErrUserNotFound) {
		return nil, fmt.Errorf("lookup %s: %w", email, err)
	}

	forcePassword := envTruthy("FORCE_SEED_PASSWORDS")
	now := time.Now().UTC()

	if existing == nil || errors.Is(err, store.ErrUserNotFound) {
		hash, err := auth.HashPassword(in.Password)
		if err != nil {
			return nil, err
		}

		user := buildUserModel(in, email, hash, now)
		if user.ReferralCode == "" {
			user.ReferralCode = randomReferralCode()
		}
		// Resolve fixed referral code conflicts
		if err := ensureUniqueReferral(ctx, identityStore, user, ""); err != nil {
			return nil, err
		}

		if err := identityStore.CreateUserFull(ctx, user); err != nil {
			return nil, fmt.Errorf("create user %s: %w", email, err)
		}

		_ = identityStore.UpsertInvitationCredits(ctx, &models.InvitationCredit{
			ID:           uuidlib.NewString(),
			UserID:       user.ID,
			TotalCredits: 25,
		})

		if err := provisionPlatform(ctx, pool, user.ID, in, now); err != nil {
			return nil, err
		}

		if in.IsGenesis {
			_, _ = identityStore.IncrementGenesisCount(ctx)
		}

		logger.Info("seed user created", zap.String("label", in.Label), zap.String("email", email))
		return &EnsureResult{User: user, Created: true, Label: in.Label}, nil
	}

	// Update existing
	updated := false
	user := existing

	// Merge roles (union)
	mergedRoles := unionRoles(user.Roles, in.Roles)
	if !sameStringSet(user.Roles, mergedRoles) {
		user.Roles = mergedRoles
		updated = true
	}

	if in.DisplayName != "" {
		d := in.DisplayName
		if user.DisplayName == nil || *user.DisplayName != d {
			user.DisplayName = &d
			updated = true
		}
	}
	if in.Username != "" {
		u := in.Username
		if user.Username == nil || *user.Username != u {
			// only set if free or already ours
			taken, _ := identityStore.IsUsernameTaken(ctx, u)
			if !taken || (user.Username != nil && *user.Username == u) {
				user.Username = &u
				updated = true
			}
		}
	}
	if in.ReferralCode != "" && user.ReferralCode != in.ReferralCode {
		taken, _ := identityStore.IsReferralCodeTaken(ctx, in.ReferralCode, user.ID)
		if !taken {
			user.ReferralCode = in.ReferralCode
			updated = true
		}
	}
	user.Status = in.Status
	user.Timezone = in.Timezone
	if in.Country != "" {
		c := in.Country
		user.Country = &c
	}
	if in.EmailVerified && user.EmailVerifiedAt == nil {
		user.EmailVerifiedAt = &now
		updated = true
	}
	if in.PhoneVerified && user.PhoneVerifiedAt == nil {
		user.PhoneVerifiedAt = &now
		updated = true
	}
	if in.IsGenesis {
		if !user.IsGenesis {
			user.IsGenesis = true
			user.GenesisDate = &now
			updated = true
		}
		if in.GenesisNumber != nil {
			user.GenesisNumber = in.GenesisNumber
			updated = true
		}
	}
	if in.IsFounder {
		user.IsFounder = true
		user.FounderBadge = true
		updated = true
	}

	if err := identityStore.ApplySeedProfile(ctx, user); err != nil {
		return nil, fmt.Errorf("update profile %s: %w", email, err)
	}

	if forcePassword {
		hash, err := auth.HashPassword(in.Password)
		if err != nil {
			return nil, err
		}
		if err := identityStore.UpdatePassword(ctx, user.ID, hash); err != nil {
			return nil, fmt.Errorf("force password %s: %w", email, err)
		}
		updated = true
		logger.Info("seed password forced", zap.String("email", email))
	}

	if err := provisionPlatform(ctx, pool, user.ID, in, now); err != nil {
		return nil, err
	}

	// Reload
	user, _ = identityStore.GetUserByEmail(ctx, email)
	logger.Info("seed user ensured",
		zap.String("label", in.Label),
		zap.String("email", email),
		zap.Bool("updated", updated),
	)
	return &EnsureResult{User: user, Created: false, Updated: updated, Label: in.Label}, nil
}

func buildUserModel(in SeedUser, email, hash string, now time.Time) *models.User {
	user := &models.User{
		ID:            uuidlib.NewString(),
		Email:         email,
		PasswordHash:  hash,
		Timezone:      in.Timezone,
		Locale:        "en",
		ReferralCode:  strings.ToUpper(strings.TrimSpace(in.ReferralCode)),
		Status:        in.Status,
		Roles:         append([]string{}, in.Roles...),
		IsGenesis:     in.IsGenesis,
		IsFounder:     in.IsFounder,
		FounderBadge:  in.IsFounder,
		GenesisNumber: in.GenesisNumber,
		LastLoginAt:   &now,
	}
	if in.Username != "" {
		u := in.Username
		user.Username = &u
	}
	if in.DisplayName != "" {
		d := in.DisplayName
		user.DisplayName = &d
	}
	if in.Country != "" {
		c := in.Country
		user.Country = &c
	}
	if in.EmailVerified {
		user.EmailVerifiedAt = &now
	}
	if in.PhoneVerified {
		user.PhoneVerifiedAt = &now
	}
	if in.IsGenesis {
		user.GenesisDate = &now
	}
	return user
}

func ensureUniqueReferral(ctx context.Context, st *store.Store, user *models.User, excludeID string) error {
	if user.ReferralCode == "" {
		user.ReferralCode = randomReferralCode()
	}
	for i := 0; i < 8; i++ {
		taken, err := st.IsReferralCodeTaken(ctx, user.ReferralCode, excludeID)
		if err != nil {
			return err
		}
		if !taken {
			return nil
		}
		// If fixed code conflicted, append suffix once then randomize
		if i == 0 && len(user.ReferralCode) < 16 {
			user.ReferralCode = user.ReferralCode + "X"
			continue
		}
		user.ReferralCode = randomReferralCode()
	}
	return fmt.Errorf("could not allocate unique referral code")
}

func randomReferralCode() string {
	return strings.ToUpper(uuidlib.NewString()[:8])
}

func unionRoles(a, b []string) []string {
	m := map[string]struct{}{}
	for _, r := range a {
		m[r] = struct{}{}
	}
	for _, r := range b {
		m[r] = struct{}{}
	}
	out := make([]string, 0, len(m))
	for r := range m {
		out = append(out, r)
	}
	return out
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := map[string]struct{}{}
	for _, x := range a {
		m[x] = struct{}{}
	}
	for _, x := range b {
		if _, ok := m[x]; !ok {
			return false
		}
	}
	return true
}

func envTruthy(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func provisionPlatform(ctx context.Context, pool *pgxpool.Pool, userID string, in SeedUser, now time.Time) error {
	if in.ProvisionWallets {
		if err := ensureWallets(ctx, pool, userID); err != nil {
			return fmt.Errorf("wallets: %w", err)
		}
	}
	if in.Merchant != nil {
		if err := ensureMerchant(ctx, pool, userID, in.Merchant); err != nil {
			return fmt.Errorf("merchant: %w", err)
		}
	}
	if in.KYCStatus != "" {
		if err := ensureKYC(ctx, pool, userID, in.KYCStatus, now); err != nil {
			return fmt.Errorf("kyc: %w", err)
		}
	}
	return nil
}

func ensureWallets(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	for _, cur := range DefaultWalletCurrencies {
		var id string
		err := pool.QueryRow(ctx,
			`SELECT id FROM wallets WHERE user_id = $1 AND currency = $2`, userID, cur,
		).Scan(&id)
		if err == nil {
			continue
		}
		if err != pgx.ErrNoRows {
			// table may not exist in some test environments
			if strings.Contains(err.Error(), "does not exist") {
				return nil
			}
			// unique race — ignore
			if !strings.Contains(err.Error(), "no rows") {
				// try insert anyway
			}
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO wallets (id, user_id, currency, address, balance, locked_balance, is_active)
			VALUES ($1, $2, $3, $4, 0, 0, true)
			ON CONFLICT (user_id, currency) DO NOTHING`,
			uuidlib.NewString(), userID, cur, fmt.Sprintf("seed-%s-%s", cur, userID[:8]),
		)
		if err != nil && !strings.Contains(err.Error(), "does not exist") {
			return err
		}
	}
	return nil
}

func ensureMerchant(ctx context.Context, pool *pgxpool.Pool, userID string, m *MerchantSeed) error {
	name := m.BusinessName
	if name == "" {
		name = "CoinDistro Store"
	}
	status := m.Status
	if status == "approved" || status == "" {
		status = "active"
	}
	verified := m.BusinessVerified
	if m.Status == "approved" {
		verified = true
	}

	var id string
	err := pool.QueryRow(ctx, `SELECT id FROM merchant_accounts WHERE user_id = $1`, userID).Scan(&id)
	if err == nil {
		_, err = pool.Exec(ctx, `
			UPDATE merchant_accounts SET
				business_name = $2, status = $3, is_verified = $4, updated_at = NOW()
			WHERE user_id = $1`, userID, name, status, verified)
		return ignoreMissingTable(err)
	}
	if err != pgx.ErrNoRows && !strings.Contains(err.Error(), "no rows") {
		return ignoreMissingTable(err)
	}

	apiKey := "mk_seed_" + uuidlib.NewString()[:16]
	apiSecret := "ms_seed_" + uuidlib.NewString()
	_, err = pool.Exec(ctx, `
		INSERT INTO merchant_accounts (
			id, user_id, business_name, business_email, status, api_key, api_secret, is_verified
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) DO UPDATE SET
			business_name = EXCLUDED.business_name,
			status = EXCLUDED.status,
			is_verified = EXCLUDED.is_verified,
			updated_at = NOW()`,
		uuidlib.NewString(), userID, name, "", status, apiKey, apiSecret, verified,
	)
	return ignoreMissingTable(err)
}

func ensureKYC(ctx context.Context, pool *pgxpool.Pool, userID, status string, now time.Time) error {
	if status == "" {
		return nil
	}
	var id string
	err := pool.QueryRow(ctx,
		`SELECT id FROM kyc_submissions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, userID,
	).Scan(&id)
	if err == nil {
		_, err = pool.Exec(ctx, `
			UPDATE kyc_submissions SET status = $2, reviewed_at = $3, updated_at = NOW() WHERE id = $1`,
			id, status, now,
		)
		return ignoreMissingTable(err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO kyc_submissions (id, user_id, status, id_type, submitted_at, reviewed_at)
		VALUES ($1, $2, $3, 'passport', $4, $5)`,
		uuidlib.NewString(), userID, status, now, now,
	)
	return ignoreMissingTable(err)
}

func ignoreMissingTable(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "does not exist") {
		return nil
	}
	return err
}

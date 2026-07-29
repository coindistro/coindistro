package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"

	apperrors "github.com/coindistro/backend/internal/errors"
	"github.com/coindistro/backend/internal/featureflags"
	ide "github.com/coindistro/backend/internal/identity/errors"
	"github.com/coindistro/backend/internal/identity/models"
)

func testFlags(t *testing.T, registrationEnabled, inviteOnly bool) *featureflags.Manager {
	t.Helper()
	ff := featureflags.New(zap.NewNop(), "test")
	_ = ff.Set(featureflags.FlagRegistration, registrationEnabled)
	_ = ff.Set(featureflags.FlagInviteOnly, inviteOnly)
	_ = ff.Set(featureflags.FlagRequiresReferral, false)
	return ff
}

func TestRegistrationFlagDefaultEnabled(t *testing.T) {
	ff := featureflags.New(zap.NewNop(), "test")
	if !ff.IsEnabled(featureflags.FlagRegistration) {
		t.Fatal("registration.enabled should default to true for public onboarding")
	}
	if ff.IsEnabled(featureflags.FlagInviteOnly) {
		t.Fatal("registration.invite_only should default to false")
	}
}

func TestRegister_Disabled(t *testing.T) {
	ff := testFlags(t, false, false)
	svc := New(nil, nil, nil, nil, nil, nil, nil, ff, nil, nil, zap.NewNop(), DefaultConfig())

	_, err := svc.Register(context.Background(), &models.RegisterRequest{
		Email:        "user@example.com",
		Password:     "Password1!",
		ReferralCode: "ABC12345",
	}, "127.0.0.1", "test-agent")

	appErr := apperrors.GetAppError(err)
	if appErr == nil {
		t.Fatal("expected AppError when registration is disabled")
	}
	if appErr.Code != "REGISTRATION_DISABLED" {
		t.Fatalf("code = %q, want REGISTRATION_DISABLED", appErr.Code)
	}
	if appErr.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", appErr.StatusCode)
	}
}

func TestRegister_EnabledDoesNotReturnDisabled(t *testing.T) {
	ff := testFlags(t, true, false)
	svc := New(nil, nil, nil, nil, nil, nil, nil, ff, nil, nil, zap.NewNop(), DefaultConfig())

	// Store is nil → ensureEmailAvailable returns internal error, not REGISTRATION_DISABLED.
	_, err := svc.Register(context.Background(), &models.RegisterRequest{
		Email:        "user@example.com",
		Password:     "Password1!",
		ReferralCode: "ABC12345",
	}, "127.0.0.1", "test-agent")

	appErr := apperrors.GetAppError(err)
	if appErr == nil {
		t.Fatal("expected error from nil store after registration gate")
	}
	if appErr.Code == "REGISTRATION_DISABLED" {
		t.Fatal("registration is enabled; must not return REGISTRATION_DISABLED")
	}
}

func TestErrRegistrationDisabledIs403(t *testing.T) {
	appErr := apperrors.GetAppError(ide.ErrRegistrationDisabled)
	if appErr == nil {
		t.Fatal("expected AppError")
	}
	if appErr.Code != "REGISTRATION_DISABLED" {
		t.Fatalf("code = %q", appErr.Code)
	}
	if appErr.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", appErr.StatusCode)
	}
}

// ─── Email availability ────────────────────────────────

type stubEmailStore struct {
	taken bool
	err   error
}

func (s stubEmailStore) IsEmailTaken(ctx context.Context, email string) (bool, error) {
	return s.taken, s.err
}

func TestEnsureEmailAvailable_OK(t *testing.T) {
	err := ensureEmailAvailable(context.Background(), stubEmailStore{taken: false}, "new@example.com")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestEnsureEmailAvailable_DuplicateEmail(t *testing.T) {
	err := ensureEmailAvailable(context.Background(), stubEmailStore{taken: true}, "dup@example.com")
	appErr := apperrors.GetAppError(err)
	if appErr == nil || appErr.Code != "EMAIL_ALREADY_EXISTS" {
		t.Fatalf("got %v, want EMAIL_ALREADY_EXISTS", err)
	}
	if appErr.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", appErr.StatusCode)
	}
}

func TestEnsureEmailAvailable_StoreError(t *testing.T) {
	err := ensureEmailAvailable(context.Background(), stubEmailStore{err: errors.New("db down")}, "a@b.com")
	appErr := apperrors.GetAppError(err)
	if appErr == nil || appErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("got %v, want 500", err)
	}
}

// ─── Referral resolution ───────────────────────────────

type stubReferralStore struct {
	user *models.User
	inv  *models.Invitation
	err  error
}

func (s stubReferralStore) GetUserByReferralCode(ctx context.Context, code string) (*models.User, error) {
	return s.user, s.err
}

func (s stubReferralStore) GetInvitationByCode(ctx context.Context, code string) (*models.Invitation, error) {
	return s.inv, s.err
}

func TestResolveReferral_EmptyIsDirect(t *testing.T) {
	ref, inv, method, err := resolveReferralCode(context.Background(), stubReferralStore{}, "")
	if err != nil || ref != nil || inv != nil || method != "direct" {
		t.Fatalf("got ref=%v inv=%v method=%q err=%v", ref, inv, method, err)
	}
}

func TestResolveReferral_Invalid(t *testing.T) {
	_, _, _, err := resolveReferralCode(context.Background(), stubReferralStore{}, "NOTREAL1")
	appErr := apperrors.GetAppError(err)
	if appErr == nil || appErr.Code != "INVALID_REFERRAL_CODE" {
		t.Fatalf("got %v, want INVALID_REFERRAL_CODE", err)
	}
}

func TestResolveReferral_ValidUserReferral(t *testing.T) {
	referrer := &models.User{ID: "ref-1", ReferralCode: "VALID001", ReferralLevel: 0}
	ref, inv, method, err := resolveReferralCode(context.Background(), stubReferralStore{user: referrer}, "VALID001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref == nil || ref.ID != "ref-1" {
		t.Fatalf("expected referrer, got %+v", ref)
	}
	if inv != nil {
		t.Fatal("expected no invitation")
	}
	if method != "referral" {
		t.Fatalf("method = %q, want referral", method)
	}
}

func TestResolveReferral_ValidInvitation(t *testing.T) {
	inv := &models.Invitation{
		ID:        "inv-1",
		Code:      "INVITE001",
		Status:    "pending",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	ref, got, method, err := resolveReferralCode(context.Background(), stubReferralStore{inv: inv}, "INVITE001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref != nil {
		t.Fatal("expected no user referrer")
	}
	if got == nil || got.ID != "inv-1" {
		t.Fatalf("expected invitation, got %+v", got)
	}
	if method != "invitation" {
		t.Fatalf("method = %q, want invitation", method)
	}
}

func TestCheckRegistrationAccess_InviteOnly(t *testing.T) {
	ff := testFlags(t, true, true)
	cfg := DefaultConfig()
	cfg.RegistrationEnabled = true
	cfg.InviteOnly = true
	svc := New(nil, nil, nil, nil, nil, nil, nil, ff, nil, nil, zap.NewNop(), cfg)
	err := svc.checkRegistrationAccess("CODE")
	appErr := apperrors.GetAppError(err)
	if appErr == nil || appErr.Code != "INVITE_ONLY" {
		t.Fatalf("got %v, want INVITE_ONLY", err)
	}
}

func TestCheckRegistrationAccess_ConfigEnabledWithoutFlags(t *testing.T) {
	// Nil feature-flag manager must NOT disable registration when config is enabled.
	cfg := DefaultConfig()
	cfg.RegistrationEnabled = true
	svc := New(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, zap.NewNop(), cfg)
	if err := svc.checkRegistrationAccess("CODE"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestCheckRegistrationAccess_ConfigDisabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RegistrationEnabled = false
	svc := New(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, zap.NewNop(), cfg)
	err := svc.checkRegistrationAccess("CODE")
	appErr := apperrors.GetAppError(err)
	if appErr == nil || appErr.Code != "REGISTRATION_DISABLED" {
		t.Fatalf("got %v, want REGISTRATION_DISABLED", err)
	}
}

func TestHashToken_IsSHA256Hex(t *testing.T) {
	svc := New(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, zap.NewNop(), DefaultConfig())
	// Long JWT-like token that overflowed VARCHAR(255) when hex-encoded (2× length).
	b := make([]byte, 600)
	for i := range b {
		b[i] = 'A' + byte(i%26)
	}
	token := string(b)

	hash := svc.hashToken(token)
	if len(hash) != 64 {
		t.Fatalf("sha256 hex length = %d, want 64 (was overflowing VARCHAR(255) with hex(JWT))", len(hash))
	}
	if len(hash) == len(token)*2 {
		t.Fatal("hashToken still appears to hex-encode raw token instead of hashing")
	}
	if svc.hashToken(token) != hash {
		t.Fatal("hashToken not deterministic")
	}
}

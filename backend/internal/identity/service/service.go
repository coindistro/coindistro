package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/audit"
	"github.com/coindistro/backend/internal/auth"
	"github.com/coindistro/backend/internal/email"
	apperrors "github.com/coindistro/backend/internal/errors"
	"github.com/coindistro/backend/internal/events"
	"github.com/coindistro/backend/internal/featureflags"
	ide "github.com/coindistro/backend/internal/identity/errors"
	"github.com/coindistro/backend/internal/identity/models"
	"github.com/coindistro/backend/internal/identity/store"
	"github.com/coindistro/backend/internal/metrics"
	"github.com/coindistro/backend/internal/rbac"
	uuidlib "github.com/coindistro/backend/internal/uuid"
	"github.com/coindistro/backend/internal/workers"
)

// Service implements the Identity Service business logic.
type Service struct {
	store        *store.Store
	auth         *auth.Auth
	rbac         *rbac.RBAC
	eventBus     *events.InMemoryBus
	jobRegistry  *workers.Registry
	workerPool   *workers.Pool
	emailSender  email.Sender
	featureFlags *featureflags.Manager
	auditLogger  *audit.Logger
	promMetrics  *metrics.Metrics
	logger       *zap.Logger
	cfg          Config
}

// Config holds the identity service configuration.
type Config struct {
	BaseURL                  string
	ReferralCodeLength       int
	InvitationCodeLength     int
	VerificationTokenLength  int
	ResetTokenLength         int
	MaxFailedLoginAttempts   int
	AccountLockoutDuration   time.Duration
	SessionDuration          time.Duration
	InvitationDuration       time.Duration
	VerificationCooldown     time.Duration
	ResetCooldown            time.Duration
	DefaultInvitationCredits int
	GenesisMaxNumber         int
}

// DefaultConfig returns default configuration for the identity service.
func DefaultConfig() Config {
	return Config{
		BaseURL:                  "http://localhost:3000",
		ReferralCodeLength:       8,
		InvitationCodeLength:     10,
		VerificationTokenLength:  32,
		ResetTokenLength:         32,
		MaxFailedLoginAttempts:   5,
		AccountLockoutDuration:   15 * time.Minute,
		SessionDuration:          7 * 24 * time.Hour,
		InvitationDuration:       7 * 24 * time.Hour,
		VerificationCooldown:     60 * time.Second,
		ResetCooldown:            60 * time.Second,
		DefaultInvitationCredits: 5,
		GenesisMaxNumber:         10000,
	}
}

// New creates a new Identity Service.
func New(
	store *store.Store,
	authService *auth.Auth,
	rbacService *rbac.RBAC,
	eventBus *events.InMemoryBus,
	jobRegistry *workers.Registry,
	workerPool *workers.Pool,
	emailSender email.Sender,
	featureFlags *featureflags.Manager,
	auditLogger *audit.Logger,
	promMetrics *metrics.Metrics,
	logger *zap.Logger,
	cfg Config,
) *Service {
	return &Service{
		store:        store,
		auth:         authService,
		rbac:         rbacService,
		eventBus:     eventBus,
		jobRegistry:  jobRegistry,
		workerPool:   workerPool,
		emailSender:  emailSender,
		featureFlags: featureFlags,
		auditLogger:  auditLogger,
		promMetrics:  promMetrics,
		logger:       logger,
		cfg:          cfg,
	}
}

// Register creates a new user account with referral validation.
func (s *Service) Register(ctx context.Context, req *models.RegisterRequest, ip, userAgent string) (*models.AuthResponse, error) {
	if !s.featureFlags.IsEnabled(featureflags.FlagRegistration) {
		return nil, ide.ErrRegistrationDisabled
	}
	if s.featureFlags.IsEnabled(featureflags.FlagInviteOnly) {
		return nil, ide.ErrInviteOnly
	}
	if s.featureFlags.IsEnabled(featureflags.FlagRequiresReferral) && req.ReferralCode == "" {
		return nil, ide.ErrReferralRequired
	}

	exists, err := s.store.IsEmailTaken(ctx, req.Email)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	if exists {
		return nil, ide.ErrEmailAlreadyExists
	}
	if req.Username != "" {
		taken, err := s.store.IsUsernameTaken(ctx, req.Username)
		if err != nil {
			return nil, apperrors.ErrInternalServer
		}
		if taken {
			return nil, ide.ErrUsernameTaken
		}
	}

	var referrer *models.User
	var invitation *models.Invitation
	referralMethod := "direct"
	if req.ReferralCode != "" {
		referrer, err = s.store.GetUserByReferralCode(ctx, req.ReferralCode)
		if err != nil {
			return nil, apperrors.ErrInternalServer
		}
		if referrer == nil {
			invitation, err = s.store.GetInvitationByCode(ctx, req.ReferralCode)
			if err != nil {
				return nil, apperrors.ErrInternalServer
			}
			if invitation == nil {
				return nil, ide.ErrInvalidReferralCode
			}
			if invitation.Status != "pending" {
				return nil, ide.ErrReferralAlreadyUsed
			}
			if invitation.ExpiresAt.Before(time.Now()) {
				return nil, ide.ErrInvitationExpired
			}
			referralMethod = "invitation"
		} else {
			referralMethod = "referral"
		}
	}

	userID := uuidlib.NewString()
	referralCode := s.generateReferralCode(ctx)
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}
	status := "active"
	autoVerify := s.featureFlags.IsEnabled(featureflags.FlagAutoVerify)
	emailVerification := s.featureFlags.IsEnabled(featureflags.FlagEmailVerification)
	if emailVerification && !autoVerify {
		status = "pending"
	}
	referralLevel := 0
	if referrer != nil {
		referralLevel = referrer.ReferralLevel + 1
	}

	user := &models.User{
		ID:            userID,
		Username:      &req.Username,
		Email:         req.Email,
		PasswordHash:  passwordHash,
		DisplayName:   &req.DisplayName,
		Country:       &req.Country,
		Timezone:      timezone,
		ReferralCode:  referralCode,
		ReferralLevel: referralLevel,
		Status:        status,
		Roles:         []string{"user"},
	}
	if referrer != nil {
		user.ReferredBy = &referrer.ID
	}
	if err := s.store.CreateUser(ctx, user); err != nil {
		return nil, apperrors.ErrInternalServer
	}

	if referrer != nil {
		ref := &models.Referral{
			ID:           uuidlib.NewString(),
			ReferrerID:   referrer.ID,
			ReferredID:   user.ID,
			ReferralCode: req.ReferralCode,
			Level:        referralLevel,
			Status:       "active",
		}
		_ = s.store.CreateReferral(ctx, ref)
	}
	if invitation != nil {
		_ = s.store.AcceptInvitation(ctx, invitation.ID, user.ID)
		_ = s.store.DeductInvitationCredit(ctx, invitation.InviterID)
	}
	if !s.featureFlags.IsEnabled(featureflags.FlagInviteOnly) {
		credits := &models.InvitationCredit{
			ID:           uuidlib.NewString(),
			UserID:       user.ID,
			TotalCredits: s.cfg.DefaultInvitationCredits,
		}
		_ = s.store.UpsertInvitationCredits(ctx, credits)
	}

	go s.checkGenesisAward(context.Background(), user)

	tokenPair, err := s.auth.GenerateTokenPair(userID, req.Email, user.Roles)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	session := s.buildSession(ctx, userID, tokenPair.RefreshToken, ip, userAgent)
	_ = s.store.CreateSession(ctx, session)

	if emailVerification && !autoVerify {
		go s.sendVerificationEmail(context.Background(), user.ID, req.Email)
	}
	s.enqueueJob(workers.JobSendEmail, map[string]interface{}{
		"to": req.Email, "subject": "Welcome to Coindistro!", "type": "welcome", "user_id": userID,
	})
	s.publishEvent(events.EventUserRegistered, map[string]interface{}{
		"user_id": userID, "email": req.Email, "referral_code": referralCode, "referral_method": referralMethod,
	})
	s.audit(ctx, audit.ActionUserCreated, audit.EntityUser, userID, ip, userAgent, map[string]interface{}{
		"referral_method": referralMethod,
	})
	if s.promMetrics != nil {
		s.promMetrics.ActiveUsers.Inc()
	}

	return &models.AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}, nil
}

// Login authenticates a user and returns tokens.
func (s *Service) Login(ctx context.Context, req *models.LoginRequest, ip, userAgent string) (*models.AuthResponse, error) {
	user, err := s.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	if user == nil {
		return nil, ide.ErrInvalidCredentials
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, ide.ErrAccountLocked
	}
	if user.Status == "suspended" {
		return nil, ide.ErrAccountSuspended
	}
	if user.Status == "banned" {
		return nil, ide.ErrAccountBanned
	}
	if user.Status == "pending" {
		return nil, ide.ErrAccountNotVerified
	}

	if err := auth.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		attempts := user.FailedLoginAttempts + 1
		var lockedUntil *time.Time
		if attempts >= s.cfg.MaxFailedLoginAttempts {
			t := time.Now().Add(s.cfg.AccountLockoutDuration)
			lockedUntil = &t
		}
		_ = s.store.RecordFailedLogin(ctx, user.ID, attempts, lockedUntil)
		s.audit(ctx, audit.ActionLoginFailed, audit.EntityUser, user.ID, ip, userAgent, map[string]interface{}{
			"attempts": attempts,
		})
		if s.promMetrics != nil {
			s.promMetrics.ActiveUsers.Dec()
		}
		return nil, ide.ErrInvalidCredentials
	}

	_ = s.store.RecordLogin(ctx, user.ID, ip, userAgent)
	tokenPair, err := s.auth.GenerateTokenPair(user.ID, user.Email, user.Roles)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	session := s.buildSession(ctx, user.ID, tokenPair.RefreshToken, ip, userAgent)
	_ = s.store.CreateSession(ctx, session)

	s.publishEvent(events.EventUserLoggedIn, map[string]interface{}{
		"user_id": user.ID, "email": user.Email,
	})
	s.audit(ctx, audit.ActionLogin, audit.EntityUser, user.ID, ip, userAgent, nil)

	return &models.AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}, nil
}

// RefreshToken validates a refresh token and returns a new token pair.
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*models.AuthResponse, error) {
	claims, err := s.auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, ide.ErrInvalidRefreshToken
	}

	user, err := s.store.GetUserByID(ctx, claims.UserID)
	if err != nil || user == nil {
		return nil, ide.ErrInvalidRefreshToken
	}
	if user.Status != "active" {
		return nil, ide.ErrAccountSuspended
	}

	tokenPair, err := s.auth.GenerateTokenPair(user.ID, user.Email, user.Roles)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}

	s.audit(ctx, audit.ActionTokenRefresh, audit.EntityUser, user.ID, "", "", nil)

	return &models.AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}, nil
}

// Logout terminates the current session.
func (s *Service) Logout(ctx context.Context, userID, ip, userAgent string) error {
	s.publishEvent(events.EventUserLoggedOut, map[string]interface{}{
		"user_id": userID,
	})
	s.audit(ctx, audit.ActionLogout, audit.EntityUser, userID, ip, userAgent, nil)
	return nil
}

// GetProfile returns a user's profile.
func (s *Service) GetProfile(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	if user == nil {
		return nil, apperrors.ErrNotFound
	}
	return user, nil
}

// UpdateProfile updates a user's profile fields.
func (s *Service) UpdateProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.User, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	if user == nil {
		return nil, apperrors.ErrNotFound
	}

	if req.Username != nil {
		taken, err := s.store.IsUsernameTaken(ctx, *req.Username)
		if err != nil {
			return nil, apperrors.ErrInternalServer
		}
		if taken && *req.Username != *user.Username {
			return nil, ide.ErrUsernameTaken
		}
		user.Username = req.Username
	}
	if req.DisplayName != nil {
		user.DisplayName = req.DisplayName
	}
	if req.Country != nil {
		user.Country = req.Country
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.store.UpdateUser(ctx, user); err != nil {
		return nil, apperrors.ErrInternalServer
	}
	return user, nil
}

// VerifyEmail verifies a user's email address.
func (s *Service) VerifyEmail(ctx context.Context, token, ip, userAgent string) error {
	user, err := s.store.GetUserByVerificationToken(ctx, token)
	if err != nil {
		return apperrors.ErrInternalServer
	}
	if user == nil {
		return ide.ErrInvalidVerificationToken
	}
	if user.EmailVerifiedAt != nil {
		return ide.ErrAlreadyVerified
	}
	if err := s.store.VerifyEmail(ctx, user.ID); err != nil {
		return apperrors.ErrInternalServer
	}
	s.publishEvent(events.EventEmailVerified, map[string]interface{}{
		"user_id": user.ID, "email": user.Email,
	})
	s.audit(ctx, audit.ActionPasswordChange, audit.EntityUser, user.ID, ip, userAgent, nil)
	return nil
}

// ResendVerification resends the verification email.
func (s *Service) ResendVerification(ctx context.Context, userID, emailAddr string) error {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return apperrors.ErrNotFound
	}
	if user.EmailVerifiedAt != nil {
		return ide.ErrAlreadyVerified
	}
	go s.sendVerificationEmail(context.Background(), user.ID, user.Email)
	return nil
}

// RequestPasswordReset initiates the password reset flow.
func (s *Service) RequestPasswordReset(ctx context.Context, emailAddr string) error {
	user, err := s.store.GetUserByEmail(ctx, emailAddr)
	if err != nil || user == nil {
		return nil // Don't reveal if email exists
	}
	token := s.generateToken(s.cfg.ResetTokenLength)
	if err := s.store.SetPasswordResetToken(ctx, user.ID, token); err != nil {
		return apperrors.ErrInternalServer
	}
	s.enqueueJob(workers.JobSendPasswordReset, map[string]interface{}{
		"to": user.Email, "user_id": user.ID, "token": token,
	})
	s.publishEvent(events.EventPasswordResetRequested, map[string]interface{}{
		"user_id": user.ID,
	})
	return nil
}

// ResetPassword completes the password reset.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword, ip, userAgent string) error {
	user, err := s.store.GetUserByResetToken(ctx, token)
	if err != nil || user == nil {
		return ide.ErrInvalidResetToken
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return apperrors.ErrInternalServer
	}
	if err := s.store.UpdatePassword(ctx, user.ID, hash); err != nil {
		return apperrors.ErrInternalServer
	}
	s.publishEvent(events.EventPasswordResetCompleted, map[string]interface{}{
		"user_id": user.ID,
	})
	s.audit(ctx, audit.ActionPasswordChange, audit.EntityUser, user.ID, ip, userAgent, nil)
	return nil
}

// ChangePassword changes the user's password.
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword, ip, userAgent string) error {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return apperrors.ErrNotFound
	}
	if err := auth.VerifyPassword(currentPassword, user.PasswordHash); err != nil {
		return ide.ErrInvalidCredentials
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return apperrors.ErrInternalServer
	}
	if err := s.store.UpdatePassword(ctx, user.ID, hash); err != nil {
		return apperrors.ErrInternalServer
	}
	s.publishEvent(events.EventPasswordChanged, map[string]interface{}{
		"user_id": user.ID,
	})
	s.audit(ctx, audit.ActionPasswordChange, audit.EntityUser, user.ID, ip, userAgent, nil)
	return nil
}

// GetSessions returns all active sessions for a user.
func (s *Service) GetSessions(ctx context.Context, userID string) ([]*models.SessionResponse, error) {
	sessions, err := s.store.GetActiveSessions(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	var resp []*models.SessionResponse
	for _, sess := range sessions {
		resp = append(resp, &models.SessionResponse{
			ID:              sess.ID,
			Browser:         sess.Browser,
			OperatingSystem: sess.OperatingSystem,
			DeviceName:      sess.DeviceName,
			DeviceType:      sess.DeviceType,
			IPAddress:       sess.IPAddress,
			Country:         sess.Country,
			IsCurrent:       sess.IsCurrent,
			LoginAt:         sess.LoginAt,
			LastActivityAt:  sess.LastActivityAt,
			ExpiresAt:       sess.ExpiresAt,
		})
	}
	return resp, nil
}

// TerminateSession terminates a specific session.
func (s *Service) TerminateSession(ctx context.Context, userID, sessionID string) error {
	return s.store.TerminateSession(ctx, sessionID)
}

// TerminateAllSessions terminates all sessions except the current one.
func (s *Service) TerminateAllSessions(ctx context.Context, userID, currentSessionID string) error {
	return s.store.TerminateAllUserSessions(ctx, userID, currentSessionID)
}

// GetDevices returns all devices for a user.
func (s *Service) GetDevices(ctx context.Context, userID string) ([]*models.DeviceResponse, error) {
	devices, err := s.store.GetUserDevices(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	var resp []*models.DeviceResponse
	for _, d := range devices {
		resp = append(resp, &models.DeviceResponse{
			ID:              d.ID,
			Name:            d.Name,
			Browser:         d.Browser,
			OperatingSystem: d.OperatingSystem,
			DeviceType:      d.DeviceType,
			IsTrusted:       d.IsTrusted,
			IsCurrent:       d.IsCurrent,
			LastSeenAt:      d.LastSeenAt,
			FirstSeenAt:     d.FirstSeenAt,
		})
	}
	return resp, nil
}

// RemoveDevice removes a device.
func (s *Service) RemoveDevice(ctx context.Context, userID, deviceID string) error {
	return s.store.RemoveDevice(ctx, deviceID, userID)
}

// GetReferralDashboard returns the referral dashboard for a user.
func (s *Service) GetReferralDashboard(ctx context.Context, userID string) (*models.ReferralDashboardResponse, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, apperrors.ErrNotFound
	}
	counts, err := s.store.GetReferralCounts(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	credits, err := s.store.GetInvitationCredits(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	conversionRate := 0.0
	if counts.Total > 0 {
		conversionRate = float64(counts.Active+counts.Converted) / float64(counts.Total) * 100
	}
	return &models.ReferralDashboardResponse{
		ReferralCode:      user.ReferralCode,
		ReferralLink:      fmt.Sprintf("%s/register?ref=%s", s.cfg.BaseURL, user.ReferralCode),
		InvitationCredits: credits.AvailableCredits(),
		TotalInvites:      counts.Total,
		SuccessfulInvites: counts.Active + counts.Converted,
		PendingInvites:    counts.Pending,
		ConversionRate:    conversionRate,
		LeaderboardRank:   0,
		RewardsEarned:     0,
	}, nil
}

// GetInvitations returns all invitations sent by a user.
func (s *Service) GetInvitations(ctx context.Context, userID string) ([]*models.InvitationResponse, error) {
	invitations, err := s.store.GetUserInvitations(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	var resp []*models.InvitationResponse
	for _, inv := range invitations {
		resp = append(resp, &models.InvitationResponse{
			ID:           inv.ID,
			InviteeEmail: inv.InviteeEmail,
			Code:         inv.Code,
			Status:       inv.Status,
			Message:      inv.Message,
			ExpiresAt:    inv.ExpiresAt,
			ConsumedAt:   inv.ConsumedAt,
			CreatedAt:    inv.CreatedAt,
		})
	}
	return resp, nil
}

// SendInvitation sends an invitation to an email.
func (s *Service) SendInvitation(ctx context.Context, userID string, req *models.SendInvitationRequest) (*models.InvitationResponse, error) {
	credits, err := s.store.GetInvitationCredits(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	if credits.AvailableCredits() <= 0 {
		return nil, ide.ErrNoInvitationCredits
	}

	code := s.generateInvitationCode(ctx)
	inv := &models.Invitation{
		ID:           uuidlib.NewString(),
		InviterID:    userID,
		InviteeEmail: req.Email,
		Code:         code,
		Status:       "pending",
		Message:      req.Message,
		Role:         "user",
		ExpiresAt:    time.Now().Add(s.cfg.InvitationDuration),
	}
	if req.Role != "" {
		inv.Role = req.Role
	}
	if err := s.store.CreateInvitation(ctx, inv); err != nil {
		return nil, apperrors.ErrInternalServer
	}

	credits.PendingCredits++
	_ = s.store.UpsertInvitationCredits(ctx, credits)

	s.enqueueJob(workers.JobSendEmail, map[string]interface{}{
		"to": req.Email, "subject": "You're invited to Coindistro!", "type": "invitation", "code": code,
	})
	s.publishEvent(events.EventInvitationAccepted, map[string]interface{}{
		"inviter_id": userID, "invitee_email": req.Email, "code": code,
	})

	return &models.InvitationResponse{
		ID:           inv.ID,
		InviteeEmail: inv.InviteeEmail,
		Code:         inv.Code,
		Status:       inv.Status,
		Message:      inv.Message,
		ExpiresAt:    inv.ExpiresAt,
		CreatedAt:    inv.CreatedAt,
	}, nil
}

// GetActivityLog returns the activity log for a user.
func (s *Service) GetActivityLog(ctx context.Context, userID string) ([]*models.ActivityLogResponse, error) {
	logs, err := s.store.GetUserActivity(ctx, userID, 50, 0)
	if err != nil {
		return nil, apperrors.ErrInternalServer
	}
	var resp []*models.ActivityLogResponse
	for _, l := range logs {
		resp = append(resp, &models.ActivityLogResponse{
			ID:        l.ID,
			Action:    l.Action,
			IPAddress: l.IPAddress,
			DeviceID:  l.DeviceID,
			Details:   l.Details,
			CreatedAt: l.CreatedAt,
		})
	}
	return resp, nil
}

// CheckEmailAvailability checks if an email is available.
func (s *Service) CheckEmailAvailability(ctx context.Context, emailAddr string) (bool, error) {
	taken, err := s.store.IsEmailTaken(ctx, emailAddr)
	if err != nil {
		return false, apperrors.ErrInternalServer
	}
	return !taken, nil
}

// CheckUsernameAvailability checks if a username is available.
func (s *Service) CheckUsernameAvailability(ctx context.Context, username string) (bool, error) {
	taken, err := s.store.IsUsernameTaken(ctx, username)
	if err != nil {
		return false, apperrors.ErrInternalServer
	}
	return !taken, nil
}

// ─── Internal helpers ─────────────────────────────────

func (s *Service) generateReferralCode(ctx context.Context) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < 10; i++ {
		code := make([]byte, s.cfg.ReferralCodeLength)
		for j := range code {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			code[j] = charset[n.Int64()]
		}
		codeStr := string(code)
		existing, err := s.store.GetUserByReferralCode(ctx, codeStr)
		if err == nil && existing == nil {
			return codeStr
		}
	}
	return strings.ToUpper(uuidlib.NewString()[:8])
}

func (s *Service) generateInvitationCode(ctx context.Context) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, s.cfg.InvitationCodeLength)
	for j := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		code[j] = charset[n.Int64()]
	}
	return string(code)
}

func (s *Service) generateToken(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	token := make([]byte, length)
	for j := range token {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		token[j] = charset[n.Int64()]
	}
	return string(token)
}

func (s *Service) buildSession(ctx context.Context, userID, refreshToken, ip, userAgent string) *models.Session {
	return &models.Session{
		ID:               uuidlib.NewString(),
		UserID:           userID,
		RefreshTokenHash: s.hashToken(refreshToken),
		Status:           "active",
		IPAddress:        &ip,
		UserAgent:        &userAgent,
		IsCurrent:        true,
		LoginAt:          time.Now(),
		LastActivityAt:   time.Now(),
		ExpiresAt:        time.Now().Add(s.cfg.SessionDuration),
	}
}

func (s *Service) hashToken(token string) string {
	return fmt.Sprintf("%x", []byte(token))
}

func (s *Service) sendVerificationEmail(ctx context.Context, userID, emailAddr string) {
	token := s.generateToken(s.cfg.VerificationTokenLength)
	_ = s.store.SetEmailVerificationToken(ctx, userID, token)
	s.enqueueJob(workers.JobSendVerification, map[string]interface{}{
		"to": emailAddr, "user_id": userID, "token": token,
	})
}

func (s *Service) enqueueJob(jobType string, payload map[string]interface{}) {
	if s.workerPool != nil {
		s.workerPool.Submit(workers.Job{
			ID:      uuidlib.NewString(),
			Type:    jobType,
			Payload: payload,
		})
	}
}

func (s *Service) publishEvent(eventType string, data map[string]interface{}) {
	if s.eventBus != nil {
		event := events.NewEvent(eventType, "identity-service", data)
		_ = s.eventBus.Publish(context.Background(), event)
	}
}

func (s *Service) audit(ctx context.Context, action audit.Action, entityType audit.EntityType, entityID, ip, userAgent string, meta map[string]interface{}) {
	if s.auditLogger != nil {
		event := audit.NewEvent("system", action).
			WithEntity(entityType, entityID).
			WithIP(ip).
			WithUserAgent(userAgent).
			WithMetadata(meta).
			Build()
		_ = s.auditLogger.Record(ctx, event)
	}
}

func (s *Service) checkGenesisAward(ctx context.Context, user *models.User) {
	if !s.featureFlags.IsEnabled(featureflags.FlagGenesis) {
		return
	}
	genesisConfig, err := s.store.GetGenesisConfig(ctx)
	if err != nil || !genesisConfig.IsActive {
		return
	}
	if genesisConfig.CurrentGenesisCount >= genesisConfig.MaxGenesisMembers {
		return
	}
	genesisNumber, err := s.store.IncrementGenesisCount(ctx)
	if err != nil {
		return
	}
	_ = s.store.MarkUserAsGenesis(ctx, user.ID, genesisNumber)
	s.publishEvent(events.EventGenesisGranted, map[string]interface{}{
		"user_id": user.ID, "genesis_number": genesisNumber,
	})
}

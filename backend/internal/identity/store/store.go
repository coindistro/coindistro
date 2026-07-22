package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/coindistro/backend/internal/identity/models"
)

// Store handles all database operations for the identity service.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a new identity store.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// CreateUser inserts a new user.
func (s *Store) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO identity_users (
			id, username, email, password_hash, display_name, avatar_url,
			country, timezone, locale, referral_code, referred_by, referral_level,
			status, roles, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16
		) RETURNING id, created_at, updated_at`

	_, err := s.pool.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.DisplayName, user.AvatarURL,
		user.Country, user.Timezone, user.Locale, user.ReferralCode, user.ReferredBy, user.ReferralLevel,
		user.Status, user.Roles, time.Now(), time.Now(),
	)
	return err
}

// GetUserByEmail retrieves a user by email.
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, email, phone, password_hash, display_name, avatar_url,
		country, timezone, locale, referral_code, referred_by, referral_level,
		status, email_verified_at, phone_verified_at,
		is_genesis, genesis_number, genesis_date, is_founder, founder_badge,
		failed_login_attempts, locked_until, last_login_at, last_login_ip, last_login_user_agent,
		roles, created_at, updated_at, deleted_at
		FROM identity_users WHERE email = $1 AND deleted_at IS NULL`

	user, err := s.scanUser(ctx, query, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves a user by ID.
func (s *Store) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT id, username, email, phone, password_hash, display_name, avatar_url,
		country, timezone, locale, referral_code, referred_by, referral_level,
		status, email_verified_at, phone_verified_at,
		is_genesis, genesis_number, genesis_date, is_founder, founder_badge,
		failed_login_attempts, locked_until, last_login_at, last_login_ip, last_login_user_agent,
		roles, created_at, updated_at, deleted_at
		FROM identity_users WHERE id = $1 AND deleted_at IS NULL`

	user, err := s.scanUser(ctx, query, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// GetUserByReferralCode retrieves a user by their referral code.
func (s *Store) GetUserByReferralCode(ctx context.Context, code string) (*models.User, error) {
	query := `SELECT id, username, email, phone, password_hash, display_name, avatar_url,
		country, timezone, locale, referral_code, referred_by, referral_level,
		status, email_verified_at, phone_verified_at,
		is_genesis, genesis_number, genesis_date, is_founder, founder_badge,
		failed_login_attempts, locked_until, last_login_at, last_login_ip, last_login_user_agent,
		roles, created_at, updated_at, deleted_at
		FROM identity_users WHERE referral_code = $1 AND deleted_at IS NULL`

	user, err := s.scanUser(ctx, query, code)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// IsEmailTaken checks if an email is already registered.
func (s *Store) IsEmailTaken(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM identity_users WHERE email = $1 AND deleted_at IS NULL)`, email,
	).Scan(&exists)
	return exists, err
}

// IsUsernameTaken checks if a username is already taken.
func (s *Store) IsUsernameTaken(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM identity_users WHERE username = $1 AND deleted_at IS NULL)`, username,
	).Scan(&exists)
	return exists, err
}

// UpdateUser updates a user's profile fields.
func (s *Store) UpdateUser(ctx context.Context, user *models.User) error {
	query := `UPDATE identity_users SET
		username = $2, display_name = $3, avatar_url = $4, country = $5, timezone = $6,
		updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`
	_, err := s.pool.Exec(ctx, query,
		user.ID, user.Username, user.DisplayName, user.AvatarURL, user.Country, user.Timezone)
	return err
}

// UpdateUserStatus updates a user's account status.
func (s *Store) UpdateUserStatus(ctx context.Context, userID, status string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE identity_users SET status = $2, updated_at = NOW() WHERE id = $1`,
		userID, status)
	return err
}

// UpdateUserRoles updates a user's roles.
func (s *Store) UpdateUserRoles(ctx context.Context, userID string, roles []string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE identity_users SET roles = $2, updated_at = NOW() WHERE id = $1`,
		userID, roles)
	return err
}

// UpdatePassword updates a user's password hash.
func (s *Store) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE identity_users SET password_hash = $2, password_reset_at = NOW(), updated_at = NOW() WHERE id = $1`,
		userID, passwordHash)
	return err
}

// SetEmailVerificationToken sets the email verification token.
func (s *Store) SetEmailVerificationToken(ctx context.Context, userID, token string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE identity_users SET email_verification_token = $2, email_verification_sent_at = NOW(), updated_at = NOW() WHERE id = $1`,
		userID, token)
	return err
}

// VerifyEmail marks a user's email as verified.
func (s *Store) VerifyEmail(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE identity_users SET email_verified_at = NOW(), email_verification_token = NULL, status = 'active', updated_at = NOW() WHERE id = $1`,
		userID)
	return err
}

// SetPasswordResetToken sets the password reset token.
func (s *Store) SetPasswordResetToken(ctx context.Context, userID, token string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE identity_users SET password_reset_token = $2, password_reset_sent_at = NOW(), updated_at = NOW() WHERE id = $1`,
		userID, token)
	return err
}

// GetUserByResetToken finds a user by their password reset token.
func (s *Store) GetUserByResetToken(ctx context.Context, token string) (*models.User, error) {
	query := `SELECT id, username, email, phone, password_hash, display_name, avatar_url,
		country, timezone, locale, referral_code, referred_by, referral_level,
		status, email_verified_at, phone_verified_at,
		is_genesis, genesis_number, genesis_date, is_founder, founder_badge,
		failed_login_attempts, locked_until, last_login_at, last_login_ip, last_login_user_agent,
		roles, created_at, updated_at, deleted_at
		FROM identity_users WHERE password_reset_token = $1 AND deleted_at IS NULL`

	user, err := s.scanUser(ctx, query, token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// GetUserByVerificationToken finds a user by their email verification token.
func (s *Store) GetUserByVerificationToken(ctx context.Context, token string) (*models.User, error) {
	query := `SELECT id, username, email, phone, password_hash, display_name, avatar_url,
		country, timezone, locale, referral_code, referred_by, referral_level,
		status, email_verified_at, phone_verified_at,
		is_genesis, genesis_number, genesis_date, is_founder, founder_badge,
		failed_login_attempts, locked_until, last_login_at, last_login_ip, last_login_user_agent,
		roles, created_at, updated_at, deleted_at
		FROM identity_users WHERE email_verification_token = $1 AND deleted_at IS NULL`

	user, err := s.scanUser(ctx, query, token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// RecordLogin updates the user's login metadata.
func (s *Store) RecordLogin(ctx context.Context, userID, ip, userAgent string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE identity_users SET
			failed_login_attempts = 0, locked_until = NULL,
			last_login_at = NOW(), last_login_ip = $2, last_login_user_agent = $3,
			updated_at = NOW()
		WHERE id = $1`,
		userID, ip, userAgent)
	return err
}

// RecordFailedLogin increments the failed login counter.
func (s *Store) RecordFailedLogin(ctx context.Context, userID string, attempts int, lockedUntil *time.Time) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE identity_users SET failed_login_attempts = $2, locked_until = $3, updated_at = NOW() WHERE id = $1`,
		userID, attempts, lockedUntil)
	return err
}

// SoftDeleteUser soft-deletes a user.
func (s *Store) SoftDeleteUser(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE identity_users SET deleted_at = NOW(), status = 'suspended', updated_at = NOW() WHERE id = $1`,
		userID)
	return err
}

// ─── Session Operations ───────────────────────────────

// CreateSession inserts a new session.
func (s *Store) CreateSession(ctx context.Context, session *models.Session) error {
	query := `INSERT INTO sessions (id, user_id, refresh_token_hash, access_token_jti, status,
		ip_address, user_agent, browser, operating_system, device_name, device_type,
		country, city, is_current, login_at, last_activity_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`
	_, err := s.pool.Exec(ctx, query,
		session.ID, session.UserID, session.RefreshTokenHash, session.AccessTokenJTI, session.Status,
		session.IPAddress, session.UserAgent, session.Browser, session.OperatingSystem,
		session.DeviceName, session.DeviceType, session.Country, session.City,
		session.IsCurrent, session.LoginAt, session.LastActivityAt, session.ExpiresAt)
	return err
}

// GetSessionByRefreshToken retrieves a session by refresh token hash.
func (s *Store) GetSessionByRefreshToken(ctx context.Context, tokenHash string) (*models.Session, error) {
	query := `SELECT id, user_id, refresh_token_hash, access_token_jti, status,
		ip_address, user_agent, browser, operating_system, device_name, device_type,
		country, city, is_current, login_at, last_activity_at, expires_at,
		terminated_at, created_at, updated_at
		FROM sessions WHERE refresh_token_hash = $1`

	session := &models.Session{}
	err := s.pool.QueryRow(ctx, query, tokenHash).Scan(
		&session.ID, &session.UserID, &session.RefreshTokenHash, &session.AccessTokenJTI,
		&session.Status, &session.IPAddress, &session.UserAgent, &session.Browser,
		&session.OperatingSystem, &session.DeviceName, &session.DeviceType,
		&session.Country, &session.City, &session.IsCurrent,
		&session.LoginAt, &session.LastActivityAt, &session.ExpiresAt,
		&session.TerminatedAt, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return session, nil
}

// GetActiveSessions retrieves all active sessions for a user.
func (s *Store) GetActiveSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	query := `SELECT id, user_id, refresh_token_hash, access_token_jti, status,
		ip_address, user_agent, browser, operating_system, device_name, device_type,
		country, city, is_current, login_at, last_activity_at, expires_at,
		terminated_at, created_at, updated_at
		FROM sessions WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()
		ORDER BY login_at DESC`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		s := &models.Session{}
		err := rows.Scan(
			&s.ID, &s.UserID, &s.RefreshTokenHash, &s.AccessTokenJTI,
			&s.Status, &s.IPAddress, &s.UserAgent, &s.Browser,
			&s.OperatingSystem, &s.DeviceName, &s.DeviceType,
			&s.Country, &s.City, &s.IsCurrent,
			&s.LoginAt, &s.LastActivityAt, &s.ExpiresAt,
			&s.TerminatedAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// TerminateSession marks a session as terminated.
func (s *Store) TerminateSession(ctx context.Context, sessionID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET status = 'terminated', terminated_at = NOW(), updated_at = NOW() WHERE id = $1`,
		sessionID)
	return err
}

// TerminateAllUserSessions terminates all sessions for a user except one.
func (s *Store) TerminateAllUserSessions(ctx context.Context, userID string, exceptSessionID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET status = 'terminated', terminated_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND id != $2 AND status = 'active'`,
		userID, exceptSessionID)
	return err
}

// DeactivateExpiredSessions marks expired sessions as expired.
func (s *Store) DeactivateExpiredSessions(ctx context.Context) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET status = 'expired', updated_at = NOW() WHERE expires_at < NOW() AND status = 'active'`)
	return err
}

// ─── Invitation Credit Operations ───────────────────

// GetInvitationCredits retrieves a user's invitation credits.
func (s *Store) GetInvitationCredits(ctx context.Context, userID string) (*models.InvitationCredit, error) {
	credit := &models.InvitationCredit{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, total_credits, used_credits, pending_credits, created_at, updated_at
		FROM invitation_credits WHERE user_id = $1`, userID,
	).Scan(&credit.ID, &credit.UserID, &credit.TotalCredits, &credit.UsedCredits,
		&credit.PendingCredits, &credit.CreatedAt, &credit.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Default: no credits
			return &models.InvitationCredit{
				UserID: userID, TotalCredits: 0, UsedCredits: 0, PendingCredits: 0,
			}, nil
		}
		return nil, err
	}
	return credit, nil
}

// UpsertInvitationCredits creates or updates invitation credits.
func (s *Store) UpsertInvitationCredits(ctx context.Context, credit *models.InvitationCredit) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO invitation_credits (user_id, total_credits, used_credits, pending_credits)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			total_credits = EXCLUDED.total_credits,
			used_credits = EXCLUDED.used_credits,
			pending_credits = EXCLUDED.pending_credits,
			updated_at = NOW()`,
		credit.UserID, credit.TotalCredits, credit.UsedCredits, credit.PendingCredits)
	return err
}

// DeductInvitationCredit decrements the available invitation credit.
func (s *Store) DeductInvitationCredit(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE invitation_credits SET used_credits = used_credits + 1, updated_at = NOW()
		WHERE user_id = $1 AND (total_credits - used_credits - pending_credits) > 0`,
		userID)
	return err
}

// ─── Invitation Operations ──────────────────────────

// CreateInvitation inserts a new invitation.
func (s *Store) CreateInvitation(ctx context.Context, inv *models.Invitation) error {
	query := `INSERT INTO invitations (id, inviter_id, invitee_email, code, status, message, role, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := s.pool.Exec(ctx, query,
		inv.ID, inv.InviterID, inv.InviteeEmail, inv.Code, inv.Status, inv.Message, inv.Role, inv.ExpiresAt)
	return err
}

// GetInvitationByCode retrieves an invitation by its code.
func (s *Store) GetInvitationByCode(ctx context.Context, code string) (*models.Invitation, error) {
	inv := &models.Invitation{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, inviter_id, invitee_email, invitee_id, code, status, message, role,
			consumed_at, expires_at, created_at, updated_at
		FROM invitations WHERE code = $1`, code,
	).Scan(&inv.ID, &inv.InviterID, &inv.InviteeEmail, &inv.InviteeID, &inv.Code,
		&inv.Status, &inv.Message, &inv.Role, &inv.ConsumedAt, &inv.ExpiresAt,
		&inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return inv, nil
}

// AcceptInvitation marks an invitation as accepted.
func (s *Store) AcceptInvitation(ctx context.Context, id, userID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE invitations SET status = 'accepted', invitee_id = $2, consumed_at = NOW(), updated_at = NOW() WHERE id = $1`,
		id, userID)
	return err
}

// GetUserInvitations retrieves all invitations sent by a user.
func (s *Store) GetUserInvitations(ctx context.Context, userID string) ([]*models.Invitation, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, inviter_id, invitee_email, invitee_id, code, status, message, role,
			consumed_at, expires_at, created_at, updated_at
		FROM invitations WHERE inviter_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*models.Invitation
	for rows.Next() {
		inv := &models.Invitation{}
		err := rows.Scan(&inv.ID, &inv.InviterID, &inv.InviteeEmail, &inv.InviteeID, &inv.Code,
			&inv.Status, &inv.Message, &inv.Role, &inv.ConsumedAt, &inv.ExpiresAt,
			&inv.CreatedAt, &inv.UpdatedAt)
		if err != nil {
			return nil, err
		}
		invitations = append(invitations, inv)
	}
	return invitations, nil
}

// ─── Referral Operations ─────────────────────────────

// CreateReferral inserts a new referral.
func (s *Store) CreateReferral(ctx context.Context, ref *models.Referral) error {
	query := `INSERT INTO referrals (id, referrer_id, referred_id, referral_code, level, status)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.pool.Exec(ctx, query,
		ref.ID, ref.ReferrerID, ref.ReferredID, ref.ReferralCode, ref.Level, ref.Status)
	return err
}

// GetReferralsByReferrer retrieves all referrals made by a user.
func (s *Store) GetReferralsByReferrer(ctx context.Context, userID string) ([]*models.Referral, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, referrer_id, referred_id, referral_code, level, status,
			reward_amount, reward_currency, converted_at, created_at, updated_at
		FROM referrals WHERE referrer_id = $1 ORDER BY level ASC, created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var referrals []*models.Referral
	for rows.Next() {
		r := &models.Referral{}
		err := rows.Scan(&r.ID, &r.ReferrerID, &r.ReferredID, &r.ReferralCode, &r.Level,
			&r.Status, &r.RewardAmount, &r.RewardCurrency, &r.ConvertedAt, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, err
		}
		referrals = append(referrals, r)
	}
	return referrals, nil
}

// GetReferralTreeDepth retrieves the maximum referral depth for a user.
func (s *Store) GetReferralTreeDepth(ctx context.Context, userID string) (int, error) {
	var maxLevel int
	err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(level), 0) FROM referrals WHERE referrer_id = $1`, userID,
	).Scan(&maxLevel)
	return maxLevel, err
}

// CountReferralsByReferrer counts referrals for a user grouped by status.
type ReferralCounts struct {
	Total     int
	Pending   int
	Active    int
	Converted int
}

// GetReferralCounts retrieves referral counts for a user.
func (s *Store) GetReferralCounts(ctx context.Context, userID string) (*ReferralCounts, error) {
	counts := &ReferralCounts{}
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*),
			COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'converted' THEN 1 ELSE 0 END), 0)
		FROM referrals WHERE referrer_id = $1`, userID,
	).Scan(&counts.Total, &counts.Pending, &counts.Active, &counts.Converted)
	if err != nil {
		return nil, err
	}
	return counts, nil
}

// ─── Device Operations ───────────────────────────────

// UpsertDevice creates or updates a device record.
func (s *Store) UpsertDevice(ctx context.Context, device *models.Device) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO devices (id, user_id, fingerprint, name, browser, operating_system, device_type, is_trusted, is_current, last_seen_at, first_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name, last_seen_at = NOW(), is_current = EXCLUDED.is_current,
			is_trusted = CASE WHEN devices.is_trusted THEN true ELSE EXCLUDED.is_trusted END,
			updated_at = NOW()`,
		device.ID, device.UserID, device.Fingerprint, device.Name, device.Browser,
		device.OperatingSystem, device.DeviceType, device.IsTrusted, device.IsCurrent,
		device.LastSeenAt, device.FirstSeenAt)
	return err
}

// GetUserDevices retrieves all devices for a user.
func (s *Store) GetUserDevices(ctx context.Context, userID string) ([]*models.Device, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, fingerprint, name, browser, operating_system, device_type,
			is_trusted, is_current, last_seen_at, first_seen_at, created_at, updated_at
		FROM devices WHERE user_id = $1 ORDER BY last_seen_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		d := &models.Device{}
		err := rows.Scan(&d.ID, &d.UserID, &d.Fingerprint, &d.Name, &d.Browser,
			&d.OperatingSystem, &d.DeviceType, &d.IsTrusted, &d.IsCurrent,
			&d.LastSeenAt, &d.FirstSeenAt, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, nil
}

// RemoveDevice deletes a device record.
func (s *Store) RemoveDevice(ctx context.Context, deviceID, userID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM devices WHERE id = $1 AND user_id = $2`, deviceID, userID)
	return err
}

// ─── Activity Log ─────────────────────────────────────

// LogActivity creates an activity log entry.
func (s *Store) LogActivity(ctx context.Context, log *models.ActivityLog) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO activity_log (id, user_id, action, ip_address, user_agent, device_id, details)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		log.ID, log.UserID, log.Action, log.IPAddress, log.UserAgent, log.DeviceID, log.Details)
	return err
}

// GetUserActivity retrieves activity log for a user.
func (s *Store) GetUserActivity(ctx context.Context, userID string, limit, offset int) ([]*models.ActivityLog, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, action, ip_address, user_agent, device_id, details, created_at
		FROM activity_log WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.ActivityLog
	for rows.Next() {
		l := &models.ActivityLog{}
		err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.IPAddress, &l.UserAgent, &l.DeviceID, &l.Details, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// ─── Genesis Operations ──────────────────────────────

// GetGenesisConfig retrieves the genesis program config.
func (s *Store) GetGenesisConfig(ctx context.Context) (*models.GenesisConfig, error) {
	config := &models.GenesisConfig{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, max_genesis_members, current_genesis_count, is_active, created_at, updated_at
		FROM genesis_config ORDER BY created_at DESC LIMIT 1`,
	).Scan(&config.ID, &config.MaxGenesisMembers, &config.CurrentGenesisCount, &config.IsActive,
		&config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &models.GenesisConfig{MaxGenesisMembers: 10000, CurrentGenesisCount: 0, IsActive: true}, nil
		}
		return nil, err
	}
	return config, nil
}

// IncrementGenesisCount atomically increments the genesis count and returns the new number.
func (s *Store) IncrementGenesisCount(ctx context.Context) (int, error) {
	var num int
	err := s.pool.QueryRow(ctx,
		`UPDATE genesis_config SET current_genesis_count = current_genesis_count + 1, updated_at = NOW()
		WHERE is_active = true AND current_genesis_count < max_genesis_members
		RETURNING current_genesis_count`,
	).Scan(&num)
	if err != nil {
		return 0, err
	}
	return num, nil
}

// MarkUserAsGenesis marks a user as a genesis member.
func (s *Store) MarkUserAsGenesis(ctx context.Context, userID string, genesisNumber int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE identity_users SET
			is_genesis = true, genesis_number = $2, genesis_date = NOW(), updated_at = NOW()
		WHERE id = $1`, userID, genesisNumber)
	return err
}

// MarkUserAsFounder marks a user as a founder.
func (s *Store) MarkUserAsFounder(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE identity_users SET is_founder = true, founder_badge = true, updated_at = NOW() WHERE id = $1`,
		userID)
	return err
}

// CountActiveGenesisMembers counts active genesis members.
func (s *Store) CountActiveGenesisMembers(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM identity_users WHERE is_genesis = true AND deleted_at IS NULL`,
	).Scan(&count)
	return count, err
}

// ─── Admin Operations ─────────────────────────────────

// ListUsers retrieves a paginated list of users.
func (s *Store) ListUsers(ctx context.Context, status string, limit, offset int) ([]*models.User, error) {
	var rows pgx.Rows
	var err error
	if status != "" {
		rows, err = s.pool.Query(ctx,
			`SELECT id, username, email, phone, password_hash, display_name, avatar_url,
				country, timezone, locale, referral_code, referred_by, referral_level,
				status, email_verified_at, phone_verified_at,
				is_genesis, genesis_number, genesis_date, is_founder, founder_badge,
				failed_login_attempts, locked_until, last_login_at, last_login_ip, last_login_user_agent,
				roles, created_at, updated_at, deleted_at
			FROM identity_users WHERE deleted_at IS NULL AND status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			status, limit, offset)
	} else {
		rows, err = s.pool.Query(ctx,
			`SELECT id, username, email, phone, password_hash, display_name, avatar_url,
				country, timezone, locale, referral_code, referred_by, referral_level,
				status, email_verified_at, phone_verified_at,
				is_genesis, genesis_number, genesis_date, is_founder, founder_badge,
				failed_login_attempts, locked_until, last_login_at, last_login_ip, last_login_user_agent,
				roles, created_at, updated_at, deleted_at
			FROM identity_users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u, err := s.scanUserFromRow(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// CountUsers counts users (optionally by status).
func (s *Store) CountUsers(ctx context.Context, status string) (int, error) {
	var count int
	var err error
	if status != "" {
		err = s.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM identity_users WHERE deleted_at IS NULL AND status = $1`, status,
		).Scan(&count)
	} else {
		err = s.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM identity_users WHERE deleted_at IS NULL`,
		).Scan(&count)
	}
	return count, err
}

// CountVerifiedUsers counts users with a verified email.
func (s *Store) CountVerifiedUsers(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM identity_users WHERE deleted_at IS NULL AND email_verified_at IS NOT NULL`,
	).Scan(&count)
	return count, err
}

// CountTotalReferrals counts all referral relationships.
func (s *Store) CountTotalReferrals(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM referrals`,
	).Scan(&count)
	return count, err
}

// CountTotalInvitations counts all invitations.
func (s *Store) CountTotalInvitations(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM invitations`,
	).Scan(&count)
	return count, err
}

// ListRecentLogins returns users ordered by last login (non-null).
func (s *Store) ListRecentLogins(ctx context.Context, limit int) ([]*models.User, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, username, email, phone, password_hash, display_name, avatar_url,
			country, timezone, locale, referral_code, referred_by, referral_level,
			status, email_verified_at, phone_verified_at,
			is_genesis, genesis_number, genesis_date, is_founder, founder_badge,
			failed_login_attempts, locked_until, last_login_at, last_login_ip, last_login_user_agent,
			roles, created_at, updated_at, deleted_at
		FROM identity_users
		WHERE deleted_at IS NULL AND last_login_at IS NOT NULL
		ORDER BY last_login_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u, err := s.scanUserFromRow(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// ListRecentActivity returns recent activity across all users (admin audit preview).
func (s *Store) ListRecentActivity(ctx context.Context, limit int) ([]*models.ActivityLog, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, action, ip_address, user_agent, device_id, details, created_at
		FROM activity_log ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.ActivityLog
	for rows.Next() {
		l := &models.ActivityLog{}
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.Action, &l.IPAddress, &l.UserAgent,
			&l.DeviceID, &l.Details, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// ─── Internal helpers ─────────────────────────────────

func (s *Store) scanUser(ctx context.Context, query string, args ...interface{}) (*models.User, error) {
	row := s.pool.QueryRow(ctx, query, args...)
	return s.scanUserFromRow(row)
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func (s *Store) scanUserFromRow(row scannable) (*models.User, error) {
	u := &models.User{}
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.Phone, &u.PasswordHash, &u.DisplayName, &u.AvatarURL,
		&u.Country, &u.Timezone, &u.Locale, &u.ReferralCode, &u.ReferredBy, &u.ReferralLevel,
		&u.Status, &u.EmailVerifiedAt, &u.PhoneVerifiedAt,
		&u.IsGenesis, &u.GenesisNumber, &u.GenesisDate, &u.IsFounder, &u.FounderBadge,
		&u.FailedLoginAttempts, &u.LockedUntil, &u.LastLoginAt, &u.LastLoginIP, &u.LastLoginUserAgent,
		&u.Roles, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}
	return u, nil
}

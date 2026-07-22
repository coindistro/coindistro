package models

import "time"

// User represents an identity user account.
type User struct {
	ID           string  `json:"id" db:"id"`
	Username     *string `json:"username,omitempty" db:"username"`
	Email        string  `json:"email" db:"email"`
	Phone        *string `json:"phone,omitempty" db:"phone"`
	PasswordHash string  `json:"-" db:"password_hash"`
	DisplayName  *string `json:"display_name,omitempty" db:"display_name"`
	AvatarURL    *string `json:"avatar_url,omitempty" db:"avatar_url"`
	Country      *string `json:"country,omitempty" db:"country"`
	Timezone     string  `json:"timezone" db:"timezone"`
	Locale       string  `json:"locale" db:"locale"`

	ReferralCode  string  `json:"referral_code" db:"referral_code"`
	ReferredBy    *string `json:"referred_by,omitempty" db:"referred_by"`
	ReferralLevel int     `json:"referral_level" db:"referral_level"`

	Status string `json:"status" db:"status"`

	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty" db:"phone_verified_at"`

	IsGenesis     bool       `json:"is_genesis" db:"is_genesis"`
	GenesisNumber *int       `json:"genesis_number,omitempty" db:"genesis_number"`
	GenesisDate   *time.Time `json:"genesis_date,omitempty" db:"genesis_date"`
	IsFounder     bool       `json:"is_founder" db:"is_founder"`
	FounderBadge  bool       `json:"founder_badge" db:"founder_badge"`

	FailedLoginAttempts int        `json:"-" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"-" db:"locked_until"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	LastLoginIP         *string    `json:"-" db:"last_login_ip"`
	LastLoginUserAgent  *string    `json:"-" db:"last_login_user_agent"`

	Roles     []string   `json:"roles" db:"roles"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Session represents a user session.
type Session struct {
	ID               string     `json:"id" db:"id"`
	UserID           string     `json:"user_id" db:"user_id"`
	RefreshTokenHash string     `json:"-" db:"refresh_token_hash"`
	AccessTokenJTI   *string    `json:"-" db:"access_token_jti"`
	Status           string     `json:"status" db:"status"`
	IPAddress        *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent        *string    `json:"user_agent,omitempty" db:"user_agent"`
	Browser          *string    `json:"browser,omitempty" db:"browser"`
	OperatingSystem  *string    `json:"operating_system,omitempty" db:"operating_system"`
	DeviceName       *string    `json:"device_name,omitempty" db:"device_name"`
	DeviceType       *string    `json:"device_type,omitempty" db:"device_type"`
	Country          *string    `json:"country,omitempty" db:"country"`
	City             *string    `json:"city,omitempty" db:"city"`
	IsCurrent        bool       `json:"is_current" db:"is_current"`
	LoginAt          time.Time  `json:"login_at" db:"login_at"`
	LastActivityAt   time.Time  `json:"last_activity_at" db:"last_activity_at"`
	ExpiresAt        time.Time  `json:"expires_at" db:"expires_at"`
	TerminatedAt     *time.Time `json:"terminated_at,omitempty" db:"terminated_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// Device represents a trusted device.
type Device struct {
	ID              string    `json:"id" db:"id"`
	UserID          string    `json:"user_id" db:"user_id"`
	Fingerprint     *string   `json:"fingerprint,omitempty" db:"fingerprint"`
	Name            *string   `json:"name,omitempty" db:"name"`
	Browser         *string   `json:"browser,omitempty" db:"browser"`
	OperatingSystem *string   `json:"operating_system,omitempty" db:"operating_system"`
	DeviceType      *string   `json:"device_type,omitempty" db:"device_type"`
	IsTrusted       bool      `json:"is_trusted" db:"is_trusted"`
	IsCurrent       bool      `json:"is_current" db:"is_current"`
	LastSeenAt      time.Time `json:"last_seen_at" db:"last_seen_at"`
	FirstSeenAt     time.Time `json:"first_seen_at" db:"first_seen_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// InvitationCredit represents a user's invitation credit balance.
type InvitationCredit struct {
	ID             string    `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	TotalCredits   int       `json:"total_credits" db:"total_credits"`
	UsedCredits    int       `json:"used_credits" db:"used_credits"`
	PendingCredits int       `json:"pending_credits" db:"pending_credits"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// AvailableCredits returns the number of credits available to use.
func (c *InvitationCredit) AvailableCredits() int {
	return c.TotalCredits - c.UsedCredits - c.PendingCredits
}

// Invitation represents an invitation sent by a user.
type Invitation struct {
	ID           string     `json:"id" db:"id"`
	InviterID    string     `json:"inviter_id" db:"inviter_id"`
	InviteeEmail string     `json:"invitee_email" db:"invitee_email"`
	InviteeID    *string    `json:"invitee_id,omitempty" db:"invitee_id"`
	Code         string     `json:"code" db:"code"`
	Status       string     `json:"status" db:"status"`
	Message      *string    `json:"message,omitempty" db:"message"`
	Role         string     `json:"role" db:"role"`
	ConsumedAt   *time.Time `json:"consumed_at,omitempty" db:"consumed_at"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Referral represents a referral relationship.
type Referral struct {
	ID             string     `json:"id" db:"id"`
	ReferrerID     string     `json:"referrer_id" db:"referrer_id"`
	ReferredID     string     `json:"referred_id" db:"referred_id"`
	ReferralCode   string     `json:"referral_code" db:"referral_code"`
	Level          int        `json:"level" db:"level"`
	Status         string     `json:"status" db:"status"`
	RewardAmount   float64    `json:"reward_amount" db:"reward_amount"`
	RewardCurrency string     `json:"reward_currency" db:"reward_currency"`
	ConvertedAt    *time.Time `json:"converted_at,omitempty" db:"converted_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// ActivityLog represents a user activity entry.
type ActivityLog struct {
	ID        string                 `json:"id" db:"id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	Action    string                 `json:"action" db:"action"`
	IPAddress *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string                `json:"user_agent,omitempty" db:"user_agent"`
	DeviceID  *string                `json:"device_id,omitempty" db:"device_id"`
	Details   map[string]interface{} `json:"details,omitempty" db:"details"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// GenesisConfig represents the genesis program configuration.
type GenesisConfig struct {
	ID                  string    `json:"id" db:"id"`
	MaxGenesisMembers   int       `json:"max_genesis_members" db:"max_genesis_members"`
	CurrentGenesisCount int       `json:"current_genesis_count" db:"current_genesis_count"`
	IsActive            bool      `json:"is_active" db:"is_active"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// ─── Request DTOs ─────────────────────────────────────

type RegisterRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	Username     string `json:"username,omitempty" binding:"omitempty,min=3,max=50,alphanum"`
	ReferralCode string `json:"referral_code" binding:"required"`
	DisplayName  string `json:"display_name,omitempty" binding:"omitempty,max=100"`
	Country      string `json:"country,omitempty" binding:"omitempty,len=3"`
	Timezone     string `json:"timezone,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordResetComplete struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SendInvitationRequest struct {
	Email   string  `json:"email" binding:"required,email"`
	Message *string `json:"message,omitempty" binding:"omitempty,max=500"`
	Role    string  `json:"role,omitempty" binding:"omitempty"`
}

type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty" binding:"omitempty,max=100"`
	Username    *string `json:"username,omitempty" binding:"omitempty,min=3,max=50,alphanum"`
	Country     *string `json:"country,omitempty" binding:"omitempty,len=3"`
	Timezone    *string `json:"timezone,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type UpdateRolesRequest struct {
	Roles []string `json:"roles" binding:"required,min=1"`
}

type UpdateInvitationCreditsRequest struct {
	Credits int `json:"credits" binding:"required,min=0"`
}

// ─── Response DTOs ────────────────────────────────────

type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int           `json:"expires_in"`
}

type UserResponse struct {
	ID            string     `json:"id"`
	Username      *string    `json:"username,omitempty"`
	Email         string     `json:"email"`
	DisplayName   *string    `json:"display_name,omitempty"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	Country       *string    `json:"country,omitempty"`
	Timezone      string     `json:"timezone"`
	ReferralCode  string     `json:"referral_code"`
	ReferredBy    *string    `json:"referred_by,omitempty"`
	Status        string     `json:"status"`
	IsVerified    bool       `json:"is_verified"`
	IsGenesis     bool       `json:"is_genesis"`
	GenesisNumber *int       `json:"genesis_number,omitempty"`
	IsFounder     bool       `json:"is_founder"`
	Roles         []string   `json:"roles"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ReferralDashboardResponse struct {
	ReferralCode      string         `json:"referral_code"`
	ReferralLink      string         `json:"referral_link"`
	InvitationCredits int            `json:"invitation_credits"`
	TotalInvites      int            `json:"total_invites"`
	SuccessfulInvites int            `json:"successful_invites"`
	PendingInvites    int            `json:"pending_invites"`
	ConversionRate    float64        `json:"conversion_rate"`
	ReferralTree      []ReferralNode `json:"referral_tree,omitempty"`
	LeaderboardRank   int            `json:"leaderboard_rank"`
	RewardsEarned     float64        `json:"rewards_earned"`
}

type ReferralNode struct {
	ID       string         `json:"id"`
	Username *string        `json:"username,omitempty"`
	Level    int            `json:"level"`
	Date     time.Time      `json:"date"`
	Children []ReferralNode `json:"children,omitempty"`
}

type SessionResponse struct {
	ID              string    `json:"id"`
	Browser         *string   `json:"browser,omitempty"`
	OperatingSystem *string   `json:"operating_system,omitempty"`
	DeviceName      *string   `json:"device_name,omitempty"`
	DeviceType      *string   `json:"device_type,omitempty"`
	IPAddress       *string   `json:"ip_address,omitempty"`
	Country         *string   `json:"country,omitempty"`
	IsCurrent       bool      `json:"is_current"`
	LoginAt         time.Time `json:"login_at"`
	LastActivityAt  time.Time `json:"last_activity_at"`
	ExpiresAt       time.Time `json:"expires_at"`
}

type DeviceResponse struct {
	ID              string    `json:"id"`
	Name            *string   `json:"name,omitempty"`
	Browser         *string   `json:"browser,omitempty"`
	OperatingSystem *string   `json:"operating_system,omitempty"`
	DeviceType      *string   `json:"device_type,omitempty"`
	IsTrusted       bool      `json:"is_trusted"`
	IsCurrent       bool      `json:"is_current"`
	LastSeenAt      time.Time `json:"last_seen_at"`
	FirstSeenAt     time.Time `json:"first_seen_at"`
}

type InvitationResponse struct {
	ID           string     `json:"id"`
	InviteeEmail string     `json:"invitee_email"`
	Code         string     `json:"code"`
	Status       string     `json:"status"`
	Message      *string    `json:"message,omitempty"`
	ExpiresAt    time.Time  `json:"expires_at"`
	ConsumedAt   *time.Time `json:"consumed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type ActivityLogResponse struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	IPAddress *string                `json:"ip_address,omitempty"`
	DeviceID  *string                `json:"device_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// AdminUserSummary is a compact user row for admin lists.
type AdminUserSummary struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Username    *string    `json:"username,omitempty"`
	DisplayName *string    `json:"display_name,omitempty"`
	Status      string     `json:"status"`
	IsVerified  bool       `json:"is_verified"`
	IsGenesis   bool       `json:"is_genesis"`
	Roles       []string   `json:"roles"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PlatformStats holds aggregate identity metrics for the admin overview.
type PlatformStats struct {
	TotalUsers         int                  `json:"total_users"`
	VerifiedUsers      int                  `json:"verified_users"`
	GenesisMembers     int                  `json:"genesis_members"`
	ActiveUsers        int                  `json:"active_users"`
	TotalReferrals     int                  `json:"total_referrals"`
	TotalInvitations   int                  `json:"total_invitations"`
	RecentRegistrations []*AdminUserSummary `json:"recent_registrations"`
	RecentLogins        []*AdminUserSummary `json:"recent_logins"`
	RecentActivity      []*ActivityLogResponse `json:"recent_activity"`
	GenesisConfig       *GenesisConfig      `json:"genesis_config,omitempty"`
}

// ToResponse converts a User model to a UserResponse.
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:            u.ID,
		Username:      u.Username,
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		AvatarURL:     u.AvatarURL,
		Country:       u.Country,
		Timezone:      u.Timezone,
		ReferralCode:  u.ReferralCode,
		ReferredBy:    u.ReferredBy,
		Status:        u.Status,
		IsVerified:    u.EmailVerifiedAt != nil,
		IsGenesis:     u.IsGenesis,
		GenesisNumber: u.GenesisNumber,
		IsFounder:     u.IsFounder,
		Roles:         u.Roles,
		LastLoginAt:   u.LastLoginAt,
		CreatedAt:     u.CreatedAt,
	}
}

package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/identity/models"
	"github.com/coindistro/backend/internal/identity/service"
	"github.com/coindistro/backend/internal/response"
)

// Handlers contains HTTP handlers for the identity service.
type Handlers struct {
	Svc    *service.Service
	logger *zap.Logger
}

// New creates a new Handlers instance.
func New(svc *service.Service, logger *zap.Logger) *Handlers {
	return &Handlers{Svc: svc, logger: logger}
}

// Register handles user registration.
// @Summary Register a new account
// @Description Register with referral code. Invite-only when enabled.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body models.RegisterRequest true "Registration details"
// @Success 201 {object} response.APIResponse{data=models.AuthResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /auth/register [post]
func (h *Handlers) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	result, err := h.Svc.Register(c.Request.Context(), &req, ip, userAgent)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, "Registration successful", result)
}

// Login handles user login.
// @Summary Login to your account
// @Description Authenticate with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body models.LoginRequest true "Login credentials"
// @Success 200 {object} response.APIResponse{data=models.AuthResponse}
// @Failure 401 {object} response.APIResponse
// @Router /auth/login [post]
func (h *Handlers) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("login: bad request",
			zap.String("step", "parse_request"),
			zap.Error(err),
		)
		response.BadRequest(c, err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	h.logger.Info("login: handler processing",
		zap.String("step", "handler_login"),
		zap.String("email", req.Email),
		zap.String("ip", ip),
	)

	result, err := h.Svc.Login(c.Request.Context(), &req, ip, userAgent)
	if err != nil {
		h.logger.Error("login: handler error",
			zap.String("step", "handler_error"),
			zap.String("email", req.Email),
			zap.Error(err),
		)
		response.HandleError(c, err)
		return
	}

	h.logger.Info("login: handler success",
		zap.String("step", "handler_success"),
		zap.String("email", req.Email),
	)
	response.OK(c, "Login successful", result)
}

// RefreshToken handles token refresh.
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.APIResponse{data=models.AuthResponse}
// @Failure 401 {object} response.APIResponse
// @Router /auth/refresh [post]
func (h *Handlers) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.Svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Token refreshed", result)
}

// Logout handles user logout.
// @Summary Logout
// @Description Terminate the current session
// @Tags Authentication
// @Security BearerAuth
// @Success 200 {object} response.APIResponse
// @Router /auth/logout [post]
func (h *Handlers) Logout(c *gin.Context) {
	userID := c.GetString("user_id")

	err := h.Svc.Logout(c.Request.Context(), userID, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Logged out successfully", nil)
}

// GetProfile returns the current user's profile.
// @Summary Get current user profile
// @Description Returns the authenticated user's profile
// @Tags Users
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=models.UserResponse}
// @Router /users/me [get]
func (h *Handlers) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := h.Svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Profile retrieved", user.ToResponse())
}

// UpdateProfile updates the current user's profile.
// @Summary Update current user profile
// @Description Update profile fields like display_name, username, country, timezone
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.UpdateProfileRequest true "Profile fields"
// @Success 200 {object} response.APIResponse{data=models.UserResponse}
// @Router /users/me [put]
func (h *Handlers) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.Svc.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Profile updated", user.ToResponse())
}

// VerifyEmail handles email verification.
// @Summary Verify email address
// @Description Verify email using the token sent to your email
// @Tags Authentication
// @Param token query string true "Verification token"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /auth/verify-email [get]
func (h *Handlers) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.BadRequest(c, "Token is required")
		return
	}

	err := h.Svc.VerifyEmail(c.Request.Context(), token, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Email verified successfully", nil)
}

// ResendVerification resends the verification email.
// @Summary Resend verification email
// @Description Resend the email verification link
// @Tags Authentication
// @Security BearerAuth
// @Success 200 {object} response.APIResponse
// @Router /auth/resend-verification [post]
func (h *Handlers) ResendVerification(c *gin.Context) {
	userID := c.GetString("user_id")
	email := c.GetString("email")

	err := h.Svc.ResendVerification(c.Request.Context(), userID, email)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Verification email sent", nil)
}

// RequestPasswordReset initiates the password reset flow.
// @Summary Request password reset
// @Description Send a password reset link to your email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body models.PasswordResetRequest true "Email address"
// @Success 200 {object} response.APIResponse
// @Router /auth/forgot-password [post]
func (h *Handlers) RequestPasswordReset(c *gin.Context) {
	var req models.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.Svc.RequestPasswordReset(c.Request.Context(), req.Email)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "If the email exists, a reset link has been sent", nil)
}

// ResetPassword completes the password reset.
// @Summary Reset password
// @Description Complete password reset with token and new password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body models.PasswordResetComplete true "Reset token and new password"
// @Success 200 {object} response.APIResponse
// @Router /auth/reset-password [post]
func (h *Handlers) ResetPassword(c *gin.Context) {
	var req models.PasswordResetComplete
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.Svc.ResetPassword(c.Request.Context(), req.Token, req.Password, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Password reset successfully", nil)
}

// ChangePassword changes the current user's password.
// @Summary Change password
// @Description Change your current password
// @Tags Security
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.ChangePasswordRequest true "Current and new password"
// @Success 200 {object} response.APIResponse
// @Router /auth/change-password [put]
func (h *Handlers) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.Svc.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Password changed successfully", nil)
}

// GetSessions returns all active sessions.
// @Summary List active sessions
// @Description Get all active sessions for the current user
// @Tags Sessions
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=[]models.SessionResponse}
// @Router /sessions [get]
func (h *Handlers) GetSessions(c *gin.Context) {
	userID := c.GetString("user_id")

	sessions, err := h.Svc.GetSessions(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Sessions retrieved", sessions)
}

// TerminateSession terminates a specific session.
// @Summary Terminate a session
// @Description Terminate a specific session by ID
// @Tags Sessions
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} response.APIResponse
// @Router /sessions/{id} [delete]
func (h *Handlers) TerminateSession(c *gin.Context) {
	userID := c.GetString("user_id")
	sessionID := c.Param("id")

	err := h.Svc.TerminateSession(c.Request.Context(), userID, sessionID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Session terminated", nil)
}

// TerminateAllSessions terminates all sessions except current.
// @Summary Terminate all sessions
// @Description Terminate all sessions except the current one
// @Tags Sessions
// @Security BearerAuth
// @Success 200 {object} response.APIResponse
// @Router /sessions/terminate-all [post]
func (h *Handlers) TerminateAllSessions(c *gin.Context) {
	userID := c.GetString("user_id")
	currentSessionID := c.GetString("session_id")

	err := h.Svc.TerminateAllSessions(c.Request.Context(), userID, currentSessionID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "All other sessions terminated", nil)
}

// GetDevices returns all trusted devices.
// @Summary List trusted devices
// @Description Get all trusted devices for the current user
// @Tags Devices
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=[]models.DeviceResponse}
// @Router /devices [get]
func (h *Handlers) GetDevices(c *gin.Context) {
	userID := c.GetString("user_id")

	devices, err := h.Svc.GetDevices(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Devices retrieved", devices)
}

// RemoveDevice removes a trusted device.
// @Summary Remove a device
// @Description Remove a trusted device by ID
// @Tags Devices
// @Security BearerAuth
// @Param id path string true "Device ID"
// @Success 200 {object} response.APIResponse
// @Router /devices/{id} [delete]
func (h *Handlers) RemoveDevice(c *gin.Context) {
	userID := c.GetString("user_id")
	deviceID := c.Param("id")

	err := h.Svc.RemoveDevice(c.Request.Context(), userID, deviceID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Device removed", nil)
}

// GetReferralDashboard returns the referral dashboard for the current user.
// @Summary Get referral dashboard
// @Description Get referral code, stats, and tree
// @Tags Referrals
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=models.ReferralDashboardResponse}
// @Router /referrals/dashboard [get]
func (h *Handlers) GetReferralDashboard(c *gin.Context) {
	userID := c.GetString("user_id")

	dashboard, err := h.Svc.GetReferralDashboard(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Referral dashboard retrieved", dashboard)
}

// GetInvitations returns all invitations sent by the user.
// @Summary List invitations
// @Description Get all invitations sent by the current user
// @Tags Invitations
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=[]models.InvitationResponse}
// @Router /invitations [get]
func (h *Handlers) GetInvitations(c *gin.Context) {
	userID := c.GetString("user_id")

	invitations, err := h.Svc.GetInvitations(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Invitations retrieved", invitations)
}

// SendInvitation sends an invitation to an email.
// @Summary Send invitation
// @Description Send an invitation email. Consumes one invitation credit.
// @Tags Invitations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.SendInvitationRequest true "Invitation details"
// @Success 201 {object} response.APIResponse{data=models.InvitationResponse}
// @Router /invitations [post]
func (h *Handlers) SendInvitation(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.SendInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	invitation, err := h.Svc.SendInvitation(c.Request.Context(), userID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, "Invitation sent", invitation)
}

// GetActivityLog returns the user's activity history.
// @Summary Get activity log
// @Description Get security and activity history for the current user
// @Tags Security
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=[]models.ActivityLogResponse}
// @Router /activity [get]
func (h *Handlers) GetActivityLog(c *gin.Context) {
	userID := c.GetString("user_id")

	logs, err := h.Svc.GetActivityLog(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Activity log retrieved", logs)
}

// CheckEmailAvailability checks if an email is available.
// @Summary Check email availability
// @Description Check if an email is already registered
// @Tags Users
// @Param email query string true "Email to check"
// @Success 200 {object} response.APIResponse
// @Router /users/check-email [get]
func (h *Handlers) CheckEmailAvailability(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.BadRequest(c, "Email is required")
		return
	}

	available, err := h.Svc.CheckEmailAvailability(c.Request.Context(), email)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Email availability checked", gin.H{"available": available})
}

// CheckUsernameAvailability checks if a username is available.
// @Summary Check username availability
// @Description Check if a username is already taken
// @Tags Users
// @Param username query string true "Username to check"
// @Success 200 {object} response.APIResponse
// @Router /users/check-username [get]
func (h *Handlers) CheckUsernameAvailability(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		response.BadRequest(c, "Username is required")
		return
	}

	available, err := h.Svc.CheckUsernameAvailability(c.Request.Context(), username)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Username availability checked", gin.H{"available": available})
}

// RegisterAuthRoutes registers public auth routes on the given group.
func RegisterAuthRoutes(rg *gin.RouterGroup, h *Handlers) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.GET("/verify-email", h.VerifyEmail)
		auth.POST("/forgot-password", h.RequestPasswordReset)
		auth.POST("/reset-password", h.ResetPassword)
	}
}

// RegisterProtectedAuthRoutes registers authenticated auth routes.
func RegisterProtectedAuthRoutes(rg *gin.RouterGroup, h *Handlers) {
	auth := rg.Group("/auth")
	{
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
		auth.POST("/resend-verification", h.ResendVerification)
		auth.PUT("/change-password", h.ChangePassword)
	}
}

// RegisterUserRoutes registers authenticated user profile routes.
func RegisterUserRoutes(rg *gin.RouterGroup, h *Handlers) {
	users := rg.Group("/users")
	{
		users.GET("/me", h.GetProfile)
		users.PUT("/me", h.UpdateProfile)
	}
}

// RegisterPublicUserRoutes registers public user availability checks.
func RegisterPublicUserRoutes(rg *gin.RouterGroup, h *Handlers) {
	users := rg.Group("/users")
	{
		users.GET("/check-email", h.CheckEmailAvailability)
		users.GET("/check-username", h.CheckUsernameAvailability)
	}
}

// RegisterSessionRoutes registers session routes.
func RegisterSessionRoutes(rg *gin.RouterGroup, h *Handlers) {
	sessions := rg.Group("/sessions")
	{
		sessions.GET("", h.GetSessions)
		sessions.POST("/terminate-all", h.TerminateAllSessions)
		sessions.DELETE("/:id", h.TerminateSession)
	}
}

// RegisterDeviceRoutes registers device routes.
func RegisterDeviceRoutes(rg *gin.RouterGroup, h *Handlers) {
	devices := rg.Group("/devices")
	{
		devices.GET("", h.GetDevices)
		devices.DELETE("/:id", h.RemoveDevice)
	}
}

// RegisterReferralRoutes registers referral routes.
func RegisterReferralRoutes(rg *gin.RouterGroup, h *Handlers) {
	referrals := rg.Group("/referrals")
	{
		referrals.GET("/dashboard", h.GetReferralDashboard)
	}
}

// RegisterInvitationRoutes registers invitation routes.
func RegisterInvitationRoutes(rg *gin.RouterGroup, h *Handlers) {
	invitations := rg.Group("/invitations")
	{
		invitations.GET("", h.GetInvitations)
		invitations.POST("", h.SendInvitation)
	}
}

// RegisterSecurityRoutes registers security routes.
func RegisterSecurityRoutes(rg *gin.RouterGroup, h *Handlers) {
	security := rg.Group("/activity")
	{
		security.GET("", h.GetActivityLog)
	}
}

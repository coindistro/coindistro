package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/identity/service"
	"github.com/coindistro/backend/internal/response"
)

// AdminHandlers contains admin-specific identity handlers.
type AdminHandlers struct {
	svc    *service.Service
	logger *zap.Logger
}

// NewAdminHandlers creates admin handlers.
func NewAdminHandlers(svc *service.Service, logger *zap.Logger) *AdminHandlers {
	return &AdminHandlers{svc: svc, logger: logger}
}

// AdminDashboardStats returns aggregate platform statistics for the admin dashboard.
func (h *AdminHandlers) AdminDashboardStats(c *gin.Context) {
	stats, err := h.svc.GetPlatformStats(c.Request.Context())
	if err != nil {
		h.logger.Error("admin dashboard stats failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Platform statistics retrieved", stats)
}

// AdminUsersList returns paginated users for admin management.
func (h *AdminHandlers) AdminUsersList(c *gin.Context) {
	status := c.Query("status")
	page := 1
	perPage := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 {
			perPage = v
		}
	}

	summaries, total, err := h.svc.AdminListUsers(c.Request.Context(), status, page, perPage)
	if err != nil {
		h.logger.Error("admin list users failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMeta(c, 200, "Users retrieved", summaries, &response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: (total + perPage - 1) / perPage,
	})
}

// AdminRecentRegistrations returns the most recent user registrations.
func (h *AdminHandlers) AdminRecentRegistrations(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	// This uses the existing ListUsers with empty status to get recent
	users, _, err := h.svc.AdminListUsers(c.Request.Context(), "", 1, limit)
	if err != nil {
		h.logger.Error("admin recent registrations failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Recent registrations retrieved", users)
}

// AdminRecentLogins returns the most recent user logins.
func (h *AdminHandlers) AdminRecentLogins(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	// GetPlatformStats already includes recent logins; trim to requested limit.
	stats, err := h.svc.GetPlatformStats(c.Request.Context())
	if err != nil {
		h.logger.Error("admin recent logins failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}
	logins := stats.RecentLogins
	if limit > 0 && len(logins) > limit {
		logins = logins[:limit]
	}
	response.OK(c, "Recent logins retrieved", logins)
}

// AdminActivityLog returns recent activity across the platform.
func (h *AdminHandlers) AdminActivityLog(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	logs, err := h.svc.GetPlatformActivityLog(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error("admin activity log failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Activity log retrieved", logs)
}

// AdminGenesisStats returns genesis program statistics.
func (h *AdminHandlers) AdminGenesisStats(c *gin.Context) {
	stats, err := h.svc.GetPlatformStats(c.Request.Context())
	if err != nil {
		h.logger.Error("admin genesis stats failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Genesis stats retrieved", stats.GenesisConfig)
}

// AdminReferralStats returns referral program statistics.
func (h *AdminHandlers) AdminReferralStats(c *gin.Context) {
	stats, err := h.svc.GetPlatformStats(c.Request.Context())
	if err != nil {
		h.logger.Error("admin referral stats failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Referral stats retrieved", gin.H{
		"total_referrals":   stats.TotalReferrals,
		"total_invitations": stats.TotalInvitations,
	})
}

// AdminEarnStats returns earn service statistics for admin dashboard.
func (h *AdminHandlers) AdminEarnStats(c *gin.Context) {
	// This would ideally call the earn service, but for now we return a placeholder
	// The frontend can call the earn admin endpoints directly
	response.OK(c, "Earn stats endpoint", gin.H{
		"message": "Call /api/v1/earn/admin/products for product stats, /api/v1/earn/admin/products/:id/analytics for details",
	})
}

// RegisterAdminRoutes registers admin identity routes (single registration point).
func RegisterAdminRoutes(rg *gin.RouterGroup, h *AdminHandlers) {
	rg.GET("/stats", h.AdminDashboardStats)
	rg.GET("/dashboard/stats", h.AdminDashboardStats)
	rg.GET("/users", h.AdminUsersList)
	rg.GET("/registrations/recent", h.AdminRecentRegistrations)
	rg.GET("/logins/recent", h.AdminRecentLogins)
	rg.GET("/activity", h.AdminActivityLog)
	rg.GET("/genesis/stats", h.AdminGenesisStats)
	rg.GET("/referrals/stats", h.AdminReferralStats)
	rg.GET("/earn/stats", h.AdminEarnStats)
}

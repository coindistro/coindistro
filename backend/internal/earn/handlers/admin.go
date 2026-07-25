package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/earn/service"
	"github.com/coindistro/backend/internal/response"
)

// EarnAdminHandlers contains admin-specific earn handlers.
type EarnAdminHandlers struct {
	svc    *service.Service
	logger *zap.Logger
}

// NewEarnAdminHandlers creates earn admin handlers.
func NewEarnAdminHandlers(svc *service.Service, logger *zap.Logger) *EarnAdminHandlers {
	return &EarnAdminHandlers{svc: svc, logger: logger}
}

// AdminEarnPlatformStats returns earn platform statistics for admin dashboard.
func (h *EarnAdminHandlers) AdminEarnPlatformStats(c *gin.Context) {
	stats, err := h.svc.GetEarnPlatformStats(c.Request.Context())
	if err != nil {
		h.logger.Error("admin earn platform stats failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Earn platform statistics retrieved", stats)
}

// RegisterEarnAdminRoutes registers earn admin routes.
func RegisterEarnAdminRoutes(rg *gin.RouterGroup, h *EarnAdminHandlers) {
	rg.GET("/earn/stats", h.AdminEarnPlatformStats)
}

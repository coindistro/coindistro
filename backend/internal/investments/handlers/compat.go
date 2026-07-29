package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/response"
)

// PortfolioCompatibilityResponse aggregates wallet + investments for the old /earn/portfolio endpoint.
type PortfolioCompatibilityResponse struct {
	TotalAssetsInEarn   float64            `json:"total_assets_in_earn"`
	EstimatedRewards    float64            `json:"estimated_rewards"`
	TodaysRewards       float64            `json:"todays_rewards"`
	LifetimeRewards     float64            `json:"lifetime_rewards"`
	ActiveProducts      int                `json:"active_products"`
	AvailableBalance    float64            `json:"available_balance"`
	LockedBalance       float64            `json:"locked_balance"`
	AllocationByProduct map[string]float64 `json:"allocation_by_product"`
	AllocationByAsset   map[string]float64 `json:"allocation_by_asset"`
}

// PortfolioCompatibility provides a backward-compatible /earn/portfolio endpoint.
// It aggregates data from Wallet Service and Investment Service (no duplicated business logic).
func (h *Handlers) PortfolioCompatibility(c *gin.Context) {
	userID := c.GetString("user_id")

	// Get wallet data
	wallet, err := h.svc.GetWallet(c.Request.Context(), userID)
	if err != nil {
		h.logger.Warn("portfolio compat: wallet load failed", zap.String("user_id", userID), zap.Error(err))
		// Return empty portfolio rather than erroring
		response.OK(c, "Portfolio overview", &PortfolioCompatibilityResponse{
			AllocationByProduct: map[string]float64{},
			AllocationByAsset:   map[string]float64{},
		})
		return
	}
	h.logger.Info("wallet loaded", zap.String("user_id", userID), zap.Float64("total", wallet.TotalBalance))

	// Get dashboard/investments data
	dash, err := h.svc.GetDashboard(c.Request.Context(), userID)
	if err != nil {
		h.logger.Warn("portfolio compat: dashboard load failed", zap.String("user_id", userID), zap.Error(err))
	} else if dash != nil {
		h.logger.Info("investments loaded",
			zap.String("user_id", userID),
			zap.Int("count", len(dash.Investments)),
			zap.Int("active", dash.ActiveInvestments),
		)
	}

	// Calculate estimated rewards from active investments
	var estimatedRewards float64
	var lifetimeRewards float64
	var activeProducts int

	if dash != nil {
		for _, inv := range dash.Investments {
			if inv.Status == "active" {
				estimatedRewards += inv.ROICDT
				activeProducts++
			}
			lifetimeRewards += inv.ROICDT
		}
	}

	resp := &PortfolioCompatibilityResponse{
		TotalAssetsInEarn:   wallet.TotalBalance,
		EstimatedRewards:    estimatedRewards,
		TodaysRewards:       0, // Daily accrual not implemented in Genesis phase
		LifetimeRewards:     lifetimeRewards,
		ActiveProducts:      activeProducts,
		AvailableBalance:    wallet.AvailableBalance,
		LockedBalance:       wallet.LockedBalance,
		AllocationByProduct: map[string]float64{},
		AllocationByAsset:   map[string]float64{"CDT": wallet.TotalBalance},
	}

	response.OK(c, "Portfolio overview", resp)
}

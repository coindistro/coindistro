package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/earn/models"
	"github.com/coindistro/backend/internal/earn/service"
	"github.com/coindistro/backend/internal/featureflags"
	"github.com/coindistro/backend/internal/middleware"
	"github.com/coindistro/backend/internal/response"
)

// Handlers exposes Earn HTTP handlers.
type Handlers struct {
	svc          *service.Service
	featureFlags *featureflags.Manager
	logger       *zap.Logger
}

// New creates Earn handlers.
func New(svc *service.Service, featureFlags *featureflags.Manager, logger *zap.Logger) *Handlers {
	return &Handlers{svc: svc, featureFlags: featureFlags, logger: logger}
}

func (h *Handlers) requireEarnEnabled(c *gin.Context) bool {
	if h.featureFlags != nil && !h.featureFlags.IsEnabled(featureflags.FlagEarn) {
		response.Error(c, 503, "EARN_DISABLED", "Earn module is disabled")
		return false
	}
	return true
}

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	return page, perPage
}

// ─── Products (discovery) ─────────────────────────────

// ListProducts godoc
// @Summary List earn products
// @Tags Earn
// @Produce json
// @Param category query string false "Product category"
// @Param status query string false "Status filter"
// @Param featured query bool false "Featured only"
// @Param asset query string false "Supported asset"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/products [get]
func (h *Handlers) ListProducts(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	page, perPage := pageParams(c)
	var featured *bool
	if v := c.Query("featured"); v != "" {
		b := v == "true" || v == "1"
		featured = &b
	}
	status := c.Query("status")
	if status == "" {
		status = models.StatusActive
	}
	list, total, err := h.svc.ListProducts(c.Request.Context(), models.ProductListFilter{
		Category: c.Query("category"),
		Status:   status,
		Featured: featured,
		Asset:    c.Query("asset"),
		Page:     page,
		PerPage:  perPage,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMeta(c, 200, "Products retrieved", list, &response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPages: (total + perPage - 1) / perPage,
	})
}

// GetProduct godoc
// @Summary Get earn product by ID or slug
// @Tags Earn
// @Produce json
// @Param id path string true "Product ID or slug"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/products/{id} [get]
func (h *Handlers) GetProduct(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	p, err := h.svc.GetProduct(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Product retrieved", p)
}

// ─── Portfolio ────────────────────────────────────────

// PortfolioOverview godoc
// @Summary Earn portfolio overview
// @Tags Earn
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse{data=models.PortfolioOverview}
// @Router /api/v1/earn/portfolio [get]
func (h *Handlers) PortfolioOverview(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	ov, err := h.svc.PortfolioOverview(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Portfolio overview", ov)
}

// ─── Participation ────────────────────────────────────

// JoinProduct godoc
// @Summary Join an earn product
// @Tags Earn
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param body body models.JoinProductRequest true "Join request"
// @Success 201 {object} response.APIResponse
// @Router /api/v1/earn/products/{id}/join [post]
func (h *Handlers) JoinProduct(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	var req models.JoinProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	part, err := h.svc.JoinProduct(c.Request.Context(), c.GetString("user_id"), c.Param("id"), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, "Joined product", part)
}

// ListParticipations godoc
// @Summary List user participations
// @Tags Earn
// @Security BearerAuth
// @Produce json
// @Param status query string false "Status"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/participations [get]
func (h *Handlers) ListParticipations(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	page, perPage := pageParams(c)
	list, total, err := h.svc.ListParticipations(c.Request.Context(), c.GetString("user_id"), c.Query("status"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMeta(c, 200, "Participations retrieved", list, &response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPages: (total + perPage - 1) / perPage,
	})
}

// GetParticipation godoc
// @Summary Get participation details
// @Tags Earn
// @Security BearerAuth
// @Produce json
// @Param id path string true "Participation ID"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/participations/{id} [get]
func (h *Handlers) GetParticipation(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	part, err := h.svc.GetParticipation(c.Request.Context(), c.GetString("user_id"), c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Participation retrieved", part)
}

// AddFunds godoc
// @Summary Add funds to a participation
// @Tags Earn
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Participation ID"
// @Param body body models.AddFundsRequest true "Amount"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/participations/{id}/add-funds [post]
func (h *Handlers) AddFunds(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	var req models.AddFundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	part, err := h.svc.AddFunds(c.Request.Context(), c.GetString("user_id"), c.Param("id"), req.Amount)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Funds added", part)
}

// Withdraw godoc
// @Summary Withdraw from a participation
// @Tags Earn
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Participation ID"
// @Param body body models.WithdrawRequest true "Amount"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/participations/{id}/withdraw [post]
func (h *Handlers) Withdraw(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	var req models.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	part, err := h.svc.Withdraw(c.Request.Context(), c.GetString("user_id"), c.Param("id"), req.Amount)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Withdrawal recorded", part)
}

// ExitParticipation godoc
// @Summary Exit a participation
// @Tags Earn
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Participation ID"
// @Param body body models.ExitParticipationRequest false "Reason"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/participations/{id}/exit [post]
func (h *Handlers) ExitParticipation(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	var req models.ExitParticipationRequest
	_ = c.ShouldBindJSON(&req)
	part, err := h.svc.ExitParticipation(c.Request.Context(), c.GetString("user_id"), c.Param("id"), req.Reason)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Participation exited", part)
}

// ─── Rewards / history ────────────────────────────────

// ListRewards godoc
// @Summary List reward history
// @Tags Earn
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/rewards [get]
func (h *Handlers) ListRewards(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	page, perPage := pageParams(c)
	list, total, err := h.svc.ListRewards(c.Request.Context(), c.GetString("user_id"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMeta(c, 200, "Rewards retrieved", list, &response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPages: (total + perPage - 1) / perPage,
	})
}

// ListTransactions godoc
// @Summary List earn transaction history
// @Tags Earn
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/history [get]
func (h *Handlers) ListTransactions(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	page, perPage := pageParams(c)
	list, total, err := h.svc.ListTransactions(c.Request.Context(), c.GetString("user_id"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMeta(c, 200, "History retrieved", list, &response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPages: (total + perPage - 1) / perPage,
	})
}

// ─── Launchpool / Learn / Referral ────────────────────

// ListLaunchpools godoc
// @Summary List launchpool campaigns
// @Tags Earn
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/launchpool [get]
func (h *Handlers) ListLaunchpools(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	list, err := h.svc.ListLaunchpools(c.Request.Context(), c.Query("status"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Launchpool campaigns", list)
}

// ListLearnCampaigns godoc
// @Summary List learn & earn campaigns
// @Tags Earn
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/learn [get]
func (h *Handlers) ListLearnCampaigns(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	list, err := h.svc.ListLearnCampaigns(c.Request.Context(), c.Query("status"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Learn campaigns", list)
}

// CompleteLearnCampaign godoc
// @Summary Record learning completion for reward eligibility
// @Tags Earn
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Campaign ID"
// @Param body body models.CompleteLearnRequest false "Metadata"
// @Success 201 {object} response.APIResponse
// @Router /api/v1/earn/learn/{id}/complete [post]
func (h *Handlers) CompleteLearnCampaign(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	var req models.CompleteLearnRequest
	_ = c.ShouldBindJSON(&req)
	comp, err := h.svc.CompleteLearnCampaign(c.Request.Context(), c.GetString("user_id"), c.Param("id"), req.Metadata)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, "Learning completion recorded", comp)
}

// ReferralRewards godoc
// @Summary Referral reward summary and history
// @Tags Earn
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse{data=models.ReferralRewardSummary}
// @Router /api/v1/earn/referral/rewards [get]
func (h *Handlers) ReferralRewards(c *gin.Context) {
	if !h.requireEarnEnabled(c) {
		return
	}
	summary, err := h.svc.ReferralRewardSummary(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Referral rewards", summary)
}

// ─── Admin ────────────────────────────────────────────

// AdminCreateProduct godoc
// @Summary Create earn product (admin)
// @Tags Earn Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.CreateProductRequest true "Product"
// @Success 201 {object} response.APIResponse
// @Router /api/v1/earn/admin/products [post]
func (h *Handlers) AdminCreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.CreateProduct(c.Request.Context(), &req, c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, "Product created", p)
}

// AdminUpdateProduct godoc
// @Summary Update earn product (admin)
// @Tags Earn Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param body body models.UpdateProductRequest true "Updates"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/admin/products/{id} [put]
func (h *Handlers) AdminUpdateProduct(c *gin.Context) {
	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.UpdateProduct(c.Request.Context(), c.Param("id"), &req, c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Product updated", p)
}

// AdminSetProductStatus godoc
// @Summary Pause/resume/archive product (admin)
// @Tags Earn Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param body body object true "Status payload" SchemaExample({"status":"paused"})
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/admin/products/{id}/status [put]
func (h *Handlers) AdminSetProductStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.SetProductStatus(c.Request.Context(), c.Param("id"), req.Status, c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Product status updated", p)
}

// AdminListParticipants godoc
// @Summary List product participants (admin)
// @Tags Earn Admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/admin/products/{id}/participants [get]
func (h *Handlers) AdminListParticipants(c *gin.Context) {
	page, perPage := pageParams(c)
	list, total, err := h.svc.ListParticipants(c.Request.Context(), c.Param("id"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMeta(c, 200, "Participants", list, &response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPages: (total + perPage - 1) / perPage,
	})
}

// AdminProductAnalytics godoc
// @Summary Product analytics (admin)
// @Tags Earn Admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} response.APIResponse{data=models.ProductAnalytics}
// @Router /api/v1/earn/admin/products/{id}/analytics [get]
func (h *Handlers) AdminProductAnalytics(c *gin.Context) {
	a, err := h.svc.ProductAnalytics(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Analytics", a)
}

// AdminCreateLaunchpool godoc
// @Summary Create launchpool campaign (admin)
// @Tags Earn Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.CreateLaunchpoolRequest true "Campaign"
// @Success 201 {object} response.APIResponse
// @Router /api/v1/earn/admin/launchpool [post]
func (h *Handlers) AdminCreateLaunchpool(c *gin.Context) {
	var req models.CreateLaunchpoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	camp, err := h.svc.CreateLaunchpool(c.Request.Context(), &req, c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, "Launchpool created", camp)
}

// AdminCreateLearnCampaign godoc
// @Summary Create learn & earn campaign (admin)
// @Tags Earn Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.CreateLearnCampaignRequest true "Campaign"
// @Success 201 {object} response.APIResponse
// @Router /api/v1/earn/admin/learn [post]
func (h *Handlers) AdminCreateLearnCampaign(c *gin.Context) {
	var req models.CreateLearnCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	camp, err := h.svc.CreateLearnCampaign(c.Request.Context(), &req, c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, "Learn campaign created", camp)
}

// AdminListProducts godoc
// @Summary List all products including drafts (admin)
// @Tags Earn Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /api/v1/earn/admin/products [get]
func (h *Handlers) AdminListProducts(c *gin.Context) {
	page, perPage := pageParams(c)
	list, total, err := h.svc.ListProducts(c.Request.Context(), models.ProductListFilter{
		Category: c.Query("category"),
		Status:   c.Query("status"), // empty = all statuses
		Page:     page,
		PerPage:  perPage,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMeta(c, 200, "Admin products", list, &response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPages: (total + perPage - 1) / perPage,
	})
}

// RegisterRoutes wires public + authenticated earn routes under /earn.
func RegisterRoutes(rg *gin.RouterGroup, h *Handlers, authMiddleware gin.HandlerFunc) {
	earn := rg.Group("/earn")
	{
		// Discovery (authenticated for dashboard consistency; still works with auth)
		earn.GET("/products", authMiddleware, h.ListProducts)
		earn.GET("/products/:id", authMiddleware, h.GetProduct)
		earn.GET("/launchpool", authMiddleware, h.ListLaunchpools)
		earn.GET("/learn", authMiddleware, h.ListLearnCampaigns)

		// User dashboard
		earn.GET("/portfolio", authMiddleware, h.PortfolioOverview)
		earn.POST("/products/:id/join", authMiddleware, h.JoinProduct)
		earn.GET("/participations", authMiddleware, h.ListParticipations)
		earn.GET("/participations/:id", authMiddleware, h.GetParticipation)
		earn.POST("/participations/:id/add-funds", authMiddleware, h.AddFunds)
		earn.POST("/participations/:id/withdraw", authMiddleware, h.Withdraw)
		earn.POST("/participations/:id/exit", authMiddleware, h.ExitParticipation)
		earn.GET("/rewards", authMiddleware, h.ListRewards)
		earn.GET("/history", authMiddleware, h.ListTransactions)
		earn.POST("/learn/:id/complete", authMiddleware, h.CompleteLearnCampaign)
		earn.GET("/referral/rewards", authMiddleware, h.ReferralRewards)

		// Admin
		admin := earn.Group("/admin")
		admin.Use(authMiddleware)
		admin.Use(middleware.RequireRole("admin", "super_admin"))
		{
			admin.GET("/products", h.AdminListProducts)
			admin.POST("/products", h.AdminCreateProduct)
			admin.PUT("/products/:id", h.AdminUpdateProduct)
			admin.PUT("/products/:id/status", h.AdminSetProductStatus)
			admin.GET("/products/:id/participants", h.AdminListParticipants)
			admin.GET("/products/:id/analytics", h.AdminProductAnalytics)
			admin.POST("/launchpool", h.AdminCreateLaunchpool)
			admin.POST("/learn", h.AdminCreateLearnCampaign)
		}
	}
}

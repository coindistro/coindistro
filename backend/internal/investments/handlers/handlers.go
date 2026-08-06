package handlers

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/investments/models"
	"github.com/coindistro/backend/internal/investments/service"
	"github.com/coindistro/backend/internal/middleware"
	"github.com/coindistro/backend/internal/response"
)

// Handlers exposes investment HTTP handlers.
type Handlers struct {
	svc    *service.Service
	logger *zap.Logger
}

// New creates investment handlers.
func New(svc *service.Service, logger *zap.Logger) *Handlers {
	return &Handlers{svc: svc, logger: logger}
}

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	return page, perPage
}

// ─── Plans ────────────────────────────────────────────

// ListPlans godoc
// @Summary List investment plans
// @Tags Investments
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /earn/plans [get]
func (h *Handlers) ListPlans(c *gin.Context) {
	plans, err := h.svc.ListPlans(c.Request.Context(), true)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Investment plans retrieved", plans)
}

// GetPlan godoc
// @Summary Get investment plan by ID
// @Tags Investments
// @Produce json
// @Param id path string true "Plan ID"
// @Success 200 {object} response.APIResponse
// @Router /earn/plans/{id} [get]
func (h *Handlers) GetPlan(c *gin.Context) {
	plan, err := h.svc.GetPlan(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Investment plan retrieved", plan)
}

// ─── Payment Initiation ───────────────────────────────

// InitPaystackPayment godoc
// @Summary Initialize Paystack payment
// @Tags Investments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.InitPaymentRequest true "Payment details"
// @Success 200 {object} response.APIResponse{data=models.InitPaymentResponse}
// @Router /payments/paystack/init [post]
func (h *Handlers) InitPaystackPayment(c *gin.Context) {
	var req models.InitPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.svc.InitPaystackPayment(c.Request.Context(), c.GetString("user_id"), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Payment initialized", resp)
}

// InitFlutterwavePayment godoc
// @Summary Initialize Flutterwave payment
// @Tags Investments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.InitPaymentRequest true "Payment details"
// @Success 200 {object} response.APIResponse{data=models.InitPaymentResponse}
// @Router /payments/flutterwave/init [post]
func (h *Handlers) InitFlutterwavePayment(c *gin.Context) {
	var req models.InitPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.svc.InitFlutterwavePayment(c.Request.Context(), c.GetString("user_id"), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Payment initialized", resp)
}

// VerifyPaystackPayment godoc
// @Summary Verify Paystack payment after redirect
// @Tags Investments
// @Security BearerAuth
// @Produce json
// @Param reference query string true "Paystack transaction reference"
// @Success 200 {object} response.APIResponse{data=models.Investment}
// @Router /payments/paystack/verify [get]
func (h *Handlers) VerifyPaystackPayment(c *gin.Context) {
	reference := c.Query("reference")
	if reference == "" {
		reference = c.Query("trxref")
	}
	if reference == "" {
		var body struct {
			Reference string `json:"reference"`
		}
		_ = c.ShouldBindJSON(&body)
		reference = body.Reference
	}
	inv, err := h.svc.VerifyPaystackPayment(c.Request.Context(), c.GetString("user_id"), reference)
	if err != nil {
		h.logger.Error("paystack verify failed",
			zap.String("reference", reference),
			zap.Error(err),
		)
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Payment verified", inv)
}

// ─── Webhooks ─────────────────────────────────────────

// PaystackWebhook godoc
// @Summary Paystack webhook handler
// @Tags Webhooks
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /payments/paystack/webhook [post]
func (h *Handlers) PaystackWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "Failed to read request body")
		return
	}

	signature := c.GetHeader("x-paystack-signature")
	if err := h.svc.ProcessPaystackWebhook(c.Request.Context(), payload, signature); err != nil {
		h.logger.Error("paystack webhook processing failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Webhook processed", nil)
}

// FlutterwaveWebhook godoc
// @Summary Flutterwave webhook handler
// @Tags Webhooks
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /payments/flutterwave/webhook [post]
func (h *Handlers) FlutterwaveWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "Failed to read request body")
		return
	}

	signature := c.GetHeader("verif-hash")
	if err := h.svc.ProcessFlutterwaveWebhook(c.Request.Context(), payload, signature); err != nil {
		h.logger.Error("flutterwave webhook processing failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Webhook processed", nil)
}

// ─── User Dashboard ───────────────────────────────────

// GetDashboard godoc
// @Summary Get investment dashboard
// @Tags Investments
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse{data=models.InvestmentDashboard}
// @Router /earn/investments [get]
func (h *Handlers) GetDashboard(c *gin.Context) {
	dash, err := h.svc.GetDashboard(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Dashboard retrieved", dash)
}

// ListInvestments godoc
// @Summary List user investments
// @Tags Investments
// @Security BearerAuth
// @Produce json
// @Param status query string false "Status filter"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {object} response.APIResponse
// @Router /earn/investments/list [get]
func (h *Handlers) ListInvestments(c *gin.Context) {
	page, perPage := pageParams(c)
	list, total, err := h.svc.GetUserInvestments(c.Request.Context(), c.GetString("user_id"), c.Query("status"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMeta(c, 200, "Investments retrieved", list, &response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPages: (total + perPage - 1) / perPage,
	})
}

// GetInvestment godoc
// @Summary Get investment details
// @Tags Investments
// @Security BearerAuth
// @Produce json
// @Param id path string true "Investment ID"
// @Success 200 {object} response.APIResponse
// @Router /earn/investments/{id} [get]
func (h *Handlers) GetInvestment(c *gin.Context) {
	inv, err := h.svc.GetInvestment(c.Request.Context(), c.GetString("user_id"), c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Investment retrieved", inv)
}

// ─── Wallet ───────────────────────────────────────────

// GetWallet godoc
// @Summary Get user wallet
// @Tags Wallet
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse{data=models.Wallet}
// @Router /wallet [get]
func (h *Handlers) GetWallet(c *gin.Context) {
	wallet, err := h.svc.GetWallet(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Wallet retrieved", wallet)
}

// GetWalletTransactions godoc
// @Summary Get wallet transactions
// @Tags Wallet
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {object} response.APIResponse
// @Router /wallet/transactions [get]
func (h *Handlers) GetWalletTransactions(c *gin.Context) {
	page, perPage := pageParams(c)
	list, total, err := h.svc.GetWalletTransactions(c.Request.Context(), c.GetString("user_id"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMeta(c, 200, "Transactions retrieved", list, &response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPages: (total + perPage - 1) / perPage,
	})
}

// ─── Admin ────────────────────────────────────────────

// AdminListPlans godoc
// @Summary List all investment plans (admin)
// @Tags Admin Investments
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /admin/plans [get]
func (h *Handlers) AdminListPlans(c *gin.Context) {
	plans, err := h.svc.ListPlans(c.Request.Context(), false)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Plans retrieved", plans)
}

// AdminCreatePlan godoc
// @Summary Create investment plan (admin)
// @Tags Admin Investments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.CreatePlanRequest true "Plan details"
// @Success 201 {object} response.APIResponse
// @Router /admin/plans [post]
func (h *Handlers) AdminCreatePlan(c *gin.Context) {
	var req models.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	plan, err := h.svc.CreatePlan(c.Request.Context(), &req, c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, "Plan created", plan)
}

// AdminUpdatePlan godoc
// @Summary Update investment plan (admin)
// @Tags Admin Investments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Plan ID"
// @Param body body models.UpdatePlanRequest true "Plan updates"
// @Success 200 {object} response.APIResponse
// @Router /admin/plans/{id} [put]
func (h *Handlers) AdminUpdatePlan(c *gin.Context) {
	var req models.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	plan, err := h.svc.UpdatePlan(c.Request.Context(), c.Param("id"), &req, c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Plan updated", plan)
}

// AdminDeletePlan godoc
// @Summary Delete investment plan (admin)
// @Tags Admin Investments
// @Security BearerAuth
// @Produce json
// @Param id path string true "Plan ID"
// @Success 200 {object} response.APIResponse
// @Router /admin/plans/{id} [delete]
func (h *Handlers) AdminDeletePlan(c *gin.Context) {
	if err := h.svc.DeletePlan(c.Request.Context(), c.Param("id"), c.GetString("user_id")); err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Plan deleted", nil)
}

// AdminListInvestments godoc
// @Summary List all investments (admin)
// @Tags Admin Investments
// @Security BearerAuth
// @Produce json
// @Param status query string false "Status filter"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {object} response.APIResponse
// @Router /admin/investments [get]
func (h *Handlers) AdminListInvestments(c *gin.Context) {
	page, perPage := pageParams(c)
	list, total, err := h.svc.AdminListInvestments(c.Request.Context(), c.Query("status"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMeta(c, 200, "Investments retrieved", list, &response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPages: (total + perPage - 1) / perPage,
	})
}

// AdminListPayments godoc
// @Summary List payment transactions (admin)
// @Tags Admin Investments
// @Security BearerAuth
// @Produce json
// @Param status query string false "Status filter"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {object} response.APIResponse
// @Router /admin/payments [get]
func (h *Handlers) AdminListPayments(c *gin.Context) {
	page, perPage := pageParams(c)
	list, total, err := h.svc.AdminListPayments(c.Request.Context(), c.Query("status"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMeta(c, 200, "Payments retrieved", list, &response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPages: (total + perPage - 1) / perPage,
	})
}

// AdminGetPricing godoc
// @Summary Get current CDT pricing (admin)
// @Tags Admin Investments
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /admin/pricing [get]
func (h *Handlers) AdminGetPricing(c *gin.Context) {
	pricing, err := h.svc.GetCurrentPricing(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Pricing retrieved", pricing)
}

// AdminSetPricing godoc
// @Summary Set CDT pricing (admin)
// @Tags Admin Investments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.SetPricingRequest true "Pricing"
// @Success 200 {object} response.APIResponse
// @Router /admin/pricing [put]
func (h *Handlers) AdminSetPricing(c *gin.Context) {
	var req models.SetPricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	pricing, err := h.svc.SetPricing(c.Request.Context(), req.PriceNGN, c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Pricing updated", pricing)
}

// AdminGetStats godoc
// @Summary Get investment statistics (admin)
// @Tags Admin Investments
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse{data=models.AdminInvestmentStats}
// @Router /admin/investments/stats [get]
func (h *Handlers) AdminGetStats(c *gin.Context) {
	stats, err := h.svc.AdminGetStats(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Stats retrieved", stats)
}

// AdminListWallets godoc
// @Summary List all wallets (admin)
// @Tags Admin Investments
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {object} response.APIResponse
// @Router /admin/wallets [get]
func (h *Handlers) AdminListWallets(c *gin.Context) {
	page, perPage := pageParams(c)
	list, total, err := h.svc.AdminListWallets(c.Request.Context(), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMeta(c, 200, "Wallets retrieved", list, &response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPages: (total + perPage - 1) / perPage,
	})
}

// AdminListWebhookEvents godoc
// @Summary List webhook events (admin)
// @Tags Admin Investments
// @Security BearerAuth
// @Produce json
// @Param provider query string false "Provider filter"
// @Param status query string false "Status filter"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {object} response.APIResponse
// @Router /admin/webhooks [get]
func (h *Handlers) AdminListWebhookEvents(c *gin.Context) {
	page, perPage := pageParams(c)
	list, total, err := h.svc.AdminListWebhookEvents(c.Request.Context(), c.Query("provider"), c.Query("status"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMeta(c, 200, "Webhook events retrieved", list, &response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPages: (total + perPage - 1) / perPage,
	})
}

// RegisterRoutes wires investment routes.
func RegisterRoutes(rg *gin.RouterGroup, h *Handlers, authMiddleware gin.HandlerFunc) {
	// Public routes
	rg.GET("/earn/plans", h.ListPlans)
	rg.GET("/earn/plans/:id", h.GetPlan)
	rg.GET("/investments/plans", h.ListPlans)
	rg.GET("/investments/plans/:id", h.GetPlan)

	// Webhooks (no auth - signature verified in handler)
	rg.POST("/payments/paystack/webhook", h.PaystackWebhook)
	rg.POST("/payments/flutterwave/webhook", h.FlutterwaveWebhook)

	// Authenticated routes
	authed := rg.Group("")
	authed.Use(authMiddleware)
	{
		// Payment initiation
		authed.POST("/payments/paystack/init", h.InitPaystackPayment)
		authed.POST("/payments/flutterwave/init", h.InitFlutterwavePayment)
		authed.POST("/payments/paystack/initiate", h.InitPaystackPayment)
		authed.POST("/payments/flutterwave/initiate", h.InitFlutterwavePayment)
		// Payment verification (after redirect from gateway)
		authed.GET("/payments/paystack/verify", h.VerifyPaystackPayment)
		authed.POST("/payments/paystack/verify", h.VerifyPaystackPayment)

		// Investment dashboard
		authed.GET("/earn/investments", h.GetDashboard)
		authed.GET("/earn/investments/list", h.ListInvestments)
		authed.GET("/earn/investments/:id", h.GetInvestment)

		// Wallet
		authed.GET("/wallet", h.GetWallet)
		authed.GET("/wallet/transactions", h.GetWalletTransactions)

		// Backward-compatible portfolio endpoint (aggregates wallet + investments)
		authed.GET("/earn/portfolio", h.PortfolioCompatibility)
	}

	// Admin routes
	admin := rg.Group("/admin")
	admin.Use(authMiddleware)
	admin.Use(middleware.RequireRole("admin", "super_admin"))
	{
		admin.GET("/plans", h.AdminListPlans)
		admin.POST("/plans", h.AdminCreatePlan)
		admin.PUT("/plans/:id", h.AdminUpdatePlan)
		admin.DELETE("/plans/:id", h.AdminDeletePlan)

		admin.GET("/investments", h.AdminListInvestments)
		admin.GET("/investments/stats", h.AdminGetStats)

		admin.GET("/payments", h.AdminListPayments)
		admin.GET("/pricing", h.AdminGetPricing)
		admin.PUT("/pricing", h.AdminSetPricing)
		admin.GET("/wallets", h.AdminListWallets)
		admin.GET("/webhooks", h.AdminListWebhookEvents)
	}
}

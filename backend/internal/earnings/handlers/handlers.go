package handlers

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/earnings/models"
	"github.com/coindistro/backend/internal/earnings/service"
	"github.com/coindistro/backend/internal/middleware"
	"github.com/coindistro/backend/internal/response"
)

// Handlers exposes earnings investment HTTP handlers.
type Handlers struct {
	svc    *service.Service
	logger *zap.Logger
}

// New creates earnings handlers.
func New(svc *service.Service, logger *zap.Logger) *Handlers {
	return &Handlers{svc: svc, logger: logger}
}

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}

// ─── Public / Settings ────────────────────────────────────

// GetExchangeRate returns the current USD→NGN rate.
func (h *Handlers) GetExchangeRate(c *gin.Context) {
	rate, err := h.svc.GetExchangeRate(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Exchange rate retrieved", rate)
}

// GetSettings returns public investment settings.
func (h *Handlers) GetSettings(c *gin.Context) {
	settings, err := h.svc.GetSettings(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Investment settings retrieved", settings)
}

// ─── Dashboard & Investments ──────────────────────────────

// GetDashboard returns the investor earnings dashboard.
func (h *Handlers) GetDashboard(c *gin.Context) {
	dash, err := h.svc.GetDashboard(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Dashboard retrieved", dash)
}

// ListInvestments lists the caller's investments.
func (h *Handlers) ListInvestments(c *gin.Context) {
	page, perPage := pageParams(c)
	items, total, err := h.svc.ListInvestments(c.Request.Context(), c.GetString("user_id"), c.Query("status"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Investments retrieved", gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// GetInvestment returns a single investment owned by the caller.
func (h *Handlers) GetInvestment(c *gin.Context) {
	inv, err := h.svc.GetInvestment(c.Request.Context(), c.GetString("user_id"), c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Investment retrieved", inv)
}

// GetInvestmentRewards returns rewards for one investment.
func (h *Handlers) GetInvestmentRewards(c *gin.Context) {
	page, perPage := pageParams(c)
	items, total, err := h.svc.GetInvestmentRewards(c.Request.Context(), c.GetString("user_id"), c.Param("id"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Investment rewards retrieved", gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// GetRewardHistory returns all daily rewards for the caller.
func (h *Handlers) GetRewardHistory(c *gin.Context) {
	page, perPage := pageParams(c)
	items, _, err := h.svc.GetInvestmentRewards(c.Request.Context(), c.GetString("user_id"), "", page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Reward history retrieved", items)
}

// GetPaymentHistory returns payment transactions.
func (h *Handlers) GetPaymentHistory(c *gin.Context) {
	page, perPage := pageParams(c)
	items, _, err := h.svc.GetPaymentHistory(c.Request.Context(), c.GetString("user_id"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Payment history retrieved", items)
}

// GetWithdrawalHistory returns withdrawal requests.
func (h *Handlers) GetWithdrawalHistory(c *gin.Context) {
	page, perPage := pageParams(c)
	items, _, err := h.svc.GetWithdrawalHistory(c.Request.Context(), c.GetString("user_id"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Withdrawal history retrieved", items)
}

// RequestWithdrawal creates a withdrawal request.
func (h *Handlers) RequestWithdrawal(c *gin.Context) {
	var req models.WithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.Normalize()
	if (req.InvestmentID == nil || *req.InvestmentID == "") && req.AmountNGN <= 0 {
		response.BadRequest(c, "amount_ngn is required")
		return
	}
	w, err := h.svc.RequestWithdrawal(c.Request.Context(), c.GetString("user_id"), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, "Withdrawal requested", w)
}

// ─── Payments ─────────────────────────────────────────────

// InitPaystackPayment initializes a Paystack checkout.
func (h *Handlers) InitPaystackPayment(c *gin.Context) {
	var req models.InitEarningsPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.Normalize()
	if req.AmountUSD <= 0 {
		response.BadRequest(c, "amount_usd is required")
		return
	}
	h.logger.Info("Initializing Paystack transaction",
		zap.String("user_id", c.GetString("user_id")),
		zap.Float64("amount_usd", req.AmountUSD),
	)
	resp, err := h.svc.InitPaystackPayment(c.Request.Context(), c.GetString("user_id"), &req)
	if err != nil {
		h.logger.Error("Paystack init failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Payment initialized", resp)
}

// VerifyPaystackPayment re-verifies a Paystack checkout after the browser returns
// from the gateway (reference query param). Complements the webhook path.
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
		h.logger.Error("Paystack verify failed",
			zap.String("reference", reference),
			zap.Error(err),
		)
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Payment verified", inv)
}

// InitFlutterwavePayment initializes a Flutterwave checkout.
func (h *Handlers) InitFlutterwavePayment(c *gin.Context) {
	var req models.InitEarningsPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.Normalize()
	if req.AmountUSD <= 0 {
		response.BadRequest(c, "amount_usd is required")
		return
	}
	resp, err := h.svc.InitFlutterwavePayment(c.Request.Context(), c.GetString("user_id"), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Payment initialized", resp)
}

// PaystackWebhook handles Paystack payment webhooks.
func (h *Handlers) PaystackWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "unable to read body")
		return
	}
	signature := c.GetHeader("x-paystack-signature")
	if err := h.svc.ProcessPaystackWebhook(c.Request.Context(), payload, signature); err != nil {
		h.logger.Error("paystack earnings webhook failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Webhook processed", nil)
}

// FlutterwaveWebhook handles Flutterwave payment webhooks.
func (h *Handlers) FlutterwaveWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "unable to read body")
		return
	}
	signature := c.GetHeader("verif-hash")
	if err := h.svc.ProcessFlutterwaveWebhook(c.Request.Context(), payload, signature); err != nil {
		h.logger.Error("flutterwave earnings webhook failed", zap.Error(err))
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Webhook processed", nil)
}

// ─── Notifications ────────────────────────────────────────

// GetNotifications lists investment notifications.
func (h *Handlers) GetNotifications(c *gin.Context) {
	page, perPage := pageParams(c)
	items, _, err := h.svc.GetNotifications(c.Request.Context(), c.GetString("user_id"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Notifications retrieved", items)
}

// MarkNotificationRead marks one notification as read.
func (h *Handlers) MarkNotificationRead(c *gin.Context) {
	if err := h.svc.MarkNotificationRead(c.Request.Context(), c.GetString("user_id"), c.Param("id")); err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Notification marked as read", nil)
}

// MarkAllNotificationsRead marks all notifications as read.
func (h *Handlers) MarkAllNotificationsRead(c *gin.Context) {
	if err := h.svc.MarkAllNotificationsRead(c.Request.Context(), c.GetString("user_id")); err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "All notifications marked as read", nil)
}

// GetUnreadNotificationCount returns unread count.
func (h *Handlers) GetUnreadNotificationCount(c *gin.Context) {
	count, err := h.svc.GetUnreadNotificationCount(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Unread count retrieved", gin.H{"count": count})
}

// ─── Admin ────────────────────────────────────────────────

// AdminGetDashboard returns admin earnings analytics.
func (h *Handlers) AdminGetDashboard(c *gin.Context) {
	dash, err := h.svc.AdminGetDashboard(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Admin dashboard retrieved", dash)
}

// AdminUpdateSettings updates investment settings.
func (h *Handlers) AdminUpdateSettings(c *gin.Context) {
	var req models.AdminUpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	settings, err := h.svc.UpdateSettings(c.Request.Context(), &req, c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Settings updated", settings)
}

// AdminSetExchangeRate updates the USD→NGN rate.
func (h *Handlers) AdminSetExchangeRate(c *gin.Context) {
	var req models.AdminUpdateExchangeRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rate, err := h.svc.SetExchangeRate(c.Request.Context(), req.USDTNGN, c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Exchange rate updated", rate)
}

// AdminListWithdrawals lists withdrawals for review.
func (h *Handlers) AdminListWithdrawals(c *gin.Context) {
	page, perPage := pageParams(c)
	items, total, err := h.svc.AdminListWithdrawals(c.Request.Context(), c.Query("status"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Withdrawals retrieved", gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// AdminProcessWithdrawal approves or rejects a withdrawal.
func (h *Handlers) AdminProcessWithdrawal(c *gin.Context) {
	var req models.AdminWithdrawalActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	w, err := h.svc.AdminProcessWithdrawal(c.Request.Context(), c.Param("id"), &req, c.GetString("user_id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Withdrawal processed", w)
}

// AdminListInvestments lists all earnings investments.
func (h *Handlers) AdminListInvestments(c *gin.Context) {
	page, perPage := pageParams(c)
	items, total, err := h.svc.AdminListInvestments(c.Request.Context(), c.Query("status"), page, perPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, "Investments retrieved", gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// RegisterRoutes wires earnings investment routes under /api/v1/investments.
func RegisterRoutes(rg *gin.RouterGroup, h *Handlers, authMiddleware gin.HandlerFunc) {
	public := rg.Group("/investments")
	{
		public.GET("/exchange-rate", h.GetExchangeRate)
		public.GET("/settings", h.GetSettings)
		public.POST("/paystack/webhook", h.PaystackWebhook)
		public.POST("/flutterwave/webhook", h.FlutterwaveWebhook)
	}

	authed := rg.Group("/investments")
	authed.Use(authMiddleware)
	{
		authed.GET("/dashboard", h.GetDashboard)
		authed.POST("/paystack/init", h.InitPaystackPayment)
		authed.GET("/paystack/verify", h.VerifyPaystackPayment)
		authed.POST("/paystack/verify", h.VerifyPaystackPayment)
		authed.POST("/flutterwave/init", h.InitFlutterwavePayment)
		authed.GET("/list", h.ListInvestments)
		authed.GET("/rewards", h.GetRewardHistory)
		authed.GET("/payments", h.GetPaymentHistory)
		authed.GET("/withdrawals", h.GetWithdrawalHistory)
		authed.POST("/withdraw", h.RequestWithdrawal)
		authed.POST("/withdrawals", h.RequestWithdrawal)
		authed.POST("/settings", h.AdminUpdateSettings)
		authed.GET("/notifications/unread-count", h.GetUnreadNotificationCount)
		authed.PUT("/notifications/read-all", h.MarkAllNotificationsRead)
		authed.GET("/notifications", h.GetNotifications)
		authed.PUT("/notifications/:id/read", h.MarkNotificationRead)
		authed.GET("/:id/rewards", h.GetInvestmentRewards)
		authed.GET("/:id", h.GetInvestment)
	}

	admin := rg.Group("/admin/earnings")
	admin.Use(authMiddleware)
	admin.Use(middleware.RequireRole("admin", "super_admin"))
	{
		admin.GET("/dashboard", h.AdminGetDashboard)
		admin.PUT("/settings", h.AdminUpdateSettings)
		admin.PUT("/exchange-rate", h.AdminSetExchangeRate)
		admin.GET("/withdrawals", h.AdminListWithdrawals)
		admin.POST("/withdrawals/:id", h.AdminProcessWithdrawal)
		admin.GET("/investments", h.AdminListInvestments)
	}
}

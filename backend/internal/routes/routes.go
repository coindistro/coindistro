package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/auth"
	"github.com/coindistro/backend/internal/cache"
	"github.com/coindistro/backend/internal/config"
	"github.com/coindistro/backend/internal/database"
	"github.com/coindistro/backend/internal/earn/handlers"
	"github.com/coindistro/backend/internal/featureflags"
	"github.com/coindistro/backend/internal/health"
	identityhandlers "github.com/coindistro/backend/internal/identity/handlers"
	"github.com/coindistro/backend/internal/metrics"
	"github.com/coindistro/backend/internal/middleware"
	"github.com/coindistro/backend/internal/rbac"
)

// SetupRouter configures all routes and returns a Gin engine.
func SetupRouter(
	cfg *config.Config,
	logger *zap.Logger,
	db *database.Database,
	redis *cache.Cache,
	authService *auth.Auth,
	rbacService *rbac.RBAC,
	featureFlags *featureflags.Manager,
	promMetrics *metrics.Metrics,
	identityHandlers *identityhandlers.Handlers,
	earnHandlers *handlers.Handlers,
) *gin.Engine {
	ginMode := gin.ReleaseMode
	if cfg.App.Debug {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	r := gin.New()

	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS(cfg.CORS))
	r.Use(middleware.Compression())
	r.Use(middleware.RateLimiter(cfg.RateLimiter, redis))

	// Prometheus metrics middleware
	if cfg.Monitoring.PrometheusEnabled && promMetrics != nil {
		r.Use(metrics.Middleware(promMetrics))
	}

	// Health checker
	healthChecker := health.New(db, redis, logger)

	// Health endpoints (no version prefix, used by orchestrators)
	r.GET("/health", healthChecker.Health)
	r.GET("/ready", healthChecker.Ready)
	r.GET("/live", healthChecker.Live)

	// Metrics endpoint
	if cfg.Monitoring.PrometheusEnabled && promMetrics != nil {
		r.GET("/metrics", gin.WrapH(metrics.Handler()))
	}

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Public routes (no authentication required)
		v1.GET("/health", healthChecker.Health)
		v1.GET("/features", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"flags": featureFlags.GetAllFlags(),
			})
		})

		// Public auth + availability routes
		identityhandlers.RegisterAuthRoutes(v1, identityHandlers)
		identityhandlers.RegisterPublicUserRoutes(v1, identityHandlers)

		// Authenticated routes
		protected := v1.Group("")
		protected.Use(middleware.Authentication(authService))
		{
			// User profile routes
			identityhandlers.RegisterUserRoutes(protected, identityHandlers)

			// Authenticated auth routes
			identityhandlers.RegisterProtectedAuthRoutes(protected, identityHandlers)

			// Session routes
			identityhandlers.RegisterSessionRoutes(protected, identityHandlers)

			// Device routes
			identityhandlers.RegisterDeviceRoutes(protected, identityHandlers)

			// Referral routes
			identityhandlers.RegisterReferralRoutes(protected, identityHandlers)

			// Invitation routes
			identityhandlers.RegisterInvitationRoutes(protected, identityHandlers)

			// Security/activity routes
			identityhandlers.RegisterSecurityRoutes(protected, identityHandlers)
		}

		// Earn module routes (/api/v1/earn/...)
		if earnHandlers != nil {
			handlers.RegisterRoutes(v1, earnHandlers, middleware.Authentication(authService))
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(middleware.Authentication(authService))
		admin.Use(middleware.RequireRole("admin", "super_admin"))
		{
			registerAdminRoutes(admin, rbacService, featureFlags)
			identityhandlers.RegisterAdminRoutes(admin, identityHandlers)
		}
	}

	return r
}

// registerAdminRoutes registers admin-related routes (non-identity).
func registerAdminRoutes(rg *gin.RouterGroup, rbacService *rbac.RBAC, ff *featureflags.Manager) {
	rg.GET("/dashboard", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Admin dashboard - not implemented"})
	})
	rg.GET("/settings", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Get system settings - not implemented"})
	})
	rg.PUT("/settings", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Update system settings - not implemented"})
	})
	rg.GET("/logs", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Get system logs - not implemented"})
	})

	// RBAC management
	rg.GET("/roles", func(c *gin.Context) {
		c.JSON(200, gin.H{"roles": rbacService.GetRoles()})
	})
	rg.GET("/roles/:role/permissions", func(c *gin.Context) {
		role, err := rbac.ParseRole(c.Param("role"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"role":        role,
			"permissions": rbacService.GetPermissions(role),
		})
	})

	// Feature flags management
	rg.GET("/features", func(c *gin.Context) {
		c.JSON(200, gin.H{"flags": ff.GetAllFlags()})
	})
	rg.PUT("/features/:flag", func(c *gin.Context) {
		flagName := c.Param("flag")
		var req struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		if err := ff.Set(flagName, req.Enabled); err != nil {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"flag": flagName, "enabled": req.Enabled})
	})
}

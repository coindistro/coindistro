package routes

import (
	"context"
	"time"

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
	"github.com/coindistro/backend/internal/response"
	"github.com/coindistro/backend/internal/scheduler"
	"github.com/coindistro/backend/internal/workers"
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
	workerPool *workers.Pool,
	sched *scheduler.Scheduler,
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
			response.OK(c, "Feature flags", gin.H{
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

		// Admin routes — super_admin, admin, and moderator
		admin := v1.Group("/admin")
		admin.Use(middleware.Authentication(authService))
		admin.Use(middleware.RequireRole("super_admin", "admin", "moderator"))
		{
			registerAdminRoutes(admin, cfg, db, redis, rbacService, featureFlags, workerPool, sched)
			identityhandlers.RegisterAdminRoutes(admin, identityHandlers)
		}
	}

	return r
}

// registerAdminRoutes registers admin-related routes (non-identity).
func registerAdminRoutes(
	rg *gin.RouterGroup,
	cfg *config.Config,
	db *database.Database,
	redis *cache.Cache,
	rbacService *rbac.RBAC,
	ff *featureflags.Manager,
	workerPool *workers.Pool,
	sched *scheduler.Scheduler,
) {
	// Live admin overview combining system + platform metadata.
	// Identity stats are also available via GET /admin/stats.
	rg.GET("/dashboard", func(c *gin.Context) {
		system := buildSystemStatus(cfg, db, redis, workerPool, sched, ff)
		response.OK(c, "Admin dashboard", gin.H{
			"system": system,
		})
	})

	rg.GET("/system", func(c *gin.Context) {
		response.OK(c, "System status", buildSystemStatus(cfg, db, redis, workerPool, sched, ff))
	})

	rg.GET("/settings", func(c *gin.Context) {
		response.OK(c, "System settings", gin.H{
			"app": gin.H{
				"name":        cfg.App.Name,
				"version":     cfg.App.Version,
				"environment": cfg.App.Environment,
				"debug":       cfg.App.Debug,
			},
		})
	})

	rg.PUT("/settings", func(c *gin.Context) {
		response.OK(c, "Update system settings - not implemented", nil)
	})

	rg.GET("/logs", func(c *gin.Context) {
		response.OK(c, "Get system logs - not implemented", nil)
	})

	// RBAC management
	rg.GET("/roles", func(c *gin.Context) {
		response.OK(c, "Roles retrieved", gin.H{"roles": rbacService.GetRoles()})
	})
	rg.GET("/roles/:role/permissions", func(c *gin.Context) {
		role, err := rbac.ParseRole(c.Param("role"))
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		response.OK(c, "Permissions retrieved", gin.H{
			"role":        role,
			"permissions": rbacService.GetPermissions(role),
		})
	})

	// Feature flags management
	rg.GET("/features", func(c *gin.Context) {
		response.OK(c, "Feature flags", gin.H{"flags": ff.GetAllFlags()})
	})
	rg.PUT("/features/:flag", func(c *gin.Context) {
		flagName := c.Param("flag")
		var req struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request")
			return
		}
		if err := ff.Set(flagName, req.Enabled); err != nil {
			response.NotFound(c, err.Error())
			return
		}
		response.OK(c, "Feature flag updated", gin.H{"flag": flagName, "enabled": req.Enabled})
	})

	// Workers status
	rg.GET("/workers", func(c *gin.Context) {
		if workerPool == nil {
			response.OK(c, "Workers", gin.H{
				"enabled": false,
				"status":  "disabled",
			})
			return
		}
		response.OK(c, "Workers", workerPool.Status())
	})

	// Scheduler status
	rg.GET("/scheduler", func(c *gin.Context) {
		if sched == nil {
			response.OK(c, "Scheduler", gin.H{
				"enabled": false,
				"status":  "disabled",
				"tasks":   []interface{}{},
			})
			return
		}
		tasks := sched.GetStatus()
		response.OK(c, "Scheduler", gin.H{
			"enabled": true,
			"status":  "running",
			"tasks":   tasks,
		})
	})
}

func buildSystemStatus(
	cfg *config.Config,
	db *database.Database,
	redis *cache.Cache,
	workerPool *workers.Pool,
	sched *scheduler.Scheduler,
	ff *featureflags.Manager,
) gin.H {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	dbStatus := "not_configured"
	if db != nil {
		if err := db.Ping(ctx); err != nil {
			dbStatus = "unhealthy"
		} else {
			dbStatus = "healthy"
		}
	}

	redisStatus := "not_configured"
	if redis != nil {
		if err := redis.Ping(ctx); err != nil {
			redisStatus = "unhealthy"
		} else {
			redisStatus = "healthy"
		}
	}

	apiStatus := "healthy"
	if dbStatus == "unhealthy" || redisStatus == "unhealthy" {
		apiStatus = "degraded"
	}

	workerStatus := gin.H{"enabled": false, "status": "disabled"}
	if workerPool != nil {
		workerStatus = workerPool.Status()
	}

	schedulerStatus := gin.H{"enabled": false, "status": "disabled", "task_count": 0}
	if sched != nil {
		tasks := sched.GetStatus()
		schedulerStatus = gin.H{
			"enabled":    true,
			"status":     "running",
			"task_count": len(tasks),
			"tasks":      tasks,
		}
	}

	flags := ff.GetAllFlags()

	return gin.H{
		"status":       apiStatus,
		"api_status":   apiStatus,
		"database":     dbStatus,
		"redis":        redisStatus,
		"backend":      "healthy",
		"docker":       "unknown",
		"version":      cfg.App.Version,
		"environment":  cfg.App.Environment,
		"app_name":     cfg.App.Name,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"workers":      workerStatus,
		"scheduler":    schedulerStatus,
		"feature_flags": flags,
	}
}

package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/cache"
	"github.com/coindistro/backend/internal/database"
	"github.com/coindistro/backend/internal/response"
)

// Checker performs health checks on dependencies.
type Checker struct {
	db     *database.Database
	redis  *cache.Cache
	logger *zap.Logger
}

// New creates a new Health checker.
func New(db *database.Database, redis *cache.Cache, logger *zap.Logger) *Checker {
	return &Checker{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Checks    map[string]string `json:"checks"`
}

// Health handles the GET /health endpoint.
func (h *Checker) Health(c *gin.Context) {
	checks := make(map[string]string)
	overallStatus := "healthy"

	// Check database
	if h.db != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if err := h.db.Ping(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			overallStatus = "degraded"
			h.logger.Error("health check: database unhealthy", zap.Error(err))
		} else {
			checks["database"] = "healthy"
		}
	} else {
		checks["database"] = "not_configured"
	}

	// Check Redis
	if h.redis != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if err := h.redis.Ping(ctx); err != nil {
			checks["redis"] = "unhealthy: " + err.Error()
			overallStatus = "degraded"
			h.logger.Error("health check: redis unhealthy", zap.Error(err))
		} else {
			checks["redis"] = "healthy"
		}
	} else {
		checks["redis"] = "not_configured"
	}

	checks["server"] = "healthy"

	statusCode := http.StatusOK
	if overallStatus == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Checks:    checks,
	})
}

// Ready handles the GET /ready endpoint.
func (h *Checker) Ready(c *gin.Context) {
	checks := make(map[string]string)
	allReady := true

	// Check database
	if h.db != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if err := h.db.Ping(ctx); err != nil {
			checks["database"] = "not_ready: " + err.Error()
			allReady = false
		} else {
			checks["database"] = "ready"
		}
	} else {
		checks["database"] = "not_configured"
	}

	// Check Redis
	if h.redis != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if err := h.redis.Ping(ctx); err != nil {
			checks["redis"] = "not_ready: " + err.Error()
			allReady = false
		} else {
			checks["redis"] = "ready"
		}
	} else {
		checks["redis"] = "not_configured"
	}

	checks["server"] = "ready"

	if !allReady {
		response.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service is not ready")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks":    checks,
	})
}

// Live handles the GET /live endpoint (simple liveness probe).
func (h *Checker) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

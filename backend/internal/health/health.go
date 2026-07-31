package health

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/cache"
	"github.com/coindistro/backend/internal/database"
	"github.com/coindistro/backend/internal/response"
)

// Checker performs health checks on dependencies.
type Checker struct {
	db      *database.Database
	redis   *cache.Cache
	logger  *zap.Logger
	version string
}

// New creates a new Health checker.
func New(db *database.Database, redis *cache.Cache, logger *zap.Logger) *Checker {
	return &Checker{
		db:      db,
		redis:   redis,
		logger:  logger,
		version: "v0.4.0-alpha",
	}
}

// NewWithVersion creates a new Health checker with a specific version string.
func NewWithVersion(db *database.Database, redis *cache.Cache, logger *zap.Logger, version string) *Checker {
	return &Checker{
		db:      db,
		redis:   redis,
		logger:  logger,
		version: version,
	}
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Checks    map[string]string `json:"checks"`
}

// Health handles GET /health for orchestrator liveness (Render, Docker, k8s).
// It always returns HTTP 200 as soon as the process can serve traffic so platforms
// do not SIGTERM a successfully started server when a dependency is briefly degraded.
// Dependency status is reported in the JSON body; use /ready for strict readiness.
func (h *Checker) Health(c *gin.Context) {
	checks := make(map[string]string)
	overallStatus := "healthy"

	// Keep probes fast so platform health timeouts (often ~5s) are never exceeded.
	if h.db != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		if err := h.db.Ping(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			overallStatus = "degraded"
			h.logger.Warn("health check: database unhealthy", zap.Error(err))
		} else {
			checks["database"] = "healthy"
			if tableErr := h.verifySchema(ctx); tableErr != nil {
				checks["schema"] = "unhealthy: " + tableErr.Error()
				overallStatus = "degraded"
				h.logger.Warn("health check: schema unhealthy", zap.Error(tableErr))
			} else {
				checks["schema"] = "healthy"
			}
		}
		cancel()
	} else {
		checks["database"] = "not_configured"
	}

	if h.redis != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		if err := h.redis.Ping(ctx); err != nil {
			status := h.redis.HealthStatus()
			checks["redis"] = status
			overallStatus = "degraded"
			h.logger.Warn("health check: redis unhealthy", zap.String("status", status), zap.Error(err))
		} else {
			checks["redis"] = "healthy"
		}
		cancel()
	} else {
		checks["redis"] = "not_configured"
	}

	checks["server"] = "healthy"

	c.JSON(http.StatusOK, HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   h.version,
		Checks:    checks,
	})
}

// Ready handles the GET /ready endpoint (strict dependency readiness).
func (h *Checker) Ready(c *gin.Context) {
	checks := make(map[string]string)
	allReady := true

	if h.db != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if err := h.db.Ping(ctx); err != nil {
			checks["database"] = "not_ready: " + err.Error()
			allReady = false
		} else if tableErr := h.verifySchema(ctx); tableErr != nil {
			checks["schema"] = "not_ready: " + tableErr.Error()
			allReady = false
		} else {
			checks["database"] = "ready"
		}
	} else {
		checks["database"] = "not_configured"
	}

	if h.redis != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if err := h.redis.Ping(ctx); err != nil {
			checks["redis"] = h.redis.HealthStatus()
			allReady = false
		} else {
			checks["redis"] = "healthy"
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

func (h *Checker) verifySchema(ctx context.Context) error {
	requiredTables := []string{"identity_users", "sessions", "refresh_tokens", "audit_logs", "schema_migrations"}
	found, err := h.existingTables(ctx, requiredTables)
	if err != nil {
		return err
	}
	missing := requiredTables[:0]
	for _, table := range requiredTables {
		if !found[table] {
			missing = append(missing, table)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required tables: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (h *Checker) existingTables(ctx context.Context, tables []string) (map[string]bool, error) {
	rows, err := h.db.Pool.Query(ctx, `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name = ANY($1)
`, tables)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	found := make(map[string]bool, len(tables))
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		found[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return found, nil
}

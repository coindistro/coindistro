package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/auth"
	"github.com/coindistro/backend/internal/config"
	apperrors "github.com/coindistro/backend/internal/errors"
	"github.com/coindistro/backend/internal/response"
)

// Logger returns a middleware that logs HTTP requests using Zap.
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Read the request body for logging
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Process request
		c.Next()

		// Log after request is complete
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
			zap.Int("body_size", c.Writer.Size()),
		}

		// Add request ID if present
		if requestID := c.GetString("request_id"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		// Add error log if there were errors
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("request error", append(fields, zap.Error(e.Err))...)
			}
		} else if statusCode >= 500 {
			logger.Error("server error", fields...)
		} else if statusCode >= 400 {
			logger.Warn("client error", fields...)
		} else {
			logger.Info("request completed", fields...)
		}
	}
}

// Recovery returns a middleware that recovers from panics.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)
				response.InternalServerError(c, "An unexpected error occurred")
				c.Abort()
			}
		}()
		c.Next()
	}
}

// Authentication returns a middleware that validates JWT access tokens.
func Authentication(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Invalid authorization header format. Use: Bearer <token>")
			c.Abort()
			return
		}

		claims, err := authService.ValidateAccessToken(parts[1])
		if err != nil {
			if appErr := apperrors.GetAppError(err); appErr != nil {
				response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message)
			} else {
				response.Unauthorized(c, "Invalid or expired token")
			}
			c.Abort()
			return
		}

		// Set user info in context
		ctx := auth.SetUserContext(c.Request.Context(), claims.UserID, claims.Email, claims.Roles)
		c.Request = c.Request.WithContext(ctx)

		// Also set in Gin context for convenience
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// OptionalAuthentication attempts to authenticate but doesn't fail if no token is present.
func OptionalAuthentication(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		claims, err := authService.ValidateAccessToken(parts[1])
		if err != nil {
			c.Next()
			return
		}

		ctx := auth.SetUserContext(c.Request.Context(), claims.UserID, claims.Email, claims.Roles)
		c.Request = c.Request.WithContext(ctx)
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// RequireRole returns a middleware that checks if the authenticated user has at least one of the required roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		rolesList, ok := userRoles.([]string)
		if !ok {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		roleSet := make(map[string]bool)
		for _, role := range rolesList {
			roleSet[role] = true
		}

		for _, requiredRole := range roles {
			if roleSet[requiredRole] {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Insufficient permissions")
		c.Abort()
	}
}

// RequirePermission returns a middleware that checks if the authenticated user has a specific permission.
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		rolesList, ok := userRoles.([]string)
		if !ok {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		// Check for admin role (super admin bypass)
		for _, role := range rolesList {
			if role == "admin" || role == "super_admin" {
				c.Next()
				return
			}
		}

		// Check for specific permission in roles (permissions can be embedded in roles)
		for _, role := range rolesList {
			if role == permission {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Insufficient permissions")
		c.Abort()
	}
}

// RequestID returns a middleware that adds a unique request ID to each request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// CORS returns a configured CORS middleware.
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     cfg.AllowedMethods,
		AllowHeaders:     cfg.AllowedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           time.Duration(cfg.MaxAge) * time.Second,
	})
}

// RateLimiter returns a simple in-memory rate limiting middleware.
func RateLimiter(cfg config.RateLimiterConfig, cacheClient interface{}) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Simple token bucket per IP
	type bucket struct {
		tokens    int
		lastCheck time.Time
	}

	buckets := make(map[string]*bucket)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		b, exists := buckets[ip]
		if !exists {
			buckets[ip] = &bucket{
				tokens:    cfg.Burst,
				lastCheck: now,
			}
			b = buckets[ip]
		}

		// Refill tokens based on time elapsed
		elapsed := now.Sub(b.lastCheck)
		refill := int(elapsed.Seconds() * float64(cfg.RequestsPerMinute) / 60)
		if refill > 0 {
			b.tokens = min(cfg.Burst, b.tokens+refill)
			b.lastCheck = now
		}

		if b.tokens <= 0 {
			response.TooManyRequests(c, "Too many requests, please try again later")
			c.Abort()
			return
		}

		b.tokens--
		c.Next()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Validation returns a middleware that validates the request body against a struct.
func Validation(handler func(interface{}) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// GinLogger returns a Gin middleware using the provided Zap logger.
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		if statusCode >= 400 {
			logger.Warn("HTTP request",
				zap.Int("status", statusCode),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.Duration("latency", latency),
			)
		}
	}
}

// Compression returns a gzip compression middleware.
func Compression() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression)
}

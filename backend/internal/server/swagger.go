package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/coindistro/backend/docs"
	"github.com/coindistro/backend/internal/config"
)

// ConfigureSwagger sets SwaggerInfo host, schemes, and base path from runtime config.
//
// Production (Render):
//
//	Host:    coindistro.onrender.com  (or SWAGGER_HOST / RENDER_EXTERNAL_HOSTNAME)
//	Schemes: https
//	BasePath: /api/v1
//
// Development:
//
//	Host:    localhost:<port>
//	Schemes: http
//	BasePath: /api/v1
//
// Route annotations use paths relative to BasePath (e.g. /auth/login), never /api/v1/...
// so Try-it-out never produces /api/v1/api/v1/...
func ConfigureSwagger(cfg *config.Config) {
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Title = "Coindistro API"
	docs.SwaggerInfo.Version = cfg.App.Version
	if docs.SwaggerInfo.Version == "" {
		docs.SwaggerInfo.Version = "1.0.0"
	}

	// Explicit override for custom domains / staging hosts.
	if host := strings.TrimSpace(os.Getenv("SWAGGER_HOST")); host != "" {
		docs.SwaggerInfo.Host = stripScheme(host)
		docs.SwaggerInfo.Schemes = schemesFromEnv(os.Getenv("SWAGGER_SCHEMES"), productionLike(cfg))
		return
	}

	if productionLike(cfg) {
		host := strings.TrimSpace(os.Getenv("RENDER_EXTERNAL_HOSTNAME"))
		if host == "" {
			host = "coindistro.onrender.com"
		}
		docs.SwaggerInfo.Host = stripScheme(host)
		docs.SwaggerInfo.Schemes = []string{"https"}
		return
	}

	// Local / development
	port := cfg.Server.Port
	if port <= 0 {
		port = 8080
	}
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", port)
	docs.SwaggerInfo.Schemes = []string{"http"}
}

func productionLike(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	if cfg.App.IsProduction() {
		return true
	}
	env := strings.ToLower(strings.TrimSpace(cfg.App.Environment))
	// Render sets RENDER=true; treat as production for Swagger host.
	if os.Getenv("RENDER") != "" {
		return true
	}
	return env == "staging" || env == "prod"
}

func stripScheme(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	return strings.TrimSuffix(host, "/")
}

func schemesFromEnv(raw string, preferHTTPS bool) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if preferHTTPS {
			return []string{"https"}
		}
		return []string{"http"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "http" || p == "https" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		if preferHTTPS {
			return []string{"https"}
		}
		return []string{"http"}
	}
	return out
}

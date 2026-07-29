package server

import (
	"testing"

	"github.com/coindistro/backend/docs"
	"github.com/coindistro/backend/internal/config"
)

func TestConfigureSwagger_Production(t *testing.T) {
	t.Setenv("RENDER", "true")
	t.Setenv("SWAGGER_HOST", "")
	t.Setenv("RENDER_EXTERNAL_HOSTNAME", "")

	cfg := &config.Config{}
	cfg.App.Environment = "production"
	cfg.App.Version = "v0.4.0-alpha"
	cfg.Server.Port = 8080

	ConfigureSwagger(cfg)

	if docs.SwaggerInfo.Host != "coindistro.onrender.com" {
		t.Fatalf("host = %q, want coindistro.onrender.com", docs.SwaggerInfo.Host)
	}
	if docs.SwaggerInfo.BasePath != "/api/v1" {
		t.Fatalf("basePath = %q, want /api/v1", docs.SwaggerInfo.BasePath)
	}
	if len(docs.SwaggerInfo.Schemes) != 1 || docs.SwaggerInfo.Schemes[0] != "https" {
		t.Fatalf("schemes = %v, want [https]", docs.SwaggerInfo.Schemes)
	}
}

func TestConfigureSwagger_Development(t *testing.T) {
	t.Setenv("RENDER", "")
	t.Setenv("SWAGGER_HOST", "")
	t.Setenv("RENDER_EXTERNAL_HOSTNAME", "")

	cfg := &config.Config{}
	cfg.App.Environment = "development"
	cfg.Server.Port = 8080

	ConfigureSwagger(cfg)

	if docs.SwaggerInfo.Host != "localhost:8080" {
		t.Fatalf("host = %q, want localhost:8080", docs.SwaggerInfo.Host)
	}
	if docs.SwaggerInfo.BasePath != "/api/v1" {
		t.Fatalf("basePath = %q, want /api/v1", docs.SwaggerInfo.BasePath)
	}
	if len(docs.SwaggerInfo.Schemes) != 1 || docs.SwaggerInfo.Schemes[0] != "http" {
		t.Fatalf("schemes = %v, want [http]", docs.SwaggerInfo.Schemes)
	}
}

func TestConfigureSwagger_OverrideHost(t *testing.T) {
	t.Setenv("SWAGGER_HOST", "https://api.example.com")
	t.Setenv("SWAGGER_SCHEMES", "https")

	cfg := &config.Config{}
	cfg.App.Environment = "development"
	cfg.Server.Port = 9999

	ConfigureSwagger(cfg)

	if docs.SwaggerInfo.Host != "api.example.com" {
		t.Fatalf("host = %q, want api.example.com", docs.SwaggerInfo.Host)
	}
	if docs.SwaggerInfo.Schemes[0] != "https" {
		t.Fatalf("schemes = %v", docs.SwaggerInfo.Schemes)
	}
}

func TestConfigureSwagger_NoDoubleAPIPrefixInBasePath(t *testing.T) {
	cfg := &config.Config{}
	cfg.App.Environment = "production"
	ConfigureSwagger(cfg)
	if docs.SwaggerInfo.BasePath == "/api/v1/api/v1" {
		t.Fatal("base path must not double /api/v1")
	}
	if docs.SwaggerInfo.BasePath != "/api/v1" {
		t.Fatalf("basePath = %q", docs.SwaggerInfo.BasePath)
	}
}

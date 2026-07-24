package bootstrap

import (
	"os"
	"testing"

	"github.com/coindistro/backend/internal/config"
)

func TestEnsureDevelopmentEnv_AllowsDev(t *testing.T) {
	cfg := &config.Config{}
	cfg.App.Environment = "development"
	_ = os.Unsetenv("APP_ENV")
	_ = os.Unsetenv("ENV")
	_ = os.Unsetenv("COINDISTRO_ENV")
	if err := EnsureDevelopmentEnv(cfg); err != nil {
		t.Fatalf("expected development allowed: %v", err)
	}
}

func TestEnsureDevelopmentEnv_BlocksProduction(t *testing.T) {
	cfg := &config.Config{}
	cfg.App.Environment = "production"
	_ = os.Unsetenv("APP_ENV")
	_ = os.Unsetenv("ENV")
	_ = os.Unsetenv("COINDISTRO_ENV")
	err := EnsureDevelopmentEnv(cfg)
	if err == nil {
		t.Fatal("expected production to be blocked")
	}
	if err.Error() != "Bootstrap is disabled in production" {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestEnsureDevelopmentEnv_EnvOverride(t *testing.T) {
	cfg := &config.Config{}
	cfg.App.Environment = "production"
	t.Setenv("APP_ENV", "development")
	if err := EnsureDevelopmentEnv(cfg); err != nil {
		t.Fatalf("APP_ENV=development should allow: %v", err)
	}
}

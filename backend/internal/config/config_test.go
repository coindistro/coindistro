package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_RedisHostPortFromEnv(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
app:
  environment: development
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("COINDISTRO_REDIS_HOST", "redis.internal")
	t.Setenv("COINDISTRO_REDIS_PORT", "6380")
	t.Setenv("COINDISTRO_REDIS_PASSWORD", "secret")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Redis.Host != "redis.internal" {
		t.Fatalf("redis host = %q, want %q", cfg.Redis.Host, "redis.internal")
	}
	if cfg.Redis.Port != 6380 {
		t.Fatalf("redis port = %d, want 6380", cfg.Redis.Port)
	}
	if cfg.Redis.Password != "secret" {
		t.Fatalf("redis password = %q, want %q", cfg.Redis.Password, "secret")
	}
}

func TestLoad_RedisURLConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
app:
  environment: development
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("COINDISTRO_REDIS_URL", "redis://:secret@redis.example.com:6380/2")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Redis.Host != "redis.example.com" {
		t.Fatalf("redis host = %q, want %q", cfg.Redis.Host, "redis.example.com")
	}
	if cfg.Redis.Port != 6380 {
		t.Fatalf("redis port = %d, want 6380", cfg.Redis.Port)
	}
	if cfg.Redis.Password != "secret" {
		t.Fatalf("redis password = %q, want %q", cfg.Redis.Password, "secret")
	}
	if cfg.Redis.DB != 2 {
		t.Fatalf("redis db = %d, want 2", cfg.Redis.DB)
	}
}

func TestLoad_RedisTLSURLConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
app:
  environment: development
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("COINDISTRO_REDIS_URL", "rediss://:secret@upstash.example.com:6380/1")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !cfg.Redis.TLSEnabled {
		t.Fatal("expected TLS to be enabled for rediss:// URLs")
	}
}

func TestLoad_ProductionRequiresRedisPassword(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
app:
  environment: production
auth:
  access_token_secret: test-access
  refresh_token_secret: test-refresh
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("COINDISTRO_DB_HOST", "db.internal")
	t.Setenv("COINDISTRO_REDIS_HOST", "redis.internal")
	t.Setenv("COINDISTRO_REDIS_PASSWORD", "")

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected production load to fail when Redis password is missing")
	}
}

func TestLoad_InvalidRedisURL(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
app:
  environment: development
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("COINDISTRO_REDIS_URL", "http://not-redis.example.com")

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected invalid redis URL to fail")
	}
}

func TestLoad_DefaultJWTMinutes(t *testing.T) {
	// Isolate from ambient env that may override TTLs.
	t.Setenv("COINDISTRO_JWT_ACCESS_TTL", "")
	t.Setenv("COINDISTRO_JWT_REFRESH_TTL", "")
	// Unset so BindEnv doesn't pick residual values — empty string may still bind.
	_ = os.Unsetenv("COINDISTRO_JWT_ACCESS_TTL")
	_ = os.Unsetenv("COINDISTRO_JWT_REFRESH_TTL")

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
app:
  environment: development
auth:
  access_token_secret: "test-access"
  refresh_token_secret: "test-refresh"
  access_token_ttl: 15
  refresh_token_ttl: 10080
  issuer: coindistro
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Auth.AccessTokenTTL != 15*time.Minute {
		t.Fatalf("access TTL = %v, want 15m", cfg.Auth.AccessTokenTTL)
	}
	if cfg.Auth.RefreshTokenTTL != 10080*time.Minute {
		t.Fatalf("refresh TTL = %v, want 10080m", cfg.Auth.RefreshTokenTTL)
	}
}

func TestLoad_DurationStringEnv(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
app:
  environment: development
auth:
  access_token_secret: "test-access"
  refresh_token_secret: "test-refresh"
  issuer: coindistro
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("COINDISTRO_JWT_ACCESS_TTL", "15m")
	t.Setenv("COINDISTRO_JWT_REFRESH_TTL", "168h")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Auth.AccessTokenTTL != 15*time.Minute {
		t.Fatalf("access TTL = %v, want 15m", cfg.Auth.AccessTokenTTL)
	}
	if cfg.Auth.RefreshTokenTTL != 168*time.Hour {
		t.Fatalf("refresh TTL = %v, want 168h", cfg.Auth.RefreshTokenTTL)
	}
}

func TestLoad_NumericStringEnv(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
app:
  environment: development
auth:
  access_token_secret: "test-access"
  refresh_token_secret: "test-refresh"
  issuer: coindistro
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Classic Render / .env form that previously crashed startup.
	t.Setenv("COINDISTRO_JWT_ACCESS_TTL", "15")
	t.Setenv("COINDISTRO_JWT_REFRESH_TTL", "10080")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Auth.AccessTokenTTL != 15*time.Minute {
		t.Fatalf("access TTL = %v, want 15m", cfg.Auth.AccessTokenTTL)
	}
	if cfg.Auth.RefreshTokenTTL != 10080*time.Minute {
		t.Fatalf("refresh TTL = %v, want 10080m", cfg.Auth.RefreshTokenTTL)
	}
}

func TestLoad_InvalidDuration(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
app:
  environment: development
auth:
  access_token_secret: "test-access"
  refresh_token_secret: "test-refresh"
  issuer: coindistro
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("COINDISTRO_JWT_ACCESS_TTL", "not-valid")
	t.Setenv("COINDISTRO_JWT_REFRESH_TTL", "168h")

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid access TTL")
	}
}

func TestLoad_FallbackDefaultsWhenZero(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
app:
  environment: development
auth:
  access_token_secret: "test-access"
  refresh_token_secret: "test-refresh"
  access_token_ttl: 0
  refresh_token_ttl: 0
  issuer: coindistro
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("COINDISTRO_JWT_ACCESS_TTL", "0")
	t.Setenv("COINDISTRO_JWT_REFRESH_TTL", "0")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Auth.AccessTokenTTL != 15*time.Minute {
		t.Fatalf("access default = %v, want 15m", cfg.Auth.AccessTokenTTL)
	}
	if cfg.Auth.RefreshTokenTTL != 10080*time.Minute {
		t.Fatalf("refresh default = %v, want 10080m", cfg.Auth.RefreshTokenTTL)
	}
}

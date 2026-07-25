package config

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	App          AppConfig          `mapstructure:"app"`
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Auth         AuthConfig         `mapstructure:"auth"`
	RateLimiter  RateLimiterConfig  `mapstructure:"rate_limiter"`
	CORS         CORSConfig         `mapstructure:"cors"`
	Logging      LoggingConfig      `mapstructure:"logging"`
	Telemetry    TelemetryConfig    `mapstructure:"telemetry"`
	Email        EmailConfig        `mapstructure:"email"`
	Storage      StorageConfig      `mapstructure:"storage"`
	FeatureFlags FeatureFlagsConfig `mapstructure:"feature_flags"`
	Workers      WorkersConfig      `mapstructure:"workers"`
	Scheduler    SchedulerConfig    `mapstructure:"scheduler"`
	Monitoring   MonitoringConfig   `mapstructure:"monitoring"`
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	MaxRequestSize int           `mapstructure:"max_request_size"`
}

// DatabaseConfig holds PostgreSQL connection configuration.
type DatabaseConfig struct {
	Host                string `mapstructure:"host"`
	Port                int    `mapstructure:"port"`
	User                string `mapstructure:"user"`
	Password            string `mapstructure:"password"`
	DBName              string `mapstructure:"dbname"`
	SSLMode             string `mapstructure:"ssl_mode"`
	MaxOpenConns        int    `mapstructure:"max_open_conns"`
	MaxIdleConns        int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime     int    `mapstructure:"conn_max_lifetime"`
	HealthCheckInterval int    `mapstructure:"health_check_interval"`
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

// IsConfigured returns true if the database has been configured with a host.
func (d DatabaseConfig) IsConfigured() bool {
	return d.Host != ""
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	URL          string `mapstructure:"url"`
	DB           int    `mapstructure:"db"`
	TLSEnabled   bool   `mapstructure:"tls_enabled"`
	DialTimeout  int    `mapstructure:"dial_timeout"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// AuthConfig holds JWT authentication configuration.
type AuthConfig struct {
	AccessTokenSecret  string        `mapstructure:"access_token_secret"`
	RefreshTokenSecret string        `mapstructure:"refresh_token_secret"`
	AccessTokenTTL     time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL    time.Duration `mapstructure:"refresh_token_ttl"`
	Issuer             string        `mapstructure:"issuer"`
}

// HasValidSecrets returns true if both JWT secrets are set to non-default values.
func (a AuthConfig) HasValidSecrets() bool {
	return a.AccessTokenSecret != "" && a.AccessTokenSecret != "change-me-in-production" &&
		a.RefreshTokenSecret != "" && a.RefreshTokenSecret != "change-me-in-production-too"
}

// RateLimiterConfig holds rate limiting configuration.
type RateLimiterConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
	Burst             int  `mapstructure:"burst"`
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level            string   `mapstructure:"level"`
	Encoding         string   `mapstructure:"encoding"`
	OutputPaths      []string `mapstructure:"output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
}

// TelemetryConfig holds OpenTelemetry configuration.
type TelemetryConfig struct {
	Enabled     bool    `mapstructure:"enabled"`
	ServiceName string  `mapstructure:"service_name"`
	Endpoint    string  `mapstructure:"endpoint"`
	SampleRate  float64 `mapstructure:"sample_rate"`
}

// EmailConfig holds email/SMTP configuration.
type EmailConfig struct {
	Provider string     `mapstructure:"provider"` // smtp, noop
	SMTP     SMTPConfig `mapstructure:"smtp"`
}

// SMTPConfig holds SMTP server configuration.
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	FromName string `mapstructure:"from_name"`
	UseTLS   bool   `mapstructure:"use_tls"`
}

// StorageConfig holds file storage configuration.
type StorageConfig struct {
	Provider string `mapstructure:"provider"` // local, s3, r2, minio
	BasePath string `mapstructure:"base_path"`
	BaseURL  string `mapstructure:"base_url"`
}

// FeatureFlagsConfig holds feature flag configuration.
type FeatureFlagsConfig struct {
	Enabled bool            `mapstructure:"enabled"`
	Flags   map[string]bool `mapstructure:"flags"`
}

// WorkersConfig holds worker pool configuration.
type WorkersConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	NumWorkers int  `mapstructure:"num_workers"`
	QueueSize  int  `mapstructure:"queue_size"`
}

// SchedulerConfig holds scheduler configuration.
type SchedulerConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// MonitoringConfig holds monitoring configuration.
type MonitoringConfig struct {
	PrometheusEnabled bool   `mapstructure:"prometheus_enabled"`
	GrafanaEndpoint   string `mapstructure:"grafana_endpoint"`
	LokiEndpoint      string `mapstructure:"loki_endpoint"`
}

// Address returns the server address string.
func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// RedisAddr returns the Redis address string.
func (r RedisConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// IsConfigured returns true if Redis has been configured with a host or URL.
func (r RedisConfig) IsConfigured() bool {
	return r.Host != "" || r.URL != ""
}

func (r *RedisConfig) ApplyURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return nil
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid redis URL: %w", err)
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "redis" && scheme != "rediss" {
		return fmt.Errorf("invalid redis URL scheme %q", parsedURL.Scheme)
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("redis URL must include a host")
	}

	host := parsedURL.Hostname()
	port := 6379
	if parsedURL.Port() != "" {
		parsedPort, err := strconv.Atoi(parsedURL.Port())
		if err != nil || parsedPort < 1 || parsedPort > 65535 {
			return fmt.Errorf("invalid redis port %q", parsedURL.Port())
		}
		port = parsedPort
	}

	password := ""
	if parsedURL.User != nil {
		if pass, ok := parsedURL.User.Password(); ok {
			password = pass
		}
	}

	db := 0
	if parsedURL.Path != "" && parsedURL.Path != "/" {
		path := strings.TrimPrefix(parsedURL.Path, "/")
		if path != "" {
			parsedDB, err := strconv.Atoi(path)
			if err != nil || parsedDB < 0 || parsedDB > 15 {
				return fmt.Errorf("invalid redis DB index %q", path)
			}
			db = parsedDB
		}
	}

	r.URL = rawURL
	r.Host = host
	r.Port = port
	r.Password = password
	r.DB = db
	r.TLSEnabled = scheme == "rediss"
	return nil
}

func (r RedisConfig) TLSConfig() *tls.Config {
	if !r.TLSEnabled {
		return nil
	}
	return &tls.Config{MinVersion: tls.VersionTLS12, ServerName: r.Host}
}

// IsProduction returns true if the environment is production.
func (c AppConfig) IsProduction() bool {
	return strings.EqualFold(c.Environment, "production")
}

// Load reads configuration from file and environment variables.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Read configuration file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath("../configs")
	}

	v.AutomaticEnv()
	v.SetEnvPrefix("COINDISTRO")

	// Map environment variables to config keys
	bindEnvKeys(v)

	if err := v.ReadInConfig(); err != nil {
		// Config file not found is acceptable; use defaults + env
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		flexibleDurationDecodeHook(),
		mapstructure.StringToTimeDurationHookFunc(),
	))); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Explicitly parse JWT TTLs from raw config values so env strings like
	// "15", "15m", and "168h" all work (Render-friendly + local .env compatible).
	accessTTL, err := ParseFlexibleDuration(v.Get("auth.access_token_ttl"), time.Minute)
	if err != nil {
		return nil, fmt.Errorf("auth.access_token_ttl: %w", err)
	}
	refreshTTL, err := ParseFlexibleDuration(v.Get("auth.refresh_token_ttl"), time.Minute)
	if err != nil {
		return nil, fmt.Errorf("auth.refresh_token_ttl: %w", err)
	}
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL <= 0 {
		refreshTTL = 10080 * time.Minute // 7 days
	}
	cfg.Auth.AccessTokenTTL = accessTTL
	cfg.Auth.RefreshTokenTTL = refreshTTL

	// Support Render's PORT environment variable
	if port := os.Getenv("PORT"); port != "" {
		var portInt int
		if _, err := fmt.Sscanf(port, "%d", &portInt); err == nil && portInt > 0 {
			cfg.Server.Port = portInt
		}
	}

	if redisURL := strings.TrimSpace(v.GetString("redis.url")); redisURL != "" {
		if err := cfg.Redis.ApplyURL(redisURL); err != nil {
			return nil, fmt.Errorf("redis URL: %w", err)
		}
	}

	// Production overrides
	if cfg.App.IsProduction() {
		// Production must not use localhost defaults for database
		if !cfg.Database.IsConfigured() {
			return nil, fmt.Errorf("database host is required in production (set COINDISTRO_DB_HOST)")
		}
		// Production must not use localhost defaults for Redis
		if !cfg.Redis.IsConfigured() {
			return nil, fmt.Errorf("redis host is required in production (set COINDISTRO_REDIS_HOST or COINDISTRO_REDIS_URL)")
		}
		if cfg.Redis.Password == "" {
			return nil, fmt.Errorf("redis password is required in production (set COINDISTRO_REDIS_PASSWORD or include it in COINDISTRO_REDIS_URL)")
		}
		// Production must have valid JWT secrets
		if !cfg.Auth.HasValidSecrets() {
			return nil, fmt.Errorf("valid JWT secrets are required in production (set COINDISTRO_JWT_ACCESS_SECRET and COINDISTRO_JWT_REFRESH_SECRET)")
		}
	}

	// Convert seconds to duration for server timeouts (YAML/env integers = seconds).
	// Only scale when the value looks like a bare unit count (not already a real duration).
	cfg.Server.ReadTimeout = scaleIfUnitCount(cfg.Server.ReadTimeout, time.Second)
	cfg.Server.WriteTimeout = scaleIfUnitCount(cfg.Server.WriteTimeout, time.Second)
	cfg.Server.IdleTimeout = scaleIfUnitCount(cfg.Server.IdleTimeout, time.Second)

	return &cfg, nil
}

// scaleIfUnitCount multiplies d by unit when d looks like a bare integer count
// (from YAML/env numbers), otherwise returns d unchanged.
func scaleIfUnitCount(d, unit time.Duration) time.Duration {
	if isLikelyUnitCount(d) {
		return time.Duration(int64(d)) * unit
	}
	return d
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "coindistro")
	v.SetDefault("app.version", "v0.4.0-alpha")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", true)

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 10)
	v.SetDefault("server.write_timeout", 10)
	v.SetDefault("server.idle_timeout", 60)
	v.SetDefault("server.max_request_size", 10)

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "coindistro")
	v.SetDefault("database.password", "coindistro")
	v.SetDefault("database.dbname", "coindistro")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 5)
	v.SetDefault("database.health_check_interval", 30)

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.dial_timeout", 5)
	v.SetDefault("redis.read_timeout", 3)
	v.SetDefault("redis.write_timeout", 3)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)

	v.SetDefault("auth.access_token_secret", "change-me-in-production")
	v.SetDefault("auth.refresh_token_secret", "change-me-in-production-too")
	v.SetDefault("auth.access_token_ttl", 15)
	v.SetDefault("auth.refresh_token_ttl", 10080)
	v.SetDefault("auth.issuer", "coindistro")

	v.SetDefault("rate_limiter.enabled", true)
	v.SetDefault("rate_limiter.requests_per_minute", 60)
	v.SetDefault("rate_limiter.burst", 100)

	v.SetDefault("cors.allowed_origins", []string{"http://localhost:3000"})
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"})
	v.SetDefault("cors.allow_credentials", true)
	v.SetDefault("cors.max_age", 86400)

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.encoding", "json")
	v.SetDefault("logging.output_paths", []string{"stdout"})
	v.SetDefault("logging.error_output_paths", []string{"stderr"})

	// Telemetry defaults
	v.SetDefault("telemetry.enabled", false)
	v.SetDefault("telemetry.service_name", "coindistro")
	v.SetDefault("telemetry.endpoint", "localhost:4318")
	v.SetDefault("telemetry.sample_rate", 0.1)

	// Email defaults
	v.SetDefault("email.provider", "noop")
	v.SetDefault("email.smtp.host", "localhost")
	v.SetDefault("email.smtp.port", 587)
	v.SetDefault("email.smtp.username", "")
	v.SetDefault("email.smtp.password", "")
	v.SetDefault("email.smtp.from", "noreply@coindistro.com")
	v.SetDefault("email.smtp.from_name", "Coindistro")
	v.SetDefault("email.smtp.use_tls", false)

	// Storage defaults
	v.SetDefault("storage.provider", "local")
	v.SetDefault("storage.base_path", "./uploads")
	v.SetDefault("storage.base_url", "")

	// Feature flags defaults
	v.SetDefault("feature_flags.enabled", true)
	v.SetDefault("feature_flags.flags", map[string]bool{})

	// Workers defaults
	v.SetDefault("workers.enabled", true)
	v.SetDefault("workers.num_workers", 5)
	v.SetDefault("workers.queue_size", 100)

	// Scheduler defaults
	v.SetDefault("scheduler.enabled", true)

	// Monitoring defaults
	v.SetDefault("monitoring.prometheus_enabled", true)
	v.SetDefault("monitoring.grafana_endpoint", "")
	v.SetDefault("monitoring.loki_endpoint", "")
}

func bindEnvKeys(v *viper.Viper) {
	envKeys := map[string]string{
		"app.environment":               "COINDISTRO_ENV",
		"app.debug":                     "COINDISTRO_DEBUG",
		"server.port":                   "COINDISTRO_PORT",
		"database.host":                 "COINDISTRO_DB_HOST",
		"database.port":                 "COINDISTRO_DB_PORT",
		"database.user":                 "COINDISTRO_DB_USER",
		"database.password":             "COINDISTRO_DB_PASSWORD",
		"database.dbname":               "COINDISTRO_DB_NAME",
		"database.ssl_mode":             "COINDISTRO_DB_SSLMODE",
		"redis.host":                    "COINDISTRO_REDIS_HOST",
		"redis.port":                    "COINDISTRO_REDIS_PORT",
		"redis.password":                "COINDISTRO_REDIS_PASSWORD",
		"redis.url":                     "COINDISTRO_REDIS_URL",
		"auth.access_token_secret":      "COINDISTRO_JWT_ACCESS_SECRET",
		"auth.refresh_token_secret":     "COINDISTRO_JWT_REFRESH_SECRET",
		"auth.access_token_ttl":         "COINDISTRO_JWT_ACCESS_TTL",
		"auth.refresh_token_ttl":        "COINDISTRO_JWT_REFRESH_TTL",
		"logging.level":                 "COINDISTRO_LOG_LEVEL",
		"telemetry.enabled":             "COINDISTRO_TELEMETRY_ENABLED",
		"telemetry.endpoint":            "COINDISTRO_TELEMETRY_ENDPOINT",
		"email.provider":                "COINDISTRO_EMAIL_PROVIDER",
		"email.smtp.host":               "COINDISTRO_SMTP_HOST",
		"email.smtp.port":               "COINDISTRO_SMTP_PORT",
		"email.smtp.username":           "COINDISTRO_SMTP_USERNAME",
		"email.smtp.password":           "COINDISTRO_SMTP_PASSWORD",
		"email.smtp.from":               "COINDISTRO_SMTP_FROM",
		"storage.provider":              "COINDISTRO_STORAGE_PROVIDER",
		"storage.base_path":             "COINDISTRO_STORAGE_PATH",
		"monitoring.prometheus_enabled": "COINDISTRO_METRICS_ENABLED",
	}

	for key, env := range envKeys {
		if err := v.BindEnv(key, env); err != nil {
			_ = err
		}
	}

	// CORS allowed origins from comma-separated env var
	_ = v.BindEnv("cors.allowed_origins", "COINDISTRO_ALLOWED_ORIGINS")
}

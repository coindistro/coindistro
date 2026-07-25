package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/coindistro/backend/internal/auth"
	"github.com/coindistro/backend/internal/config"
	"github.com/coindistro/backend/internal/database"
	earnservice "github.com/coindistro/backend/internal/earn/service"
	earnstore "github.com/coindistro/backend/internal/earn/store"
	"github.com/coindistro/backend/internal/email"
	"github.com/coindistro/backend/internal/events"
	"github.com/coindistro/backend/internal/featureflags"
	idservice "github.com/coindistro/backend/internal/identity/service"
	"github.com/coindistro/backend/internal/identity/store"
	"github.com/coindistro/backend/internal/logger"
	"github.com/coindistro/backend/internal/rbac"
)

// Runtime holds shared infrastructure for bootstrap and seed CLI tools.
type Runtime struct {
	Config    *config.Config
	Logger    *logger.Logger
	DB        *database.Database
	Auth      *auth.Auth
	Identity  *idservice.Service
	Earn      *earnservice.Service
	EarnStore *earnstore.Store
}

// Close releases resources.
func (r *Runtime) Close() {
	if r.DB != nil {
		r.DB.Close()
	}
	if r.Logger != nil {
		_ = r.Logger.Sync()
	}
}

// NewRuntime loads config, connects to the database, and wires Identity + Earn services.
func NewRuntime(configPath string) (*Runtime, error) {
	if configPath == "" {
		configPath = resolveConfigPath()
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log, err := logger.New(
		cfg.Logging.Level,
		cfg.Logging.Encoding,
		cfg.Logging.OutputPaths,
		cfg.Logging.ErrorOutputPaths,
	)
	if err != nil {
		return nil, fmt.Errorf("logger: %w", err)
	}

	if err := EnsureDevelopmentEnv(cfg); err != nil {
		return nil, err
	}

	db, err := database.New(cfg.Database, log.Logger)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	authService := auth.New(cfg.Auth, log.Logger)
	rbacService := rbac.New()
	eventBus := events.NewInMemoryBus(log.Logger)
	ff := featureflags.New(log.Logger, cfg.App.Environment)
	// Ensure registration/identity flags are usable for seed-related paths.
	ff.LoadFromConfig(context.Background(), map[string]bool{
		featureflags.FlagRegistration:      true,
		featureflags.FlagRequiresReferral:  false,
		featureflags.FlagInviteOnly:        false,
		featureflags.FlagEmailVerification: false,
		featureflags.FlagAutoVerify:        true,
		featureflags.FlagGenesis:           true,
		featureflags.FlagEarn:              true,
		featureflags.FlagEarnFlexible:      true,
		featureflags.FlagEarnFixed:         true,
		featureflags.FlagEarnStablecoin:    true,
		featureflags.FlagEarnAI:            true,
		featureflags.FlagEarnSignalVault:   true,
	})

	emailSender := email.NewNoopSender(log.Logger)
	identityStore := store.New(db.Pool)
	identityCfg := idservice.DefaultConfig()
	identitySvc := idservice.New(
		identityStore,
		authService,
		rbacService,
		eventBus,
		nil,
		nil,
		emailSender,
		ff,
		nil,
		nil,
		log.Logger,
		identityCfg,
	)

	eStore := earnstore.New(db.Pool)
	earnSvc := earnservice.New(
		eStore,
		eventBus,
		nil,
		nil,
		ff,
		nil,
		nil,
		log.Logger,
	)

	return &Runtime{
		Config:    cfg,
		Logger:    log,
		DB:        db,
		Auth:      authService,
		Identity:  identitySvc,
		Earn:      earnSvc,
		EarnStore: eStore,
	}, nil
}

func resolveConfigPath() string {
	if p := os.Getenv("COINDISTRO_CONFIG"); p != "" {
		return p
	}
	candidates := []string{
		"./configs/config.yaml",
		"../configs/config.yaml",
		"configs/config.yaml",
		"backend/configs/config.yaml",
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return ""
}

// Must is a small helper for CLI tools.
func Must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// OK prints a green-style check line for CLI progress.
func OK(msg string) {
	fmt.Printf("✓ %s\n", msg)
}

// Info prints an informational line.
func Info(msg string) {
	fmt.Println(msg)
}

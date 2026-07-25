// Package bootstrap provides first-time platform initialization for DEVELOPMENT only.
// It creates the Genesis Super Admin through the Identity Service and never runs in production.
package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/config"
	"github.com/coindistro/backend/internal/database"
	"github.com/coindistro/backend/internal/identity/models"
	idservice "github.com/coindistro/backend/internal/identity/service"
)

// SuperAdminCredentials are the default Genesis Super Admin credentials (development only).
const (
	SuperAdminEmail    = "admin@coindistro.com"
	SuperAdminPassword = "Admin@123456"
	SuperAdminUsername = "emmanuel"
	SuperAdminName     = "Emmanuel Ekanem"
)

// Result is the outcome of a bootstrap run.
type Result struct {
	AlreadyCompleted bool
	User             *models.User
	Message          string
}

// Dependencies required to bootstrap the platform.
type Dependencies struct {
	Config   *config.Config
	DB       *database.Database
	Identity *idservice.Service
	Logger   *zap.Logger
}

// EnsureDevelopmentEnv rejects production environments.
// Accepts APP_ENV / ENV / COINDISTRO_ENV / config app.environment = development|dev|local.
func EnsureDevelopmentEnv(cfg *config.Config) error {
	env := strings.ToLower(strings.TrimSpace(cfg.App.Environment))
	if env == "" {
		env = "development"
	}

	// Explicit overrides from common env var names (not always bound into config).
	for _, key := range []string{"APP_ENV", "ENV", "COINDISTRO_ENV"} {
		if v := strings.ToLower(strings.TrimSpace(os.Getenv(key))); v != "" {
			env = v
			break
		}
	}

	switch env {
	case "development", "dev", "local", "test":
		return nil
	default:
		return fmt.Errorf("Bootstrap is disabled in production")
	}
}

// Run performs first platform initialization: Genesis Super Admin only.
// If a super_admin already exists, returns success without creating duplicates.
func Run(ctx context.Context, deps Dependencies) (*Result, error) {
	if err := EnsureDevelopmentEnv(deps.Config); err != nil {
		return nil, err
	}
	if deps.Identity == nil {
		return nil, fmt.Errorf("identity service is required")
	}
	if deps.DB == nil || deps.DB.Pool == nil {
		return nil, fmt.Errorf("database is required")
	}

	exists, err := deps.Identity.HasSuperAdmin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check super admin: %w", err)
	}
	if exists {
		msg := "Bootstrap already completed."
		if deps.Logger != nil {
			deps.Logger.Info(msg)
		}
		return &Result{AlreadyCompleted: true, Message: msg}, nil
	}

	// Also treat existing admin@coindistro.com as completed bootstrap.
	if existing, _ := deps.Identity.GetProfileByEmail(ctx, SuperAdminEmail); existing != nil {
		msg := "Bootstrap already completed."
		return &Result{AlreadyCompleted: true, Message: msg, User: existing}, nil
	}

	genesisN := 1
	user, err := deps.Identity.BootstrapCreateUser(ctx, idservice.BootstrapUserInput{
		Email:             SuperAdminEmail,
		Username:          SuperAdminUsername,
		DisplayName:       SuperAdminName,
		Password:          SuperAdminPassword,
		Roles:             []string{"super_admin", "user"},
		Country:           "NGA",
		Timezone:          "Africa/Lagos",
		InvitationCredits: 1000,
		IsGenesis:         true,
		IsFounder:         true,
		EmailVerified:     true,
		GenesisNumber:     &genesisN,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create super admin: %w", err)
	}

	msg := "Bootstrap completed. Genesis Super Admin created."
	if deps.Logger != nil {
		deps.Logger.Info(msg, zap.String("email", user.Email), zap.String("id", user.ID))
	}
	return &Result{AlreadyCompleted: false, User: user, Message: msg}, nil
}

// RunMigrations applies SQL files from migrationsDir in lexical order.
// Safe to re-run when migrations use IF NOT EXISTS.
func RunMigrations(ctx context.Context, db *database.Database, migrationsDir string) error {
	if db == nil || db.Pool == nil {
		return fmt.Errorf("database is required")
	}
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".sql") {
			files = append(files, filepath.Join(migrationsDir, name))
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return fmt.Errorf("no migration files found in %s", migrationsDir)
	}

	for _, path := range files {
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", path, err)
		}
		sql := string(sqlBytes)
		if strings.TrimSpace(sql) == "" {
			continue
		}
		if _, err := db.Pool.Exec(ctx, sql); err != nil {
			// Allow partially re-applied migrations (e.g. duplicate index names).
			if isIgnorableMigrationError(err) {
				continue
			}
			return fmt.Errorf("apply migration %s: %w", filepath.Base(path), err)
		}
	}
	return nil
}

func isIgnorableMigrationError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "duplicate key")
}

// ResolveMigrationsDir finds the migrations folder relative to common working directories.
func ResolveMigrationsDir() string {
	candidates := []string{
		"migrations",
		"./migrations",
		"../migrations",
		"backend/migrations",
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return "migrations"
}

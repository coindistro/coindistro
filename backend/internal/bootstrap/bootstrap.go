// Package bootstrap provides first-time platform initialization for DEVELOPMENT only.
// It creates the Genesis Super Admin through the Identity Service and never runs in production.
package bootstrap

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/coindistro/backend/internal/config"
	"github.com/coindistro/backend/internal/database"
	"github.com/coindistro/backend/internal/identity/models"
	idservice "github.com/coindistro/backend/internal/identity/service"
)

// SuperAdminCredentials are the default Genesis Super Admin credentials.
// Production can override password via COINDISTRO_SUPER_ADMIN_PASSWORD.
const (
	SuperAdminEmail        = "admin@coindistro.com"
	SuperAdminPassword     = "Admin@123456" // legacy dev default (seed CLI)
	SuperAdminPasswordProd = "Admin123!"    // default production bootstrap password
	SuperAdminUsername     = "emmanuel"
	SuperAdminName         = "Emmanuel Ekanem"
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

// Run performs development-only CLI bootstrap (blocked outside development).
func Run(ctx context.Context, deps Dependencies) (*Result, error) {
	if err := EnsureDevelopmentEnv(deps.Config); err != nil {
		return nil, err
	}
	return EnsureSuperAdmin(ctx, deps, SuperAdminPassword)
}

// EnsureSuperAdmin creates the platform super admin if one does not already exist.
// Safe to call in every environment (including production) after migrations.
// Idempotent: never creates duplicates.
//
// Password resolution order:
//  1. COINDISTRO_SUPER_ADMIN_PASSWORD env
//  2. passwordOverride argument (if non-empty)
//  3. SuperAdminPasswordProd ("Admin123!")
func EnsureSuperAdmin(ctx context.Context, deps Dependencies, passwordOverride string) (*Result, error) {
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

	// Treat existing admin@coindistro.com as completed bootstrap.
	if existing, _ := deps.Identity.GetProfileByEmail(ctx, SuperAdminEmail); existing != nil {
		msg := "Bootstrap already completed."
		if deps.Logger != nil {
			deps.Logger.Info(msg, zap.String("email", SuperAdminEmail))
		}
		return &Result{AlreadyCompleted: true, Message: msg, User: existing}, nil
	}

	password := strings.TrimSpace(os.Getenv("COINDISTRO_SUPER_ADMIN_PASSWORD"))
	if password == "" {
		password = strings.TrimSpace(passwordOverride)
	}
	if password == "" {
		password = SuperAdminPasswordProd
	}

	genesisN := 1
	user, err := deps.Identity.BootstrapCreateUser(ctx, idservice.BootstrapUserInput{
		Email:             SuperAdminEmail,
		Username:          SuperAdminUsername,
		DisplayName:       SuperAdminName,
		Password:          password,
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
// It records applied versions in schema_migrations and is safe to re-run.
func RunMigrations(ctx context.Context, db *database.Database, migrationsDir string, logger *zap.Logger) error {
	if db == nil || db.Pool == nil {
		return fmt.Errorf("database is required")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	logger.Info("starting database migrations", zap.String("migrations_dir", migrationsDir))
	if err := runMigrations(ctx, db.Pool, migrationsDir, logger); err != nil {
		return err
	}
	logger.Info("database schema ready")
	return nil
}

type migrationTx interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type migrationClient interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// migrationRenames maps previously deployed filenames to the current unique sequence.
// Existing production DBs that recorded the old names must not re-run SQL.
var migrationRenames = map[string]string{
	"003_genesis_investor_program.sql":     "004_genesis_investor_program.sql",
	"004_identity_platform_links.sql":      "005_identity_platform_links.sql",
	"004_investor_earnings_dashboard.sql":  "006_investor_earnings_dashboard.sql",
	"005_sessions_token_columns.sql":       "007_sessions_token_columns.sql",
}

func runMigrations(ctx context.Context, db migrationClient, migrationsDir string, logger *zap.Logger) error {
	if err := ensureSchemaMigrationsTable(ctx, db); err != nil {
		return err
	}
	files, err := findMigrationFiles(migrationsDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no migration files found in %s", migrationsDir)
	}
	if err := validateUniqueMigrationPrefixes(files); err != nil {
		return err
	}
	applied, err := appliedMigrations(ctx, db)
	if err != nil {
		return fmt.Errorf("read applied migrations: %w", err)
	}
	if err := reconcileRenamedMigrations(ctx, db, applied, logger); err != nil {
		return err
	}
	// Reload after alias inserts so pending detection sees the new names.
	applied, err = appliedMigrations(ctx, db)
	if err != nil {
		return fmt.Errorf("read applied migrations: %w", err)
	}
	pending := make([]string, 0, len(files))
	for _, path := range files {
		if _, ok := applied[filepath.Base(path)]; !ok {
			pending = append(pending, path)
		}
	}
	logger.Info("found pending migrations", zap.Int("count", len(pending)))
	for _, path := range pending {
		name := filepath.Base(path)
		logger.Info("applying migration", zap.String("migration", name))
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", path, err)
		}
		sql := string(sqlBytes)
		if strings.TrimSpace(sql) == "" {
			logger.Info("skipping empty migration", zap.String("migration", name))
			continue
		}
		checksum := checksum(sqlBytes)
		tx, err := db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration transaction: %w", err)
		}
		if _, err := tx.Exec(ctx, sql); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, checksum, applied_at) VALUES ($1, $2, NOW())`,
			name,
			checksum,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
		logger.Info("migration successful", zap.String("migration", name))
	}
	return nil
}

// reconcileRenamedMigrations records the new filename when an older alias was already applied,
// preventing partial re-application after renumbering duplicate prefixes.
func reconcileRenamedMigrations(ctx context.Context, db migrationClient, applied map[string]string, logger *zap.Logger) error {
	for oldName, newName := range migrationRenames {
		oldChecksum, hasOld := applied[oldName]
		_, hasNew := applied[newName]
		if !hasOld || hasNew {
			continue
		}
		if _, err := db.Exec(ctx,
			`INSERT INTO schema_migrations (version, checksum, applied_at) VALUES ($1, $2, NOW()) ON CONFLICT (version) DO NOTHING`,
			newName,
			oldChecksum,
		); err != nil {
			return fmt.Errorf("record renamed migration %s -> %s: %w", oldName, newName, err)
		}
		logger.Info("recorded renamed migration without re-applying",
			zap.String("from", oldName),
			zap.String("to", newName),
		)
		applied[newName] = oldChecksum
	}
	return nil
}

func validateUniqueMigrationPrefixes(files []string) error {
	seen := make(map[string]string, len(files))
	for _, path := range files {
		name := filepath.Base(path)
		prefix := migrationPrefix(name)
		if prefix == "" {
			return fmt.Errorf("migration %s missing numeric prefix", name)
		}
		if existing, ok := seen[prefix]; ok {
			return fmt.Errorf("duplicate migration prefix %s: %s and %s", prefix, existing, name)
		}
		seen[prefix] = name
	}
	return nil
}

func migrationPrefix(filename string) string {
	base := filepath.Base(filename)
	parts := strings.SplitN(base, "_", 2)
	if len(parts) < 2 {
		return ""
	}
	for _, ch := range parts[0] {
		if ch < '0' || ch > '9' {
			return ""
		}
	}
	return parts[0]
}

func findMigrationFiles(migrationsDir string) ([]string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
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
	return files, nil
}

func AppliedMigrations(ctx context.Context, db *database.Database) (map[string]string, error) {
	if db == nil || db.Pool == nil {
		return nil, fmt.Errorf("database is required")
	}
	return appliedMigrations(ctx, db.Pool)
}

func appliedMigrations(ctx context.Context, db migrationClient) (map[string]string, error) {
	rows, err := db.Query(ctx, `SELECT version, checksum FROM schema_migrations ORDER BY version ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	migrations := make(map[string]string)
	for rows.Next() {
		var version, checksum string
		if err := rows.Scan(&version, &checksum); err != nil {
			return nil, err
		}
		migrations[version] = checksum
	}
	return migrations, rows.Err()
}

func ensureSchemaMigrationsTable(ctx context.Context, db migrationClient) error {
	_, err := db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    checksum TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`)
	return err
}

func checksum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ResolveMigrationsDir finds the migrations folder relative to common working directories.
func ResolveMigrationsDir() string {
	if env := strings.TrimSpace(os.Getenv("COINDISTRO_MIGRATIONS_DIR")); env != "" {
		return env
	}
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

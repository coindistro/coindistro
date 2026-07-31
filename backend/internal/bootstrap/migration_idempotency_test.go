package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenesisMigrationIsIdempotent(t *testing.T) {
	migrationPath := filepath.Join("..", "..", "migrations", "004_genesis_investor_program.sql")
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration file: %v", err)
	}

	sql := string(content)

	for _, expected := range []string{
		"CREATE TABLE IF NOT EXISTS",
		"CREATE INDEX IF NOT EXISTS",
		"CREATE UNIQUE INDEX IF NOT EXISTS",
		"ON CONFLICT (id) DO NOTHING",
		"ON CONFLICT (name) DO NOTHING",
		"DROP TRIGGER IF EXISTS",
	} {
		if !strings.Contains(sql, expected) {
			t.Fatalf("expected migration to contain %q", expected)
		}
	}

	for _, forbidden := range []string{
		"CREATE INDEX idx_investment_plans_enabled",
		"CREATE INDEX idx_wallets_user_id",
		"CREATE UNIQUE INDEX idx_investments_payment_ref_provider",
	} {
		if strings.Contains(sql, forbidden) {
			t.Fatalf("expected migration to avoid non-idempotent statement %q", forbidden)
		}
	}
}

func TestMigrationPrefixesAreUnique(t *testing.T) {
	dir := filepath.Join("..", "..", "migrations")
	files, err := findMigrationFiles(dir)
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}
	if err := validateUniqueMigrationPrefixes(files); err != nil {
		t.Fatalf("unique prefixes: %v", err)
	}
	if len(files) < 7 {
		t.Fatalf("expected at least 7 migrations, got %d", len(files))
	}
}

func TestMigrationPrefixParser(t *testing.T) {
	if got := migrationPrefix("004_genesis_investor_program.sql"); got != "004" {
		t.Fatalf("prefix = %q, want 004", got)
	}
	if got := migrationPrefix("bad.sql"); got != "" {
		t.Fatalf("prefix = %q, want empty", got)
	}
}

func TestValidateUniqueMigrationPrefixesDetectsDuplicates(t *testing.T) {
	err := validateUniqueMigrationPrefixes([]string{
		"003_earn_service.sql",
		"003_genesis_investor_program.sql",
	})
	if err == nil {
		t.Fatal("expected duplicate prefix error")
	}
}

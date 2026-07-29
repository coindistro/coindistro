package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenesisMigrationIsIdempotent(t *testing.T) {
	migrationPath := filepath.Join("..", "..", "migrations", "003_genesis_investor_program.sql")
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

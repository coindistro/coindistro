package service_test

import (
	"testing"

	"github.com/coindistro/backend/internal/earn/models"
	"github.com/coindistro/backend/internal/featureflags"
	"github.com/coindistro/backend/internal/rbac"
	"go.uber.org/zap"
)

func TestUserHasEarnPermissions(t *testing.T) {
	r := rbac.New()
	if !r.HasPermission(rbac.RoleUser, rbac.PermEarnRead) {
		t.Fatal("user should have earn.read")
	}
	if !r.HasPermission(rbac.RoleUser, rbac.PermEarnJoin) {
		t.Fatal("user should have earn.join")
	}
	if !r.HasPermission(rbac.RoleAdmin, rbac.PermEarnAdmin) {
		t.Fatal("admin should have earn.admin")
	}
	if r.HasPermission(rbac.RoleStudent, rbac.PermEarnAdmin) {
		t.Fatal("student should not have earn.admin")
	}
}

func TestEarnFeatureFlagsRegistered(t *testing.T) {
	logger := zap.NewNop()
	m := featureflags.New(logger, "test")
	flags := []string{
		featureflags.FlagEarn,
		featureflags.FlagEarnFlexible,
		featureflags.FlagEarnFixed,
		featureflags.FlagEarnStablecoin,
		featureflags.FlagEarnAI,
		featureflags.FlagEarnSignalVault,
		featureflags.FlagEarnLaunchpool,
		featureflags.FlagEarnLearn,
		featureflags.FlagEarnReferral,
	}
	for _, f := range flags {
		if !m.IsEnabled(f) {
			t.Fatalf("expected flag %s enabled by default", f)
		}
	}
}

func TestProductCategories(t *testing.T) {
	cats := []string{
		models.CategoryFlexible,
		models.CategoryFixed,
		models.CategoryStablecoin,
		models.CategoryAISmart,
		models.CategorySignalVault,
		models.CategoryLaunchpool,
		models.CategoryLearnEarn,
		models.CategoryReferral,
	}
	if len(cats) != 8 {
		t.Fatalf("expected 8 categories")
	}
}

func TestFixedDurations(t *testing.T) {
	if len(models.FixedDurations) != 5 {
		t.Fatalf("expected 5 fixed durations")
	}
}

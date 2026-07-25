// Package seed populates development databases with realistic demo data.
// It must only run in development environments.
package seed

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/bootstrap"
	"github.com/coindistro/backend/internal/config"
	"github.com/coindistro/backend/internal/database"
	earnmodels "github.com/coindistro/backend/internal/earn/models"
	earnservice "github.com/coindistro/backend/internal/earn/service"
	earnstore "github.com/coindistro/backend/internal/earn/store"
	"github.com/coindistro/backend/internal/identity/models"
	idservice "github.com/coindistro/backend/internal/identity/service"
	uuidlib "github.com/coindistro/backend/internal/uuid"
)

// Credentials printed after a successful seed.
const (
	AdminPassword = "Admin@123456"
	UserPassword  = "User@123456"
)

// Dependencies for the seeder.
type Dependencies struct {
	Config    *config.Config
	DB        *database.Database
	Identity  *idservice.Service
	Earn      *earnservice.Service
	EarnStore *earnstore.Store
	Logger    *zap.Logger
}

// Progress is called with human-readable status lines.
type Progress func(step string)

// Result summarizes what was seeded.
type Result struct {
	SuperAdmin     *models.User
	Admins         []*models.User
	Moderator      *models.User
	Users          []*models.User
	Products       []*earnmodels.Product
	Participations int
}

// Run executes the full development seed pipeline.
func Run(ctx context.Context, deps Dependencies, progress Progress) (*Result, error) {
	if err := bootstrap.EnsureDevelopmentEnv(deps.Config); err != nil {
		return nil, err
	}
	if progress == nil {
		progress = func(string) {}
	}
	log := deps.Logger
	if log == nil {
		log = zap.NewNop()
	}

	progress("Checking database...")
	if deps.DB == nil || deps.DB.Pool == nil {
		return nil, fmt.Errorf("database is required")
	}
	if err := deps.DB.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	progress("Running migrations...")
	migDir := bootstrap.ResolveMigrationsDir()
	if err := bootstrap.RunMigrations(ctx, deps.DB, migDir); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	progress("Creating Super Admin...")
	boot, err := bootstrap.Run(ctx, bootstrap.Dependencies{
		Config: deps.Config, DB: deps.DB, Identity: deps.Identity, Logger: log,
	})
	if err != nil {
		return nil, err
	}

	superAdmin := boot.User
	if superAdmin == nil {
		superAdmin, err = deps.Identity.GetProfileByEmail(ctx, bootstrap.SuperAdminEmail)
		if err != nil || superAdmin == nil {
			return nil, fmt.Errorf("super admin missing after bootstrap")
		}
	}

	result := &Result{SuperAdmin: superAdmin}

	// Idempotency: if demo user1 already exists, skip heavy re-seed of users.
	if existing, _ := deps.Identity.GetProfileByEmail(ctx, "user1@coindistro.com"); existing != nil {
		progress("Demo users already present — refreshing earn/activity if needed...")
		result.Users = []*models.User{existing}
		if err := seedEarnIfEmpty(ctx, deps, result, progress); err != nil {
			return nil, err
		}
		progress("Seeder completed successfully.")
		return result, nil
	}

	progress("Creating Admins...")
	admins, err := seedAdmins(ctx, deps, superAdmin)
	if err != nil {
		return nil, err
	}
	result.Admins = admins

	progress("Creating Moderator...")
	mod, err := seedModerator(ctx, deps, superAdmin)
	if err != nil {
		return nil, err
	}
	result.Moderator = mod

	progress("Creating Users...")
	users, err := seedStandardUsers(ctx, deps, superAdmin, admins)
	if err != nil {
		return nil, err
	}
	result.Users = users

	allUsers := append([]*models.User{}, superAdmin)
	allUsers = append(allUsers, admins...)
	if mod != nil {
		allUsers = append(allUsers, mod)
	}
	allUsers = append(allUsers, users...)

	progress("Creating Sessions...")
	if err := seedSessionsAndDevices(ctx, deps, allUsers); err != nil {
		return nil, err
	}

	progress("Creating Activity...")
	if err := seedActivityAndNotifications(ctx, deps, allUsers, superAdmin); err != nil {
		return nil, err
	}

	progress("Creating Invitations...")
	if err := seedInvitations(ctx, deps, superAdmin, users); err != nil {
		log.Warn("invitations seed partial", zap.Error(err))
	}

	progress("Creating Earn Products...")
	products, err := seedEarnProducts(ctx, deps, superAdmin.ID)
	if err != nil {
		return nil, err
	}
	result.Products = products

	progress("Creating Demo Portfolio...")
	n, err := seedEarnParticipations(ctx, deps, products, allUsers)
	if err != nil {
		return nil, err
	}
	result.Participations = n

	progress("Creating Notifications...")
	// Notifications are represented as activity_log events (see seedActivityAndNotifications).
	_ = seedSystemAnnouncements(ctx, deps, allUsers)

	progress("Seeder completed successfully.")
	return result, nil
}

func seedEarnIfEmpty(ctx context.Context, deps Dependencies, result *Result, progress Progress) error {
	if deps.EarnStore == nil {
		return nil
	}
	list, total, err := deps.EarnStore.ListProducts(ctx, earnmodels.ProductListFilter{Status: earnmodels.StatusActive, Page: 1, PerPage: 5})
	if err != nil {
		return err
	}
	if total > 0 && len(list) > 0 {
		result.Products = list
		return nil
	}
	progress("Creating Earn Products...")
	products, err := seedEarnProducts(ctx, deps, result.SuperAdmin.ID)
	if err != nil {
		return err
	}
	result.Products = products
	return nil
}

func seedAdmins(ctx context.Context, deps Dependencies, super *models.User) ([]*models.User, error) {
	defs := []struct {
		email, username, name string
		genesisN              int
	}{
		{"admin1@coindistro.com", "admin_ada", "Ada Okonkwo", 2},
		{"admin2@coindistro.com", "admin_chidi", "Chidi Nwosu", 3},
	}
	var out []*models.User
	for _, d := range defs {
		n := d.genesisN
		u, err := deps.Identity.BootstrapCreateUser(ctx, idservice.BootstrapUserInput{
			Email: d.email, Username: d.username, DisplayName: d.name,
			Password: AdminPassword, Roles: []string{"admin", "user"},
			Country: "NGA", Timezone: "Africa/Lagos",
			InvitationCredits: 50, IsGenesis: true, EmailVerified: true,
			GenesisNumber: &n, ReferredBy: &super.ID, ReferralLevel: 1,
		})
		if err != nil {
			return nil, fmt.Errorf("admin %s: %w", d.email, err)
		}
		out = append(out, u)
	}
	return out, nil
}

func seedModerator(ctx context.Context, deps Dependencies, super *models.User) (*models.User, error) {
	return deps.Identity.BootstrapCreateUser(ctx, idservice.BootstrapUserInput{
		Email: "moderator@coindistro.com", Username: "mod_amina", DisplayName: "Amina Yusuf",
		Password: AdminPassword, Roles: []string{"moderator", "user"},
		Country: "NGA", Timezone: "Africa/Lagos",
		InvitationCredits: 20, EmailVerified: true,
		ReferredBy: &super.ID, ReferralLevel: 1,
	})
}

func seedStandardUsers(ctx context.Context, deps Dependencies, super *models.User, admins []*models.User) ([]*models.User, error) {
	people := []struct {
		email, username, name, country string
	}{
		{"user1@coindistro.com", "user_kemi", "Kemi Adebayo", "NGA"},
		{"user2@coindistro.com", "user_tunde", "Tunde Bakare", "NGA"},
		{"user3@coindistro.com", "user_zainab", "Zainab Bello", "NGA"},
		{"user4@coindistro.com", "user_kwame", "Kwame Mensah", "GHA"},
		{"user5@coindistro.com", "user_ama", "Ama Serwaa", "GHA"},
		{"user6@coindistro.com", "user_fatou", "Fatou Diallo", "SEN"},
		{"user7@coindistro.com", "user_ibrahim", "Ibrahim Traoré", "CIV"},
		{"user8@coindistro.com", "user_noura", "Noura Hassan", "EGY"},
		{"user9@coindistro.com", "user_thabo", "Thabo Dlamini", "ZAF"},
		{"user10@coindistro.com", "user_aisha", "Aisha Kamau", "KEN"},
	}

	var out []*models.User
	for i, p := range people {
		// Build a small referral tree: first half under super, rest under admins/users.
		var referrerID *string
		level := 1
		if i < 4 {
			referrerID = &super.ID
		} else if i < 7 && len(admins) > 0 {
			referrerID = &admins[i%len(admins)].ID
			level = 2
		} else if len(out) > 0 {
			prev := out[i-1]
			referrerID = &prev.ID
			level = prev.ReferralLevel + 1
		} else {
			referrerID = &super.ID
		}

		isGenesis := i < 3
		var genN *int
		if isGenesis {
			n := 10 + i
			genN = &n
		}

		u, err := deps.Identity.BootstrapCreateUser(ctx, idservice.BootstrapUserInput{
			Email: p.email, Username: p.username, DisplayName: p.name,
			Password: UserPassword, Roles: []string{"user"},
			Country: p.country, Timezone: "Africa/Lagos",
			InvitationCredits: 5 + i, IsGenesis: isGenesis, EmailVerified: true,
			GenesisNumber: genN, ReferredBy: referrerID, ReferralLevel: level,
		})
		if err != nil {
			return nil, fmt.Errorf("user %s: %w", p.email, err)
		}
		out = append(out, u)
	}
	return out, nil
}

func seedSessionsAndDevices(ctx context.Context, deps Dependencies, users []*models.User) error {
	browsers := []string{"Chrome", "Firefox", "Safari", "Edge"}
	oses := []string{"Windows 11", "macOS", "Android", "iOS", "Ubuntu"}
	devices := []string{"desktop", "mobile", "tablet"}
	ips := []string{"102.89.23.14", "41.58.102.9", "197.210.45.3", "105.112.48.22"}

	for _, u := range users {
		if u == nil {
			continue
		}
		b := browsers[rand.Intn(len(browsers))]
		o := oses[rand.Intn(len(oses))]
		d := devices[rand.Intn(len(devices))]
		ip := ips[rand.Intn(len(ips))]
		ua := fmt.Sprintf("%s on %s", b, o)

		if _, err := deps.Identity.CreateDemoSession(ctx, u.ID, ip, ua, b, o, "Primary Device", d, true); err != nil {
			return err
		}
		// Extra historical session for some users
		if rand.Intn(2) == 0 {
			_, _ = deps.Identity.CreateDemoSession(ctx, u.ID, ips[rand.Intn(len(ips))], ua, b, o, "Secondary Device", "mobile", false)
		}
		if _, err := deps.Identity.CreateDemoDevice(ctx, u.ID, "Primary Device", b, o, d, true, true); err != nil {
			return err
		}
	}
	return nil
}

func seedActivityAndNotifications(ctx context.Context, deps Dependencies, users []*models.User, super *models.User) error {
	actions := []struct {
		action  string
		details map[string]interface{}
	}{
		{"user.registered", map[string]interface{}{"source": "seed"}},
		{"auth.login_success", map[string]interface{}{"method": "password"}},
		{"auth.logout", map[string]interface{}{}},
		{"device.added", map[string]interface{}{"device": "Primary Device"}},
		{"security.password_changed", map[string]interface{}{}},
		{"referral.joined", map[string]interface{}{"via": "referral_code"}},
		{"referral.invite_sent", map[string]interface{}{}},
		{"genesis.granted", map[string]interface{}{}},
		{"earn.joined", map[string]interface{}{"product": "flexible-usdt"}},
		{"earn.reward_claimed", map[string]interface{}{"asset": "USDT"}},
		{"system.announcement", map[string]interface{}{"title": "Welcome to Coindistro"}},
	}

	for i, u := range users {
		if u == nil {
			continue
		}
		// Stagger timestamps into the past
		for j, a := range actions {
			// Not every action for every user
			if j > 0 && rand.Intn(3) == 0 {
				continue
			}
			if a.action == "genesis.granted" && !u.IsGenesis {
				continue
			}
			created := time.Now().UTC().Add(-time.Duration(i*3+j*5) * time.Hour)
			_ = deps.Identity.LogDemoActivityAt(ctx, u.ID, a.action, "102.89.23.14", "Coindistro Seeder", a.details, created)
		}
	}

	// Super admin platform events
	_ = deps.Identity.LogDemoActivityAt(ctx, super.ID, "admin.platform_bootstrapped", "127.0.0.1", "bootstrap",
		map[string]interface{}{"env": "development"}, time.Now().UTC().Add(-72*time.Hour))
	return nil
}

func seedSystemAnnouncements(ctx context.Context, deps Dependencies, users []*models.User) error {
	for _, u := range users {
		if u == nil {
			continue
		}
		_ = deps.Identity.LogDemoActivity(ctx, u.ID, "system.announcement", "127.0.0.1", "system",
			map[string]interface{}{
				"title":   "Welcome to Coindistro Demo",
				"message": "Explore dashboards, referrals, and Earn with preloaded demo data.",
			})
	}
	return nil
}

func seedInvitations(ctx context.Context, deps Dependencies, super *models.User, users []*models.User) error {
	// Create a few pending invitations from super admin using store via SendInvitation if credits allow
	// Direct store path through service SendInvitation
	for i := 0; i < 3 && i < len(users); i++ {
		email := fmt.Sprintf("pending.invite%d@example.com", i+1)
		msg := "Join me on Coindistro!"
		_, err := deps.Identity.SendInvitation(ctx, super.ID, &models.SendInvitationRequest{
			Email:   email,
			Message: &msg,
		})
		if err != nil {
			// Non-fatal if credits or feature flags block
			continue
		}
	}
	return nil
}

func seedEarnProducts(ctx context.Context, deps Dependencies, actorID string) ([]*earnmodels.Product, error) {
	if deps.Earn == nil {
		return nil, fmt.Errorf("earn service is required")
	}
	capTotal := 1_000_000.0
	maxAlloc := 50_000.0
	d30, d90 := 30, 90

	defs := []earnmodels.CreateProductRequest{
		{
			Name: "Flexible USDT Earn", Slug: "flexible-usdt-earn",
			Description: "Flexible USDT savings with daily rewards. Withdraw anytime.",
			Category:    earnmodels.CategoryFlexible, SupportedAssets: []string{"USDT"},
			CapacityTotal: &capTotal, RiskLevel: "low", MinAllocation: 10, MaxAllocation: &maxAlloc,
			RewardModel: "flexible", RewardAPR: 8.5, Featured: true, Status: earnmodels.StatusActive,
		},
		{
			Name: "Fixed BTC Earn", Slug: "fixed-btc-earn",
			Description: "Lock Bitcoin for fixed duration yield.",
			Category:    earnmodels.CategoryFixed, SupportedAssets: []string{"BTC"},
			DurationDays: &d90, CapacityTotal: &capTotal, RiskLevel: "medium", MinAllocation: 0.001,
			RewardModel: "fixed", RewardAPR: 5.2, Featured: true, Status: earnmodels.StatusActive,
		},
		{
			Name: "ETH Staking", Slug: "eth-staking",
			Description: "Stake Ethereum and earn network rewards.",
			Category:    earnmodels.CategoryFixed, SupportedAssets: []string{"ETH"},
			DurationDays: &d30, CapacityTotal: &capTotal, RiskLevel: "medium", MinAllocation: 0.01,
			RewardModel: "fixed", RewardAPR: 4.8, Featured: false, Status: earnmodels.StatusActive,
		},
		{
			Name: "AI Smart Earn", Slug: "ai-smart-earn",
			Description: "AI-optimized multi-asset yield strategies.",
			Category:    earnmodels.CategoryAISmart, SupportedAssets: []string{"USDT", "USDC", "BTC"},
			CapacityTotal: &capTotal, RiskLevel: "medium", MinAllocation: 50, MaxAllocation: &maxAlloc,
			RewardModel: "flexible", RewardAPR: 12.0,
			StrategyProfiles: []string{"conservative", "balanced", "growth", "aggressive"},
			Featured:         true, Status: earnmodels.StatusActive,
		},
		{
			Name: "Signal Vault", Slug: "signal-vault",
			Description: "Vault linked to curated trading signal performance.",
			Category:    earnmodels.CategorySignalVault, SupportedAssets: []string{"USDT"},
			CapacityTotal: &capTotal, RiskLevel: "high", MinAllocation: 25, MaxAllocation: &maxAlloc,
			RewardModel: "promotional", RewardAPR: 15.0, Featured: true, Status: earnmodels.StatusActive,
		},
	}

	var products []*earnmodels.Product
	for _, d := range defs {
		// Skip if slug exists
		if deps.EarnStore != nil {
			if existing, _ := deps.EarnStore.GetProductBySlug(ctx, d.Slug); existing != nil {
				products = append(products, existing)
				continue
			}
		}
		p, err := deps.Earn.CreateProduct(ctx, &d, actorID)
		if err != nil {
			return nil, fmt.Errorf("product %s: %w", d.Slug, err)
		}
		products = append(products, p)
	}
	return products, nil
}

func seedEarnParticipations(ctx context.Context, deps Dependencies, products []*earnmodels.Product, users []*models.User) (int, error) {
	if deps.EarnStore == nil || len(products) == 0 || len(users) == 0 {
		return 0, nil
	}

	target := 100
	count := 0
	rng := rand.New(rand.NewSource(42))

	for count < target {
		u := users[rng.Intn(len(users))]
		if u == nil {
			continue
		}
		p := products[rng.Intn(len(products))]
		asset := "USDT"
		if len(p.SupportedAssets) > 0 {
			asset = p.SupportedAssets[rng.Intn(len(p.SupportedAssets))]
		}

		amount := 50.0 + float64(rng.Intn(5000))
		if asset == "BTC" {
			amount = 0.01 + float64(rng.Intn(100))/1000.0
		} else if asset == "ETH" {
			amount = 0.1 + float64(rng.Intn(50))/10.0
		}

		est := amount * (p.RewardAPR / 100.0) / 12.0
		accrued := est * (0.2 + rng.Float64()*0.6)
		lifetime := accrued * (1.0 + rng.Float64())
		now := time.Now().UTC()
		joined := now.Add(-time.Duration(rng.Intn(60)) * 24 * time.Hour)

		part := &earnmodels.Participation{
			ID:               uuidlib.NewString(),
			UserID:           u.ID,
			ProductID:        p.ID,
			Asset:            asset,
			AllocatedAmount:  amount,
			CurrentBalance:   amount + accrued*0.1,
			EstimatedRewards: est,
			AccruedRewards:   accrued,
			LifetimeRewards:  lifetime,
			Status:           earnmodels.ParticipationActive,
			JoinedAt:         joined,
			CreatedAt:        joined,
			UpdatedAt:        now,
			Metadata:         map[string]interface{}{"source": "seed"},
		}
		if p.Category == earnmodels.CategoryFixed && p.DurationDays != nil {
			start := joined
			end := joined.Add(time.Duration(*p.DurationDays) * 24 * time.Hour)
			part.LockStartAt = &start
			part.LockEndAt = &end
			part.Status = earnmodels.ParticipationLocked
		}
		if p.Category == earnmodels.CategoryAISmart {
			profiles := []string{"conservative", "balanced", "growth", "aggressive"}
			pr := profiles[rng.Intn(len(profiles))]
			part.StrategyProfile = &pr
		}

		if err := deps.EarnStore.CreateParticipation(ctx, part); err != nil {
			// Continue on unique/race issues
			continue
		}
		_ = deps.EarnStore.IncrementCapacity(ctx, p.ID, amount)

		// Reward history
		for r := 0; r < 1+rng.Intn(3); r++ {
			granted := joined.Add(time.Duration(r+1) * 24 * time.Hour)
			amt := accrued / float64(2+r)
			pid := part.ID
			reward := &earnmodels.Reward{
				ID:              uuidlib.NewString(),
				UserID:          u.ID,
				ProductID:       p.ID,
				ParticipationID: &pid,
				Asset:           asset,
				Amount:          amt,
				RewardType:      "daily",
				Status:          "granted",
				Description:     "Seed demo reward",
				GrantedAt:       &granted,
				CreatedAt:       granted,
				UpdatedAt:       granted,
				Metadata:        map[string]interface{}{"source": "seed"},
			}
			_ = deps.EarnStore.CreateReward(ctx, reward)
		}

		// Transaction history
		_ = deps.EarnStore.CreateTransaction(ctx, &earnmodels.Transaction{
			ID: uuidlib.NewString(), UserID: u.ID, ProductID: &p.ID, ParticipationID: &part.ID,
			Type: "join", Asset: asset, Amount: amount, BalanceAfter: &part.CurrentBalance,
			Status: "completed", Description: "Joined product (seed)", CreatedAt: joined,
			Metadata: map[string]interface{}{"source": "seed"},
		})

		count++
	}
	return count, nil
}

// PrintCredentials writes the demo login table to stdout via the progress callback style.
func PrintCredentials() string {
	return fmt.Sprintf(`
-------------------------------------
Super Admin
Email:
%s
Password:
%s
-------------------------------------
Admin
Email:
admin1@coindistro.com
Password:
%s
-------------------------------------
User
Email:
user1@coindistro.com
Password:
%s
-------------------------------------
`, bootstrap.SuperAdminEmail, bootstrap.SuperAdminPassword, AdminPassword, UserPassword)
}

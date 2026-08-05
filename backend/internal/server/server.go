package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/auth"
	"github.com/coindistro/backend/internal/bootstrap"
	"github.com/coindistro/backend/internal/cache"
	"github.com/coindistro/backend/internal/config"
	"github.com/coindistro/backend/internal/database"
	earningshandlers "github.com/coindistro/backend/internal/earnings/handlers"
	earningsservice "github.com/coindistro/backend/internal/earnings/service"
	earningsstore "github.com/coindistro/backend/internal/earnings/store"
	"github.com/coindistro/backend/internal/email"
	"github.com/coindistro/backend/internal/events"
	"github.com/coindistro/backend/internal/featureflags"
	"github.com/coindistro/backend/internal/identity/handlers"
	idservice "github.com/coindistro/backend/internal/identity/service"
	"github.com/coindistro/backend/internal/identity/store"
	investhandlers "github.com/coindistro/backend/internal/investments/handlers"
	investservice "github.com/coindistro/backend/internal/investments/service"
	investstore "github.com/coindistro/backend/internal/investments/store"
	"github.com/coindistro/backend/internal/logger"
	"github.com/coindistro/backend/internal/metrics"
	"github.com/coindistro/backend/internal/rbac"
	"github.com/coindistro/backend/internal/routes"
	"github.com/coindistro/backend/internal/scheduler"
	"github.com/coindistro/backend/internal/seed"
	"github.com/coindistro/backend/internal/storage"
	"github.com/coindistro/backend/internal/telemetry"
	"github.com/coindistro/backend/internal/workers"
)

// Server represents the HTTP server with all infrastructure components.
type Server struct {
	cfg                *config.Config
	logger             *logger.Logger
	db                 *database.Database
	redis              *cache.Cache
	auth               *auth.Auth
	rbac               *rbac.RBAC
	eventBus           *events.InMemoryBus
	workerPool         *workers.Pool
	jobRegistry        *workers.Registry
	sched              *scheduler.Scheduler
	featureFlags       *featureflags.Manager
	promMetrics        *metrics.Metrics
	tracer             *telemetry.TracerProvider
	emailSender        email.Sender
	storageProv        storage.Provider
	identitySvc        *idservice.Service
	investmentSvc      *investservice.Service
	investmentHandlers *investhandlers.Handlers
	earningsSvc        *earningsservice.Service
	earningsHandlers   *earningshandlers.Handlers
	engine             *gin.Engine
	http               *http.Server
}

// New creates a new Server instance with all infrastructure components.
func New(cfg *config.Config) (*Server, error) {
	// Initialize logger
	log, err := logger.New(
		cfg.Logging.Level,
		cfg.Logging.Encoding,
		cfg.Logging.OutputPaths,
		cfg.Logging.ErrorOutputPaths,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	buildVersion, buildCommit, buildDate := readBuildInfo(cfg.App.Version)
	log.Info("initializing server",
		zap.String("app", cfg.App.Name),
		zap.String("version", cfg.App.Version),
		zap.String("environment", cfg.App.Environment),
		zap.String("build_version", buildVersion),
		zap.String("git_commit", buildCommit),
		zap.String("build_date", buildDate),
	)

	// Initialize database
	var db *database.Database
	if cfg.Database.Host != "" {
		db, err = database.New(cfg.Database, log.Logger)
		if err != nil {
			log.Warn("database connection failed, continuing without database", zap.Error(err))
			db = nil
		} else {
			log.Info("starting database migrations")
			migDir := bootstrap.ResolveMigrationsDir()
			log.Info("resolved migrations directory", zap.String("path", migDir))
			if err := bootstrap.RunMigrations(context.Background(), db, migDir, log.Logger); err != nil {
				return nil, fmt.Errorf("database migrations failed: %w", err)
			}
			log.Info("database schema ready")
		}
	} else {
		log.Info("database not configured, running without database")
	}

	log.Info("Redis environment variables",
		zap.String("COINDISTRO_REDIS_URL", envStatus(os.Getenv("COINDISTRO_REDIS_URL"))),
		zap.String("COINDISTRO_REDIS_HOST", envStatus(os.Getenv("COINDISTRO_REDIS_HOST"))),
		zap.String("COINDISTRO_REDIS_PORT", envStatus(os.Getenv("COINDISTRO_REDIS_PORT"))),
		zap.String("COINDISTRO_REDIS_PASSWORD", envStatus(os.Getenv("COINDISTRO_REDIS_PASSWORD"))),
	)

	// Initialize Redis
	var redis *cache.Cache
	if cfg.Redis.IsConfigured() {
		redis, err = cache.New(cfg.Redis, log.Logger)
		if err != nil {
			log.Warn("redis initialization failed, continuing without redis", zap.Error(err))
			redis = nil
		}
	} else {
		log.Info("redis not configured, running without redis")
	}

	// Initialize auth service
	authService := auth.New(cfg.Auth, log.Logger)

	// Initialize RBAC
	rbacService := rbac.New()
	log.Info("rbac initialized", zap.Int("roles", len(rbacService.GetRoles())))

	// Initialize event bus
	eventBus := events.NewInMemoryBus(log.Logger)
	log.Info("event bus initialized")

	// Initialize feature flags (registration defaults to enabled for public onboarding).
	ff := featureflags.New(log.Logger, cfg.App.Environment)
	if cfg.FeatureFlags.Enabled && len(cfg.FeatureFlags.Flags) > 0 {
		ff.LoadFromConfig(context.Background(), cfg.FeatureFlags.Flags)
	}
	// Explicit registration config / COINDISTRO_REGISTRATION_ENABLED always wins.
	if err := ff.Set(featureflags.FlagRegistration, cfg.Registration.Enabled); err != nil {
		log.Warn("failed to set registration.enabled flag", zap.Error(err))
	}
	if err := ff.Set(featureflags.FlagInviteOnly, cfg.Registration.InviteOnly); err != nil {
		log.Warn("failed to set registration.invite_only flag", zap.Error(err))
	}
	log.Info("feature flags initialized",
		zap.Int("flags", len(ff.GetAllFlags())),
		zap.Bool("flag_registration", ff.IsEnabled(featureflags.FlagRegistration)),
		zap.Bool("flag_invite_only", ff.IsEnabled(featureflags.FlagInviteOnly)),
		zap.Bool("flag_requires_referral", ff.IsEnabled(featureflags.FlagRequiresReferral)),
	)
	log.Info(fmt.Sprintf("Registration enabled: %t", cfg.Registration.Enabled))
	log.Info(fmt.Sprintf("Invite only mode: %t", cfg.Registration.InviteOnly))

	// Initialize Prometheus metrics
	var promMetrics *metrics.Metrics
	if cfg.Monitoring.PrometheusEnabled {
		promMetrics = metrics.New()
		log.Info("prometheus metrics initialized")
	}

	// Initialize OpenTelemetry tracing
	tracer, err := telemetry.NewTracerProvider(
		telemetry.Config{
			Enabled:     cfg.Telemetry.Enabled,
			ServiceName: cfg.Telemetry.ServiceName,
			Endpoint:    cfg.Telemetry.Endpoint,
			Environment: cfg.App.Environment,
			SampleRate:  cfg.Telemetry.SampleRate,
		},
		log.Logger,
	)
	if err != nil {
		log.Warn("telemetry initialization failed, continuing without tracing", zap.Error(err))
		tracer = nil
	} else if cfg.Telemetry.Enabled {
		log.Info("telemetry initialized", zap.String("endpoint", cfg.Telemetry.Endpoint))
	}

	// Initialize email sender
	var emailSender email.Sender
	switch cfg.Email.Provider {
	case "smtp":
		emailSender = email.NewSMTPSender(email.SMTPConfig{
			Host: cfg.Email.SMTP.Host, Port: cfg.Email.SMTP.Port,
			Username: cfg.Email.SMTP.Username, Password: cfg.Email.SMTP.Password,
			From: cfg.Email.SMTP.From, FromName: cfg.Email.SMTP.FromName,
			UseTLS: cfg.Email.SMTP.UseTLS,
		}, log.Logger)
		log.Info("email sender initialized", zap.String("provider", "smtp"))
	default:
		emailSender = email.NewNoopSender(log.Logger)
		log.Info("email sender initialized", zap.String("provider", "noop"))
	}

	// Initialize storage provider
	var storageProv storage.Provider
	switch cfg.Storage.Provider {
	case "local":
		prov, err := storage.NewLocalProvider(cfg.Storage.BasePath, cfg.Storage.BaseURL, log.Logger)
		if err != nil {
			log.Warn("local storage initialization failed", zap.Error(err))
			storageProv = storage.NewInMemoryProvider(log.Logger)
		} else {
			storageProv = prov
		}
	default:
		storageProv = storage.NewInMemoryProvider(log.Logger)
	}
	log.Info("storage provider initialized", zap.String("provider", cfg.Storage.Provider))

	// Initialize worker pool
	var workerPool *workers.Pool
	var jobRegistry *workers.Registry
	if cfg.Workers.Enabled {
		workerPool = workers.NewPool(workers.PoolConfig{
			NumWorkers: cfg.Workers.NumWorkers,
			QueueSize:  cfg.Workers.QueueSize,
			Logger:     log.Logger,
		})
		jobRegistry = workers.NewRegistry()
		log.Info("worker pool initialized",
			zap.Int("workers", cfg.Workers.NumWorkers),
			zap.Int("queue_size", cfg.Workers.QueueSize),
		)
	}

	// Initialize scheduler
	var sched *scheduler.Scheduler
	if cfg.Scheduler.Enabled {
		sched = scheduler.New(log.Logger)
		log.Info("scheduler initialized")
	}

	// Initialize Identity Service
	var identitySvc *idservice.Service
	if db != nil && db.Pool != nil {
		identityStore := store.New(db.Pool)
		identityCfg := idservice.DefaultConfig()
		// Wire COINDISTRO_REGISTRATION_ENABLED into the identity service (source of truth).
		identityCfg.RegistrationEnabled = cfg.Registration.Enabled
		identityCfg.InviteOnly = cfg.Registration.InviteOnly
		identitySvc = idservice.New(
			identityStore,
			authService,
			rbacService,
			eventBus,
			jobRegistry,
			workerPool,
			emailSender,
			ff,
			nil, // auditLogger - will be wired when audit store is implemented
			promMetrics,
			log.Logger,
			identityCfg,
		)
		log.Info("identity service initialized",
			zap.Bool("registration_enabled", identityCfg.RegistrationEnabled),
			zap.Bool("invite_only", identityCfg.InviteOnly),
		)

		// Ensure platform super admin exists (idempotent; safe in production).
		if result, err := bootstrap.EnsureSuperAdmin(context.Background(), bootstrap.Dependencies{
			Config:   cfg,
			DB:       db,
			Identity: identitySvc,
			Logger:   log.Logger,
		}, ""); err != nil {
			log.Warn("super admin bootstrap failed", zap.Error(err))
		} else if result != nil {
			log.Info("super admin bootstrap",
				zap.Bool("already_completed", result.AlreadyCompleted),
				zap.String("message", result.Message),
			)
		}

		// Development demo accounts (idempotent). Disabled in production unless explicitly allowed.
		if seed.ShouldSeedDemoUsers(cfg) {
			log.Info("seeding development demo accounts")
			if err := seed.SeedDemoAccounts(context.Background(), db, log.Logger); err != nil {
				log.Warn("demo account seed failed", zap.Error(err))
			}
		} else {
			log.Info("demo account seed skipped",
				zap.String("environment", cfg.App.Environment),
			)
		}
	}

	// Initialize Investment Service
	var investmentSvc *investservice.Service
	var investmentHandlers *investhandlers.Handlers
	var earningsSvc *earningsservice.Service
	var earningsHandlers *earningshandlers.Handlers

	// Paystack credentials are loaded exclusively from PAYSTACK_* env vars so
	// the PurpleSoftHub (dev) or future CoinDistro account can be swapped by
	// changing environment variables only — no code changes.
	paystack := config.LoadPaystackCredentials()
	flutterwaveSecretKey := os.Getenv("COINDISTRO_FLUTTERWAVE_SECRET_KEY")
	flutterwavePublicKey := os.Getenv("COINDISTRO_FLUTTERWAVE_PUBLIC_KEY")
	flutterwaveSecretHash := os.Getenv("COINDISTRO_FLUTTERWAVE_SECRET_HASH")

	// Startup-time validation: fail fast with a descriptive warning instead of
	// only returning 503 EARNINGS_GATEWAY_NOT_CONFIGURED when a user hits
	// /api/v1/investments/paystack/init. Missing keys mean payment init will
	// fail for every caller.
	if paystack.SecretKey == "" {
		log.Warn("payment gateway not configured: PAYSTACK_SECRET_KEY is empty",
			zap.String("env_var", "PAYSTACK_SECRET_KEY"),
			zap.String("impact", "POST /api/v1/investments/paystack/init and POST /api/v1/payments/paystack/init will return 503 EARNINGS_GATEWAY_NOT_CONFIGURED"),
		)
	} else {
		log.Info("Paystack gateway configured from environment",
			zap.Bool("has_public_key", paystack.PublicKey != ""),
			zap.Bool("has_callback_url", paystack.CallbackURL != ""),
			zap.Bool("has_webhook_secret", paystack.WebhookSecret != ""),
		)
	}
	if paystack.PublicKey == "" {
		log.Warn("payment gateway: PAYSTACK_PUBLIC_KEY is empty",
			zap.String("env_var", "PAYSTACK_PUBLIC_KEY"),
		)
	}
	if paystack.CallbackURL == "" {
		log.Warn("payment gateway: PAYSTACK_CALLBACK_URL is empty; falling back to BaseURL-derived callback",
			zap.String("env_var", "PAYSTACK_CALLBACK_URL"),
		)
	}
	if paystack.WebhookSecret == "" {
		log.Warn("payment gateway: PAYSTACK_WEBHOOK_SECRET is empty; webhook HMAC will use PAYSTACK_SECRET_KEY",
			zap.String("env_var", "PAYSTACK_WEBHOOK_SECRET"),
		)
	}
	if flutterwaveSecretKey == "" {
		log.Warn("payment gateway not configured: COINDISTRO_FLUTTERWAVE_SECRET_KEY is empty",
			zap.String("env_var", "COINDISTRO_FLUTTERWAVE_SECRET_KEY"),
			zap.String("impact", "POST /api/v1/investments/flutterwave/init and POST /api/v1/payments/flutterwave/init will return 503 EARNINGS_GATEWAY_NOT_CONFIGURED"),
		)
	}
	if flutterwavePublicKey == "" {
		log.Warn("payment gateway not configured: COINDISTRO_FLUTTERWAVE_PUBLIC_KEY is empty",
			zap.String("env_var", "COINDISTRO_FLUTTERWAVE_PUBLIC_KEY"),
		)
	}
	if flutterwaveSecretHash == "" {
		log.Warn("payment gateway not configured: COINDISTRO_FLUTTERWAVE_SECRET_HASH is empty",
			zap.String("env_var", "COINDISTRO_FLUTTERWAVE_SECRET_HASH"),
			zap.String("impact", "Flutterwave webhook signature verification will fail"),
		)
	}

	investmentCfg := investservice.Config{
		BaseURL:               cfg.App.BaseURL,
		PaystackSecretKey:     paystack.SecretKey,
		PaystackPublicKey:     paystack.PublicKey,
		PaystackCallbackURL:   paystack.CallbackURL,
		PaystackWebhookSecret: paystack.SignatureSecret(),
		FlutterwaveSecretKey:  flutterwaveSecretKey,
		FlutterwavePublicKey:  flutterwavePublicKey,
		FlutterwaveSecretHash: flutterwaveSecretHash,
	}

	earningsCfg := earningsservice.Config{
		BaseURL:               cfg.App.BaseURL,
		AppURL:                cfg.App.BaseURL,
		PaystackSecretKey:     paystack.SecretKey,
		PaystackPublicKey:     paystack.PublicKey,
		PaystackCallbackURL:   paystack.CallbackURL,
		PaystackWebhookSecret: paystack.SignatureSecret(),
		FlutterwaveSecretKey:  flutterwaveSecretKey,
		FlutterwavePublicKey:  flutterwavePublicKey,
		FlutterwaveSecretHash: flutterwaveSecretHash,
	}

	if db != nil && db.Pool != nil {
		investmentStore := investstore.New(db.Pool)
		investmentSvc = investservice.New(
			investmentStore,
			eventBus,
			jobRegistry,
			workerPool,
			nil, // auditLogger - will be wired when audit store is implemented
			promMetrics,
			log.Logger,
			investmentCfg,
		)

		investmentHandlers = investhandlers.New(investmentSvc, log.Logger)
		log.Info("investment service initialized")

		// Register Genesis maturity processor as early as possible so scheduler Start logs tasks >= 1.
		if sched != nil {
			sched.AddTask(scheduler.Task{
				ID:       "investment_maturity_check",
				Name:     "Investment Maturity Check",
				Interval: 1 * time.Minute,
				Handler: func(ctx context.Context) error {
					return investmentSvc.ProcessMaturedInvestments(ctx)
				},
			})
			log.Info("Scheduler task registered",
				zap.String("task_id", "investment_maturity_check"),
				zap.String("name", "Investment Maturity Check"),
				zap.Duration("interval", time.Minute),
			)
		}

		// Initialize Earnings Investor Dashboard service
		earningsStore := earningsstore.New(db.Pool)
		earningsSvc = earningsservice.New(
			earningsStore,
			eventBus,
			jobRegistry,
			workerPool,
			nil,
			log.Logger,
			earningsCfg,
		)
		earningsHandlers = earningshandlers.New(earningsSvc, log.Logger)
		log.Info("earnings investment service initialized")

		if sched != nil {
			sched.AddTask(scheduler.Task{
				ID:       "earnings_daily_rewards",
				Name:     "Earnings Daily Rewards",
				Interval: 1 * time.Hour,
				Handler: func(ctx context.Context) error {
					return earningsSvc.ProcessDailyRewards(ctx)
				},
			})
			log.Info("Scheduler task registered",
				zap.String("task_id", "earnings_daily_rewards"),
				zap.String("name", "Earnings Daily Rewards"),
				zap.Duration("interval", time.Hour),
			)
		}
	} else {
		log.Warn("database not configured; starting investment API in fallback mode")
		investmentSvc = investservice.New(nil, eventBus, jobRegistry, workerPool, nil, promMetrics, log.Logger, investmentCfg)
		investmentHandlers = investhandlers.New(investmentSvc, log.Logger)
		earningsSvc = earningsservice.New(nil, eventBus, jobRegistry, workerPool, nil, log.Logger, earningsCfg)
		earningsHandlers = earningshandlers.New(earningsSvc, log.Logger)
	}

	// Create identity handlers
	identityHandlers := handlers.New(identitySvc, log.Logger)

	// Setup routes
	engine := routes.SetupRouter(cfg, log.Logger, db, redis, authService, rbacService, ff, promMetrics, identityHandlers, nil, investmentHandlers, earningsHandlers, workerPool, sched)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		cfg:                cfg,
		logger:             log,
		db:                 db,
		redis:              redis,
		auth:               authService,
		rbac:               rbacService,
		eventBus:           eventBus,
		workerPool:         workerPool,
		jobRegistry:        jobRegistry,
		sched:              sched,
		featureFlags:       ff,
		promMetrics:        promMetrics,
		tracer:             tracer,
		emailSender:        emailSender,
		storageProv:        storageProv,
		identitySvc:        identitySvc,
		investmentSvc:      investmentSvc,
		investmentHandlers: investmentHandlers,
		earningsSvc:        earningsSvc,
		earningsHandlers:   earningsHandlers,
		engine:             engine,
		http:               httpServer,
	}, nil
}

func envStatus(value string) string {
	if value == "" {
		return "missing"
	}
	return "present"
}

// readBuildInfo returns version, git commit, and build date for startup diagnostics.
// Commit/date come from Go module build info (vcs.revision / vcs.time) when available.
func readBuildInfo(appVersion string) (version, commit, buildDate string) {
	version = appVersion
	if version == "" {
		version = "unknown"
	}
	commit = "unknown"
	buildDate = "unknown"

	if bi, ok := debug.ReadBuildInfo(); ok {
		if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			version = bi.Main.Version
		}
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				if len(s.Value) > 12 {
					commit = s.Value[:12]
				} else if s.Value != "" {
					commit = s.Value
				}
			case "vcs.time":
				if s.Value != "" {
					buildDate = s.Value
				}
			}
		}
	}
	// Allow CI/CD to inject explicit values without ldflags complexity.
	if v := os.Getenv("COINDISTRO_BUILD_VERSION"); v != "" {
		version = v
	}
	if v := os.Getenv("COINDISTRO_GIT_COMMIT"); v != "" {
		commit = v
	}
	if v := os.Getenv("COINDISTRO_BUILD_DATE"); v != "" {
		buildDate = v
	}
	return version, commit, buildDate
}

// Start starts the HTTP server and all background services.
func (s *Server) Start() error {
	ctx := context.Background()

	// Start system metrics collection
	if s.promMetrics != nil {
		go s.promMetrics.RecordSystemMetrics(ctx)
	}

	// Start worker pool
	if s.workerPool != nil {
		s.workerPool.Start(ctx, func(ctx context.Context, job workers.Job) error {
			return s.jobRegistry.Run(ctx, job)
		})
		s.logger.Info("worker pool started")
	}

	// Start scheduler (maturity task is registered during New when investment service is available)
	if s.sched != nil {
		// Defensive re-register if New path skipped (e.g. late investment wiring)
		if s.investmentSvc != nil {
			statuses := s.sched.GetStatus()
			hasMaturity := false
			for _, st := range statuses {
				if st.TaskID == "investment_maturity_check" {
					hasMaturity = true
					break
				}
			}
			if !hasMaturity {
				s.sched.AddTask(scheduler.Task{
					ID:       "investment_maturity_check",
					Name:     "Investment Maturity Check",
					Interval: 1 * time.Minute,
					Handler: func(ctx context.Context) error {
						return s.investmentSvc.ProcessMaturedInvestments(ctx)
					},
				})
				s.logger.Info("Scheduler task registered",
					zap.String("task_id", "investment_maturity_check"),
					zap.Duration("interval", time.Minute),
				)
			}
		}
		s.sched.Start()
	}

	// Channel to listen for errors
	errChan := make(chan error, 1)

	// Bind the socket before serving so health probes can succeed as soon as we log "accepting".
	ln, err := net.Listen("tcp", s.cfg.Server.Address())
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.cfg.Server.Address(), err)
	}

	s.logger.Info("server starting",
		zap.String("address", s.cfg.Server.Address()),
		zap.String("environment", s.cfg.App.Environment),
		zap.String("health_path", "/health"),
	)
	s.logger.Info("HTTP server accepting connections",
		zap.String("address", ln.Addr().String()),
	)

	go func() {
		if err := s.http.Serve(ln); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		s.logger.Info("shutdown signal received", zap.String("signal", sig.String()))
	case err := <-errChan:
		return err
	}

	return s.Shutdown()
}

// Shutdown gracefully shuts down the server and all background services.
func (s *Server) Shutdown() error {
	s.logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if s.sched != nil {
		s.sched.Stop()
		s.logger.Info("scheduler stopped")
	}
	if s.workerPool != nil {
		s.workerPool.Stop()
		s.logger.Info("worker pool stopped")
	}
	if err := s.http.Shutdown(ctx); err != nil {
		s.logger.Error("server shutdown error", zap.Error(err))
		return fmt.Errorf("server shutdown error: %w", err)
	}
	s.logger.Info("HTTP server stopped")

	if s.eventBus != nil {
		_ = s.eventBus.Close()
	}
	if s.tracer != nil {
		_ = s.tracer.Shutdown(ctx)
	}
	if s.emailSender != nil {
		_ = s.emailSender.Close()
	}
	if s.storageProv != nil {
		_ = s.storageProv.Close()
	}
	if s.db != nil {
		s.db.Close()
	}
	if s.redis != nil {
		s.redis.Close()
	}
	_ = s.logger.Sync()

	s.logger.Info("server shutdown complete")
	return nil
}

// Engine returns the Gin engine (useful for testing).
func (s *Server) Engine() *gin.Engine                          { return s.engine }
func (s *Server) Logger() *logger.Logger                       { return s.logger }
func (s *Server) Database() *database.Database                 { return s.db }
func (s *Server) Redis() *cache.Cache                          { return s.redis }
func (s *Server) Auth() *auth.Auth                             { return s.auth }
func (s *Server) RBAC() *rbac.RBAC                             { return s.rbac }
func (s *Server) EventBus() *events.InMemoryBus                { return s.eventBus }
func (s *Server) WorkerPool() *workers.Pool                    { return s.workerPool }
func (s *Server) JobRegistry() *workers.Registry               { return s.jobRegistry }
func (s *Server) Scheduler() *scheduler.Scheduler              { return s.sched }
func (s *Server) FeatureFlags() *featureflags.Manager          { return s.featureFlags }
func (s *Server) Metrics() *metrics.Metrics                    { return s.promMetrics }
func (s *Server) Tracer() *telemetry.TracerProvider            { return s.tracer }
func (s *Server) EmailSender() email.Sender                    { return s.emailSender }
func (s *Server) Storage() storage.Provider                    { return s.storageProv }
func (s *Server) IdentityService() *idservice.Service          { return s.identitySvc }
func (s *Server) InvestmentService() *investservice.Service    { return s.investmentSvc }
func (s *Server) InvestmentHandlers() *investhandlers.Handlers { return s.investmentHandlers }

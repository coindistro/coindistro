package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/auth"
	"github.com/coindistro/backend/internal/cache"
	"github.com/coindistro/backend/internal/config"
	"github.com/coindistro/backend/internal/database"
	earnhandlers "github.com/coindistro/backend/internal/earn/handlers"
	earnservice "github.com/coindistro/backend/internal/earn/service"
	earnstore "github.com/coindistro/backend/internal/earn/store"
	"github.com/coindistro/backend/internal/email"
	"github.com/coindistro/backend/internal/events"
	"github.com/coindistro/backend/internal/featureflags"
	"github.com/coindistro/backend/internal/identity/handlers"
	idservice "github.com/coindistro/backend/internal/identity/service"
	"github.com/coindistro/backend/internal/identity/store"
	"github.com/coindistro/backend/internal/logger"
	"github.com/coindistro/backend/internal/metrics"
	"github.com/coindistro/backend/internal/rbac"
	"github.com/coindistro/backend/internal/routes"
	"github.com/coindistro/backend/internal/scheduler"
	"github.com/coindistro/backend/internal/storage"
	"github.com/coindistro/backend/internal/telemetry"
	"github.com/coindistro/backend/internal/workers"
)

// Server represents the HTTP server with all infrastructure components.
type Server struct {
	cfg          *config.Config
	logger       *logger.Logger
	db           *database.Database
	redis        *cache.Cache
	auth         *auth.Auth
	rbac         *rbac.RBAC
	eventBus     *events.InMemoryBus
	workerPool   *workers.Pool
	jobRegistry  *workers.Registry
	sched        *scheduler.Scheduler
	featureFlags *featureflags.Manager
	promMetrics  *metrics.Metrics
	tracer       *telemetry.TracerProvider
	emailSender  email.Sender
	storageProv  storage.Provider
	identitySvc  *idservice.Service
	earnSvc      *earnservice.Service
	engine       *gin.Engine
	http         *http.Server
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

	log.Info("initializing server",
		zap.String("app", cfg.App.Name),
		zap.String("version", cfg.App.Version),
		zap.String("environment", cfg.App.Environment),
	)

	// Initialize database
	var db *database.Database
	if cfg.Database.Host != "" {
		db, err = database.New(cfg.Database, log.Logger)
		if err != nil {
			log.Warn("database connection failed, continuing without database", zap.Error(err))
			db = nil
		}
	} else {
		log.Info("database not configured, running without database")
	}

	// Initialize Redis
	var redis *cache.Cache
	if cfg.Redis.Host != "" {
		redis, err = cache.New(cfg.Redis, log.Logger)
		if err != nil {
			log.Warn("redis connection failed, continuing without redis", zap.Error(err))
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

	// Initialize feature flags
	ff := featureflags.New(log.Logger, cfg.App.Environment)
	if cfg.FeatureFlags.Enabled && len(cfg.FeatureFlags.Flags) > 0 {
		ff.LoadFromConfig(context.Background(), cfg.FeatureFlags.Flags)
	}
	log.Info("feature flags initialized", zap.Int("flags", len(ff.GetAllFlags())))

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
		log.Info("identity service initialized")
	}

	// Create identity handlers
	identityHandlers := handlers.New(identitySvc, log.Logger)

	// Initialize Earn Service
	var earnSvc *earnservice.Service
	var earnHandlers *earnhandlers.Handlers
	if db != nil && db.Pool != nil {
		earnSvc = earnservice.New(
			earnstore.New(db.Pool),
			eventBus,
			jobRegistry,
			workerPool,
			ff,
			nil, // audit logger wired when store is available
			promMetrics,
			log.Logger,
		)
		earnHandlers = earnhandlers.New(earnSvc, ff, log.Logger)
		log.Info("earn service initialized")

		// Register earn scheduler tasks
		if sched != nil {
			sched.AddTask(scheduler.Task{
				ID: "earn_daily_rewards", Name: "Earn Daily Reward Calculations",
				Interval: 24 * time.Hour,
				Handler:  func(ctx context.Context) error { return earnSvc.RunDailyRewardCalculations(ctx) },
			})
			sched.AddTask(scheduler.Task{
				ID: "earn_lifecycle", Name: "Earn Product Lifecycle Updates",
				Interval: 15 * time.Minute,
				Handler:  func(ctx context.Context) error { return earnSvc.RunLifecycleUpdates(ctx) },
			})
			sched.AddTask(scheduler.Task{
				ID: "earn_performance_snapshots", Name: "Earn Performance Snapshots",
				Interval: 1 * time.Hour,
				Handler:  func(ctx context.Context) error { return earnSvc.RunPerformanceSnapshots(ctx) },
			})
			sched.AddTask(scheduler.Task{
				ID: "earn_metrics_refresh", Name: "Earn Metrics Refresh",
				Interval: 1 * time.Minute,
				Handler: func(ctx context.Context) error {
					earnSvc.RefreshMetrics(ctx)
					return nil
				},
			})
			log.Info("earn scheduler tasks registered")
		}
	}

	// Setup routes
	engine := routes.SetupRouter(cfg, log.Logger, db, redis, authService, rbacService, ff, promMetrics, identityHandlers, earnHandlers, workerPool, sched)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		cfg:          cfg,
		logger:       log,
		db:           db,
		redis:        redis,
		auth:         authService,
		rbac:         rbacService,
		eventBus:     eventBus,
		workerPool:   workerPool,
		jobRegistry:  jobRegistry,
		sched:        sched,
		featureFlags: ff,
		promMetrics:  promMetrics,
		tracer:       tracer,
		emailSender:  emailSender,
		storageProv:  storageProv,
		identitySvc:  identitySvc,
		earnSvc:      earnSvc,
		engine:       engine,
		http:         httpServer,
	}, nil
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

	// Start scheduler
	if s.sched != nil {
		s.sched.Start()
	}

	// Channel to listen for errors
	errChan := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		s.logger.Info("server starting",
			zap.String("address", s.cfg.Server.Address()),
			zap.String("environment", s.cfg.App.Environment),
		)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
func (s *Server) Engine() *gin.Engine                 { return s.engine }
func (s *Server) Logger() *logger.Logger              { return s.logger }
func (s *Server) Database() *database.Database        { return s.db }
func (s *Server) Redis() *cache.Cache                 { return s.redis }
func (s *Server) Auth() *auth.Auth                    { return s.auth }
func (s *Server) RBAC() *rbac.RBAC                    { return s.rbac }
func (s *Server) EventBus() *events.InMemoryBus       { return s.eventBus }
func (s *Server) WorkerPool() *workers.Pool           { return s.workerPool }
func (s *Server) JobRegistry() *workers.Registry      { return s.jobRegistry }
func (s *Server) Scheduler() *scheduler.Scheduler     { return s.sched }
func (s *Server) FeatureFlags() *featureflags.Manager { return s.featureFlags }
func (s *Server) Metrics() *metrics.Metrics           { return s.promMetrics }
func (s *Server) Tracer() *telemetry.TracerProvider   { return s.tracer }
func (s *Server) EmailSender() email.Sender           { return s.emailSender }
func (s *Server) Storage() storage.Provider           { return s.storageProv }
func (s *Server) IdentityService() *idservice.Service { return s.identitySvc }

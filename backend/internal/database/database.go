package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/config"
)

// Database wraps the PostgreSQL connection pool.
type Database struct {
	Pool   *pgxpool.Pool
	config config.DatabaseConfig
	logger *zap.Logger
}

// New creates a new Database connection pool.
func New(cfg config.DatabaseConfig, logger *zap.Logger) (*Database, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = time.Duration(cfg.ConnMaxLifetime) * time.Minute
	poolConfig.HealthCheckPeriod = time.Duration(cfg.HealthCheckInterval) * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection established",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("dbname", cfg.DBName),
		zap.Int32("max_conns", poolConfig.MaxConns),
	)

	return &Database{
		Pool:   pool,
		config: cfg,
		logger: logger,
	}, nil
}

// Ping checks if the database is reachable.
func (d *Database) Ping(ctx context.Context) error {
	return d.Pool.Ping(ctx)
}

// Close closes the database connection pool.
func (d *Database) Close() {
	if d.Pool != nil {
		d.Pool.Close()
		d.logger.Info("database connection closed")
	}
}

// Stats returns pool statistics.
func (d *Database) Stats() *pgxpool.Stat {
	if d.Pool != nil {
		return d.Pool.Stat()
	}
	return nil
}

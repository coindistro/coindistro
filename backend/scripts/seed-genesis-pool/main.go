// Seed Genesis pool profits for active $30 investors.
//
// Distributes a $500 profit pool equally ($62.50 each) across up to 8 active
// Genesis investments using the earnings service (rewards ledger, wallet
// transactions, audit, notifications). Idempotent.
//
// Usage (from backend/):
//
//	go run ./scripts/seed-genesis-pool
//
// Requires the same database env as the API (Render/local).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/config"
	"github.com/coindistro/backend/internal/database"
	earningsservice "github.com/coindistro/backend/internal/earnings/service"
	earningsstore "github.com/coindistro/backend/internal/earnings/store"
	"github.com/coindistro/backend/internal/logger"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cfg, err := config.Load("")
	if err != nil {
		// Config file optional — still try with env defaults
		cfg, err = config.Load("configs/config.yaml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "config load: %v\n", err)
			os.Exit(1)
		}
	}

	log, err := logger.New(cfg.Logging.Level, cfg.Logging.Encoding, cfg.Logging.OutputPaths, cfg.Logging.ErrorOutputPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v\n", err)
		os.Exit(1)
	}

	if !cfg.Database.IsConfigured() {
		fmt.Fprintln(os.Stderr, "database not configured")
		os.Exit(1)
	}

	db, err := database.New(cfg.Database, log.Logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	st := earningsstore.New(db.Pool)
	svc := earningsservice.New(st, nil, nil, nil, nil, log.Logger, earningsservice.Config{
		BaseURL: cfg.App.BaseURL,
		AppURL:  cfg.App.BaseURL,
	})

	log.Info("Seeding Genesis pool profits",
		zap.Float64("pool_usd", earningsservice.GenesisPoolTotalUSD),
		zap.Float64("per_investor_usd", earningsservice.GenesisPoolProfitPerInvestorUSD),
		zap.Float64("investment_usd", earningsservice.GenesisPoolInvestmentUSD),
		zap.Int("max_investors", earningsservice.GenesisPoolInvestors),
	)

	summary, err := svc.SeedGenesisPoolProfits(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "seed failed: %v\n", err)
		os.Exit(1)
	}

	out, _ := json.MarshalIndent(summary, "", "  ")
	fmt.Println(string(out))
	log.Info("Genesis pool seed complete",
		zap.Int("credited", summary.InvestorsCredited),
		zap.Int("skipped", summary.InvestorsSkipped),
		zap.Float64("total_profit_usd", summary.TotalProfitUSD),
	)
}

// Platform wallet reconciliation report.
//
// Usage (from backend/):
//
//	go run ./scripts/wallet-recon
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/coindistro/backend/internal/config"
	"github.com/coindistro/backend/internal/database"
	earningsservice "github.com/coindistro/backend/internal/earnings/service"
	earningsstore "github.com/coindistro/backend/internal/earnings/store"
	"github.com/coindistro/backend/internal/logger"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		cfg, err = config.Load("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "config: %v\n", err)
			os.Exit(1)
		}
	}
	log, err := logger.New(cfg.Logging.Level, cfg.Logging.Encoding, cfg.Logging.OutputPaths, cfg.Logging.ErrorOutputPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v\n", err)
		os.Exit(1)
	}
	db, err := database.New(cfg.Database, log.Logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	svc := earningsservice.New(earningsstore.New(db.Pool), nil, nil, nil, nil, log.Logger, earningsservice.Config{})
	// Sync first so report reflects locked capital+profit.
	_, _ = svc.SyncAllInvestorWallets(ctx)
	recon, err := svc.PlatformReconciliation(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "recon: %v\n", err)
		os.Exit(1)
	}
	out, _ := json.MarshalIndent(recon, "", "  ")
	fmt.Println(string(out))
}

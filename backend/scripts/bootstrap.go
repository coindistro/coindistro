// Bootstrap CLI — development only.
//
// Usage (from backend/):
//
//	go run ./scripts/bootstrap.go
//
// Creates the Genesis Super Admin if one does not already exist.
// Never runs in production environments.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coindistro/backend/internal/bootstrap"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	rt, err := bootstrap.NewRuntime("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer rt.Close()

	bootstrap.OK("Environment: development")
	bootstrap.OK("Checking database...")
	if err := rt.DB.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: database unreachable: %v\n", err)
		os.Exit(1)
	}

	bootstrap.OK("Running migrations...")
	if err := bootstrap.RunMigrations(ctx, rt.DB, bootstrap.ResolveMigrationsDir()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: migrations failed: %v\n", err)
		os.Exit(1)
	}

	bootstrap.OK("Checking Super Admin...")
	result, err := bootstrap.Run(ctx, bootstrap.Dependencies{
		Config:   rt.Config,
		DB:       rt.DB,
		Identity: rt.Identity,
		Logger:   rt.Logger.Logger,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if result.AlreadyCompleted {
		fmt.Println("Platform already bootstrapped.")
		fmt.Println(result.Message)
		printCredentials()
		return
	}

	bootstrap.OK("Creating Super Admin...")
	fmt.Println(result.Message)
	printCredentials()
}

func printCredentials() {
	fmt.Printf(`
-------------------------------------
Super Admin
Email:
%s
Password:
%s
-------------------------------------
`, bootstrap.SuperAdminEmail, bootstrap.SuperAdminPassword)
}

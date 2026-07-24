// Seed CLI — development only.
//
// Usage (from backend/):
//
//	go run ./scripts/seed.go
//
// Populates the database with Super Admin, demo users, sessions, activity,
// referrals, Earn products, participations, and rewards.
// Never runs in production environments.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coindistro/backend/internal/bootstrap"
	"github.com/coindistro/backend/internal/seed"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	rt, err := bootstrap.NewRuntime("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer rt.Close()

	result, err := seed.Run(ctx, seed.Dependencies{
		Config:    rt.Config,
		DB:        rt.DB,
		Identity:  rt.Identity,
		Earn:      rt.Earn,
		EarnStore: rt.EarnStore,
		Logger:    rt.Logger.Logger,
	}, func(step string) {
		bootstrap.OK(step)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if result != nil {
		fmt.Printf("\nSeeded: super_admin=1 admins=%d moderator=%d users=%d products=%d participations=%d\n",
			len(result.Admins),
			boolToInt(result.Moderator != nil),
			len(result.Users),
			len(result.Products),
			result.Participations,
		)
	}

	fmt.Print(seed.PrintCredentials())
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

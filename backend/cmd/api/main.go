package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/coindistro/backend/docs"
	"github.com/coindistro/backend/internal/config"
	"github.com/coindistro/backend/internal/server"
)

// @title           Coindistro API
// @version         1.0.0
// @description     Coindistro — One Platform. Everything Crypto. API backend for Africa's next-generation crypto financial ecosystem.
// @termsOfService  https://coindistro.com/terms

// @contact.name   Coindistro Support
// @contact.url    https://coindistro.com/support
// @contact.email  support@coindistro.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345".

// @securityDefinitions.basic BasicAuth
func main() {
	configPath := flag.String("config", "", "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create and start server
	srv, err := server.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Start server (blocks until shutdown)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

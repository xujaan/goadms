package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/jan/goadms/internal/config"
	"github.com/jan/goadms/internal/server"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	// Validate critical config.
	if cfg.Database.DSN() == "" {
		logger.Error("database DSN is empty")
		os.Exit(1)
	}
	if cfg.Auth.JWTSecret == "" || cfg.Auth.JWTSecret == "change-me" {
		logger.Warn("JWT secret is default — set ADMS_JWT_SECRET env or change config.yaml")
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Error("create server", "error", err)
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("ADMS starting on :%d", cfg.Server.Port))
	if err := srv.Run(); err != nil {
		logger.Error("server run", "error", err)
		os.Exit(1)
	}
}

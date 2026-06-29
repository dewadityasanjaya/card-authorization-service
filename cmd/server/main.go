package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/dewadityasanjaya/card-authorization-service/config"
	"github.com/dewadityasanjaya/card-authorization-service/pkg/database"
	"github.com/dewadityasanjaya/card-authorization-service/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Step 1 — Load config
	cfg := config.Load()

	// Step 2 — Init logger
	logger.Init(cfg.App.Env)
	defer logger.Sync()

	logger.Info("Application starting...",
		zap.String("env", cfg.App.Env),
		zap.String("port", cfg.App.Port),
	)

	// Step 3 — Connect to database
	db := database.Connect(&cfg.Database)
	defer database.Close(db)

	// Step 4 — Wait for shutdown signal (Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gracefully...")
}

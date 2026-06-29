package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/dewadityasanjaya/card-authorization-service/config"
	"github.com/dewadityasanjaya/card-authorization-service/internal/handler"
	"github.com/dewadityasanjaya/card-authorization-service/internal/middleware"
	"github.com/dewadityasanjaya/card-authorization-service/internal/repository"
	"github.com/dewadityasanjaya/card-authorization-service/internal/service"
	"github.com/dewadityasanjaya/card-authorization-service/pkg/database"
	"github.com/dewadityasanjaya/card-authorization-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// ── Config & Logger ──────────────────────────────
	cfg := config.Load()
	logger.Init(cfg.App.Env)
	defer logger.Sync()

	logger.Info("Application starting...",
		zap.String("env", cfg.App.Env),
		zap.String("port", cfg.App.Port),
	)

	// ── Database ─────────────────────────────────────
	db := database.Connect(&cfg.Database)
	defer database.Close(db)

	// ── Repositories ─────────────────────────────────
	cardRepo := repository.NewCardRepository(db)
	authRepo := repository.NewAuthorizationRepository(db)

	// ── Services ─────────────────────────────────────
	cardSvc := service.NewCardService(cardRepo)
	txManager := database.NewTxManager(db)
	authSvc := service.NewAuthorizationService(txManager, cardRepo, authRepo)

	// ── Handlers ─────────────────────────────────────
	cardHandler := handler.NewCardHandler(cardSvc)
	transactionHandler := handler.NewTransactionHandler(authSvc)

	// ── Gin Router ───────────────────────────────────
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery()) // recover from panics

	// ── Routes ───────────────────────────────────────
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Card routes
	cards := r.Group("/cards")
	{
		cards.POST("", cardHandler.CreateCard)
		cards.GET("/:id", cardHandler.GetCard)
		cards.POST("/:id/freeze", cardHandler.FreezeCard)
		cards.POST("/:id/unfreeze", cardHandler.UnfreezeCard)
		cards.POST("/:id/topup", cardHandler.TopUp)
		cards.GET("/:id/transactions", transactionHandler.GetTransactionHistory)
	}

	// Transaction routes
	transactions := r.Group("/transactions")
	transactions.Use(middleware.IdempotencyKey())
	{
		transactions.POST("/authorize", transactionHandler.Authorize)
		transactions.POST("/:authorizationId/reverse", transactionHandler.Reverse)
	}

	// ── Start Server ─────────────────────────────────
	go func() {
		logger.Info("Server listening", zap.String("port", cfg.App.Port))
		if err := r.Run(":" + cfg.App.Port); err != nil {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// ── Graceful Shutdown ────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gracefully...")
}

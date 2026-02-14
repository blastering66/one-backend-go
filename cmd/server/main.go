// Package main is the entry point for the food service API server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/one-backend-go/internal/config"
	"github.com/one-backend-go/internal/db"
	"github.com/one-backend-go/internal/domain/auth"
	"github.com/one-backend-go/internal/domain/product"
	"github.com/one-backend-go/internal/domain/user"
	apphttp "github.com/one-backend-go/internal/http"
	"github.com/one-backend-go/internal/pkg/validate"
)

func main() {
	// ── Logger ─────────────────────────────────────────────────────────
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	// ── Config ─────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// ── MongoDB ────────────────────────────────────────────────────────
	ctx := context.Background()
	mongoDB, err := db.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		slog.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Disconnect(ctx, mongoDB); err != nil {
			slog.Error("failed to disconnect from MongoDB", "error", err)
		}
	}()

	if err := db.EnsureIndexes(ctx, mongoDB); err != nil {
		slog.Error("failed to create indexes", "error", err)
		os.Exit(1)
	}

	// ── Dependencies ───────────────────────────────────────────────────
	validator := validate.New()

	// Repositories
	userRepo := user.NewRepository(mongoDB)
	authRepo := auth.NewRepository(mongoDB)
	productRepo := product.NewRepository(mongoDB)

	// JWT Manager
	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.AccessTokenTTL)

	// Services
	userSvc := user.NewService(userRepo)
	authSvc := auth.NewService(cfg, jwtMgr, authRepo, userSvc)
	productSvc := product.NewService(productRepo)

	// Handlers
	userHandler := user.NewHandler(userSvc, validator)
	authHandler := auth.NewHandler(authSvc, validator)
	productHandler := product.NewHandler(productSvc, validator)

	// ── HTTP Server ────────────────────────────────────────────────────
	router := apphttp.NewRouter(cfg, jwtMgr, userRepo, userHandler, authHandler, productHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── Graceful Shutdown ──────────────────────────────────────────────
	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutting down server", "signal", sig.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server exited gracefully")
}

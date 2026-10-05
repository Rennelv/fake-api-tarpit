package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fake-api-server/internal/config"
	"fake-api-server/internal/stats"
	"fake-api-server/internal/tarpit"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize structured logger
	var logLevel slog.Level
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting fake API tarpit server",
		"addr", cfg.Addr,
		"delay", cfg.TarpitDelay.String(),
		"max_concurrent", cfg.MaxConcurrent,
		"max_bytes_per_stream", cfg.MaxBytesPerStream,
		"post_limit_behavior", cfg.PostLimitBehavior,
		"daily_limit_mb", cfg.DailyEgressLimitMB,
		"gzip_bomb", cfg.EnableGzipBomb, "gzip_bomb_wire_cap", cfg.GzipBombMaxBytes,
	)

	metrics := stats.NewMetrics(cfg.DailyEgressLimitMB)
	handler := tarpit.NewServerHandler(cfg, metrics)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: handler,
		// Protect against Slowloris attacks on request headers
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Graceful shutdown on SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	case sig := <-quit:
		slog.Info("Shutting down server gracefully...", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	} else {
		slog.Info("Server stopped cleanly")
	}
}

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/rjnemo/auth/internal/config"
	"github.com/rjnemo/auth/internal/driver/logging"
	"github.com/rjnemo/auth/internal/server"
)

func main() {
	baseLogger := logging.New(os.Stdout, logging.ModeText, &slog.HandlerOptions{AddSource: true})

	cfg, err := config.New()
	if err != nil {
		baseLogger.Error("configuration error", slog.Any("error", err))
		os.Exit(1)
	}

	logger := logging.New(os.Stdout, cfg.LogMode, &slog.HandlerOptions{AddSource: cfg.Environment == "development"})
	logger = logger.With(slog.String("env", cfg.Environment))

	if err := run(cfg, logger); err != nil {
		logger.Error("server exited", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(cfg *config.Config, logger *slog.Logger) error {
	srv, err := server.New(*cfg, logger)
	if err != nil {
		return fmt.Errorf("initialise server: %w", err)
	}

	logger.Info("starting server", slog.String("addr", fmt.Sprintf("http://localhost%s", cfg.ListenAddr)))
	if err := http.ListenAndServe(cfg.ListenAddr, srv.Router()); err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	return nil
}

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/rjnemo/auth/internal/logging"
	"github.com/rjnemo/auth/internal/server"
	"gorm.io/gorm/logger"
)

const listenAddr = ":8000"

func main() {
	if err := run(logger); err != nil {
		logger.Error("server exited", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	logger := logging.New(os.Stdout, logging.ModeText, &slog.HandlerOptions{AddSource: true})
	srv, err := server.New(logger)
	if err != nil {
		return fmt.Errorf("initialise server: %w", err)
	}

	logger.Info("starting server", slog.String("addr", listenAddr))
	if err := http.ListenAndServe(listenAddr, srv.Router()); err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	return nil
}

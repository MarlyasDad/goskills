package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	app "<module>/internal/app/<appname>"
	appconfig "<module>/internal/config"
	"<module>/internal/platform/logger"
)

func main() {
	configPath := flag.String("config", "", "path to config file (local run)")
	flag.Parse()

	if err := run(*configPath); err != nil {
		slog.Error("service failed", "error", err)
		os.Exit(1)
	}
}

// run держит основной сценарий отдельно от main: defer stop() успевает
// выполниться, а ошибки возвращаются с контекстом.
func run(configPath string) error {
	if err := appconfig.LoadFile(configPath); err != nil {
		return fmt.Errorf("load config file: %w", err)
	}

	cfg, err := appconfig.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.Logger)
	slog.SetDefault(log)

	a, err := app.New(cfg, log)
	if err != nil {
		return fmt.Errorf("new app: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := a.Run(ctx); err != nil {
		return fmt.Errorf("run app: %w", err)
	}

	log.Info("stopped", "timeout", cfg.ShutdownTimeout.String())
	return nil
}

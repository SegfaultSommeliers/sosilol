package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/SegfaultSommeliers/sosilol/internal/app"
	"github.com/SegfaultSommeliers/sosilol/internal/config"
	"github.com/SegfaultSommeliers/sosilol/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("error loading config", "error", err)
		os.Exit(1)
	}

	l := logger.New(cfg.Environment)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	a, err := app.NewApp(ctx, cfg, l)
	if err != nil {
		l.Error("error creating app", "error", err)
		os.Exit(1)
	}
	defer func(a *app.App) {
		if err := a.Close(); err != nil {
			l.Error("error closing app", "error", err)
		}
	}(a)

	if err = a.Start(ctx, cfg); err != nil {
		l.Error("failed to start app", "error", err)
	}
}

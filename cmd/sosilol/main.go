package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SegfaultSommeliers/sosilol/internal/app"
	"github.com/SegfaultSommeliers/sosilol/internal/config"
	"github.com/labstack/echo/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	a, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("error creating app: %v", err)
	}
	defer func(a *app.App) {
		err := a.Close()
		if err != nil {
			log.Printf("error closing app: %v", err)
		}
	}(a)

	startConfig := echo.StartConfig{
		Address:         cfg.HttpAddress,
		GracefulTimeout: cfg.GracefulTimeout,
	}

	if err := startConfig.Start(ctx, a.Echo); err != nil {
		a.Logger.Error("failed to start app", "error", err)
	}

}

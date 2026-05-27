package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/SegfaultSommeliers/sosilol/internal/app"
	"github.com/SegfaultSommeliers/sosilol/internal/config"
	"github.com/gofiber/fiber/v3"
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
		return
	}
	defer func(a *app.App) {
		err := a.Close()
		if err != nil {
			log.Printf("error closing app: %v", err)
			return
		}
	}(a)

	if err := a.Fiber.Listen(cfg.HttpAddress, fiber.ListenConfig{
		GracefulContext: ctx,
		ShutdownTimeout: cfg.GracefulTimeout,
	}); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.Logger.Error("failed to start app", "error", err)
	}

}

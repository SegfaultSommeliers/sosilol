package logger

import (
	"log/slog"
	"os"

	"github.com/SegfaultSommeliers/sosilol/internal/config"
)

func New(env config.Environment) *slog.Logger {
	level := slog.LevelInfo
	if env.IsDev() {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)
}

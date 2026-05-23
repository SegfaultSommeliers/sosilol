package middleware

import (
	"log/slog"

	"github.com/SegfaultSommeliers/sosilol/internal/logger"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Register(e *echo.Echo, l *slog.Logger) {
	e.Use(middleware.RequestID())
	e.Use(logger.RequestLogger(logger.RequestLoggerConfig{
		Logger: l,
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.BodyLimit(4 * 1024 * 1024))
}

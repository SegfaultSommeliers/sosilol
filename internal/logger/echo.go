package logger

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type RequestLoggerConfig struct {
	Logger  *slog.Logger
	Skipper middleware.Skipper
}

var DefaultRequestLoggerConfig = RequestLoggerConfig{
	Logger:  slog.Default(),
	Skipper: middleware.DefaultSkipper,
}

func RequestLogger(config RequestLoggerConfig) echo.MiddlewareFunc {
	if config.Logger == nil {
		config.Logger = DefaultRequestLoggerConfig.Logger
	}
	if config.Skipper == nil {
		config.Skipper = DefaultRequestLoggerConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			start := time.Now()

			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			reqLogger := config.Logger.With(
				slog.String("request_id", reqID),
			)

			err := next(c)

			resp, unwrapErr := echo.UnwrapResponse(c.Response())
			if unwrapErr != nil {
				reqLogger.Error("failed to unwrap response", "error", unwrapErr)
				return err
			}

			status := resp.Status
			latency := time.Since(start)

			reqLogger.Info("http_request",
				slog.String("method", c.Request().Method),
				slog.String("path", c.Path()),
				slog.String("uri", c.Request().RequestURI),
				slog.Int("status", status),
				slog.Duration("latency", latency),
				slog.String("remote_ip", c.RealIP()),
				slog.Int64("bytes_out", resp.Size),
			)

			return err
		}
	}
}

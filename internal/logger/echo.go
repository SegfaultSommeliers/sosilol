package logger

import (
	"context"
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

type ctxLoggerKey struct{}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
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

			req := c.Request()
			ctx := context.WithValue(req.Context(), ctxLoggerKey{}, reqLogger)
			c.SetRequest(req.WithContext(ctx))
			c.SetLogger(reqLogger)

			err := next(c)

			resp, unwrapErr := echo.UnwrapResponse(c.Response())
			if unwrapErr != nil {
				reqLogger.Error("failed to unwrap response", "error", unwrapErr)
				return err
			}

			latency := time.Since(start)
			status := resp.Status

			attrs := []any{
				slog.String("method", c.Request().Method),
				slog.String("path", c.Path()),
				slog.String("uri", c.Request().URL.Path),
				slog.Int("status", status),
				slog.Duration("latency", latency),
				slog.String("remote_ip", c.RealIP()),
				slog.Int64("bytes_out", resp.Size),
			}

			if err != nil {
				attrs = append(attrs, slog.String("error", err.Error()))
			}

			if status >= 500 {
				reqLogger.Error("http_request_failed", attrs...)
			} else if status >= 400 {
				reqLogger.Warn("http_request_warn", attrs...)
			} else {
				reqLogger.Info("http_request", attrs...)
			}

			return err
		}
	}
}

package logger

import (
	"context"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

type Config struct {
	Next   func(c fiber.Ctx) bool
	Logger *slog.Logger
}

var ConfigDefault = Config{
	Next:   nil,
	Logger: slog.Default(),
}

func configDefault(config ...Config) Config {
	if len(config) < 1 {
		return ConfigDefault
	}

	cfg := config[0]
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return cfg
}

type ctxLoggerKey struct{}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

var sensitivePathPrefixes = []string{"/auth/"}

func sanitizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	for _, prefix := range sensitivePathPrefixes {
		if strings.HasPrefix(u.Path, prefix) && u.RawQuery != "" {
			u.RawQuery = "[redacted]"
			return u.String()
		}
	}
	return rawURL
}

func NewRequestLogger(config ...Config) fiber.Handler {
	cfg := configDefault(config...)

	return func(c fiber.Ctx) error {
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}
		start := time.Now()

		reqID := requestid.FromContext(c)
		reqLogger := cfg.Logger.With(
			slog.String("request_id", reqID),
		)

		ctx := context.WithValue(c.Context(), ctxLoggerKey{}, reqLogger)
		c.SetContext(ctx)

		err := c.Next()

		latency := time.Since(start)
		status := 200
		if resp := c.Response(); resp != nil {
			status = resp.StatusCode()
		}

		attrs := []any{
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.String("uri", sanitizeURL(c.OriginalURL())),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("remote_ip", c.IP()),
		}

		if resp := c.Response(); resp != nil {
			attrs = append(attrs, slog.Int("bytes_out", len(resp.Body())))
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

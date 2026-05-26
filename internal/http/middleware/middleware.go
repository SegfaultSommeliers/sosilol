package middleware

import (
	"log/slog"

	"github.com/SegfaultSommeliers/sosilol/internal/config"
	"github.com/SegfaultSommeliers/sosilol/internal/logger"
	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Register(
	e *echo.Echo,
	l *slog.Logger,
	cfg config.Config,
	sessionManager *scs.SessionManager,
) {
	e.Use(middleware.RequestID())
	e.Use(logger.RequestLogger(logger.RequestLoggerConfig{
		Logger: l,
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XFrameOptions: "DENY",
		HSTSMaxAge: func() int {
			if cfg.Environment == "dev" {
				return 0
			}

			return 31536000
		}(),
		HSTSPreloadEnabled:    cfg.Environment != "dev",
		HSTSExcludeSubdomains: false,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' https: data:; font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'; upgrade-insecure-requests",
	}))
	e.Use(middleware.BodyLimit(4 * 1024 * 1024))
	e.Use(Session(sessionManager))
}

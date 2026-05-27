package middleware

import (
	"log/slog"
	"time"

	"github.com/SegfaultSommeliers/sosilol/internal/config"
	"github.com/SegfaultSommeliers/sosilol/internal/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func Register(
	app *fiber.App,
	l *slog.Logger,
	cfg config.Config,
	sessionConfig session.Config,
) {
	app.Use(requestid.New())
	app.Use(logger.NewRequestLogger(logger.Config{
		Logger: l,
	}))
	app.Use(recoverer.New())

	var hstsMaxAge int
	if cfg.Environment != "dev" {
		hstsMaxAge = 31536000
	}
	app.Use(helmet.New(helmet.Config{
		XFrameOptions:         "DENY",
		HSTSMaxAge:            hstsMaxAge,
		HSTSPreloadEnabled:    cfg.Environment != "dev",
		HSTSExcludeSubdomains: false,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; img-src 'self' https: data:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'; upgrade-insecure-requests",
	}))

	app.Use(session.New(sessionConfig))

	app.Use(limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,
	}))

	app.Use(csrf.New(csrf.Config{
		CookieSecure:   cfg.Environment != "dev",
		CookieSameSite: "Lax",
	}))
}

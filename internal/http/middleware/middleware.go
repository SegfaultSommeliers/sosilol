package middleware

import (
	"log/slog"
	"strings"
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
	"github.com/gofiber/storage/redis/v3"
)

func Register(
	app *fiber.App,
	l *slog.Logger,
	cfg config.Config,
	sessionConfig session.Config,
	redisStorage *redis.Storage,
) {
	app.Use(requestid.New())
	app.Use(logger.NewRequestLogger(logger.Config{
		Logger: l,
	}))
	app.Use(recoverer.New())

	var hstsMaxAge int
	if !cfg.Environment.IsDev() {
		hstsMaxAge = 31536000
	}

	mainHelmet := helmet.New(helmet.Config{
		XFrameOptions:         "DENY",
		HSTSMaxAge:            hstsMaxAge,
		HSTSPreloadEnabled:    !cfg.Environment.IsDev(),
		HSTSExcludeSubdomains: false,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' https://avatars.githubusercontent.com https://github.com; font-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'; upgrade-insecure-requests",
	})
	swaggerHelmet := helmet.New(helmet.Config{
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            hstsMaxAge,
		HSTSPreloadEnabled:    !cfg.Environment.IsDev(),
		HSTSExcludeSubdomains: false,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; img-src 'self' data:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'self';",
	})

	app.Use(func(c fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/v1/swagger-ui/") {
			return swaggerHelmet(c)
		}
		return mainHelmet(c)
	})

	app.Use(session.New(sessionConfig))

	app.Use(limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,
		Storage:    redisStorage,
	}))

	app.Use(csrf.New(csrf.Config{
		Storage:        redisStorage,
		CookieSecure:   !cfg.Environment.IsDev(),
		CookieSameSite: "Lax",
		CookieHTTPOnly: false,
	}))
}

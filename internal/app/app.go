package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/SegfaultSommeliers/sosilol/internal/config"
	"github.com/SegfaultSommeliers/sosilol/internal/db"
	"github.com/SegfaultSommeliers/sosilol/internal/github"
	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/internal/http/middleware"
	"github.com/SegfaultSommeliers/sosilol/internal/http/router"
	"github.com/SegfaultSommeliers/sosilol/internal/http/validator"
	"github.com/SegfaultSommeliers/sosilol/internal/paste"
	"github.com/SegfaultSommeliers/sosilol/internal/paste/cache"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/storage/redis/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

// App
//
// @title sosi.lol
// @version v1
// @description A simple paste service
// @host sosi.lol
type App struct {
	Logger      *slog.Logger
	Fiber       *fiber.App
	RedisClient *goredis.Client
	DbPool      *pgxpool.Pool
}

func NewApp(
	ctx context.Context,
	cfg config.Config,
	l *slog.Logger,
) (*App, error) {
	redisOpts := &goredis.Options{
		Addr: net.JoinHostPort(cfg.RedisHost, cfg.RedisPort),
	}
	if cfg.RedisTLS {
		redisOpts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	if cfg.RedisPassword != "" {
		redisOpts.Password = cfg.RedisPassword
	}
	redisClient := goredis.NewClient(redisOpts)

	redisStorage := redis.NewFromConnection(redisClient)
	sessionConfig := session.Config{
		Storage:           redisStorage,
		IdleTimeout:       30 * time.Minute,
		AbsoluteTimeout:   24 * time.Hour,
		CookieHTTPOnly:    true,
		CookiePath:        "/",
		CookieSecure:      !cfg.Environment.IsDev(),
		CookieSessionOnly: true,
		CookieSameSite:    "Lax",
		Extractor:         extractors.FromCookie("sosilol_session"),
	}

	poolCfg, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, err
	}
	poolCfg.ConnConfig.Host = cfg.PostgresHost
	poolCfg.ConnConfig.Port = func() uint16 {
		var port uint16
		if _, err := fmt.Sscanf(cfg.PostgresPort, "%d", &port); err != nil || port == 0 {
			return 5432
		}
		return port
	}()
	poolCfg.ConnConfig.User = cfg.PostgresUsername
	poolCfg.ConnConfig.Password = cfg.PostgresPassword
	poolCfg.ConnConfig.Database = cfg.PostgresDatabase
	if cfg.PostgresTLS {
		poolCfg.ConnConfig.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: cfg.PostgresHost,
		}
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		err := redisClient.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	if err = dbPool.Ping(ctx); err != nil {
		err := redisClient.Close()
		if err != nil {
			return nil, err
		}
		dbPool.Close()
		return nil, fmt.Errorf("database unreachable: %w", err)
	}
	queries := db.New(dbPool)

	githubService := github.NewService(
		cfg.GithubClientId,
		cfg.GithubClientSecret,
		cfg.GithubRedirectUrl,
		queries,
	)

	pasteService := paste.NewService(
		queries,
		l,

		cache.NewService(
			redisClient,
			cfg.PasteCacheTTL,
		),
	)

	app := fiber.New(fiber.Config{
		ErrorHandler:    apphttp.NewCustomErrorHandler(l),
		StructValidator: validator.NewCustomValidator(),
		CaseSensitive:   true,
		ProxyHeader:     fasthttp.HeaderXForwardedFor,
		TrustProxy:      cfg.TrustedProxy != "",
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: func() []string {
				if cfg.TrustedProxy != "" {
					return []string{cfg.TrustedProxy}
				}
				return nil
			}(),
		},
	})

	app.Hooks().OnListen(func(listenData fiber.ListenData) error {
		l.Info("Server is starting",
			"host", listenData.Host,
			"port", listenData.Port,
			"tls", listenData.TLS,
		)
		return nil
	})

	middleware.Register(app, l, cfg, sessionConfig, redisStorage)
	router.RegisterRoutes(
		app,
		redisStorage,
		githubService,
		pasteService,
	)

	return &App{
		Logger:      l,
		Fiber:       app,
		RedisClient: redisClient,
		DbPool:      dbPool,
	}, nil
}

func (app *App) Start(ctx context.Context, cfg config.Config) error {
	if err := app.Fiber.Listen(cfg.HttpAddress, fiber.ListenConfig{
		GracefulContext:       ctx,
		ShutdownTimeout:       cfg.GracefulTimeout,
		DisableStartupMessage: true,
		EnablePrefork:         cfg.EnablePrefork,
	}); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (app *App) Close() error {
	err := app.RedisClient.Close()
	app.DbPool.Close()
	if err != nil {
		return err
	}

	return nil
}

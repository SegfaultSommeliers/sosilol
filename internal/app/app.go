package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/SegfaultSommeliers/sosilol/internal/config"
	"github.com/SegfaultSommeliers/sosilol/internal/db"
	"github.com/SegfaultSommeliers/sosilol/internal/github"
	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/internal/http/middleware"
	"github.com/SegfaultSommeliers/sosilol/internal/http/router"
	"github.com/SegfaultSommeliers/sosilol/internal/http/validator"
	"github.com/SegfaultSommeliers/sosilol/internal/logger"
	"github.com/SegfaultSommeliers/sosilol/internal/paste"
	"github.com/SegfaultSommeliers/sosilol/internal/paste/cache"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/storage/redis/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

// App
//
// @title соси лол
// @version v1
// @description Лучшая паста на свете
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
) (*App, error) {
	l := logger.New(cfg.Environment)

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
		CookieSecure:      cfg.Environment != "dev",
		CookieSessionOnly: false,
		CookieSameSite:    "Lax",
		Extractor:         extractors.FromCookie("sosilol_session"),
	}

	dbPool, err := pgxpool.New(
		ctx,
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s",
			cfg.PostgresUsername,
			cfg.PostgresPassword,
			cfg.PostgresHost,
			cfg.PostgresPort,
			cfg.PostgresDatabase,
		),
	)
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

		githubService,
		cache.NewService(
			redisClient,
			cfg.PasteCacheTTL,
		),
	)

	app := fiber.New(fiber.Config{
		ErrorHandler:    apphttp.NewCustomErrorHandler(l),
		StructValidator: validator.NewCustomValidator(),
		CaseSensitive:   true,
	})

	middleware.Register(app, l, cfg, sessionConfig)
	router.RegisterRoutes(
		app,
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

func (app *App) Close() error {
	err := app.RedisClient.Close()
	app.DbPool.Close()
	if err != nil {
		return err
	}

	return nil
}

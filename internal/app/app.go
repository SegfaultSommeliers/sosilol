package app

import (
	"context"
	"crypto/tls"
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
	"github.com/SegfaultSommeliers/sosilol/internal/logger"
	"github.com/SegfaultSommeliers/sosilol/internal/paste"
	"github.com/SegfaultSommeliers/sosilol/internal/paste/cache"
	"github.com/alexedwards/scs/goredisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
)

// App
//
// @title соси лол
// @version v1
// @description Лучшая паста на свете
// @host sosi.lol
type App struct {
	Logger      *slog.Logger
	Echo        *echo.Echo
	RedisClient *redis.Client
	DbPool      *pgxpool.Pool
}

func NewApp(
	ctx context.Context,
	cfg config.Config,
) (*App, error) {
	l := logger.New(cfg.Environment)

	redisOpts := &redis.Options{
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
	redisClient := redis.NewClient(redisOpts)
	sessionManager := scs.New()
	sessionManager.Store = goredisstore.New(redisClient)
	sessionManager.HashTokenInStore = true
	sessionManager.IdleTimeout = 30 * time.Minute
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Name = "sosilol_session"
	sessionManager.Cookie.Path = "/"
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = cfg.Environment != "dev"
	sessionManager.Cookie.Persist = true

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

	e := echo.New()
	e.Validator = validator.NewCustomValidator()
	e.HTTPErrorHandler = apphttp.CustomErrorHandler

	middleware.Register(e, l, cfg, sessionManager)
	router.RegisterRoutes(
		e,
		sessionManager,

		githubService,
		pasteService,
	)

	return &App{
		Logger:      l,
		Echo:        e,
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

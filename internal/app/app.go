package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/SegfaultSommeliers/sosilol/internal/config"
	"github.com/SegfaultSommeliers/sosilol/internal/github"
	"github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/internal/http/middleware"
	"github.com/SegfaultSommeliers/sosilol/internal/http/router"
	"github.com/SegfaultSommeliers/sosilol/internal/http/validator"
	"github.com/SegfaultSommeliers/sosilol/internal/logger"
	"github.com/boj/redistore/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

type App struct {
	Logger       *slog.Logger
	Echo         *echo.Echo
	SessionStore *redistore.RediStore
	DbPool       *pgxpool.Pool
}

func NewApp(
	ctx context.Context,
	cfg config.Config,
) (*App, error) {
	l := logger.New(cfg.Environment)

	sessionStore, err := redistore.NewStore(
		redistore.KeysFromStrings(cfg.SessionSecret),
		redistore.WithAddress("tcp", cfg.RedisHost+":"+cfg.RedisPort),
	)
	if err != nil {
		return nil, err
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
		sessionStore.Close()
		return nil, err
	}

	githubService := github.NewService(
		cfg.GithubClientId,
		cfg.GithubClientSecret,
	)

	e := echo.New()
	e.Validator = validator.NewCustomValidator()
	e.HTTPErrorHandler = http.CustomErrorHandler

	middleware.Register(e, l)
	router.RegisterRoutes(
		e,
		sessionStore,
		cfg,

		githubService,
	)

	return &App{
		Logger:       l,
		Echo:         e,
		SessionStore: sessionStore,
		DbPool:       dbPool,
	}, nil
}

func (app *App) Close() error {
	err := app.SessionStore.Close()
	if err != nil {
		return err
	}
	app.DbPool.Close()

	return nil
}

package router

import (
	"github.com/SegfaultSommeliers/sosilol/internal/config"
	"github.com/SegfaultSommeliers/sosilol/internal/embed"
	"github.com/SegfaultSommeliers/sosilol/internal/github"
	"github.com/SegfaultSommeliers/sosilol/internal/health"
	"github.com/boj/redistore/v2"
	"github.com/labstack/echo/v5"
)

func RegisterRoutes(
	e *echo.Echo,
	sessionStore *redistore.RediStore,
	cfg config.Config,

	githubService *github.Service,
) {
	e.GET("/health", health.Handler)
	e.StaticFS("/", echo.MustSubFS(embed.Static, "static"))

	githubHandler := github.NewHandler(
		sessionStore,
		cfg.GithubClientId,
		cfg.GithubRedirectUrl,

		githubService,
	)
	e.GET("/auth", githubHandler.Auth)
	e.GET("/requestAuth", githubHandler.RequestAuth)
	e.GET("/profile", github.NewProfileHandler(
		githubService,
		sessionStore,
	))
}

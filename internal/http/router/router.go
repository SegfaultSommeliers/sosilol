package router

import (
	"net/http"

	"github.com/SegfaultSommeliers/sosilol/internal/config"
	"github.com/SegfaultSommeliers/sosilol/internal/embed"
	"github.com/SegfaultSommeliers/sosilol/internal/github"
	"github.com/SegfaultSommeliers/sosilol/internal/health"
	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/internal/paste"
	"github.com/SegfaultSommeliers/sosilol/view/page"
	"github.com/boj/redistore/v2"
	"github.com/labstack/echo/v5"
)

func RegisterRoutes(
	e *echo.Echo,
	sessionStore *redistore.RediStore,
	cfg config.Config,

	githubService *github.Service,
	pasteService *paste.Service,
) {
	e.GET("/health", health.Handler)
	e.GET("/", func(c *echo.Context) error {
		return apphttp.Render(c, http.StatusOK, page.Main())
	})
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
	pasteHandler := paste.NewHandler(
		pasteService,
	)
	e.POST("/save", pasteHandler.Save)
	e.GET("/view/:id", pasteHandler.View)
	e.GET("/raw/:id", pasteHandler.Raw)
}

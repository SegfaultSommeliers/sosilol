package router

import (
	"net/http"

	"github.com/SegfaultSommeliers/sosilol/internal/embed"
	"github.com/SegfaultSommeliers/sosilol/internal/github"
	"github.com/SegfaultSommeliers/sosilol/internal/github/auth"
	"github.com/SegfaultSommeliers/sosilol/internal/github/profile"
	"github.com/SegfaultSommeliers/sosilol/internal/health"
	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/internal/paste"
	"github.com/SegfaultSommeliers/sosilol/view/page"
	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v5"
	echoSwagger "github.com/swaggo/echo-swagger/v2"

	_ "github.com/SegfaultSommeliers/sosilol/docs"
)

func RegisterRoutes(
	e *echo.Echo,
	sessionManager *scs.SessionManager,

	githubService *github.Service,
	pasteService *paste.Service,
) {
	e.GET("/health", health.Handler)
	e.GET("/", func(c *echo.Context) error {
		return apphttp.Render(c, http.StatusOK, page.Main())
	})
	e.GET("/v1/swagger-ui/*", echoSwagger.WrapHandlerV3)
	e.StaticFS("/", echo.MustSubFS(embed.Static, "static"))

	authHandler := auth.NewHandler(
		githubService,
		sessionManager,
	)

	e.GET("/auth/request", authHandler.RequestAuth)
	e.GET("/auth/redirect", authHandler.RedirectAuth)

	e.GET("/profile", profile.NewHandler(
		githubService,
		sessionManager,
	))
	pasteHandler := paste.NewHandler(
		sessionManager,
		pasteService,
	)
	e.POST("/save", pasteHandler.Save)
	e.GET("/view/:id", pasteHandler.View)
	e.GET("/raw/:id", pasteHandler.Raw)
}

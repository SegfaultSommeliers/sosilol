package router

import (
	"net/http"
	"time"

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
	"github.com/labstack/echo/v5/middleware"
	echoSwagger "github.com/swaggo/echo-swagger/v2"

	_ "github.com/SegfaultSommeliers/sosilol/docs"
)

func RegisterRoutes(
	e *echo.Echo,
	sessionManager *scs.SessionManager,

	githubService *github.Service,
	pasteService *paste.Service,
) {
	authRateConfig := middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      5,
				Burst:     10,
				ExpiresIn: time.Minute,
			},
		),
		ErrorHandler: func(c *echo.Context, err error) error {
			return err
		},
		DenyHandler: func(c *echo.Context, identifier string, err error) error {
			return &apphttp.AppError{
				StatusCode: http.StatusTooManyRequests,
				Code:       "rate_limit_exceeded",
				Message:    "rate limit exceeded",
			}
		},
	}

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

	authGroup := e.Group("/auth", middleware.RateLimiterWithConfig(authRateConfig))
	authGroup.GET("/request", authHandler.RequestAuth)
	authGroup.GET("/redirect", authHandler.RedirectAuth)

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

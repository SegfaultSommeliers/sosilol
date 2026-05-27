package router

import (
	"io/fs"
	"net/http"
	"time"

	_ "github.com/SegfaultSommeliers/sosilol/docs"
	"github.com/SegfaultSommeliers/sosilol/internal/embed"
	"github.com/SegfaultSommeliers/sosilol/internal/github"
	"github.com/SegfaultSommeliers/sosilol/internal/github/auth"
	"github.com/SegfaultSommeliers/sosilol/internal/github/profile"
	"github.com/SegfaultSommeliers/sosilol/internal/health"
	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/internal/paste"
	"github.com/SegfaultSommeliers/sosilol/view/page"
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/static"
)

var saveLimiter = limiter.New(limiter.Config{
	Max:        10,
	Expiration: 1 * time.Minute,
})

var authRequestLimiter = limiter.New(limiter.Config{
	Max:        5,
	Expiration: 1 * time.Minute,
})

func RegisterRoutes(
	app *fiber.App,

	githubService *github.Service,
	pasteService *paste.Service,
) {
	app.Get("/health", health.Handler)
	app.Get("/", func(c fiber.Ctx) error {
		return apphttp.Render(c, http.StatusOK, page.Main())
	})
	app.Get("/v1/swagger-ui/*", swaggo.HandlerDefault)
	subFS, err := fs.Sub(embed.Static, "static")
	if err != nil {
		panic(err)
	}

	app.Use("/", static.New("", static.Config{
		FS:         subFS,
		Browse:     false,
		IndexNames: []string{},
	}))

	authHandler := auth.NewHandler(
		githubService,
	)

	authGroup := app.Group("/auth")
	authGroup.Get("/request", authRequestLimiter, authHandler.RequestAuth)
	authGroup.Get("/redirect", authHandler.RedirectAuth)
	authGroup.Post("/logout", authHandler.Logout)

	app.Get("/profile", profile.NewHandler(
		githubService,
	))
	pasteHandler := paste.NewHandler(
		pasteService,
	)
	app.Post("/save", saveLimiter, pasteHandler.Save)
	app.Get("/view/:id", pasteHandler.View)
	app.Get("/raw/:id", pasteHandler.Raw)
}

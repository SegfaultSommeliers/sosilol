package router

import (
	"io/fs"
	"net/http"

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
	"github.com/gofiber/fiber/v3/middleware/static"
)

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
	authGroup.Get("/request", authHandler.RequestAuth)
	authGroup.Get("/redirect", authHandler.RedirectAuth)

	app.Get("/profile", profile.NewHandler(
		githubService,
	))
	pasteHandler := paste.NewHandler(
		pasteService,
	)
	app.Post("/save", pasteHandler.Save)
	app.Get("/view/:id", pasteHandler.View)
	app.Get("/raw/:id", pasteHandler.Raw)
}

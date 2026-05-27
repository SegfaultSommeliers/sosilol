package profile

import (
	"net/http"

	"github.com/SegfaultSommeliers/sosilol/internal/github"
	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/view/page"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func NewHandler(
	githubService *github.Service,
) func(c fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		ctx := c.Context()
		sess := session.FromContext(c)

		accountType, ok := sess.Get("account_type").(string)
		if !ok || accountType != "github" {
			return c.Redirect().To("/auth/request")
		}
		accessToken, ok := sess.Get("access_token").(string)
		if !ok || accessToken == "" {
			return c.Redirect().To("/auth/request")
		}

		profile, err := githubService.GetProfile(ctx, accessToken)
		if err != nil {
			return err
		}

		return apphttp.Render(c, http.StatusOK, page.Profile(
			profile.Login,
			profile.AvatarURL,
			profile.Pastes,
		))
	}
}

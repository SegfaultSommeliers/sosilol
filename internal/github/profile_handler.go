package github

import (
	"net/http"

	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/view/page"
	"github.com/boj/redistore/v2"
	"github.com/labstack/echo/v5"
)

func NewProfileHandler(
	githubService *Service,
	sessionStore *redistore.RediStore,
) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()

		session, err := sessionStore.Get(c.Request(), "github_oauth")
		if err != nil {
			return err
		}

		accountType, ok := session.Values["account_type"].(string)
		if !ok || accountType == "" {
			return c.Redirect(http.StatusMovedPermanently, "/requestAuth")
		}

		accessToken, ok := session.Values["access_token"].(string)
		if !ok || accessToken == "" {
			return c.Redirect(http.StatusMovedPermanently, "/requestAuth")
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

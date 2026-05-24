package profile

import (
	"net/http"

	"github.com/SegfaultSommeliers/sosilol/internal/github"
	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/view/page"
	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v5"
)

func NewHandler(
	githubService *github.Service,
	sessionManager *scs.SessionManager,
) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()

		accountType := sessionManager.GetString(ctx, "account_type")
		if accountType != "github" {
			return c.Redirect(http.StatusFound, "/auth/request")
		}
		accessToken := sessionManager.GetString(ctx, "access_token")
		if accessToken == "" {
			return c.Redirect(http.StatusFound, "/auth/request")
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

package profile

import (
	"net/http"

	"github.com/SegfaultSommeliers/sosilol/internal/github"
	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/internal/shared/model"
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
		userID, ok := sess.Get("github_user_id").(int64)
		if !ok || userID == 0 {
			return c.Redirect().To("/auth/request")
		}

		login, _ := sess.Get("github_login").(string)
		avatarURL, _ := sess.Get("github_avatar_url").(string)

		pastes, err := githubService.GetPastesByUserID(ctx, userID)
		if err != nil {
			return err
		}

		if pastes == nil {
			pastes = make([]model.Paste, 0)
		}

		return apphttp.Render(c, http.StatusOK, page.Profile(
			login,
			avatarURL,
			pastes,
		))
	}
}

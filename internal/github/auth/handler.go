package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"

	"github.com/SegfaultSommeliers/sosilol/internal/github"
	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

type Handler struct {
	service *github.Service
}

func NewHandler(
	service *github.Service,
) *Handler {
	return &Handler{
		service: service,
	}
}

func isSafeRedirect(s string) bool {
	if s == "" {
		return false
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "" || u.Host != "" || u.User != nil || u.Opaque != "" {
		return false
	}
	return strings.HasPrefix(u.Path, "/") && !strings.HasPrefix(u.Path, "//")
}

type redirectQuery struct {
	Code  string `query:"code"  validate:"required" message:"missing code"`
	State string `query:"state" validate:"required" message:"missing state"`
}

// Logout
// @Summary      Log out
// @Description  Destroys the current session and redirects to the home page.
// @Tags         auth
// @Success      302  {string}  string  "Redirect to /"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /auth/logout [post]
func (h *Handler) Logout(c fiber.Ctx) error {
	sess := session.FromContext(c)
	if err := sess.Destroy(); err != nil {
		return err
	}
	return c.Redirect().To("/")
}

func (h *Handler) RequestAuth(c fiber.Ctx) error {
	nonce := rand.Text()
	sess := session.FromContext(c)

	sess.Set("oauth_state", nonce)

	redirectTo := c.Query("redirect")
	if !isSafeRedirect(redirectTo) {
		redirectTo = "/profile"
	}
	sess.Set("redirect_after_login", redirectTo)

	return c.Redirect().To(h.service.GetAuthURL(nonce))
}

func (h *Handler) RedirectAuth(c fiber.Ctx) error {
	ctx := c.Context()

	q := new(redirectQuery)
	if err := c.Bind().Query(q); err != nil {
		return err
	}

	sess := session.FromContext(c)

	expectedState, ok := sess.Get("oauth_state").(string)
	if !ok {
		return &apphttp.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "bad_request",
			Message:    "missing state",
		}
	}

	if expectedState == "" || subtle.ConstantTimeCompare([]byte(q.State), []byte(expectedState)) != 1 {
		return &apphttp.AppError{
			StatusCode: http.StatusUnauthorized,
			Code:       "unauthorized",
			Message:    "state mismatch",
		}
	}
	sess.Delete("oauth_state")

	accessToken, err := h.service.Authorize(ctx, q.Code)
	if err != nil {
		return err
	}

	profile, err := h.service.GetRawProfile(ctx, accessToken)
	if err != nil {
		return err
	}

	redirectTo := "/profile"
	if redirectToString, ok := sess.Get("redirect_after_login").(string); ok &&
		isSafeRedirect(redirectToString) {
		redirectTo = redirectToString
	}
	sess.Delete("redirect_after_login")

	if err := sess.Regenerate(); err != nil {
		return err
	}

	sess.Set("account_type", "github")
	sess.Set("github_user_id", profile.ID)
	sess.Set("github_login", profile.Login)
	sess.Set("github_avatar_url", profile.AvatarURL)

	return c.Redirect().To(redirectTo)
}

package auth

import (
	"crypto/rand"
	"net/http"
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

func (h *Handler) RequestAuth(c fiber.Ctx) error {
	nonce := rand.Text()
	sess := session.FromContext(c)

	sess.Set("oauth_state", nonce)

	redirectTo := c.Query("redirect")
	if !strings.HasPrefix(redirectTo, "/") ||
		strings.HasPrefix(redirectTo, "//") ||
		strings.HasPrefix(redirectTo, "\\\\") {
		redirectTo = "/profile"
	}

	sess.Set("redirect_after_login", redirectTo)

	return c.Redirect().To(h.service.GetAuthURL(nonce))
}

func (h *Handler) RedirectAuth(c fiber.Ctx) error {
	ctx := c.Context()

	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		return &apphttp.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "bad_request",
			Message:    "missing code or state",
		}
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
	if expectedState == "" || state != expectedState {
		return &apphttp.AppError{
			StatusCode: http.StatusUnauthorized,
			Code:       "unauthorized",
			Message:    "state mismatch",
		}
	}
	sess.Delete("oauth_state")

	accessToken, err := h.service.Authorize(ctx, code)
	if err != nil {
		return err
	}

	sess.Set("account_type", "github")
	sess.Set("access_token", accessToken)

	redirectTo := "/profile"
	if redirectToString, ok := sess.Get("redirect_after_login").(string); ok &&
		redirectToString != "" {
		redirectTo = redirectToString
	}
	sess.Delete("redirect_after_login")

	return c.Redirect().To(redirectTo)
}

package auth

import (
	"crypto/rand"
	"net/http"
	"strings"

	"github.com/SegfaultSommeliers/sosilol/internal/github"
	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	service        *github.Service
	sessionManager *scs.SessionManager
}

func NewHandler(
	service *github.Service,
	sessionManager *scs.SessionManager,
) *Handler {
	return &Handler{
		service:        service,
		sessionManager: sessionManager,
	}
}

func (h *Handler) RequestAuth(c *echo.Context) error {
	ctx := c.Request().Context()
	nonce := rand.Text()
	h.sessionManager.Put(c.Request().Context(), "oauth_state", nonce)

	redirectTo := c.QueryParam("redirect")
	if !strings.HasPrefix(redirectTo, "/") ||
		strings.HasPrefix(redirectTo, "//") ||
		strings.HasPrefix(redirectTo, "\\\\") {
		redirectTo = "/profile"
	}

	h.sessionManager.Put(ctx, "redirect_after_login", redirectTo)

	return c.Redirect(
		http.StatusFound,
		h.service.GetAuthURL(nonce),
	)
}

func (h *Handler) RedirectAuth(c *echo.Context) error {
	ctx := c.Request().Context()

	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if code == "" || state == "" {
		return &apphttp.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "bad_request",
			Message:    "missing code or state",
		}
	}

	expectedState := h.sessionManager.GetString(ctx, "oauth_state")
	if expectedState == "" || state != expectedState {
		return &apphttp.AppError{
			StatusCode: http.StatusUnauthorized,
			Code:       "unauthorized",
			Message:    "state mismatch",
		}
	}

	h.sessionManager.Remove(ctx, "oauth_state")

	accessToken, err := h.service.Authorize(ctx, code)
	if err != nil {
		return err
	}

	h.sessionManager.Put(ctx, "account_type", "github")
	h.sessionManager.Put(ctx, "access_token", accessToken)

	redirectTo := h.sessionManager.PopString(ctx, "redirect_after_login")
	if redirectTo == "" {
		redirectTo = "/profile"
	}

	return c.Redirect(http.StatusFound, redirectTo)
}

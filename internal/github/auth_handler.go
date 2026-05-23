package github

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/boj/redistore/v2"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	sessionStore *redistore.RediStore

	githubClientId    string
	githubRedirectUri string
	service           *Service
}

func NewHandler(
	sessionStore *redistore.RediStore,
	githubClientId string,
	githubRedirectUri string,

	service *Service,
) *Handler {
	return &Handler{
		sessionStore:      sessionStore,
		githubClientId:    githubClientId,
		githubRedirectUri: githubRedirectUri,

		service: service,
	}
}

func (h *Handler) Auth(c *echo.Context) error {
	ctx := c.Request().Context()

	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if code == "" || state == "" {
		return &apphttp.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "bad_request",
			Message:    fmt.Sprintf("missing code or state"),
		}
	}

	session, err := h.sessionStore.Get(c.Request(), "github_oauth")
	if err != nil {
		return err
	}

	expectedState, ok := session.Values["oauth_state"].(string)
	if !ok || expectedState == "" || state != expectedState {
		return &apphttp.AppError{
			StatusCode: http.StatusUnauthorized,
			Code:       "unauthorized",
			Message:    fmt.Sprintf("state mismatch"),
		}
	}

	// Clear the nonce after use
	delete(session.Values, "oauth_state")

	accessToken, err := h.service.Authorize(ctx, code)
	if err != nil {
		return err
	}

	session.Values["account_type"] = "github"
	session.Values["access_token"] = accessToken

	if err = session.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/redirect")
}

func (h *Handler) RequestAuth(c *echo.Context) error {
	session, err := h.sessionStore.Get(c.Request(), "github_oauth")
	if err != nil {
		return err
	}

	// Generate a cryptographically random state nonce for CSRF protection
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return err
	}
	nonce := hex.EncodeToString(b)
	session.Values["oauth_state"] = nonce

	if err := session.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return c.Redirect(
		http.StatusFound,
		fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user&state=%s",
			h.githubClientId,
			h.githubRedirectUri,
			nonce,
		),
	)
}

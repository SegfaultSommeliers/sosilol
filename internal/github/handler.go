package github

import (
	"fmt"
	"net/http"

	"github.com/SegfaultSommeliers/sosilol"
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
		return sosilol.BadRequest("code and state are required")
	}

	session, err := h.sessionStore.Get(c.Request(), "github_oauth")
	if err != nil {
		return err
	}

	if state != session.ID {
		return sosilol.BadRequest("Invalid state")
	}

	accessToken, err := h.service.Authorize(ctx, code)
	if err != nil {
		return err
	}

	session.Values["account_type"] = "github"
	session.Values["access_token"] = accessToken

	if err = session.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return c.Redirect(http.StatusMovedPermanently, "/redirect")
}

func (h *Handler) RequestAuth(c *echo.Context) error {
	session, err := h.sessionStore.Get(c.Request(), "github_oauth")
	if err != nil {
		return err
	}

	return c.Redirect(
		http.StatusMovedPermanently,
		fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user&state=%s",
			h.githubClientId,
			h.githubRedirectUri,
			session.ID,
		),
	)
}

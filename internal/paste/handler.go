package paste

import (
	"fmt"
	"net/http"

	"github.com/SegfaultSommeliers/sosilol"
	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/view/page"
	"github.com/boj/redistore/v2"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	sessionStore *redistore.RediStore

	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Save(c *echo.Context) error {
	ctx := c.Request().Context()
	text := c.Param("text")

	session, err := h.sessionStore.Get(c.Request(), "github_oauth")
	if err != nil {
		return err
	}

	if accountType, ok := session.Values["account_type"].(string); ok {
		if accessToken, ok := session.Values["access_token"].(string); accountType != "" && ok {
			id, err := h.service.Save(ctx, text, accessToken)
			if err != nil {
				return err
			}

			return c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/view/%s", id))
		}
	}

	id, err := h.service.Save(ctx, text, "")
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/view/%s", id))
}

func (h *Handler) View(c *echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.Redirect(http.StatusMovedPermanently, "/")
	}

	code, err := h.service.LoadRaw(ctx, id)
	if err != nil {
		if err := c.Redirect(http.StatusMovedPermanently, "/"); err != nil {
			return err
		}
		return err
	}

	return apphttp.Render(c, http.StatusOK, page.Paste(code))
}

func (h *Handler) Raw(c *echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return sosilol.BadRequest("id required")
	}

	code, err := h.service.LoadRaw(ctx, id)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, code)
}

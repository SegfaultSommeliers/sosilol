package paste

import (
	"fmt"
	"net/http"

	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/view/page"
	"github.com/boj/redistore/v2"
	"github.com/labstack/echo/v5"
)

const maxPasteSize = 1 * 1024 * 1024

type Handler struct {
	sessionStore *redistore.RediStore

	service *Service
}

func NewHandler(sessionStore *redistore.RediStore, service *Service) *Handler {
	return &Handler{sessionStore: sessionStore, service: service}
}

// Save
// @Summary      Сохранение новой пасты
// @Description  Сохранение новой пасты. Если пользователь авторизован через GitHub — паста привязывается к аккаунту.
// @Tags         Контроллер паст
// @Accept       plain
// @Produce      plain
// @Param        text  formData  string  true  "Текст пасты"
// @Success      302   {string}  string  "Редирект на /view/{id}"
// @Failure      500   {string}  string  "Внутренняя ошибка сервера"
// @Router       /save [post]
func (h *Handler) Save(c *echo.Context) error {
	ctx := c.Request().Context()
	text := c.FormValue("text")

	if text == "" {
		return &apphttp.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "Bad Request",
			Message:    "Empty text",
		}
	}

	if len(text) > maxPasteSize {
		return &apphttp.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "bad_request",
			Message:    "text too large",
		}
	}

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

			return c.Redirect(http.StatusFound, fmt.Sprintf("/view/%s", id))
		}
	}

	id, err := h.service.Save(ctx, text, "")
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/view/%s", id))
}

// View
// @Summary      Просмотр пасты
// @Description  Возвращает HTML-страницу с содержимым пасты
// @Tags         Контроллер паст
// @Produce      html
// @Param        id   path      string  true  "ID пасты"
// @Success      200  {string}  string  "HTML-страница с пастой"
// @Failure      302  {string}  string  "Редирект на / при отсутствии id или ошибке"
// @Router       /view/{id} [get]
func (h *Handler) View(c *echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.Redirect(http.StatusFound, "/")
	}

	code, err := h.service.LoadRaw(ctx, id)
	if err != nil {
		return c.Redirect(http.StatusFound, "/")
	}

	return apphttp.Render(c, http.StatusOK, page.Paste(code))
}

// Raw
// @Summary      Получение пасты в сыром виде
// @Description  Возвращает содержимое пасты в виде plain text
// @Tags         Контроллер паст
// @Produce      plain
// @Param        id   path      string  true  "ID пасты"
// @Success      200  {string}  string  "Паста в сыром виде"
// @Failure      400  {string}  string  "id обязателен"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /raw/{id} [get]
func (h *Handler) Raw(c *echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return &apphttp.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "bad_request",
			Message:    "id is required",
		}
	}

	code, err := h.service.LoadRaw(ctx, id)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, code)
}

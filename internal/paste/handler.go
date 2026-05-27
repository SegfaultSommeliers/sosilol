package paste

import (
	"fmt"
	"net/http"

	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/view/page"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

const maxPasteSize = 1 * 1024 * 1024

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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
func (h *Handler) Save(c fiber.Ctx) error {
	ctx := c.Context()
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

	sess := session.FromContext(c)
	accountType, _ := sess.Get("account_type").(string)
	accessToken, _ := sess.Get("access_token").(string)

	token := ""
	if accountType != "" && accessToken != "" {
		token = accessToken
	}

	id, err := h.service.Save(ctx, text, token)
	if err != nil {
		return err
	}

	return c.Redirect().To(fmt.Sprintf("/view/%s", id))
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
func (h *Handler) View(c fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	if id == "" {
		return c.Redirect().To("/")
	}

	code, err := h.service.LoadRaw(ctx, id)
	if err != nil {
		return c.Redirect().To("/")
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
func (h *Handler) Raw(c fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
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

	return c.SendString(code)
}

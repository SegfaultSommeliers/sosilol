package paste

import (
	"net/http"
	"regexp"

	apphttp "github.com/SegfaultSommeliers/sosilol/internal/http"
	"github.com/SegfaultSommeliers/sosilol/view/page"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

const maxPasteSize = 1 * 1024 * 1024

// validPasteID matches exactly the 7-character alphanumeric IDs produced by
// go-nanoid. Reject anything else before hitting Redis or PostgreSQL.
var validPasteID = regexp.MustCompile(`^[a-zA-Z0-9]{7}$`)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Save
// @Summary      Save a new paste
// @Description  Saves a paste. If the user is authenticated via GitHub, the paste is linked to their account.
// @Tags         pastes
// @Accept       json
// @Produce      json
// @Param        body  body      map[string]string  true  "Paste text"
// @Success      201   {object}  map[string]string  "JSON with the paste id"
// @Failure      400   {string}  string             "Invalid request body or empty text"
// @Failure      500   {string}  string             "Internal server error"
// @Router       /save [post]
func (h *Handler) Save(c fiber.Ctx) error {
	ctx := c.Context()

	var body struct {
		Text string `json:"text"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return &apphttp.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "bad_request",
			Message:    "invalid JSON body",
		}
	}

	if body.Text == "" {
		return &apphttp.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "bad_request",
			Message:    "text is required",
		}
	}

	if len(body.Text) > maxPasteSize {
		return &apphttp.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "bad_request",
			Message:    "text too large",
		}
	}

	sess := session.FromContext(c)
	accountType, _ := sess.Get("account_type").(string)
	userID, _ := sess.Get("github_user_id").(int64)

	// Pass userID=0 when the user is not authenticated; service treats 0 as anonymous.
	var authorID int64
	if accountType == "github" && userID != 0 {
		authorID = userID
	}

	id, err := h.service.Save(ctx, body.Text, authorID)
	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"id": id})
}

// View
// @Summary      View a paste
// @Description  Returns an HTML page with the paste content
// @Tags         pastes
// @Produce      html
// @Param        id   path      string  true  "Paste ID"
// @Success      200  {string}  string  "HTML page with the paste"
// @Failure      302  {string}  string  "Redirect to / if id is missing or not found"
// @Router       /view/{id} [get]
func (h *Handler) View(c fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	if id == "" || !validPasteID.MatchString(id) {
		return c.Redirect().To("/")
	}

	code, err := h.service.LoadRaw(ctx, id)
	if err != nil {
		return c.Redirect().To("/")
	}

	return apphttp.Render(c, http.StatusOK, page.Paste(code))
}

// Raw
// @Summary      Get raw paste
// @Description  Returns the paste content as plain text
// @Tags         pastes
// @Produce      plain
// @Param        id   path      string  true  "Paste ID"
// @Success      200  {string}  string  "Raw paste content"
// @Failure      400  {string}  string  "id is required"
// @Failure      500  {string}  string  "Internal server error"
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

	if !validPasteID.MatchString(id) {
		return &apphttp.AppError{
			StatusCode: http.StatusBadRequest,
			Code:       "bad_request",
			Message:    "invalid id",
		}
	}

	code, err := h.service.LoadRaw(ctx, id)
	if err != nil {
		return err
	}

	return c.SendString(code)
}

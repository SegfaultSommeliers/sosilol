package http

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
)

func Render(
	c fiber.Ctx,
	statusCode int,
	t templ.Component,
) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Status(statusCode)
	return t.Render(c.Context(), c.Response().BodyWriter())
}

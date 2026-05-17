package http

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v5"
)

func Render(
	c *echo.Context,
	statusCode int,
	t templ.Component,
) error {
	ctx := c.Request().Context()

	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx, buf); err != nil {
		return err
	}

	return c.HTML(statusCode, buf.String())
}

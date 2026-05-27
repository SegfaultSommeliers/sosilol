package health

import (
	"github.com/gofiber/fiber/v3"
)

func Handler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

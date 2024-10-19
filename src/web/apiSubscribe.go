package web

import "github.com/gofiber/fiber/v2"

func (a *API) subscribe(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) Unsubscribe(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

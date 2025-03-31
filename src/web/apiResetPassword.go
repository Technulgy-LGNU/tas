package web

import "github.com/gofiber/fiber/v2"

func (a *API) resetPassword(c *fiber.Ctx) error {

	return c.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (a *API) resetPasswordCode(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{})
}

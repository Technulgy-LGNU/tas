package web

import "github.com/gofiber/fiber/v2"

func (a *API) getForms(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) getForm(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) postForm(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) deleteForm(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

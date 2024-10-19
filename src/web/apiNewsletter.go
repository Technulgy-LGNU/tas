package web

import "github.com/gofiber/fiber/v2"

func (a *API) getNewsletters(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) getNewsletter(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) createNewsletter(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) updateNewsletter(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) deleteNewsletter(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) sendNewsletter(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

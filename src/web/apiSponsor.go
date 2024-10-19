package web

import "github.com/gofiber/fiber/v2"

func (a *API) getSponsors(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) getSponsor(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) createSponsor(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) updateSponsor(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) deleteSponsor(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

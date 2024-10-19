package web

import "github.com/gofiber/fiber/v2"

func (a *API) getEvents(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) getEvent(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) createEvent(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) updateEvent(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) deleteEvent(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

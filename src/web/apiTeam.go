package web

import "github.com/gofiber/fiber/v2"

func (a *API) getTeams(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) getTeam(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) createTeam(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) updateTeam(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) deleteTeam(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

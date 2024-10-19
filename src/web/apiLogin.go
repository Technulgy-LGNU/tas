package web

import "github.com/gofiber/fiber/v2"

func (a *API) login(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) logout(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func checkIfUserIsLoggedIn(key string) bool {
	return false
}

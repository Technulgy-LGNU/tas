package web

import "github.com/gofiber/fiber/v2"

func (a *API) getFiles(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) uploadFile(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) updateFile(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) downloadFile(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) deleteFile(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

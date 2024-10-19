package web

import "github.com/gofiber/fiber/v2"

func (a *API) getGifts(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) createGift(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) deleteGift(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

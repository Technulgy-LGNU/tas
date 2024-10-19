package web

import "github.com/gofiber/fiber/v2"

func (a *API) getOrderEntries(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) createOrderEntry(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) updateOrderEntry(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) deleteOrderEntry(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

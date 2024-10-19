package web

import "github.com/gofiber/fiber/v2"

func (a *API) getOrders(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) getOrder(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) createOrder(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) updateOrder(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) deleteOrder(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

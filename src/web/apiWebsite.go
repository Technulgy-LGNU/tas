package web

import "github.com/gofiber/fiber/v2"

func (a *API) getPosts(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) getPost(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) createPost(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) updatePost(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) deletePost(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

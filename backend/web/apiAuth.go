package web

import "github.com/gofiber/fiber/v3"

func (a *API) authLogin(c fiber.Ctx) error {
  return c.Status(fiber.StatusOK).JSON("")
}

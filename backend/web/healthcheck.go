package web

import "github.com/gofiber/fiber/v3"

func getHealthCheck(c fiber.Ctx) error {
  return c.Status(fiber.StatusOK).JSON("HT Ok")
}

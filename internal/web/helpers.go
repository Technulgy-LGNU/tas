package web

import "github.com/gofiber/fiber/v2"

func fail(status int, message string) error {
	return fiber.NewError(status, message)
}

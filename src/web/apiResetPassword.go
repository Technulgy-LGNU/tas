package web

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"tas/src/database"
	"tas/src/mail"
	"tas/src/util"
)

// Sends
func (a *API) resetPassword(c *fiber.Ctx) error {
	var (
		data = struct { // Incoming data
			Email string `json:"email"`
		}{}
		user  []database.User
		reset database.ResetPassword
	)
	// Parse body
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fmt.Sprintf("Error parsing incoming data: %v\n", err))
	}

	result := a.DB.Find(&user, data.Email)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fmt.Sprintf("Error finding user: %v\n", result.Error))
	} else if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotAcceptable).JSON("User not found")
	}

	reset.Email = data.Email
	reset.Code, _ = util.GenerateResetCode()
	a.DB.Create(&reset)

	err := mail.SendEmailPWDReset(data.Email, "T.A.S. Email Reset", reset.Code, a.CFG)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fmt.Sprintf("Error sending email: %v\n", err))
	}

	return c.Status(fiber.StatusOK).JSON("Email sent")
}

func (a *API) resetPasswordCode(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("")
}

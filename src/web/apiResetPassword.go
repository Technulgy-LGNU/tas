package web

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
	"strings"
	"tas/src/database"
	"tas/src/mail"
	"tas/src/util"
	"time"
)

func (a *API) resetPassword(c *fiber.Ctx) error {
	var (
		data = struct {
			Email string `json:"email"`
		}{}

		err error
	)
	// Parse the request body && Validate the data
	if err = c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid request body")
	}
	if data.Email == "" || !strings.Contains(data.Email, "@") {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid email")
	}

	// Check if the email exists in the database
	var user database.Member
	if err = a.DB.Where("email = ?", data.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON("Email not found")
		}
		return c.Status(fiber.StatusInternalServerError).JSON("Database error")
	}

	// Generate a reset password code
	code, err := util.GenerateResetCode()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("Failed to generate reset code")
	}
	// Save the reset code to the database
	var resetPassword = database.ResetPassword{
		UserId: user.ID,
		Code:   code,
		Email:  data.Email,
	}
	if err = a.DB.Create(&resetPassword).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("Failed to save reset code")
	}

	// Send the reset password email
	err = mail.SendEmailPWDReset(data.Email, code, a.CFG)
	if err != nil {
		log.Printf("Error sending email reset password to %s: %v\n", data.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON("Failed to send email")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (a *API) resetPasswordCode(c *fiber.Ctx) error {
	var (
		data = struct {
			Code     string `json:"code"`
			Password string `json:"password"`
		}{}

		err error
	)
	// Parse the request body && Validate the data
	if err = c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid request body")
	}
	if data.Code == "" || data.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid code or password")
	}

	// Check if the code exists in the database
	var resetPassword database.ResetPassword
	if err = a.DB.Where("code = ?", data.Code).First(&resetPassword).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON("Code not found")
		}
		return c.Status(fiber.StatusInternalServerError).JSON("Database error")
	}

	// Check if the code is still valid
	if resetPassword.CreatedAt.Add(10 * time.Minute).Before(time.Now()) {
		err = a.DB.Delete(&resetPassword).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON("Failed to delete expired code")
		}
		return c.Status(fiber.StatusBadRequest).JSON("Code expired")
	}

	// Update the user's password
	var user database.Member
	if err = a.DB.Where("id = ?", resetPassword.UserId).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON("Member not found")
		}
		return c.Status(fiber.StatusInternalServerError).JSON("Database error")
	}
	user.Password = util.HashString(data.Password)
	if err = a.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("Failed to update password")
	}
	// Delete the reset password code from the database
	err = a.DB.Delete(&resetPassword).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("Failed to delete reset code")
	}

	// Send a confirmation email to the user
	err = mail.SendEmailPWDResetSuccess(user.Email, a.CFG)
	if err != nil {
		log.Printf("Error sending email reset password to %s: %v\n", user.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON("Failed to send email")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{})
}

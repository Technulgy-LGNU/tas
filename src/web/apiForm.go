package web

import (
	discordwebhook "github.com/bensch777/discord-webhook-golang"
	"github.com/gofiber/fiber/v2"
	"log"
	"tas/src/database"
	"tas/src/mail"
	"tas/src/notifications"
	"tas/src/util"
)

func (a *API) getForms(c *fiber.Ctx) error {
	// Check if bearer token are present & valid
	if !util.CheckPermissions(c.GetReqHeaders(), 1, util.Forms, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("Unauthorized")
	}

	// Get the forms from the database
	var forms []database.Form
	if err := a.DB.Find(&forms).Error; err != nil {
		log.Printf("Error getting forms: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Error getting forms")
	}

	return c.Status(fiber.StatusOK).JSON(forms)
}

func (a *API) getForm(c *fiber.Ctx) error {
	// Check if bearer token are present & valid
	if !util.CheckPermissions(c.GetReqHeaders(), 1, util.Forms, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("Unauthorized")
	}

	// Get the form from the database
	var form database.Form
	if err := a.DB.Where("id = ?", c.Params("id")).First(&form).Error; err != nil {
		log.Printf("Error getting form: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Error getting form")
	}

	return c.Status(fiber.StatusOK).JSON(form)
}

func (a *API) postForm(c *fiber.Ctx) error {
	var (
		data = struct {
			Name    string `json:"name"`
			Email   string `json:"email"`
			Content string `json:"content"`
		}{}
	)
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	if data.Name == "" || data.Email == "" || data.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	// Store the form data in the database
	form := database.Form{
		Name:    data.Name,
		Email:   data.Email,
		Message: data.Content,
	}
	if err := a.DB.Create(&form).Error; err != nil {
		log.Printf("Error creating form: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Error creating form")
	}

	// Discord notification
	embed := discordwebhook.Embed{
		Title:       "Form Submission",
		Color:       0xFFA500,
		Description: "New form submitted",
	}
	notifications.SendDiscordEmbed(embed, a.CFG)
	// Mail notification
	go func() {
		err := mail.SendEmailForm("contact@technulgy.com", data.Name, data.Email, data.Content, a.CFG)
		if err != nil {
			log.Printf("Error sending email: %v\n", err)
		}
	}()

	return c.Status(fiber.StatusOK).JSON("Form submitted successfully")
}

func (a *API) deleteForm(c *fiber.Ctx) error {
	// Check if bearer token are present & valid
	if !util.CheckPermissions(c.GetReqHeaders(), 3, util.Forms, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("Unauthorized")
	}

	// Delete the form from the database
	var form database.Form
	if err := a.DB.Where("id = ?", c.Params("id")).Delete(&form).Error; err != nil {
		log.Printf("Error deleting form: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Error deleting form")
	}

	return c.Status(fiber.StatusOK).JSON("Form deleted successfully")
}

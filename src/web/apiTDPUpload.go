package web

import (
	"errors"
	"fmt"
	discordwebhook "github.com/bensch777/discord-webhook-golang"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
	"tas/src/database"
	"tas/src/notifications"
)

func (a *API) postTDPUpload(c *fiber.Ctx) error {
	var (
		data = struct {
			Team string `json:"team"`
			Year int    `json:"year"`
			URL  string `json:"url"`
		}{}

		err error
	)
	// Check for TDPUpload_Key
	headers := c.GetReqHeaders()
	if headers["Authorization"][0] != "Bearer "+a.CFG.TDPUploadKey {
		log.Println("Invalid TDPUpload_Key")
		return c.Status(fiber.StatusUnauthorized).JSON("Invalid Key")
	}

	// Parse the request body
	if err = c.BodyParser(&data); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON("Error parsing request body")
	}

	// Check if all values are not null
	if data.Team == "" || data.Year == 0 || data.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON("All fields are required")
	}

	// Check if there is already an entry for this team and year
	var existingTDP database.TDPList
	err = a.DB.Where("team = ? AND year = ?", data.Team, data.Year).First(&existingTDP).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error getting existing TDP: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Error getting existing TDP")
	}
	if err == nil {
		err = a.DB.Delete(&existingTDP).Error
		if err != nil {
			log.Printf("Error deleting existing TDP: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON("Error deleting existing TDP")
		}
	}

	// Enter the TDP URL into the database
	tpd := database.TDPList{
		Team: data.Team,
		Year: data.Year,
		URL:  data.URL,
	}
	result := a.DB.Create(&tpd)
	if result.Error != nil {
		log.Printf("Error creating TDP entry: %v", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("Error creating TDP entry")
	}

	// Send Discord notification
	embed := discordwebhook.Embed{
		Title:       "TDP Upload",
		Color:       0x00FF00,
		Description: fmt.Sprintf("TDP for %s has been successfully uploaded", data.Team),
		Fields: []discordwebhook.Field{
			{
				Name:   fmt.Sprintf("Year: %d", data.Year),
				Value:  fmt.Sprintf("You can find the TDP [here](%s)", data.URL),
				Inline: true,
			},
		},
		Footer: discordwebhook.Footer{
			Text: "Technulgy Admin Software",
		},
	}
	notifications.SendDiscordEmbed(embed, a.CFG)

	return c.Status(fiber.StatusOK).JSON("TDP url stored")
}

func (a *API) getTDPs(c *fiber.Ctx) error {
	type TDP struct {
		Team string `json:"team"`
		Year int    `json:"year"`
		URL  string `json:"url"`
	}

	var (
		tdpList []database.TDPList
		err     error
	)

	// Get all TDPs from the database
	result := a.DB.Find(&tdpList)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("Error getting TDP list: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Error getting TDP list")
	}

	// Check if there are no TDPs
	if len(tdpList) == 0 {
		return c.Status(fiber.StatusNotFound).JSON("No TDPs found")
	}

	// Convert TDPs to a slice of TDP structs
	var data []TDP
	for _, tdp := range tdpList {
		tdpJSON := TDP{
			Team: tdp.Team,
			Year: tdp.Year,
			URL:  tdp.URL,
		}
		data = append(data, tdpJSON)
	}

	// Return the TDPs as JSON
	return c.Status(fiber.StatusOK).JSON(data)
}

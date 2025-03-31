package web

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
	"tas/src/database"
	"tas/src/util"
)

func (a *API) login(c *fiber.Ctx) error {
	var (
		data = struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			DeviceId string `json:"deviceId"`
		}{}

		err error
	)
	// Parse body && validate
	if err = c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}
	if data.Email == "" || data.Password == "" || len(data.DeviceId) == 16 {
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}

	// Check if user exists && password is correct
	var user database.User
	if err = a.DB.Where("email = ?", data.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("invalid credentials")
	}
	if user.Password != data.Password {
		return c.Status(fiber.StatusBadRequest).JSON("invalid password")
	}

	// Check if user is already logged in
	var browserTokens []database.BrowserTokens
	err = a.DB.Where("user_id = ?", user.ID).Find(&browserTokens).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error: No browser tokens found for user: %v\n\n", user.ID)
		return c.Status(fiber.StatusInternalServerError).JSON("Error getting all browser tokens")
	}
	if len(browserTokens) > 0 {
		var found bool
		for _, token := range browserTokens {
			if token.DeviceId == data.DeviceId {
				fmt.Printf("User %v is already logged in on device %v\n", user.ID, data.DeviceId)
				fmt.Println("All tokens will be deleted")
				// Delete all tokens for this user
				found = true
			}
		}
		if found {
			err = a.DB.Where("user_id = ? AND device_id = ?", user.ID, data.DeviceId).Delete(&database.BrowserTokens{}).Error
			if err != nil {
				log.Printf("Error deleting browser tokens: %v\n", err)
				return c.Status(fiber.StatusInternalServerError).JSON("Error deleting old browser tokens")
			}
		}
	}

	// Create new browser token
	var token = util.GenerateSessionToken()
	if token == "" {
		return c.Status(fiber.StatusInternalServerError).JSON("Error generating token")
	}
	browserToken := database.BrowserTokens{
		DeviceId: data.DeviceId,
		Key:      token,
		User:     user,
		UserID:   user.ID,
	}
	err = a.DB.Create(&browserToken).Error
	if err != nil {
		log.Printf("Error creating browser token: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Error creating browser token")
	}

	// Get user permissions
	var perms database.Permission
	err = a.DB.Where("user_id = ?", user.ID).First(&perms).Error
	if err != nil {
		log.Printf("Error getting user permissions: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Error getting user permissions")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": token,
		"perms": fiber.Map{
			"login":      perms.Login,
			"admin":      perms.Admin,
			"members":    perms.Members,
			"teams":      perms.Teams,
			"events":     perms.Events,
			"newsletter": perms.Newsletter,
			"form":       perms.Form,
			"website":    perms.Website,
			"orders":     perms.Orders,
			"sponsors":   perms.Sponsors,
		},
	})
}

func (a *API) logout(c *fiber.Ctx) error {
	var (
		data = struct {
			DeviceId string `json:"deviceId"`
			Token    string `json:"token"`
		}{}

		err error
	)
	// Parse body && validate
	if err = c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}
	if len(data.DeviceId) == 16 || data.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}

	// Delete browser token
	err = a.DB.Where("device_id = ? AND key = ?", data.DeviceId, data.Token).Delete(&database.BrowserTokens{}).Error
	if err != nil {
		log.Printf("Error deleting browser token: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Error deleting browser token")
	}

	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) checkIfUserIsLoggedIn(c *fiber.Ctx) error {
	var (
		data = struct {
			DeviceId string `json:"deviceId"`
			Token    string `json:"token"`
		}{}

		err error
	)
	// Parse body && validate
	if err = c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}
	if len(data.DeviceId) == 16 || data.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}

	// Check if user is logged in
	var browserToken database.BrowserTokens
	err = a.DB.Where("device_id = ? AND key = ?", data.DeviceId, data.Token).First(&browserToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON("invalid credentials")
		}
		log.Printf("Error getting browser token: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Error getting browser token")
	}

	// Get user permissions
	var perms database.Permission
	err = a.DB.Where("user_id = ?", browserToken.UserID).First(&perms).Error
	if err != nil {
		log.Printf("Error getting user permissions: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Error getting user permissions")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"perms": fiber.Map{
			"login":      perms.Login,
			"admin":      perms.Admin,
			"members":    perms.Members,
			"teams":      perms.Teams,
			"events":     perms.Events,
			"newsletter": perms.Newsletter,
			"form":       perms.Form,
			"website":    perms.Website,
			"orders":     perms.Orders,
			"sponsors":   perms.Sponsors,
		},
	})
}

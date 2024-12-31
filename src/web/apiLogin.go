package web

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"tas/src/database"
	"tas/src/util"
)

func (a *API) login(c *fiber.Ctx) error {
	var (
		data = struct { // incoming data
			Email    string `json:"email"`
			Password string `json:"password"`
			Id       string `json:"id"`
		}{}
		returnData = struct { // Permissions and the session key will be returned to the client
			Perms struct {
				Login      *bool `json:"login"`
				Admin      *bool `json:"admin"`
				Members    *int  `json:"members"`
				Teams      *int  `json:"teams"`
				Events     *int  `json:"events"`
				Newsletter *int  `json:"newsletter"`
				Form       *int  `json:"form"`
				Website    *int  `json:"website"`
				Orders     *int  `json:"orders"`
			} `json:"perms"`
			Key string `json:"key"`
		}{}
		user  database.User
		perms database.Permission
	)
	// Parsing Body
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fmt.Sprintf("Error parsing request body: %v\n", err))
	}

	// Finding User
	result := a.DB.Find(&user).Where("email = ? AND password = ?", data.Email, data.Password)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fmt.Sprintf("Error finding user: %v\n", result.Error))
	} else if result.RowsAffected == 0 {
		return c.Status(fiber.StatusBadRequest).JSON("User not found")
	}

	// Creating a new session token and stores it in the UserKeys database, together with the client id
	sessionToken, err := util.GenerateSessionToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fmt.Sprintf("Error generating session token: %v\n", err))
	}
	userKey := database.UserKey{
		DeviceId: data.Id,
		Key:      sessionToken,
		UserID:   user.ID,
	}
	if err := a.DB.Create(&userKey).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fmt.Sprintf("Error creating user key: %v\n", err))
	}

	// Get permissions
	err = a.DB.Find(&perms).Where("user_id = ?", user.ID).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fmt.Sprintf("Error finding permissions: %v\n", err))
	}

	// Defines return data, and sends it to the client
	returnData.Key = sessionToken
	returnData.Perms.Login = perms.Login
	returnData.Perms.Admin = perms.Admin
	returnData.Perms.Members = perms.Members
	returnData.Perms.Teams = perms.Teams
	returnData.Perms.Events = perms.Events
	returnData.Perms.Newsletter = perms.Newsletter
	returnData.Perms.Form = perms.Form
	returnData.Perms.Website = perms.Website

	return c.Status(fiber.StatusOK).JSON(returnData)
}

func (a *API) logout(c *fiber.Ctx) error {
	var (
		data = struct { // Incoming Data
			ID string `json:"id"`
		}{}
		user []database.UserKey
	)
	// Parsing body
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fmt.Sprintf("Error parsing request body: %v\n", err))
	}

	// Find the devices
	err := a.DB.Find(&user).Where("user_id = ?", data.ID).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fmt.Sprintf("Error finding user: %v\n", err))
	}
	// Deletes every auth key associated with that device
	for _, u := range user {
		err = a.DB.Delete(&u).Error
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Printf("Error deleting user: %v\n", err)
		}
	}

	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) checkIfUserIsLoggedIn(c *fiber.Ctx) error {
	var (
		data = struct { // Incoming data
			ID  string `json:"id"`
			Key string `json:"key"`
		}{}
		user  database.UserKey
		perms database.Permission
	)
	// Parsing body
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fmt.Sprintf("Error parsing request body: %v\n", err))
	}
	// Finding the user
	result := a.DB.Find(&user).Where("device_id = ? AND  key = ?", data.ID, data.Key)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(result.Error)
	} else if result.RowsAffected == 0 { // If no user is found, the client gets the command to delete the stored session token and return to the login site
		return c.Status(fiber.StatusForbidden).JSON("User not found, instant logout")
	}

	// Finds the permissions of that user and sends them to the client
	err := a.DB.Find(&perms).Where("user_id = ?", user.UserID).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fmt.Sprintf("Error finding permissions: %v\n", err))
	}

	return c.Status(fiber.StatusOK).JSON(perms)
}

package web

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
	"tas/src/database"
	"tas/src/integrations"
	"tas/src/util"
)

func (a *API) getTeams(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 1, util.Teams, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	// Get all teams
	var teams []database.Team
	result := a.DB.Find(&teams)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("no teams found")
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON("error finding teams")
		}
	}

	return c.Status(fiber.StatusOK).JSON(teams)
}

func (a *API) getTeam(c *fiber.Ctx) error {
	var (
		data = struct {
			ID      uint64 `json:"Id"`
			Name    string `json:"Name"`
			League  string `json:"League"`
			Members []struct {
				ID   uint64 `json:"Id"`
				Name string `json:"Name"`
			}
		}{}
	)
	if !util.CheckPermissions(c.GetReqHeaders(), 1, util.Teams, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	// Get team by ID
	var team database.Team
	result := a.DB.First(&team, c.Params("id"))
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("team not found")
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON("error finding team")
		}
	}

	// Get team members
	var members []database.Member
	result = a.DB.Where("team_id = ?", team.ID).Find(&members)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON("error finding team members")
	}

	for _, member := range members {
		var memberData struct {
			ID   uint64 `json:"Id"`
			Name string `json:"Name"`
		}
		memberData.ID = member.ID
		memberData.Name = member.Name
		data.Members = append(data.Members, memberData)
	}

	// Prepare response data
	data.ID = team.ID
	data.Name = team.Name
	data.League = team.League

	return c.Status(fiber.StatusOK).JSON(data)
}

func (a *API) createTeam(c *fiber.Ctx) error {
	var (
		data = struct {
			Name            string `json:"name"`
			Email           string `json:"email"`
			League          string `json:"league"`
			Password        string `json:"password"`
			CreateMail      bool   `json:"createMail"`
			CreateNextCloud bool   `json:"createNextcloud"`
		}{}

		err error
	)
	if !util.CheckPermissions(c.GetReqHeaders(), 2, util.Teams, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}
	// Parse and validate the request body
	if err = c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("invalid request body")
	}
	if data.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON("name and league are required")
	}
	// Check if league is okay
	var leagues = []string{"Soccer Entry", "Soccer LightWeight Entry", "Soccer LightWeight int.", "Soccer Open int.",
		"Rescue Line Entry", "Rescue Line int.", "Rescue Maze Entry", "Rescue Maze int.",
		"Onstage Entry", "Onstage int."}
	var isLeagueValid = false
	for _, l := range leagues {
		if l == data.League {
			isLeagueValid = true
			break
		}
	}
	if !isLeagueValid {
		return c.Status(fiber.StatusBadRequest).JSON("league is not valid")
	}

	// Create a new team
	team := database.Team{
		Name:     data.Name,
		Email:    data.Email,
		League:   data.League,
		Password: data.Password,
	}
	result := a.DB.Create(&team)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("error creating team")
	}

	// Create team in mailcow and nextcloud
	if data.CreateMail {
		err = integrations.CreateMailcowUser(data.Name, data.Email, data.Password, a.CFG)
		if err != nil {
			log.Printf("error creating mailcow user: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON("error creating mailcow user")
		}
	}
	if data.CreateNextCloud {
		err = integrations.CreateNextCloudUser(data.Name, data.Email, data.Password, a.CFG)
		if err != nil {
			log.Printf("error creating nextcloud user: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON("error creating nextcloud user")
		}
	}

	return c.Status(fiber.StatusOK).JSON(team)
}

func (a *API) updateTeam(c *fiber.Ctx) error {
	var (
		data = struct {
			Name   string `json:"name"`
			Email  string `json:"email"`
			League string `json:"league"`
		}{}

		err error
	)
	if !util.CheckPermissions(c.GetReqHeaders(), 2, util.Teams, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}
	// Parse and validate the request body
	if err = c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("invalid request body")
	}
	if data.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON("invalid request body")
	}
	// Check if league is okay
	var leagues = []string{"Soccer Entry", "Soccer LightWeight Entry", "Soccer LightWeight int.", "Soccer Open int.",
		"Rescue Line Entry", "Rescue Line int.", "Rescue Maze Entry", "Rescue Maze int.",
		"Onstage Entry", "Onstage int."}
	var isLeagueValid = false
	for _, l := range leagues {
		if l == data.League {
			isLeagueValid = true
			break
		}
	}
	if !isLeagueValid {
		return c.Status(fiber.StatusBadRequest).JSON("league is not valid")
	}

	// Get team by ID
	var team database.Team
	result := a.DB.First(&team, c.Params("id"))
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("team not found")
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON("error finding team")
		}
	}

	// Update team
	team.Name = data.Name
	team.League = data.League

	result = a.DB.Save(&team)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("error updating team")
	}

	return c.Status(fiber.StatusOK).JSON(team)
}

func (a *API) deleteTeam(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 3, util.Teams, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	// Delete Team by ID
	var team database.Team
	result := a.DB.Where("id = ?", c.Params("id")).Delete(&team)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("team not found")
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON("error deleting team")
		}
	}

	return c.Status(fiber.StatusOK).JSON("")
}

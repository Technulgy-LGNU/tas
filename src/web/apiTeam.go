package web

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"tas/src/database"
	"tas/src/util"
)

func (a *API) getTeams(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 1, "teams", a.DB) {
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
	if !util.CheckPermissions(c.GetReqHeaders(), 1, "teams", a.DB) {
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

	return c.Status(fiber.StatusOK).JSON(team)
}

func (a *API) createTeam(c *fiber.Ctx) error {
	var (
		data = struct {
			Name   string `json:"name"`
			League string `json:"league"`
		}{}

		err error
	)
	if !util.CheckPermissions(c.GetReqHeaders(), 2, "teams", a.DB) {
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
		Name:   data.Name,
		League: data.League,
	}
	result := a.DB.Create(&team)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("error creating team")
	}

	return c.Status(fiber.StatusOK).JSON(team)
}

func (a *API) updateTeam(c *fiber.Ctx) error {
	var (
		data = struct {
			Name   string `json:"name"`
			League string `json:"league"`
		}{}

		err error
	)
	if !util.CheckPermissions(c.GetReqHeaders(), 2, "teams", a.DB) {
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
	if !util.CheckPermissions(c.GetReqHeaders(), 3, "teams", a.DB) {
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

package web

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
	"tas/src/database"
	"tas/src/mail"
	"tas/src/util"
	"time"
)

func (a *API) getEvents(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 1, util.Events, a.DB) {
		return c.Status(fiber.StatusForbidden).JSON("")
	}

	// Get all events
	var events []database.Event
	result := a.DB.Find(&events)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON("")
	}

	return c.Status(fiber.StatusOK).JSON(events)
}

func (a *API) getEvent(c *fiber.Ctx) error {
	var (
		data = struct {
			Name            string    `json:"Name"`
			Location        string    `json:"Location"`
			StartDate       time.Time `json:"StartDate"`
			EndDate         time.Time `json:"EndDate"`
			RegisteredTeams []struct {
				Id     uint64 `json:"Id"`
				Name   string `json:"Name"`
				League string `json:"League"`
			}
		}{}
	)
	if !util.CheckPermissions(c.GetReqHeaders(), 1, util.Events, a.DB) {
		return c.Status(fiber.StatusForbidden).JSON("")
	}

	// Get event
	var event database.Event
	result := a.DB.Where("id = ?", c.Params("id")).First(&event)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("Event not found")
		}
		return c.Status(fiber.StatusInternalServerError).JSON("Error fetching event")
	}
	data.Name = event.Name
	data.Location = event.Location
	data.StartDate = event.StartDate
	data.EndDate = event.EndDate

	// Get registered teams
	var teams []database.Team
	result = a.DB.Where("event_id = ?", event.ID).Find(&teams)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON("Error fetching teams")
	}
	for _, team := range teams {
		data.RegisteredTeams = append(data.RegisteredTeams, struct {
			Id     uint64 `json:"Id"`
			Name   string `json:"Name"`
			League string `json:"League"`
		}{
			Id:     team.ID,
			Name:   team.Name,
			League: team.League,
		})
	}

	return c.Status(fiber.StatusOK).JSON(data)
}

func (a *API) createEvent(c *fiber.Ctx) error {
	var (
		data = struct {
			Name      string    `json:"Name"`
			Location  string    `json:"Location"`
			StartDate time.Time `json:"StartDate"`
			EndDate   time.Time `json:"EndDate"`
		}{}
	)
	if !util.CheckPermissions(c.GetReqHeaders(), 2, util.Events, a.DB) {
		return c.Status(fiber.StatusForbidden).JSON("")
	}
	// Parse & validate request body
	if err := c.BodyParser(&data); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON("Invalid request")
	}
	if data.Name == "" || data.Location == "" {
		log.Printf("Invalid request data: %v", data)
		return c.Status(fiber.StatusBadRequest).JSON("Invalid request")
	}

	// Create event
	event := database.Event{
		Name:      data.Name,
		Location:  data.Location,
		StartDate: data.StartDate,
		EndDate:   data.EndDate,
	}
	result := a.DB.Create(&event)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("Error creating event")
	}

	return c.Status(fiber.StatusOK).JSON(event)
}

func (a *API) addTeamToEvent(c *fiber.Ctx) error {
	var (
		data = struct {
			TeamId uint64 `json:"teamId"`
		}{}
	)
	if !util.CheckPermissions(c.GetReqHeaders(), 2, util.Events, a.DB) {
		return c.Status(fiber.StatusForbidden).JSON("")
	}
	// Parse & validate request body
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid request")
	}
	if data.TeamId == 0 {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid request")
	}

	// Get event
	var event database.Event
	result := a.DB.Where("id = ?", c.Params("id")).First(&event)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("Event not found")
		}
		return c.Status(fiber.StatusInternalServerError).JSON("Error fetching event")
	}

	// Get team
	var team database.Team
	result = a.DB.Where("id = ?", data.TeamId).First(&team)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("Team not found")
		}
		return c.Status(fiber.StatusInternalServerError).JSON("Error fetching team")
	}
	// Add team to event
	event.RegisteredTeams = append(event.RegisteredTeams, team)
	result = a.DB.Save(&event)

	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) updateEvent(c *fiber.Ctx) error {
	var (
		data = struct {
			Name      string    `json:"Name"`
			Location  string    `json:"Location"`
			StartDate time.Time `json:"StartDate"`
			EndDate   time.Time `json:"EndDate"`
		}{}
	)
	if !util.CheckPermissions(c.GetReqHeaders(), 2, util.Events, a.DB) {
		return c.Status(fiber.StatusForbidden).JSON("")
	}

	// Parse request body
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid request")
	}

	// Get event
	var event database.Event
	result := a.DB.Where("id = ?", c.Params("id")).First(&event)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("Event not found")
		}
		return c.Status(fiber.StatusInternalServerError).JSON("Error fetching event")
	}

	// Update event
	event.Name = data.Name
	event.Location = data.Location
	event.StartDate = data.StartDate
	event.EndDate = data.EndDate

	result = a.DB.Save(&event)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("Error updating event")
	}

	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) deleteEvent(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 3, util.Events, a.DB) {
		return c.Status(fiber.StatusForbidden).JSON("")
	}

	// Delete event
	var event database.Event
	result := a.DB.Where("id = ?", c.Params("id")).Delete(&event)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("Event not found")
		}
		return c.Status(fiber.StatusInternalServerError).JSON("Error deleting event")
	}

	return c.Status(fiber.StatusOK).JSON("")
}

func (a *API) sendEventNotification(c *fiber.Ctx) error {
	var data = struct {
		Title   string `json:"Title"`
		Message string `json:"Message"`
	}{}
	if !util.CheckPermissions(c.GetReqHeaders(), 2, util.Events, a.DB) {
		return c.Status(fiber.StatusForbidden).JSON("")
	}
	// Parse & validate request body
	if err := c.BodyParser(&data); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON("Invalid request")
	}
	if data.Title == "" || data.Message == "" {
		log.Printf("Invalid request data: %v", data)
		return c.Status(fiber.StatusBadRequest).JSON("Invalid request")
	}

	// Get all teams
	var teams []database.Team
	result := a.DB.Find(&teams).Where("event_id = ?", c.Params("id"))
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON("Error fetching teams")
	}

	// Get all members of these teams
	var members []database.Member
	result = a.DB.Find(&members).Where("team_id IN ?", teams)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON("Error fetching members")
	}

	// Send notification to all teams & members
	var emailList []string
	for _, member := range members {
		emailList = append(emailList, member.Email)
	}
	for _, team := range teams {
		emailList = append(emailList, team.Email)
	}

	// Send emails
	for _, email := range emailList {
		err := mail.SendSystemEmail(email, data.Title, data.Message, a.CFG)
		if err != nil {
			log.Printf("Error sending email to %s: %v", email, err)
		}
	}

	return c.Status(fiber.StatusOK).JSON("")
}

package web

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
	"tas/src/database"
	"tas/src/util"
	"time"
)

// -> All members
func (a *API) getMembers(c *fiber.Ctx) error {
	type DataToSend struct {
		Id          uint64 `json:"Id"`
		Name        string `json:"Name"`
		Email       string `json:"Email"`
		Gender      string `json:"Gender"`
		Birthday    string `json:"´Birthday"`
		TeamID      uint64 `json:"TeamId"`
		Permissions struct {
			Login      bool `json:"Login"`
			Admin      bool `json:"Admin"`
			Members    int  `json:"Members"`
			Teams      int  `json:"Teams"`
			Events     int  `json:"Events"`
			Newsletter int  `json:"Newsletter"`
			Form       int  `json:"Form"`
			Website    int  `json:"Website"`
			Orders     int  `json:"Orders"`
			Sponsors   int  `json:"Sponsors"`
		} `json:"permissions"`
	}
	if !util.CheckPermissions(c.GetReqHeaders(), 1, "members", a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	// Get all members
	var members []database.Member
	result := a.DB.Find(&members)
	if result.Error != nil {
		log.Printf("Error getting all members from the database: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("")
	}

	// Get all members permissions
	var perms []database.Permission
	result = a.DB.Find(&perms)
	if result.Error != nil {
		log.Printf("Error getting all members permissions from the database: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("Error getting all members permissions from the database")
	}

	// Send the data in a nice format
	var membersWithPerms []DataToSend
	for _, member := range members {
		// Add permissions to each user
		for _, perm := range perms {
			if member.ID == perm.MemberID {
				var dataToSend DataToSend
				dataToSend.Id = member.ID
				dataToSend.Name = member.Name
				dataToSend.Email = member.Email
				dataToSend.Gender = member.Gender
				dataToSend.Birthday = member.Birthday.Format("02-04-2006")
				dataToSend.TeamID = member.TeamID
				dataToSend.Permissions.Login = perm.Login
				dataToSend.Permissions.Admin = perm.Admin
				dataToSend.Permissions.Members = perm.Members
				dataToSend.Permissions.Teams = perm.Teams
				dataToSend.Permissions.Events = perm.Events
				dataToSend.Permissions.Newsletter = perm.Newsletter
				dataToSend.Permissions.Form = perm.Form
				dataToSend.Permissions.Website = perm.Website
				dataToSend.Permissions.Orders = perm.Orders
				dataToSend.Permissions.Sponsors = perm.Sponsors
				membersWithPerms = append(membersWithPerms, dataToSend)
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(membersWithPerms)
}

// -> Member by ID
func (a *API) getMember(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 1, "members", a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	// Get member by ID
	var member database.Member
	result := a.DB.Where("id = ?", c.Params("id")).First(&member)
	if result.Error != nil {
		log.Printf("Error getting member from the database: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("")
	}

	// Get members permissions
	var perms database.Permission
	result = a.DB.Where("member_id = ?", member.ID).First(&perms)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON("no permissions found for this user")
		}
		log.Printf("Error checking if permissions exist: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("Error checking if permissions exist")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"Id":       member.ID,
		"Name":     member.Name,
		"Email":    member.Email,
		"Gender":   member.Gender,
		"Birthday": member.Birthday.Format("02-04-2006"),
		"TeamId":   member.TeamID,
		"Permissions": fiber.Map{
			"Login":      perms.Login,
			"Admin":      perms.Admin,
			"Members":    perms.Members,
			"Teams":      perms.Teams,
			"Events":     perms.Events,
			"Newsletter": perms.Newsletter,
			"Form":       perms.Form,
			"Website":    perms.Website,
			"Orders":     perms.Orders,
			"Sponsors":   perms.Sponsors,
		},
	})
}

// <- Create member
func (a *API) createMember(c *fiber.Ctx) error {
	var (
		data = struct {
			Name        string `json:"Name"`
			Email       string `json:"Email"`
			Password    string `json:"Password"`
			Gender      string `json:"Gender"`
			Birthday    string `json:"Birthday"`
			TeamID      uint64 `json:"Team_id"`
			Permissions struct {
				Login      bool `json:"Login"`
				Members    int  `json:"Members"`
				Teams      int  `json:"Teams"`
				Events     int  `json:"Events"`
				Newsletter int  `json:"Newsletter"`
				Form       int  `json:"Form"`
				Website    int  `json:"Website"`
				Orders     int  `json:"Orders"`
				Sponsors   int  `json:"Sponsors"`
			} `json:"permissions"`
		}{}

		err error
	)
	if !util.CheckPermissions(c.GetReqHeaders(), 2, "members", a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}
	// Parse & Validate the body
	if err = c.BodyParser(&data); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}
	if data.Name == "" || data.Email == "" || data.Gender != "male" && data.Gender != "female" && data.Gender != "divers" || data.Birthday == "" {
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}

	// Check if user with this email already exists
	var existingMember database.Member
	result := a.DB.Where("email = ?", data.Email).First(&existingMember)
	if result.Error == nil && result.RowsAffected != 0 {
		return c.Status(fiber.StatusBadRequest).JSON("user with this email already exists")
	} else if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("Error checking if user with this email already exists: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("Error checking if user with this email already exists")
	}

	// Reformat the birthday to the correct format (Incoming: dd-mm-yyyy, Database: time.Time)
	var formatedDate time.Time
	formatedDate, err = time.Parse("2006-04-02", data.Birthday)
	if err != nil {
		log.Printf("Error parsing birthday: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}
	// Create the new member
	newMember := database.Member{
		Name:     data.Name,
		Email:    data.Email,
		Password: data.Password,
		Gender:   data.Gender,
		Birthday: formatedDate,
		TeamID:   data.TeamID,
		Perms: &database.Permission{
			Login:      data.Permissions.Login,
			Admin:      false,
			Members:    data.Permissions.Members,
			Teams:      data.Permissions.Teams,
			Events:     data.Permissions.Events,
			Newsletter: data.Permissions.Newsletter,
			Form:       data.Permissions.Form,
			Website:    data.Permissions.Website,
			Orders:     data.Permissions.Orders,
			Sponsors:   data.Permissions.Sponsors,
		},
	}
	result = a.DB.Create(&newMember)
	if result.Error != nil {
		log.Printf("Error creating new member: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("")
	}

	return c.Status(fiber.StatusOK).JSON("")
}

// <- Updates member
func (a *API) updateMember(c *fiber.Ctx) error {
	var (
		data = struct {
			Name        string `json:"Name"`
			Email       string `json:"Email"`
			Gender      string `json:"Gender"`
			Birthday    string `json:"Birthday"`
			TeamID      uint64 `json:"TeamId"`
			Permissions struct {
				Login      bool `json:"Login"`
				Members    int  `json:"Members"`
				Teams      int  `json:"Teams"`
				Events     int  `json:"Events"`
				Newsletter int  `json:"Newsletter"`
				Form       int  `json:"Form"`
				Website    int  `json:"Website"`
				Orders     int  `json:"Orders"`
				Sponsors   int  `json:"Sponsors"`
			} `json:"Permissions"`
		}{}

		err error
	)
	if !util.CheckPermissions(c.GetReqHeaders(), 2, "members", a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}
	// Parse & Validate the body
	if err = c.BodyParser(&data); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}
	if data.Name == "" || data.Email == "" || data.Gender == "" || data.Birthday == "" {
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}

	// Check if user with given id exists
	var existingMember database.Member
	result := a.DB.Where("id = ?", c.Params("id")).First(&existingMember)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON("no user found with this id")
		}
		log.Printf("Error checking if user with this id exists: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("Error checking if user with this id exists")
	}

	// Reformat the birthday to the correct format (Incoming: dd-mm-yyyy, Database: time.Time)
	var formatedDate time.Time
	formatedDate, err = time.Parse("2006-04-02", data.Birthday)
	if err != nil {
		log.Printf("Error parsing birthday: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}

	// Update the member
	if data.Name != existingMember.Name {
		existingMember.Name = data.Name
	}
	if data.Email != existingMember.Email {
		existingMember.Email = data.Email
	}
	if data.Gender != existingMember.Gender {
		existingMember.Gender = data.Gender
	}
	if formatedDate != existingMember.Birthday {
		existingMember.Birthday = formatedDate
	}
	if data.TeamID != existingMember.TeamID {
		existingMember.TeamID = data.TeamID
	}

	result = a.DB.Save(&existingMember)
	if result.Error != nil {
		log.Printf("Error updating member: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("Error updating member")
	}

	// Check if the permissions have changed
	var perms database.Permission
	result = a.DB.Where("member_id = ?", existingMember.ID).First(&perms)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON("no permissions found for this user")
		}
		log.Printf("Error checking if permissions exist: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("Error checking if permissions exist")
	}
	if data.Permissions.Login != perms.Login {
		perms.Login = data.Permissions.Login
	}
	if data.Permissions.Members != perms.Members {
		perms.Members = data.Permissions.Members
	}
	if data.Permissions.Teams != perms.Teams {
		perms.Teams = data.Permissions.Teams
	}
	if data.Permissions.Events != perms.Events {
		perms.Events = data.Permissions.Events
	}
	if data.Permissions.Newsletter != perms.Newsletter {
		perms.Newsletter = data.Permissions.Newsletter
	}
	if data.Permissions.Form != perms.Form {
		perms.Form = data.Permissions.Form
	}
	if data.Permissions.Website != perms.Website {
		perms.Website = data.Permissions.Website
	}
	if data.Permissions.Orders != perms.Orders {
		perms.Orders = data.Permissions.Orders
	}
	if data.Permissions.Sponsors != perms.Sponsors {
		perms.Sponsors = data.Permissions.Sponsors
	}

	result = a.DB.Save(&perms)
	if result.Error != nil {
		log.Printf("Error updating permissions: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("Error updating permissions")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"Name":     existingMember.Name,
		"Email":    existingMember.Email,
		"Gender":   existingMember.Gender,
		"Birthday": existingMember.Birthday.Format("02-04-2006"),
		"TeamID":   existingMember.TeamID,
		"Permissions": fiber.Map{
			"Login":      perms.Login,
			"Admin":      perms.Admin,
			"Members":    perms.Members,
			"Teams":      perms.Teams,
			"Events":     perms.Events,
			"Newsletter": perms.Newsletter,
			"Form":       perms.Form,
			"Website":    perms.Website,
			"Orders":     perms.Orders,
			"Sponsors":   perms.Sponsors,
		},
	})
}

// <- Deletes member by id
func (a *API) deleteMember(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 3, "members", a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}
	// Check if user with given id exists
	var existingMember database.Member
	result := a.DB.Where("id = ?", c.Params("id")).First(&existingMember)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON("no user found with this id")
		}
		log.Printf("Error checking if user with this id exists: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("Error checking if user with this id exists")
	}

	// Delete the member
	result = a.DB.Delete(&existingMember)
	if result.Error != nil {
		log.Printf("Error deleting member: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("")
	}

	// Delete the permissions
	var perms database.Permission
	result = a.DB.Where("member_id = ?", c.Params("id")).Delete(&perms)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("Error checking if permissions exist: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("Error checking if permissions exist")
	}

	// Delete all browser tokens
	var tokens []database.BrowserToken
	result = a.DB.Where("member_id = ?", c.Params("id")).Delete(&tokens)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("Error checking if tokens exist: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON("Error checking if tokens exist")
	}

	return c.Status(fiber.StatusOK).JSON("")
}

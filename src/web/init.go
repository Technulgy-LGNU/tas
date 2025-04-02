package web

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
	"log"
	"strings"
	"tas/src/config"
)

type API struct {
	DB      *gorm.DB
	CFG     *config.CFG
	Clients map[*websocket.Conn]bool
}

func InitWeb(cfg *config.CFG, db *gorm.DB) {
	var (
		addrTASBackend = "0.0.0.0:3001"
		addrTASLinks   = "0.0.0.0:3002"

		err error

		// Backend app
		tasBackend = fiber.New(fiber.Config{
			ServerHeader: "tas_backend:fiber",
			AppName:      "tas_backend",
		})

		// Links app
		links = fiber.New(fiber.Config{
			ServerHeader: "tas_links:fiber",
			AppName:      "tas_links",
		})

		c = cors.New(cors.Config{
			AllowOrigins: strings.Join([]string{
				"*",
			}, ","),

			AllowMethods: strings.Join([]string{
				fiber.MethodGet,
				fiber.MethodPost,
				fiber.MethodPatch,
				fiber.MethodDelete,
			}, ","),

			AllowHeaders: strings.Join([]string{
				"application/json",
			}, ","),

			AllowCredentials: false,
		})

		// Monitor
		mon = monitor.New(monitor.Config{
			Title: "TAS Monitor",
		})
	)

	// Internal tools
	// TAS-Backend
	tasBackend.Use(c)                                          // Cors middleware
	tasBackend.Use(healthcheck.New(healthcheck.ConfigDefault)) // Healthcheck
	tasBackend.Use("/ws", func(c *fiber.Ctx) error {           // Websocket middleware
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	tasBackend.Get("/monitor", mon)                // Monitor
	tasBackend.Get("/healthcheck", getHealthcheck) // Healthcheck

	// TAS-Links
	links.Use(c)
	links.Use(healthcheck.New(healthcheck.ConfigDefault))
	links.Get("/monitor", mon)

	// API
	api := fiber.New()
	tasBackend.Mount("/api", api)
	a := API{
		DB:      db,
		CFG:     cfg,
		Clients: make(map[*websocket.Conn]bool),
	}
	// Websocket
	api.Get("/ws", websocket.New(a.WebsocketConnection))
	// Login / Password reset
	api.Post("/login", a.login)                         // <- Email&Password, returns new session token
	api.Delete("/logout", a.logout)                     // <- Token, deletes session
	api.Post("/checkLogin", a.checkIfUserIsLoggedIn)    // -> Bool&Perms, checks if the session is valid and returns the users permissions
	api.Post("/resetPassword", a.resetPassword)         // <- Email, checks if email exists, if yes, sends an email with a code to reset your password and a link to the specific site
	api.Post("/resetPasswordCode", a.resetPasswordCode) // <- Code&NewPassword, checks if the code is still valid, if yes, changes the password to the one provided and returns 200
	// Members
	api.Get("/getMembers", a.getMembers)            // -> Members, returns all members
	api.Get("/getMember/:id", a.getMember)          // -> Member, returns the member with the given id
	api.Post("/createMember", a.createMember)       // <- Member, creates a new member and returns the new member
	api.Patch("/updateMember/:id", a.updateMember)  // <- Member, updates the member with the given id and returns the updated member
	api.Delete("/deleteMember/:id", a.deleteMember) // <- Member, deletes the member with the given id and returns 200
	// Teams
	api.Get("/getTeams", a.getTeams)            // -> Teams, returns all teams
	api.Get("/getTeam/:id", a.getTeam)          // -> Team, returns the team with the given id
	api.Post("/createTeam", a.createTeam)       // <- Team, creates a new team and returns the new team
	api.Patch("/updateTeam/:id", a.updateTeam)  // <- Team, updates the team with the given id and returns the updated team
	api.Delete("/deleteTeam/:id", a.deleteTeam) // <- Team, deletes the team with the given id and returns 200
	// Website
	api.Post("/tdpUpload", a.postTDPUpload)     // <- TDP Upload, returns 200 if successful
	api.Get("/getTDPs", a.getTDPs)              // -> TDPs, returns all TDPs
	api.Post("/newForm", a.postForm)            // <- Form, returns 200 if successful
	api.Get("/getForms", a.getForms)            // -> Forms, returns all forms
	api.Get("/getForm/:id", a.getForm)          // -> Form, returns the form with the given id
	api.Delete("/deleteForm/:id", a.deleteForm) // <- Form, deletes a specific form, returns 200 if successful
	// Newsletter

	// Orders

	// Events
	api.Get("/getEvents", a.getEvents)                // -> Events, returns all events
	api.Get("/getEvent/:id", a.getEvent)              // -> Event, returns the event with the given id
	api.Post("/createEvent", a.createEvent)           // <- Event, creates a new event and returns the new event
	api.Patch("/updateEvent/:id", a.updateEvent)      // <- Event, updates the event with the given id and returns the updated event
	api.Delete("/deleteEvent/:id", a.deleteEvent)     // <- Event, deletes the event with the given id and returns 200
	api.Post("/addTeamToEvent/:id", a.addTeamToEvent) // <- Event, adds a team to the event with the given id and returns 200
	// Sponsors

	// TAS-Web
	tasBackend.Static("/", "./web/dist")

	// Start TAS-Backend
	go func() {
		log.Println("Started T.A.S. Backend V1")
		err = tasBackend.Listen(addrTASBackend)
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Fatalf("Error starting webserver: %v\n", err)
		}
	}()

	// Start TAS-Links
	log.Println("Started T.A.S. Links V1")
	err = links.Listen(addrTASLinks)
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Fatalf("Error starting webserver: %v\n", err)
	}
}

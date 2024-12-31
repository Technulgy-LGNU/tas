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
	cLog "tas/src/log"
)

type API struct {
	Logger  *cLog.Logger
	DB      *gorm.DB
	CFG     *config.CFG
	Clients map[*websocket.Conn]bool
}

func InitWeb(logger *cLog.FiberLogger, mainLog *cLog.Logger, cfg *config.CFG, db *gorm.DB) error {
	var (
		addrTASBackend = "0.0.0.0:3001" // Tas Frontend is on 3002 and Website on 3000
		addrTASCDN     = "0.0.0.0:3003"
		addrTASLinks   = "0.0.0.0:3004"

		err error

		// Backend app
		tasBackend = fiber.New(fiber.Config{
			ServerHeader: "tas_backend:fiber",
			AppName:      "tas_backend",
		})

		// CDN app
		cdn = fiber.New(fiber.Config{
			ServerHeader: "tas_cdn:fiber",
			AppName:      "tas_cdn",
		})

		// Links app
		links = fiber.New(fiber.Config{
			ServerHeader: "tas_links:fiber",
			AppName:      "tas_links",
		})

		c = cors.New(cors.Config{
			AllowOrigins: strings.Join([]string{
				"https://technulgy.com",
				"https://tas.technulgy.com",
				"https://cdn.technulgy.com",
				"https://links.technulgy.com",
				"http://localhost:3000",
				"http://localhost:3001",
				"http://localhost:3002",
				"http://localhost:3003",
				"http://localhost:3004",
			}, ","),

			AllowMethods: strings.Join([]string{
				fiber.MethodGet,
				fiber.MethodPost,
				fiber.MethodPatch,
				fiber.MethodDelete,
			}, ","),

			AllowHeaders: strings.Join([]string{
				"application/json",
				"text/html",
				"text/plain",
				"text/css",
				"text/javascript",
				"application/javascript",
				"application/x-javascript",
				"application/ecmascript",
				"application/x-ecmascript",
			}, ","),

			AllowCredentials: true,
		})

		// Monitor
		mon = monitor.New(monitor.Config{
			Title: "TAS Monitor",
		})
	)

	// Internal tools
	// TAS-Backend
	tasBackend.Use(c)                                          // Cors middleware
	tasBackend.Use(logger.FiberLoggerMiddleware())             // Logger   // Logger
	tasBackend.Use(healthcheck.New(healthcheck.ConfigDefault)) // Healthcheck
	tasBackend.Use("/ws", func(c *fiber.Ctx) error {           // Websocket middleware
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	tasBackend.Get("/monitor", mon) // Monitor

	// TAS-CDN
	cdn.Use(c)
	cdn.Use(logger.FiberLoggerMiddleware())
	cdn.Use(healthcheck.New(healthcheck.ConfigDefault))
	cdn.Get("/monitor", mon)
	// Technulgy Admin Software: Website

	// TAS-Links
	links.Use(c)
	links.Use(logger.FiberLoggerMiddleware())
	links.Use(healthcheck.New(healthcheck.ConfigDefault))
	links.Get("/monitor", mon)

	// API
	api := fiber.New()
	tasBackend.Mount("/api", api)
	a := API{
		Logger:  mainLog,
		DB:      db,
		CFG:     cfg,
		Clients: make(map[*websocket.Conn]bool),
	}
	// Websocket
	api.Get("/ws", websocket.New(a.WebsocketConnection))
	// Login / Password reset
	api.Post("/login", a.login)                          // <- Email&Password, returns new session token
	api.Delete("/logout", a.logout)                      // <- Token, deletes session
	api.Post("/checkLogin", a.checkIfUserIsLoggedIn)     // -> Bool&Perms, checks if the session is valid and returns the users permissions
	api.Post("/resetPassword", a.resetPassword)          // <- Email, checks if email exists, if yes, sends an email with a code to reset your password and a link to the specific site
	api.Post("/resetPassword/code", a.resetPasswordCode) // <- Code&NewPassword, checks if the code is still valid, if yes, changes the password to the one provided and returns 200
	// Users

	// Website

	// Newsletter

	// Orders

	// Events

	// Sponsors

	// Start TAS-Backend
	go func() {
		log.Println("Started T.A.S. Backend V1")
		err = tasBackend.Listen(addrTASBackend)
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Fatalf("Error starting webserver: %v\n", err)
		}
	}()

	// Start TAS-CDN
	go func() {
		log.Println("Started T.A.S. CDN V1")
		err = cdn.Listen(addrTASCDN)
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Fatalf("Error starting webserver: %v\n", err)
		}
	}()

	// Start TAS-Links
	go func() {
		log.Println("Started T.A.S. Links V1")
		err = links.Listen(addrTASLinks)
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Fatalf("Error starting webserver: %v\n", err)
		}
	}()

	return nil
}

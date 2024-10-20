package web

import (
	"errors"
	"fmt"
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
	DB      *gorm.DB
	Clients map[*websocket.Conn]bool
}

func InitWeb(logger *cLog.FiberCustomLogger, cfg *config.CFG, db *gorm.DB) error {
	var (
		addr = fmt.Sprintf("%s:%d", cfg.Website.Host, cfg.Website.Port)

		err error

		// Fiber App
		app = fiber.New(fiber.Config{
			ServerHeader: "tas:fiber",
			AppName:      "tas",
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
			Title: "ShowMaster Monitor",
		})
	)

	// Internal tools
	app.Use(c)                                          // Cors middleware
	app.Use(logger.FiberLoggerMiddleware())             // Logger
	app.Use(healthcheck.New(healthcheck.ConfigDefault)) // Healthcheck
	app.Use("/ws", func(c *fiber.Ctx) error {           // Websocket middleware
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/monitor", mon) // Monitor

	// API
	api := fiber.New()
	app.Mount("/api", api)
	a := API{
		DB:      db,
		Clients: make(map[*websocket.Conn]bool),
	}
	// Websocket
	api.Get("/ws", websocket.New(a.WebsocketConnection))
	// Login
	api.Post("/login", a.login)                      // <- Email&Password, returns new session token
	api.Delete("/logout", a.logout)                  // <- Token, deletes session
	api.Get("/check-login", a.checkIfUserIsLoggedIn) // -> Bool&Perms, checks if the session is valid and returns the users permissions
	// Users

	// Website

	// Newsletter

	// Orders

	// Events

	// Sponsors

	// Frontend
	app.Static("/", cfg.Website.Files)
	// CDN
	// app.Static("/cdn", "./data/cdn")

	// Start fiber
	log.Println("Started T.A.S. V1")
	err = app.Listen(addr)
	if err != nil {
		return errors.New(fmt.Sprintf("error starting web server %s\n", err))
	}
	return nil
}

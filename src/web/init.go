package web

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
	"log"
	"strings"
	"tas/src/config"
)

type API struct {
	DB      *gorm.DB
	Clients map[*websocket.Conn]bool
}

var clients = make(map[*websocket.Conn]bool)

func InitWeb(cfg *config.CFG, db *gorm.DB) {
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

		logs = logger.New(logger.Config{
			Next:     noLog,
			Format:   "[${ip}]:${port} ${status} - ${method} ${path}\n",
			TimeZone: "Europe/Berlin",
		})

		// Monitor
		mon = monitor.New(monitor.Config{
			Title: "ShowMaster Monitor",
		})
	)

	// Internal tools
	app.Use(c)                                          // Cors middleware
	app.Use(logs)                                       // Logger
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
		Clients: clients,
	}
	// Websocket
	api.Get("/ws", websocket.New(a.WebsocketConnection))
	// Users

	// Website

	// Newsletter

	// Orders

	// Events

	// Sponsors

	// Frontend
	app.Static("/", "./public")
	// CDN
	app.Static("/cdn", "./var/lib/tas/data/cdn")

	// Start fiber
	log.Println("Started ShowMaster V3")
	err = app.Listen(addr)
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Fatalf("Web init error: %d\n", err)
	}
}

package web

import (
	"crypto/rsa"
	"fmt"
	"log"
	"strings"
	"sync"
	"tas/internal/config"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

type API struct {
	DB          *gorm.DB
	CFG         *config.Config
	Clients     map[*websocket.Conn]bool
	jwksMu      sync.Mutex
	jwksKeys    map[string]*rsa.PublicKey
	jwksFetched time.Time
}

func InitWeb(cfg *config.Config, db *gorm.DB) {
	var (
		addr = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

		err error

		tasApp = fiber.New(fiber.Config{
			ServerHeader: "tas:fiber",
			AppName:      "tas",
		})

		tasLinks = fiber.New(fiber.Config{
			DisableStartupMessage: true,
			ServerHeader:          "tasLinks:fiber",
			AppName:               "tas",
		})

		c = cors.New(cors.Config{
			AllowOrigins: strings.Join([]string{
				"tas.technulgy.com",
				"links.technulgy.com",
				"technulgy.com",
			}, ","),

			AllowMethods: strings.Join([]string{
				fiber.MethodGet,
				fiber.MethodPost,
				fiber.MethodPatch,
				fiber.MethodDelete,
				fiber.MethodOptions,
			}, ","),

			AllowHeaders: strings.Join([]string{
				"Content-Type",
				"Accept",
				"Origin",
				"Authorization",
			}, ","),

			AllowCredentials: true,
			MaxAge:           86400,
		})
	)
	// Internal
	tasApp.Use(c) // Cors Middleware
	tasApp.Use(healthcheck.New(healthcheck.ConfigDefault))
	tasLinks.Use(c)
	tasLinks.Use(healthcheck.New(healthcheck.ConfigDefault))

	// API
	api := fiber.New()
	tasApp.Mount("/api", api)
	a := API{
		DB:      db,
		CFG:     cfg,
		Clients: make(map[*websocket.Conn]bool),
	}
	// API
	api.Get("/healthcheck", a.getHealthcheck)
	api.Get("/auth/config", a.authConfig)

	// Links

	// Static
	tasApp.Static("/", "./frontend/dist")
	tasApp.Get("*", func(c *fiber.Ctx) error {
		return c.SendFile("./frontend/dist/index.html")
	})

	go func() {
		err := tasLinks.Listen(addr)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Starting RCJV Paperless")
	err = tasApp.Listen(addr)
	if err != nil {
		log.Fatal(err)
	}
}

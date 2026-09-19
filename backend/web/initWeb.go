package web

import (
	"fmt"
	"log"
	"tas/backend/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/static"
	"gorm.io/gorm"
)

type API struct {
  CFG *config.Config
  DB *gorm.DB
}

func InitWeb(cfg *config.Config, db *gorm.DB) {
  var (
    addr = fmt.Sprintf("%s:%d", "0.0.0.0", 2005)
    a = API{
      CFG: cfg,
      DB: db,
    }

    tas = fiber.New(fiber.Config{
      ServerHeader: "tas:fiber",
      AppName: "TAS",
    })

    // Cors
    c = cors.New(cors.Config{
      AllowOrigins: []string{
        "*",
      },

			AllowHeaders: []string{
				"Origin",
				"Content-Type",
				"Accept",
			},

			AllowMethods: []string{
				fiber.MethodGet,
				fiber.MethodPost,
				fiber.MethodDelete,
			},

			AllowCredentials: false,
    })
  )

  tas.Use(c)
  tas.Get("/healthcheck", getHealthCheck)

  api := tas.Group("/api")
  v1 := api.Group("/v1")
  v1.Get("/auth/login", a.authLogin)

  // Site
  tas.Get("/", static.New("frontend/dist"))

  log.Fatal(tas.Listen(addr))
}

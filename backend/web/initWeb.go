package web

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"gorm.io/gorm"
	"log"
	"tas/backend/cloudflare"
	"tas/backend/config"
	"tas/backend/database"
)

type API struct {
	CFG           *config.Config
	DB            *gorm.DB
	Auth          *Auth
	Images        database.ImageStore
	ImageProvider imageProvider
}

func NewApp(cfg *config.Config, db *gorm.DB) (*fiber.App, error) {
	auth, err := NewAuth(cfg.Auth)
	if err != nil {
		return nil, err
	}
	a := &API{CFG: cfg, DB: db, Auth: auth, ImageProvider: cloudflare.NewImages(cfg.Cloudflare.ImagesAccountId, cfg.Cloudflare.ImagesAPIToken)}
	if db != nil {
		a.Images = database.PostgresImages{DB: db}
	}
	app := fiber.New(fiber.Config{ServerHeader: "tas:fiber", AppName: "TAS", Immutable: true, BodyLimit: maxImageBytes + 1024*1024})
	app.Use(func(c fiber.Ctx) error {
		c.Set("Referrer-Policy", "no-referrer")
		c.Set("X-Content-Type-Options", "nosniff")
		if len(c.Path()) >= 5 && (c.Path()[:5] == "/auth" || c.Path()[:5] == "/api/") {
			c.Set("Cache-Control", "no-store")
		}
		return c.Next()
	})
	app.Get("/healthcheck", getHealthCheck)
	app.Get("/auth/login", auth.Login)
	app.Get("/auth/callback", auth.Callback)

	v1 := app.Group("/api/v1", auth.CSRF)
	v1.Get("/auth/login", auth.Login)
	v1.Post("/auth/logout", auth.Logout)
	v1.Use(a.Auth.RequireAuth)
	v1.Get("/auth/me", auth.Me)
	// Register all future private API routes here, after RequireAuth.
	a.registerImages(v1)
	app.Use("/api", func(c fiber.Ctx) error { return c.Status(404).JSON(fiber.Map{"error": "not_found"}) })
	app.Use("/auth", func(c fiber.Ctx) error { return c.SendStatus(404) })
	app.Use("/", static.New("frontend/dist"))
	return app, nil
}

func InitWeb(cfg *config.Config, db *gorm.DB) {
	app, err := NewApp(cfg, db)
	if err != nil {
		log.Fatalf("Invalid authentication configuration: %v", err)
	}
	if cfg.Auth.DisableFusionAuth {
		log.Print("Local development mode: FusionAuth disabled; requests run as local admin. Listening on loopback only.")
	}
	log.Fatal(app.Listen(listenAddress(cfg)))
}

func listenAddress(cfg *config.Config) string {
	if cfg.Auth.DisableFusionAuth {
		return "127.0.0.1:2005"
	}
	return "0.0.0.0:2005"
}

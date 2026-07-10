package web

import (
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"tas/internal/config"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"gorm.io/gorm"
)

type API struct {
	DB          *gorm.DB
	CFG         *config.Config
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

		c = cors.New(cors.Config{
			AllowOrigins: strings.Join(cfg.CORS.Origins, ","),
			AllowMethods: strings.Join([]string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete, fiber.MethodOptions}, ","),
			AllowHeaders: strings.Join([]string{"Content-Type", "Accept", "Origin", "Authorization"}, ","),

			AllowCredentials: true,
			MaxAge:           86400,
		})
	)
	// Internal
	tasApp.Use(c) // Cors Middleware
	tasApp.Use(healthcheck.New(healthcheck.ConfigDefault))

	// API
	api := fiber.New()
	tasApp.Mount("/api", api)
	a := API{
		DB:  db,
		CFG: cfg,
	}
	a.registerRoutes(api)

	// Static
	if _, err := os.Stat("./frontend/dist"); err == nil {
		tasApp.Static("/", "./frontend/dist")
	}
	tasApp.Get("*", func(c *fiber.Ctx) error {
		return c.SendFile("./frontend/dist/index.html")
	})

	log.Printf("Starting T.A.S. on %s", addr)
	err = tasApp.Listen(addr)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *API) registerRoutes(api fiber.Router) {
	api.Get("/healthcheck", a.getHealthcheck)
	api.Get("/auth/config", a.authConfig)
	api.Get("/me", a.requireAuthenticated, a.getMe)

	api.Get("/roles", a.requirePermission("members:view", "members:manage"), a.listRoles)
	api.Get("/members", a.requirePermission("members:view", "members:manage"), a.listMembers)
	api.Patch("/members/:id", a.requirePermission("members:manage"), a.updateMember)

	api.Get("/inventory/categories", a.requirePermission("inventory:view"), a.listInventoryCategories)
	api.Post("/inventory/categories", a.requirePermission("inventory:edit", "inventory:manage"), a.createInventoryCategory)
	api.Patch("/inventory/categories/:id", a.requirePermission("inventory:edit", "inventory:manage"), a.updateInventoryCategory)
	api.Delete("/inventory/categories/:id", a.requirePermission("inventory:manage"), a.deleteInventoryCategory)
	api.Get("/inventory/items", a.requirePermission("inventory:view"), a.listInventoryItems)
	api.Post("/inventory/items", a.requirePermission("inventory:edit", "inventory:manage"), a.createInventoryItem)
	api.Patch("/inventory/items/:id", a.requirePermission("inventory:edit", "inventory:manage"), a.updateInventoryItem)
	api.Delete("/inventory/items/:id", a.requirePermission("inventory:manage"), a.deleteInventoryItem)
	api.Post("/inventory/items/:id/reorder", a.requirePermission("orders:request", "orders:manage"), a.reorderInventoryItem)

	api.Get("/orders/requests", a.requirePermission("orders:view"), a.listOrderRequests)
	api.Post("/orders/requests", a.requirePermission("orders:request", "orders:manage"), a.createOrderRequest)
	api.Patch("/orders/requests/:id", a.requirePermission("orders:edit", "orders:manage"), a.updateOrderRequest)
	api.Delete("/orders/requests/:id", a.requirePermission("orders:manage"), a.deleteOrderRequest)
	api.Post("/orders/requests/:id/approve", a.requirePermission("orders:manage"), a.approveOrderRequest)
	api.Post("/orders/requests/:id/reject", a.requirePermission("orders:manage"), a.rejectOrderRequest)
	api.Get("/orders/lists", a.requirePermission("orders:view"), a.listOrderLists)
	api.Post("/orders/lists", a.requirePermission("orders:edit", "orders:manage"), a.createOrderList)
	api.Get("/orders/lists/:id", a.requirePermission("orders:view"), a.getOrderList)
	api.Patch("/orders/lists/:id", a.requirePermission("orders:edit", "orders:manage"), a.updateOrderList)
	api.Delete("/orders/lists/:id", a.requirePermission("orders:manage"), a.deleteOrderList)
	api.Post("/orders/lists/:id/publish", a.requirePermission("orders:manage"), a.publishOrderList)
	api.Post("/orders/lists/:id/items", a.requirePermission("orders:edit", "orders:manage"), a.createOrderListItem)
	api.Patch("/orders/lists/:id/items/:item_id", a.requirePermission("orders:edit", "orders:manage"), a.updateOrderListItem)
	api.Post("/orders/lists/:id/items/:item_id/ordered", a.requirePermission("orders:manage"), a.markOrderListItemOrdered)
	api.Get("/orders/lists/:id/items/:item_id/matches", a.requirePermission("orders:manage"), a.matchOrderListItem)
	api.Post("/orders/lists/:id/items/:item_id/receive", a.requirePermission("orders:manage"), a.receiveOrderListItem)

	api.Post("/images", a.requirePermission("website:edit", "website:manage"), a.uploadImage)
	api.Get("/images", a.requirePermission("website:view", "website:edit", "website:manage"), a.listImages)

	api.Get("/website/teams", a.requirePermission("website:view"), a.listTeams)
	api.Post("/website/teams", a.requirePermission("website:edit", "website:manage"), a.createTeam)
	api.Patch("/website/teams/:id", a.requirePermission("website:edit", "website:manage"), a.updateTeam)
	api.Delete("/website/teams/:id", a.requirePermission("website:manage"), a.deleteTeam)
	api.Get("/website/competitions", a.requirePermission("website:view"), a.listCompetitions)
	api.Post("/website/competitions", a.requirePermission("website:edit", "website:manage"), a.createCompetition)
	api.Patch("/website/competitions/:id", a.requirePermission("website:edit", "website:manage"), a.updateCompetition)
	api.Delete("/website/competitions/:id", a.requirePermission("website:manage"), a.deleteCompetition)
	api.Get("/website/prizes", a.requirePermission("website:view"), a.listPrizes)
	api.Post("/website/prizes", a.requirePermission("website:edit", "website:manage"), a.createPrize)
	api.Patch("/website/prizes/:id", a.requirePermission("website:edit", "website:manage"), a.updatePrize)
	api.Delete("/website/prizes/:id", a.requirePermission("website:manage"), a.deletePrize)
	api.Get("/website/sponsor-categories", a.requirePermission("website:view"), a.listSponsorCategories)
	api.Post("/website/sponsor-categories", a.requirePermission("website:edit", "website:manage"), a.createSponsorCategory)
	api.Patch("/website/sponsor-categories/:id", a.requirePermission("website:edit", "website:manage"), a.updateSponsorCategory)
	api.Delete("/website/sponsor-categories/:id", a.requirePermission("website:manage"), a.deleteSponsorCategory)
	api.Get("/website/sponsors", a.requirePermission("website:view"), a.listSponsors)
	api.Post("/website/sponsors", a.requirePermission("website:edit", "website:manage"), a.createSponsor)
	api.Patch("/website/sponsors/:id", a.requirePermission("website:edit", "website:manage"), a.updateSponsor)
	api.Delete("/website/sponsors/:id", a.requirePermission("website:manage"), a.deleteSponsor)
	api.Get("/website/home", a.requirePermission("website:view"), a.listHomeArticles)
	api.Put("/website/home/:slot", a.requirePermission("website:edit", "website:manage"), a.upsertHomeArticle)

	api.Get("/public/teams", a.publicTeams)
	api.Get("/public/participation-history", a.publicParticipationHistory)
	api.Get("/public/sponsors", a.publicSponsors)
	api.Get("/public/home", a.publicHome)
}

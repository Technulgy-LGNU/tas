package web

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"tas/backend/cloudflare"
	"tas/backend/contact"
	"tas/backend/database"
)

// Serializes content/reference mutations with image deletion across TAS processes.
func websiteWrite(db *gorm.DB, fn func(*gorm.DB) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(736281905)").Error; err != nil {
			return err
		}
		return fn(tx)
	})
}
func (a *API) registerWebsite(app *fiber.App, v1 fiber.Router) {
	admin := v1.Group("/website", func(c fiber.Ctx) error {
		if a.DB == nil {
			return c.Status(503).JSON(fiber.Map{"error": "Database unavailable."})
		}
		return c.Next()
	})
	admin.Get("/settings", func(c fiber.Ctx) error { return c.JSON(fiber.Map{"contactEnabled": a.CFG.Website.Contact.Ready()}) })
	admin.Get("/:kind", a.adminWebsiteList)
	admin.Post("/:kind", RequireRoles("editor", "admin"), a.saveWebsiteEntry)
	admin.Put("/:kind/:id", RequireRoles("editor", "admin"), a.saveWebsiteEntry)
	admin.Delete("/:kind/:id", RequireRoles("admin"), a.deleteWebsiteEntry)

	public := app.Group("/website")
	// Public content has no session cookies. CORS applies only to /website, never /api.
	origins := a.CFG.Website.AllowedOrigins
	if len(origins) > 0 {
		public.Use(cors.New(cors.Config{AllowOrigins: origins, AllowMethods: []string{"GET", "POST", "OPTIONS"}, AllowHeaders: []string{"Content-Type"}}))
	}
	public.Use(func(c fiber.Ctx) error {
		c.Set("Cache-Control", "no-store")
		if c.Get("Origin") != "" && len(origins) > 0 {
			allowed := false
			for _, o := range origins {
				if o == c.Get("Origin") {
					allowed = true
				}
			}
			if !allowed {
				return c.Status(403).JSON(fiber.Map{"error": "Origin is not allowed."})
			}
		}
		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}
		if a.DB == nil {
			return c.Status(503).JSON(fiber.Map{"error": "Website content is unavailable."})
		}
		lang, err := websiteLanguage(c.Query("lang"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		c.Locals("language", lang)
		c.Set("Content-Language", lang)
		return c.Next()
	})
	public.Get("/home", a.publicHome)
	public.Post("/contact", limiter.New(limiter.Config{Max: 5, Expiration: time.Hour, LimitReached: func(c fiber.Ctx) error {
		return c.Status(429).JSON(fiber.Map{"error": "Too many messages. Please try again later."})
	}}), a.websiteContact)
	public.Get("/participation-history", a.publicWebsiteList)
	public.Get("/:kind/:slug", a.publicWebsiteDetail)
	public.Get("/:kind", a.publicWebsiteList)
	public.Use(func(c fiber.Ctx) error { return c.Status(404).JSON(fiber.Map{"error": "Website endpoint not found."}) })
}
func websiteError(c fiber.Ctx, err error) error {
	var apiErr *fiber.Error
	if errors.As(err, &apiErr) {
		return c.Status(apiErr.Code).JSON(fiber.Map{"error": apiErr.Message})
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(404).JSON(fiber.Map{"error": "Content not found."})
	}
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) && pgerr.Code == "23505" {
		return c.Status(409).JSON(fiber.Map{"error": "This URL slug is already in use in this section."})
	}
	log.Printf("Website database operation failed: %v", err)
	return c.Status(500).JSON(fiber.Map{"error": "Website content could not be saved or loaded."})
}
func (a *API) adminWebsiteList(c fiber.Ctx) error {
	kind := c.Params("kind")
	if !websiteKinds[kind] {
		return c.SendStatus(404)
	}
	entries := []database.WebsiteEntry{}
	if err := a.DB.WithContext(c.Context()).Where("kind = ?", kind).Order("sort_order ASC, created_at DESC").Find(&entries).Error; err != nil {
		return websiteError(c, err)
	}
	return c.JSON(fiber.Map{"entries": entries})
}
func (a *API) saveWebsiteEntry(c fiber.Ctx) error {
	kind := c.Params("kind")
	if !websiteKinds[kind] {
		return c.SendStatus(404)
	}
	if len(c.Body()) > 2*1024*1024 {
		return c.Status(413).JSON(fiber.Map{"error": "Content exceeds 2 MB."})
	}
	var entry database.WebsiteEntry
	if err := json.Unmarshal(c.Body(), &entry); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON content."})
	}
	entry.Kind = kind
	if err := validateWebsiteEntry(&entry); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	creating := c.Method() == "POST"
	if creating {
		entry.ID = uuid.NewString()
		entry.Version = 1
		entry.CreatedAt = time.Now()
	} else {
		id, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return c.SendStatus(400)
		}
		entry.ID = id.String()
	}
	err := websiteWrite(a.DB.WithContext(c.Context()), func(tx *gorm.DB) error {
		references := []database.WebsiteReference{}
		seen := map[string]bool{}
		for _, image := range entry.Content.AllImages() {
			if seen[image.ID] {
				continue
			}
			seen[image.ID] = true
			var count int64
			if err := tx.Model(&database.Image{}).Where("id = ? AND status = ?", image.ID, "ready").Count(&count).Error; err != nil {
				return err
			}
			if count != 1 {
				return fiber.NewError(400, "A selected image is missing or its upload is incomplete.")
			}
			references = append(references, database.WebsiteReference{EntryID: entry.ID, TargetID: image.ID, Kind: "image"})
		}
		seen = map[string]bool{}
		for _, award := range entry.Content.Awards {
			if seen[award.EventID] {
				continue
			}
			seen[award.EventID] = true
			var event database.WebsiteEntry
			if err := tx.First(&event, "id = ? AND kind = ?", award.EventID, "events").Error; err != nil {
				return fiber.NewError(400, "A selected event no longer exists.")
			}
			if entry.Published && !event.Published {
				return fiber.NewError(400, "Publish the selected events before publishing this team's results.")
			}
			references = append(references, database.WebsiteReference{EntryID: entry.ID, TargetID: award.EventID, Kind: "event"})
		}
		if creating {
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		} else {
			var old database.WebsiteEntry
			if err := tx.First(&old, "id = ? AND kind = ?", entry.ID, kind).Error; err != nil {
				return err
			}
			if old.Version != entry.Version {
				return fiber.NewError(409, "Someone changed this content. Reload it before saving your changes.")
			}
			entry.Version++
			entry.CreatedAt = old.CreatedAt
			if err := tx.Save(&entry).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("entry_id = ?", entry.ID).Delete(&database.WebsiteReference{}).Error; err != nil {
			return err
		}
		if len(references) > 0 {
			return tx.Create(&references).Error
		}
		return nil
	})
	if err != nil {
		return websiteError(c, err)
	}
	status := 200
	if creating {
		status = 201
	}
	return c.Status(status).JSON(fiber.Map{"entry": entry})
}
func (a *API) deleteWebsiteEntry(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.SendStatus(400)
	}
	err = websiteWrite(a.DB.WithContext(c.Context()), func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&database.WebsiteReference{}).Where("target_id = ? AND kind = ?", id.String(), "event").Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fiber.NewError(409, "This event is used in team results. Remove those results first.")
		}
		var entry database.WebsiteEntry
		if err := tx.First(&entry, "id = ? AND kind = ?", id.String(), c.Params("kind")).Error; err != nil {
			return err
		}
		if err := tx.Where("entry_id = ?", entry.ID).Delete(&database.WebsiteReference{}).Error; err != nil {
			return err
		}
		return tx.Delete(&entry).Error
	})
	if err != nil {
		return websiteError(c, err)
	}
	return c.SendStatus(204)
}
func publishedQuery(db *gorm.DB) *gorm.DB {
	return db.Where("published = ? AND (publish_at IS NULL OR publish_at <= ?)", true, time.Now())
}
func (a *API) websiteURL(path, lang string) string {
	return strings.TrimRight(a.CFG.Website.PublicURL, "/") + path + "?lang=" + lang
}
func (a *API) websiteImage(image database.Image, ref database.WebsiteImage, lang string) fiber.Map {
	cfg := a.CFG.Cloudflare
	return fiber.Map{"id": image.ID, "url": cloudflare.DeliveryURL(cfg.ImagesDeliveryURL, image.CloudflareID, cfg.ImagesTransformOrigin, 1920), "alt": ref.Alt.Get(lang)}
}

type websiteProjection struct {
	a      *API
	lang   string
	images map[string]database.Image
	events map[string]database.WebsiteEntry
	teams  []database.WebsiteEntry
}

func (a *API) projection(c fiber.Ctx, entries []database.WebsiteEntry) (*websiteProjection, error) {
	p := &websiteProjection{a: a, lang: c.Locals("language").(string), images: map[string]database.Image{}, events: map[string]database.WebsiteEntry{}}
	// Only published events/teams may be joined into public responses.
	related := []database.WebsiteEntry{}
	if err := publishedQuery(a.DB.WithContext(c.Context())).Where("kind IN ?", []string{"events", "teams"}).Find(&related).Error; err != nil {
		return nil, err
	}
	for _, e := range related {
		if e.Kind == "events" {
			p.events[e.ID] = e
		} else {
			p.teams = append(p.teams, e)
		}
	}
	ids := []string{}
	for _, e := range entries {
		for _, im := range e.Content.AllImages() {
			ids = append(ids, im.ID)
		}
	}
	if len(ids) > 0 {
		images := []database.Image{}
		if err := a.DB.WithContext(c.Context()).Where("id IN ? AND status = ?", ids, "ready").Find(&images).Error; err != nil {
			return nil, err
		}
		for _, im := range images {
			p.images[im.ID] = im
		}
	}
	return p, nil
}
func (p *websiteProjection) imageList(refs []database.WebsiteImage) []fiber.Map {
	images := []fiber.Map{}
	for _, ref := range refs {
		if image, ok := p.images[ref.ID]; ok {
			images = append(images, p.a.websiteImage(image, ref, p.lang))
		}
	}
	return images
}
func (p *websiteProjection) entry(e database.WebsiteEntry, detail bool) fiber.Map {
	c := e.Content
	result := fiber.Map{"id": e.ID, "slug": e.Slug, "name": c.Name.Get(p.lang), "description": c.Description.Get(p.lang), "images": p.imageList(c.Images)}
	result["image"] = nil
	for _, image := range c.Images {
		if image.ID == c.CoverImageID {
			if im, ok := p.images[image.ID]; ok {
				result["image"] = p.a.websiteImage(im, image, p.lang)
			}
		}
	}
	switch e.Kind {
	case "teams":
		result["status"] = c.TeamStatus
		result["startImage"] = result["image"]
		awards := []fiber.Map{}
		for _, award := range c.Awards {
			if event, ok := p.events[award.EventID]; ok {
				awards = append(awards, fiber.Map{"eventId": event.ID, "event": event.Content.Name.Get(p.lang), "date": event.Content.Date, "year": event.Content.Date[:4], "league": award.League.Get(p.lang), "result": award.Result.Get(p.lang)})
			}
		}
		result["awards"] = awards
	case "events":
		result["date"] = c.Date
		result["year"] = c.Date[:4]
		teams := []fiber.Map{}
		for _, team := range p.teams {
			for _, award := range team.Content.Awards {
				if award.EventID == e.ID {
					teams = append(teams, fiber.Map{"teamId": team.ID, "team": team.Content.Name.Get(p.lang), "teamSlug": team.Slug, "league": award.League.Get(p.lang), "result": award.Result.Get(p.lang)})
				}
			}
		}
		result["results"] = teams
	case "sponsors", "publications":
		result["url"] = c.URL
	case "blog":
		result["publishedAt"] = e.PublishAt
		result["url"] = p.a.websiteURL("/blog/"+e.Slug, p.lang)
		if detail {
			blocks := []fiber.Map{}
			for _, b := range c.Blocks {
				blocks = append(blocks, fiber.Map{"id": b.ID, "type": b.Type, "text": b.Text.Get(p.lang), "level": b.Level, "images": p.imageList(b.Images), "url": b.URL})
			}
			result["blocks"] = blocks
			result["textFormat"] = "markdown"
		}
	}
	return result
}
func (a *API) publicHome(c fiber.Ctx) error {
	var home database.WebsiteEntry
	err := publishedQuery(a.DB.WithContext(c.Context())).First(&home, "kind = ?", "home").Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(404).JSON(fiber.Map{"error": "The home page has not been published yet."})
	}
	if err != nil {
		return websiteError(c, err)
	}
	blogs := []database.WebsiteEntry{}
	if err := publishedQuery(a.DB.WithContext(c.Context())).Where("kind = ?", "blog").Order("publish_at DESC, id DESC").Limit(3).Find(&blogs).Error; err != nil {
		return websiteError(c, err)
	}
	p, err := a.projection(c, append(blogs, home))
	if err != nil {
		return websiteError(c, err)
	}
	articles := []fiber.Map{}
	for _, blog := range blogs {
		articles = append(articles, p.entry(blog, false))
	}
	videos := []fiber.Map{}
	for _, v := range home.Content.Videos {
		videos = append(videos, fiber.Map{"title": v.Title.Get(p.lang), "url": v.URL})
	}
	return c.JSON(fiber.Map{"language": p.lang, "images": p.imageList(home.Content.Images), "aboutUs": home.Content.About.Get(p.lang), "blogs": articles, "videos": videos, "contact": fiber.Map{"endpoint": "/website/contact?lang=" + p.lang, "enabled": a.CFG.Website.Contact.Ready()}})
}
func (a *API) publicWebsiteList(c fiber.Ctx) error {
	kind := c.Params("kind")
	if c.Path() == "/website/participation-history" {
		kind = "events"
	}
	if !websiteKinds[kind] || kind == "home" {
		return c.SendStatus(404)
	}
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 || page > 100000 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid page."})
	}
	query := publishedQuery(a.DB.WithContext(c.Context())).Model(&database.WebsiteEntry{}).Where("kind = ?", kind)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return websiteError(c, err)
	}
	order := "sort_order ASC, created_at DESC"
	if kind == "blog" {
		order = "publish_at DESC, id DESC"
	}
	if kind == "events" {
		order = "content->>'date' DESC, id DESC"
	}
	entries := []database.WebsiteEntry{}
	if err := query.Order(order).Limit(50).Offset((page - 1) * 50).Find(&entries).Error; err != nil {
		return websiteError(c, err)
	}
	p, err := a.projection(c, entries)
	if err != nil {
		return websiteError(c, err)
	}
	items := []fiber.Map{}
	for _, e := range entries {
		items = append(items, p.entry(e, false))
	}
	return c.JSON(fiber.Map{"language": p.lang, "items": items, "page": page, "pageSize": 50, "total": total})
}
func (a *API) publicWebsiteDetail(c fiber.Ctx) error {
	kind := c.Params("kind")
	if !websiteKinds[kind] || kind == "home" {
		return c.SendStatus(404)
	}
	var entry database.WebsiteEntry
	if err := publishedQuery(a.DB.WithContext(c.Context())).First(&entry, "kind = ? AND slug = ?", kind, c.Params("slug")).Error; err != nil {
		return websiteError(c, err)
	}
	p, err := a.projection(c, []database.WebsiteEntry{entry})
	if err != nil {
		return websiteError(c, err)
	}
	result := p.entry(entry, true)
	result["language"] = p.lang
	return c.JSON(result)
}
func (a *API) websiteContact(c fiber.Ctx) error {
	if len(c.Body()) > 20000 {
		return c.Status(413).JSON(fiber.Map{"error": "Message is too large."})
	}
	if !strings.HasPrefix(c.Get("Content-Type"), "application/json") {
		return c.Status(415).JSON(fiber.Map{"error": "Send the form as application/json."})
	}
	var input struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Subject string `json:"subject"`
		Message string `json:"message"`
		Website string `json:"website"`
	}
	if json.Unmarshal(c.Body(), &input) != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid contact form."})
	}
	if input.Website != "" {
		return c.Status(202).JSON(fiber.Map{"sent": true})
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.Subject = strings.TrimSpace(input.Subject)
	input.Message = strings.TrimSpace(input.Message)
	for _, field := range []struct {
		value string
		max   int
	}{{input.Name, 200}, {input.Subject, 200}, {input.Message, 10000}} {
		if field.value == "" || !utf8.ValidString(field.value) || strings.ContainsRune(field.value, 0) || utf8.RuneCountInString(field.value) > field.max {
			return c.Status(400).JSON(fiber.Map{"error": "Provide a name, subject and message within the permitted lengths."})
		}
	}
	if !contact.ValidAddress(input.Email) || strings.ContainsAny(input.Name+input.Subject, "\r\n") {
		return c.Status(400).JSON(fiber.Map{"error": "Provide a valid email address, name and subject."})
	}
	if !a.CFG.Website.Contact.Ready() {
		return c.Status(503).JSON(fiber.Map{"error": "The contact form is temporarily unavailable."})
	}
	message := contact.Message{Name: input.Name, Email: input.Email, Subject: input.Subject, Body: input.Message, Language: c.Locals("language").(string)}
	if err := a.SendContact(a.CFG.Website.Contact, message); err != nil {
		log.Printf("Contact email failed: %v", err)
		return c.Status(502).JSON(fiber.Map{"error": "Your message could not be sent. Please try again later."})
	}
	return c.Status(202).JSON(fiber.Map{"sent": true})
}

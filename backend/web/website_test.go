package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"tas/backend/config"
	"tas/backend/contact"
	"tas/backend/database"
)

func translated(text string) database.LocalizedText {
	return database.LocalizedText{DE: "DE " + text, EN: "EN " + text}
}
func TestWebsiteValidation(t *testing.T) {
	for _, kind := range []string{"home", "teams", "events", "sponsors", "publications", "blog"} {
		e := database.WebsiteEntry{Kind: kind, Slug: "example", Content: database.WebsiteContent{Name: translated("name"), TeamStatus: "active", Date: "2026-07-01"}}
		if err := validateWebsiteEntry(&e); err != nil {
			t.Fatalf("valid %s: %v", kind, err)
		}
	}
	e := database.WebsiteEntry{Kind: "blog", Slug: "example", Published: true, Content: database.WebsiteContent{Name: translated("article"), Description: translated("summary")}}
	if validateWebsiteEntry(&e) == nil {
		t.Fatal("published blog accepted without image/date/blocks")
	}
	e.Published = false
	e.Slug = "../../admin"
	if validateWebsiteEntry(&e) == nil {
		t.Fatal("invalid slug accepted")
	}
	for _, raw := range []string{"https://evil.example/watch?v=dQw4w9WgXcQ", "javascript:alert(1)", "https://youtube.com@evil.example/watch?v=dQw4w9WgXcQ", "https://youtube.com/watch?v=x"} {
		if _, err := youtubeURL(raw); err == nil {
			t.Fatalf("bad video accepted: %s", raw)
		}
	}
	for _, raw := range []string{"https://youtu.be/dQw4w9WgXcQ", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", "https://www.youtube.com/shorts/dQw4w9WgXcQ"} {
		if value, err := youtubeURL(raw); err != nil || value != "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
			t.Fatalf("valid video rejected: %s", raw)
		}
	}
	if _, err := websiteLanguage("fr"); err == nil {
		t.Fatal("unsupported language accepted")
	}
	if lang, _ := websiteLanguage(""); lang != "en" {
		t.Fatal("default language changed")
	}
}

// Optional integration suite: uses an isolated schema in a rolled-back PostgreSQL transaction.
func websiteDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	path := os.Getenv("TAS_TEST_CONFIG")
	if path == "" {
		t.Skip("Set TAS_TEST_CONFIG to test website persistence against PostgreSQL.")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cfg config.Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	db := database.GetDB(&cfg).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	conn, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	schema := "tas_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := tx.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Exec("SET LOCAL search_path TO " + schema).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.AutoMigrate(&database.Image{}, &database.WebsiteEntry{}, &database.WebsiteReference{}); err != nil {
		t.Fatal(err)
	}
	return tx
}
func websiteFixture(t *testing.T) (*API, *fiber.App, *string, string) {
	t.Helper()
	db := websiteDatabase(t)
	cfg := &config.Config{}
	cfg.Website.PublicURL = "https://www.example.org"
	cfg.Website.AllowedOrigins = []string{"https://www.example.org"}
	cfg.Cloudflare.ImagesDeliveryURL = "https://imagedelivery.net/hash"
	cfg.Cloudflare.ImagesAccountId = "account"
	cfg.Cloudflare.ImagesAPIToken = "token"
	cfg.Cloudflare.ImagesVariant = "public"
	cfg.Auth.DisableFusionAuth = true
	auth, err := NewAuth(cfg.Auth)
	if err != nil {
		t.Fatal(err)
	}
	provider := &fakeImageProvider{store: &memoryImages{images: map[string]database.Image{}}}
	a := &API{CFG: cfg, DB: db, Auth: auth, Images: database.PostgresImages{DB: db}, ImageProvider: provider, SendContact: func(contact.Config, contact.Message) error { return nil }}
	role := "admin"
	app := fiber.New(fiber.Config{Immutable: true})
	v1 := app.Group("/api/v1", auth.CSRF, func(c fiber.Ctx) error { c.Locals("user", User{ID: "user", Roles: []string{role}}); return c.Next() })
	a.registerImages(v1)
	a.registerWebsite(app, v1)
	imageID := uuid.NewString()
	if err := db.Create(&database.Image{ID: imageID, CloudflareID: "tas-" + imageID, Name: "test", AltText: "test", Status: "ready"}).Error; err != nil {
		t.Fatal(err)
	}
	return a, app, &role, imageID
}
func siteReq(t *testing.T, app *fiber.App, method, path string, body any) (int, map[string]any) {
	t.Helper()
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TAS-CSRF", "1")
	res, err := app.Test(req, fiber.TestConfig{Timeout: 20 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	result := map[string]any{}
	raw, _ := io.ReadAll(res.Body)
	_ = json.Unmarshal(raw, &result)
	return res.StatusCode, result
}
func createSiteEntry(t *testing.T, app *fiber.App, e database.WebsiteEntry) database.WebsiteEntry {
	t.Helper()
	status, body := siteReq(t, app, "POST", "/api/v1/website/"+e.Kind, e)
	if status != 201 {
		t.Fatalf("create %s: %d %v", e.Kind, status, body)
	}
	raw, _ := json.Marshal(body["entry"])
	var saved database.WebsiteEntry
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	return saved
}
func TestWebsitePersistenceAndPublicProjection(t *testing.T) {
	a, app, role, imageID := websiteFixture(t)
	image := database.WebsiteImage{ID: imageID, Alt: translated("image")}
	event := createSiteEntry(t, app, database.WebsiteEntry{Kind: "events", Slug: "german-open-2026", Published: true, Content: database.WebsiteContent{Name: translated("German Open"), Description: translated("event description"), Date: "2026-07-01", Images: []database.WebsiteImage{image}}})
	team := createSiteEntry(t, app, database.WebsiteEntry{Kind: "teams", Slug: "alpha", Published: true, Content: database.WebsiteContent{Name: translated("Alpha"), Description: translated("team description"), TeamStatus: "active", Images: []database.WebsiteImage{image}, CoverImageID: imageID, Awards: []database.TeamAward{{EventID: event.ID, League: translated("league"), Result: translated("first")}}}})
	createSiteEntry(t, app, database.WebsiteEntry{Kind: "home", Slug: "home", Published: true, Content: database.WebsiteContent{About: translated("about"), Images: []database.WebsiteImage{image}, Videos: []database.WebsiteVideo{{Title: translated("video"), URL: "https://youtu.be/dQw4w9WgXcQ"}}}})
	for _, kind := range []string{"sponsors", "publications"} {
		createSiteEntry(t, app, database.WebsiteEntry{Kind: kind, Slug: "one", Published: true, Content: database.WebsiteContent{Name: translated(kind), Description: translated("description"), Images: []database.WebsiteImage{image}, URL: "https://example.org"}})
	}
	var firstBlog database.WebsiteEntry
	for i, slug := range []string{"old", "newer", "newest", "latest", "future", "draft"} {
		date := time.Now().Add(time.Duration(i-10) * time.Hour)
		if slug == "future" {
			date = time.Now().Add(time.Hour)
		}
		e := createSiteEntry(t, app, database.WebsiteEntry{Kind: "blog", Slug: slug, Published: slug != "draft", PublishAt: &date, Content: database.WebsiteContent{Name: translated(slug), Description: translated("description"), Images: []database.WebsiteImage{image}, CoverImageID: imageID, Blocks: []database.ArticleBlock{{ID: uuid.NewString(), Type: "text", Text: translated("**article**")}, {ID: uuid.NewString(), Type: "image", Images: []database.WebsiteImage{image}, Text: translated("caption")}}}})
		if i == 0 {
			firstBlog = e
		}
	}
	for _, lang := range []string{"de", "en"} {
		status, home := siteReq(t, app, "GET", "/website/home?lang="+lang, nil)
		if status != 200 || home["aboutUs"] != map[string]string{"de": "DE about", "en": "EN about"}[lang] {
			t.Fatalf("home language: %d %v", status, home)
		}
		images := home["images"].([]any)
		if !strings.HasSuffix(images[0].(map[string]any)["url"].(string), "/width=1920,fit=scale-down,quality=80,format=webp") {
			t.Fatal("website image transformation missing")
		}
		blogs := home["blogs"].([]any)
		if len(blogs) != 3 || blogs[0].(map[string]any)["slug"] != "latest" {
			t.Fatalf("latest published blogs incorrect: %v", blogs)
		}
		if strings.Contains(blogs[0].(map[string]any)["url"].(string), "future") {
			t.Fatal("future article leaked")
		}
		_, teams := siteReq(t, app, "GET", "/website/teams?lang="+lang, nil)
		awards := teams["items"].([]any)[0].(map[string]any)["awards"].([]any)
		if len(awards) != 1 {
			t.Fatal("team awards missing")
		}
		_, history := siteReq(t, app, "GET", "/website/participation-history?lang="+lang, nil)
		if len(history["items"].([]any)[0].(map[string]any)["results"].([]any)) != 1 {
			t.Fatal("event results missing")
		}
	}
	if status, _ := siteReq(t, app, "GET", "/website/blog/future", nil); status != 404 {
		t.Fatal("scheduled article leaked")
	}
	if status, _ := siteReq(t, app, "GET", "/website/blog/draft", nil); status != 404 {
		t.Fatal("draft article leaked")
	}
	status, detail := siteReq(t, app, "GET", "/website/blog/old?lang=de", nil)
	if status != 200 || len(detail["blocks"].([]any)) != 2 {
		t.Fatal("article blocks missing")
	}
	if status, _ := siteReq(t, app, "GET", "/website/home?lang=fr", nil); status != 400 {
		t.Fatal("invalid language accepted")
	}
	// Referenced images/events cannot be deleted, including references inside article blocks.
	if status, _ := siteReq(t, app, "DELETE", "/api/v1/images/"+imageID, nil); status != 409 {
		t.Fatal("referenced image could be deleted")
	}
	if status, _ := siteReq(t, app, "DELETE", "/api/v1/website/events/"+event.ID, nil); status != 409 {
		t.Fatal("referenced event could be deleted")
	}
	// Optimistic versioning prevents one translation editor overwriting another's work.
	team.Content.Description = translated("updated")
	if status, _ := siteReq(t, app, "PUT", "/api/v1/website/teams/"+team.ID, team); status != 200 {
		t.Fatal("update failed")
	}
	if status, _ := siteReq(t, app, "PUT", "/api/v1/website/teams/"+team.ID, team); status != 409 {
		t.Fatal("stale update overwrote content")
	}
	*role = "viewer"
	if status, _ := siteReq(t, app, "PUT", "/api/v1/website/blog/"+firstBlog.ID, firstBlog); status != 403 {
		t.Fatal("viewer could edit")
	}
	*role = "editor"
	if status, _ := siteReq(t, app, "DELETE", "/api/v1/website/blog/"+firstBlog.ID, nil); status != 403 {
		t.Fatal("editor could delete")
	}
	*role = "admin"
	if status, _ := siteReq(t, app, "DELETE", "/api/v1/website/blog/"+firstBlog.ID, nil); status != 204 {
		t.Fatal("admin deletion failed")
	}
	// Existing authentication defaults remain private; the public endpoint is separate.
	_ = a
}
func TestWebsiteContactAndCORS(t *testing.T) {
	a, app, _, _ := websiteFixture(t)
	body := map[string]string{"name": "Visitor", "email": "visitor@example.org", "subject": "Hello", "message": "A message"}
	if status, _ := siteReq(t, app, "POST", "/website/contact?lang=de", body); status != 503 {
		t.Fatal("unconfigured email reported success")
	}
	a.CFG.Website.Contact = contact.Config{Host: "smtp.example.org", Port: 587, TLSMode: "starttls", From: "noreply@example.org", To: "team@example.org"}
	calls := 0
	a.SendContact = func(_ contact.Config, m contact.Message) error {
		calls++
		if m.Language != "de" || m.Email != body["email"] {
			t.Error("incorrect contact fields")
		}
		return nil
	}
	if status, _ := siteReq(t, app, "POST", "/website/contact?lang=de", body); status != 202 || calls != 1 {
		t.Fatal("valid email not delivered")
	}
	body["website"] = "bot"
	if status, _ := siteReq(t, app, "POST", "/website/contact?lang=de", body); status != 202 || calls != 1 {
		t.Fatal("honeypot sent email")
	}
	body["website"] = ""
	body["email"] = "victim@example.org\r\nBcc: other@example.org"
	if status, _ := siteReq(t, app, "POST", "/website/contact?lang=de", body); status != 400 {
		t.Fatal("header injection accepted")
	}
	body["email"] = "visitor@example.org"
	a.SendContact = func(contact.Config, contact.Message) error { return errors.New("SMTP failed") }
	if status, _ := siteReq(t, app, "POST", "/website/contact?lang=de", body); status != 502 {
		t.Fatal("SMTP failure reported as success")
	}
	if status, _ := siteReq(t, app, "POST", "/website/contact?lang=de", body); status != 429 {
		t.Fatal("contact rate limit not enforced")
	}
	for _, origin := range []string{"https://www.example.org", "https://evil.example"} {
		req := httptest.NewRequest("GET", "/website/teams", nil)
		req.Header.Set("Origin", origin)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		if origin == "https://www.example.org" {
			if res.Header.Get("Access-Control-Allow-Origin") != origin {
				t.Fatal("website CORS missing")
			}
		} else if res.StatusCode != 403 {
			t.Fatal("unapproved origin accepted")
		}
	}
}

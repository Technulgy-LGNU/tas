package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"tas/backend/config"
	"tas/backend/database"
)

const testImageID = "eb0c2729-326a-4e1a-a4c8-bcbf85edba23"

var testPNG = []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("x", 40))

type memoryImages struct {
	images                         map[string]database.Image
	createErr, readyErr, deleteErr error
	lastLimit, lastOffset          int
	lastSearch                     string
}

func (s *memoryImages) List(_ context.Context, limit, offset int, search string) ([]database.Image, int64, error) {
	s.lastLimit, s.lastOffset, s.lastSearch = limit, offset, search
	result := []database.Image{}
	for _, image := range s.images {
		result = append(result, image)
	}
	return result, int64(len(result)), nil
}
func (s *memoryImages) Create(_ context.Context, image *database.Image) error {
	if s.createErr != nil {
		return s.createErr
	}
	image.CreatedAt = time.Now()
	image.UpdatedAt = image.CreatedAt
	s.images[image.ID] = *image
	return nil
}
func (s *memoryImages) Get(_ context.Context, id string) (database.Image, error) {
	image, ok := s.images[id]
	if !ok {
		return image, gorm.ErrRecordNotFound
	}
	return image, nil
}
func (s *memoryImages) MarkReady(_ context.Context, id string) error {
	if s.readyErr != nil {
		return s.readyErr
	}
	image := s.images[id]
	image.Status = "ready"
	s.images[id] = image
	return nil
}
func (s *memoryImages) Update(ctx context.Context, id, name, alt string) (database.Image, error) {
	image, err := s.Get(ctx, id)
	if err != nil {
		return image, err
	}
	image.Name = name
	image.AltText = alt
	s.images[id] = image
	return image, nil
}
func (s *memoryImages) Delete(_ context.Context, id string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.images, id)
	return nil
}

type fakeImageProvider struct {
	uploads, deletes     int
	uploadErr, deleteErr error
	store                *memoryImages
}

func (p *fakeImageProvider) Upload(_ context.Context, id, filename, creator string, data []byte) error {
	p.uploads++
	tracked := false
	for _, image := range p.store.images {
		if image.CloudflareID == id {
			tracked = true
		}
	}
	if !tracked {
		panic("upload must have a durable record first")
	}
	if _, err := uuid.Parse(id); err == nil {
		return errors.New("Cloudflare custom IDs must not be UUIDs")
	}
	return p.uploadErr
}
func (p *fakeImageProvider) Delete(_ context.Context, id string) error {
	p.deletes++
	return p.deleteErr
}

func imageApp(roles ...string) (*fiber.App, *memoryImages, *fakeImageProvider) {
	store := &memoryImages{images: map[string]database.Image{}}
	provider := &fakeImageProvider{store: store}
	cfg := &config.Config{}
	cfg.Cloudflare.ImagesAccountId = "account"
	cfg.Cloudflare.ImagesAPIToken = "secret"
	cfg.Cloudflare.ImagesDeliveryURL = "https://imagedelivery.net/hash"
	a := &API{CFG: cfg, Images: store, ImageProvider: provider}
	app := fiber.New(fiber.Config{BodyLimit: maxImageBytes + 1024*1024})
	v1 := app.Group("/api/v1", func(c fiber.Ctx) error { c.Locals("user", User{ID: "user", Roles: roles}); return c.Next() })
	a.registerImages(v1)
	return app, store, provider
}

func imageRequest(t *testing.T, app *fiber.App, method, path, contentType string, body io.Reader) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	return res
}
func uploadBody(t *testing.T, name string, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	_ = form.WriteField("name", name)
	_ = form.WriteField("altText", "Accessible description")
	part, err := form.CreateFormFile("file", "photo.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(data)
	_ = form.Close()
	return &body, form.FormDataContentType()
}
func seedImage(store *memoryImages) {
	store.images[testImageID] = database.Image{ID: testImageID, CloudflareID: testImageID, Name: "Original", AltText: "Existing alt", Status: "ready", CreatedAt: time.Now().Add(-time.Hour)}
}

func TestImageRolePermissions(t *testing.T) {
	for _, role := range []string{"viewer", "editor", "admin"} {
		for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
			t.Run(role+"_"+method, func(t *testing.T) {
				app, store, provider := imageApp(role)
				seedImage(store)
				path := "/api/v1/images"
				var body io.Reader
				contentType := ""
				if method == "POST" {
					body, contentType = uploadBody(t, "An image", testPNG)
				}
				if method == "PATCH" {
					path += "/" + testImageID
					body = strings.NewReader(`{"name":"Renamed","altText":""}`)
					contentType = "application/json"
				}
				if method == "DELETE" {
					path += "/" + testImageID
				}
				res := imageRequest(t, app, method, path, contentType, body)
				forbidden := (role == "viewer" && method != "GET") || (role == "editor" && method == "DELETE")
				if forbidden {
					if res.StatusCode != 403 || provider.uploads != 0 || provider.deletes != 0 || store.images[testImageID].Name != "Original" {
						t.Fatal("unauthorized mutation reached a dependency")
					}
				} else if res.StatusCode < 200 || res.StatusCode >= 300 {
					t.Fatalf("authorized request failed: %d", res.StatusCode)
				}
			})
		}
	}
}

func TestUploadPersistsMetadataAndRejectsInvalidFiles(t *testing.T) {
	app, store, _ := imageApp("editor")
	body, contentType := uploadBody(t, "  Website hero  ", testPNG)
	res := imageRequest(t, app, "POST", "/api/v1/images", contentType, body)
	var result struct {
		Image imageResponse `json:"image"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	stored := store.images[result.Image.ID]
	if stored.CloudflareID != "tas-"+stored.ID {
		t.Fatalf("Cloudflare ID must be distinct from the database UUID: %q", stored.CloudflareID)
	}
	if res.StatusCode != 201 || stored.Name != "Website hero" || stored.AltText != "Accessible description" || stored.UploadedBy != "user" || stored.ContentType != "image/png" || stored.Size != int64(len(testPNG)) || stored.Status != "ready" || !strings.HasSuffix(result.Image.URL, "/"+stored.CloudflareID+"/width=640,fit=scale-down,quality=80,format=webp") {
		t.Fatalf("wrong image record: %+v", result)
	}
	for _, tc := range []struct {
		name string
		data []byte
	}{{"", testPNG}, {strings.Repeat("x", 201), testPNG}, {"fake", []byte("<script>alert(1)</script>")}, {"empty", nil}, {"oversized", make([]byte, maxImageBytes+1)}} {
		app, store, provider := imageApp("editor")
		body, contentType := uploadBody(t, tc.name, tc.data)
		res := imageRequest(t, app, "POST", "/api/v1/images", contentType, body)
		if res.StatusCode != 400 || provider.uploads != 0 || len(store.images) != 0 {
			t.Fatalf("invalid image accepted: %s, %d", tc.name, res.StatusCode)
		}
	}
}

func TestImageFailureRecovery(t *testing.T) {
	for _, stage := range []string{"create", "upload", "finalize", "remote_delete", "database_delete"} {
		t.Run(stage, func(t *testing.T) {
			app, store, provider := imageApp("admin")
			failure := errors.New("test failure")
			if strings.Contains(stage, "delete") {
				seedImage(store)
				if stage == "remote_delete" {
					provider.deleteErr = failure
				} else {
					store.deleteErr = failure
				}
				res := imageRequest(t, app, "DELETE", "/api/v1/images/"+testImageID, "", nil)
				if res.StatusCode < 500 || len(store.images) != 1 {
					t.Fatal("failed delete lost its database record")
				}
				provider.deleteErr, store.deleteErr = nil, nil
				res = imageRequest(t, app, "DELETE", "/api/v1/images/"+testImageID, "", nil)
				if res.StatusCode != 204 || len(store.images) != 0 {
					t.Fatal("delete could not be retried")
				}
				return
			}
			switch stage {
			case "create":
				store.createErr = failure
			case "upload":
				provider.uploadErr = failure
			case "finalize":
				store.readyErr = failure
			}
			body, contentType := uploadBody(t, "Upload", testPNG)
			res := imageRequest(t, app, "POST", "/api/v1/images", contentType, body)
			if res.StatusCode < 500 {
				t.Fatal("failure reported as success")
			}
			if stage == "create" {
				if provider.uploads != 0 {
					t.Fatal("uploaded without a durable record")
				}
			} else {
				if len(store.images) != 1 {
					t.Fatal("lost incomplete upload record")
				}
				for _, image := range store.images {
					if image.Status != "pending" || image.CloudflareID == "" {
						t.Fatal("missing recovery metadata")
					}
					res := imageRequest(t, app, "DELETE", "/api/v1/images/"+image.ID, "", nil)
					if res.StatusCode != 409 {
						t.Fatal("allowed deletion of an in-flight upload")
					}
					image.CreatedAt = time.Now().Add(-3 * time.Minute)
					store.images[image.ID] = image
					res = imageRequest(t, app, "DELETE", "/api/v1/images/"+image.ID, "", nil)
					if res.StatusCode != 204 {
						t.Fatal("could not clean up incomplete upload")
					}
				}
			}
		})
	}
}

func TestImageListingAndMetadataEdits(t *testing.T) {
	app, store, _ := imageApp("editor")
	seedImage(store)
	res := imageRequest(t, app, "GET", "/api/v1/images?page=2&search=hero", "", nil)
	if res.StatusCode != 200 || store.lastLimit != 24 || store.lastOffset != 24 || store.lastSearch != "hero" {
		t.Fatal("pagination/search parameters not applied")
	}
	res = imageRequest(t, app, "PATCH", "/api/v1/images/"+testImageID, "application/json", strings.NewReader(`{"name":"New name","altText":""}`))
	if res.StatusCode != 200 || store.images[testImageID].AltText != "" || store.images[testImageID].CloudflareID != testImageID {
		t.Fatal("rename or clearing alt text failed")
	}
	for _, path := range []string{"/api/v1/images?page=0", "/api/v1/images?page=wrong"} {
		if imageRequest(t, app, "GET", path, "", nil).StatusCode != 400 {
			t.Fatal("bad pagination accepted")
		}
	}
	if imageRequest(t, app, "PATCH", "/api/v1/images/not-a-uuid", "application/json", strings.NewReader(`{}`)).StatusCode != 400 {
		t.Fatal("invalid UUID accepted")
	}
	if imageRequest(t, app, "PATCH", "/api/v1/images/"+testImageID, "application/json", strings.NewReader(`{"name":" "}`)).StatusCode != 400 {
		t.Fatal("empty name accepted")
	}
}

func TestImageRoutesUseAuthenticationAndCSRF(t *testing.T) {
	fixture := newFixture(t)
	cfg := &config.Config{Auth: fixture.auth.cfg}
	app, err := NewApp(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
		path := "/api/v1/images"
		if method == "PATCH" || method == "DELETE" {
			path += "/" + testImageID
		}
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("X-TAS-CSRF", "1")
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		if res.StatusCode != 401 {
			t.Fatalf("%s is not behind authentication: %d", method, res.StatusCode)
		}
		if method != "GET" {
			res = imageRequest(t, app, method, path, "", nil)
			if res.StatusCode != 403 {
				t.Fatal("missing CSRF header accepted")
			}
		}
	}
}

func TestTransformedImagesInLibraryAndWebsite(t *testing.T) {
	cfg := &config.Config{}
	cfg.Cloudflare.ImagesDeliveryURL = "https://imagedelivery.net/hash"
	cfg.Cloudflare.ImagesTransformOrigin = "https://images.example.org"
	a := &API{CFG: cfg}
	im := database.Image{ID: testImageID, CloudflareID: "tas-photo", Status: "ready", ContentType: "image/jpeg"}
	library := a.imageResponse(im)
	if library.URL != "https://images.example.org/cdn-cgi/imagedelivery/hash/tas-photo/width=640,fit=scale-down,quality=80,format=webp" {
		t.Fatal(library.URL)
	}
	if library.ContentType != "image/jpeg" {
		t.Fatal("source metadata was rewritten")
	}
	website := a.websiteImage(im, database.WebsiteImage{ID: testImageID, Alt: database.LocalizedText{DE: "Teamfoto", EN: "Team photo"}}, "de")
	if website["url"] != "https://images.example.org/cdn-cgi/imagedelivery/hash/tas-photo/width=1920,fit=scale-down,quality=80,format=webp" || website["alt"] != "Teamfoto" {
		t.Fatal(website)
	}
	im.Status = "uploading"
	if a.imageResponse(im).URL != "" {
		t.Fatal("incomplete upload exposed a delivery URL")
	}
}

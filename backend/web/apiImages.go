package web

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"tas/backend/database"
)

const maxImageBytes = 10 * 1024 * 1024

type imageProvider interface {
	Upload(context.Context, string, string, string, []byte) error
	Delete(context.Context, string) error
}

// RequireRoles must run after RequireAuth. Any one of these roles grants access.
func RequireRoles(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		user, ok := c.Locals("user").(User)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Authentication required."})
		}
		for _, role := range roles {
			if slices.Contains(user.Roles, role) {
				return c.Next()
			}
		}
		return c.Status(403).JSON(fiber.Map{"error": "You do not have permission to perform this action."})
	}
}

func (a *API) registerImages(v1 fiber.Router) {
	v1.Get("/images", a.listImages)
	v1.Post("/images", RequireRoles("editor", "admin"), a.uploadImage)
	v1.Patch("/images/:id", RequireRoles("editor", "admin"), a.updateImage)
	v1.Delete("/images/:id", RequireRoles("admin"), a.deleteImage)
}

type imageResponse struct {
	database.Image
	URL string `json:"url"`
}

func (a *API) imageResponse(image database.Image) imageResponse {
	link := ""
	if image.Status == "ready" {
		link = strings.TrimRight(a.CFG.Cloudflare.ImagesDeliveryURL, "/") + "/" + url.PathEscape(image.CloudflareID) + "/" + url.PathEscape(a.CFG.Cloudflare.ImagesVariant)
	}
	return imageResponse{Image: image, URL: link}
}

func (a *API) imagesConfigured(c fiber.Ctx) bool {
	cfg := a.CFG.Cloudflare
	u, err := url.Parse(cfg.ImagesDeliveryURL)
	if a.Images == nil || a.ImageProvider == nil || cfg.ImagesAccountId == "" || cfg.ImagesAPIToken == "" || cfg.ImagesVariant == "" || err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.RawQuery != "" || u.Fragment != "" {
		_ = c.Status(503).JSON(fiber.Map{"error": "The image library is not configured. Contact your administrator."})
		return false
	}
	return true
}

func (a *API) listImages(c fiber.Ctx) error {
	if !a.imagesConfigured(c) {
		return nil
	}
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 || page > 1000000 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid page."})
	}
	search := strings.TrimSpace(c.Query("search"))
	if utf8.RuneCountInString(search) > 200 {
		return c.Status(400).JSON(fiber.Map{"error": "Search must be at most 200 characters."})
	}
	const pageSize = 24
	images, total, err := a.Images.List(c.Context(), pageSize, (page-1)*pageSize, search)
	if err != nil {
		return imageDBError(c, err)
	}
	result := make([]imageResponse, 0, len(images))
	for _, image := range images {
		result = append(result, a.imageResponse(image))
	}
	return c.JSON(fiber.Map{"images": result, "total": total, "page": page, "pageSize": pageSize})
}

func imageFields(name, alt string) (string, string, error) {
	name, alt = strings.TrimSpace(name), strings.TrimSpace(alt)
	if !utf8.ValidString(name) || strings.ContainsRune(name, 0) || name == "" || utf8.RuneCountInString(name) > 200 {
		return "", "", errors.New("Image name must contain between 1 and 200 characters.")
	}
	if !utf8.ValidString(alt) || strings.ContainsRune(alt, 0) || utf8.RuneCountInString(alt) > 1000 {
		return "", "", errors.New("Alt text must be at most 1000 characters.")
	}
	return name, alt, nil
}

func (a *API) uploadImage(c fiber.Ctx) error {
	if !a.imagesConfigured(c) {
		return nil
	}
	name, alt, err := imageFields(c.FormValue("name"), c.FormValue("altText"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	form, err := c.MultipartForm()
	if err != nil || len(form.File["file"]) != 1 {
		return c.Status(400).JSON(fiber.Map{"error": "Choose one image file."})
	}
	file := form.File["file"][0]
	if file.Size <= 0 || file.Size > maxImageBytes {
		return c.Status(400).JSON(fiber.Map{"error": "Images must be between 1 byte and 10 MB."})
	}
	reader, err := file.Open()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "The image could not be read."})
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, maxImageBytes+1))
	if err != nil || len(data) == 0 || len(data) > maxImageBytes {
		return c.Status(400).JSON(fiber.Map{"error": "The image could not be read or exceeds 10 MB."})
	}
	contentType := http.DetectContentType(data)
	if !slices.Contains([]string{"image/jpeg", "image/png", "image/gif", "image/webp"}, contentType) {
		return c.Status(400).JSON(fiber.Map{"error": "Choose a JPEG, PNG, GIF or WebP image."})
	}
	id := uuid.NewString()
	// Cloudflare reserves bare UUIDs for its own generated IDs (error 5411).
	// Keep our database UUID, but use a namespaced custom ID for the remote file.
	image := database.Image{ID: id, CloudflareID: "tas-" + id, Name: name, AltText: alt, Filename: filepath.Base(strings.ReplaceAll(file.Filename, "\\", "/")), ContentType: contentType,
		Size: int64(len(data)), UploadedBy: c.Locals("user").(User).ID, Status: "pending"}
	if err := a.Images.Create(c.Context(), &image); err != nil {
		return imageDBError(c, err)
	}
	// Persist the ID before uploading: timeouts/crashes must not leave untracked files.
	if err := a.ImageProvider.Upload(c.Context(), image.CloudflareID, image.Filename, image.UploadedBy, data); err != nil {
		log.Printf("Image upload %s: %v", id, err)
		return c.Status(502).JSON(fiber.Map{"error": "Upload could not be completed. Its library entry is marked incomplete; an administrator can remove it before you try again."})
	}
	if err := a.Images.MarkReady(c.Context(), id); err != nil {
		return imageDBError(c, err)
	}
	image.Status = "ready"
	return c.Status(201).JSON(fiber.Map{"image": a.imageResponse(image)})
}

func imageID(c fiber.Ctx) (string, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return "", fiber.NewError(400, "Invalid image ID.")
	}
	return id.String(), nil
}

func (a *API) updateImage(c fiber.Ctx) error {
	if !a.imagesConfigured(c) {
		return nil
	}
	id, err := imageID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid image ID."})
	}
	var fields struct {
		Name    string `json:"name"`
		AltText string `json:"altText"`
	}
	if len(c.Body()) > 16384 || json.Unmarshal(c.Body(), &fields) != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid image details."})
	}
	name, alt, err := imageFields(fields.Name, fields.AltText)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	image, err := a.Images.Update(c.Context(), id, name, alt)
	if err != nil {
		return imageDBError(c, err)
	}
	return c.JSON(fiber.Map{"image": a.imageResponse(image)})
}

func (a *API) deleteImage(c fiber.Ctx) error {
	if !a.imagesConfigured(c) {
		return nil
	}
	id, err := imageID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid image ID."})
	}
	image, err := a.Images.Get(c.Context(), id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.SendStatus(204)
	}
	if err != nil {
		return imageDBError(c, err)
	}
	// Avoid deleting while the provider's upload request can still be in flight.
	if image.Status == "pending" && time.Since(image.CreatedAt) < 2*time.Minute {
		return c.Status(409).JSON(fiber.Map{"error": "This upload may still be running. Wait two minutes before removing it."})
	}
	if err := a.ImageProvider.Delete(c.Context(), image.CloudflareID); err != nil {
		log.Printf("Image deletion %s: %v", id, err)
		return c.Status(502).JSON(fiber.Map{"error": "Cloudflare could not delete this image. Please try again."})
	}
	if err := a.Images.Delete(c.Context(), id); err != nil {
		return imageDBError(c, err)
	}
	return c.SendStatus(204)
}

func imageDBError(c fiber.Ctx, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(404).JSON(fiber.Map{"error": "This image no longer exists."})
	}
	log.Printf("Image library database operation failed: %v", err)
	return c.Status(500).JSON(fiber.Map{"error": "The image library could not be saved or loaded. Please refresh and try again."})
}

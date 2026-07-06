package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"tas/internal/database"

	"github.com/gofiber/fiber/v2"
)

type cloudflareImageResponse struct {
	Success bool `json:"success"`
	Result  struct {
		ID       string   `json:"id"`
		Filename string   `json:"filename"`
		Variants []string `json:"variants"`
	} `json:"result"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (a *API) listImages(c *fiber.Ctx) error {
	var images []database.UploadedImage
	if err := a.DB.Order("created_at desc").Find(&images).Error; err != nil {
		return err
	}
	return c.JSON(images)
}

func (a *API) uploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return fail(fiber.StatusBadRequest, "file field is required")
	}
	member, _ := currentMember(c)
	uploaded, err := a.uploadToCloudflare(file)
	if err != nil {
		return err
	}
	variants, _ := json.Marshal(uploaded.Result.Variants)
	image := database.UploadedImage{
		CloudflareImageID: uploaded.Result.ID,
		Filename:          firstNonEmpty(uploaded.Result.Filename, file.Filename),
		ContentType:       file.Header.Get("Content-Type"),
		DeliveryURL:       a.imageDeliveryURL(uploaded.Result.ID, uploaded.Result.Variants),
		VariantsJSON:      string(variants),
	}
	if member != nil {
		image.UploadedByID = &member.ID
	}
	if err := a.DB.Create(&image).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(image)
}

func (a *API) uploadToCloudflare(fileHeader *multipart.FileHeader) (cloudflareImageResponse, error) {
	var result cloudflareImageResponse
	accountID := strings.TrimSpace(a.CFG.Cloudflare.ImagesAccountID)
	token := strings.TrimSpace(a.CFG.Cloudflare.ImagesAPIToken)
	if accountID == "" || token == "" {
		return result, fail(fiber.StatusServiceUnavailable, "Cloudflare Images is not configured")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return result, fail(fiber.StatusBadRequest, "could not open upload")
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		return result, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return result, err
	}
	if err := writer.Close(); err != nil {
		return result, err
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/images/v1", accountID)
	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, fail(fiber.StatusBadGateway, "Cloudflare upload failed")
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || !result.Success {
		message := "Cloudflare upload was rejected"
		if len(result.Errors) > 0 && strings.TrimSpace(result.Errors[0].Message) != "" {
			message = result.Errors[0].Message
		}
		return result, fail(fiber.StatusBadGateway, message)
	}
	return result, nil
}

func (a *API) imageDeliveryURL(id string, variants []string) string {
	if len(variants) > 0 {
		return variants[0]
	}
	base := strings.TrimRight(strings.TrimSpace(a.CFG.Cloudflare.ImagesDeliveryURL), "/")
	variant := strings.TrimSpace(a.CFG.Cloudflare.ImagesVariant)
	if base == "" || id == "" || variant == "" {
		return ""
	}
	return base + "/" + id + "/" + variant
}

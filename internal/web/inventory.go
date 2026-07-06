package web

import (
	"strings"
	"tas/internal/database"

	"github.com/gofiber/fiber/v2"
)

type inventoryCategoryPayload struct {
	Name      string `json:"name"`
	ParentID  *uint  `json:"parent_id"`
	SortOrder int    `json:"sort_order"`
}

func (a *API) listInventoryCategories(c *fiber.Ctx) error {
	var categories []database.InventoryCategory
	if err := a.DB.Preload("Children").Order("sort_order asc, name asc").Find(&categories).Error; err != nil {
		return err
	}
	return c.JSON(categories)
}

func (a *API) createInventoryCategory(c *fiber.Ctx) error {
	payload, err := bindJSON[inventoryCategoryPayload](c)
	if err != nil {
		return err
	}
	category, err := inventoryCategoryFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Create(&category).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(category)
}

func (a *API) updateInventoryCategory(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	payload, err := bindJSON[inventoryCategoryPayload](c)
	if err != nil {
		return err
	}
	category, err := inventoryCategoryFromPayload(payload)
	if err != nil {
		return err
	}
	category.ID = id
	if err := a.DB.Model(&database.InventoryCategory{}).Where("id = ?", id).Updates(category).Error; err != nil {
		return err
	}
	if err := a.DB.First(&category, id).Error; err != nil {
		return notFoundOrError(err, "category not found")
	}
	return c.JSON(category)
}

func (a *API) deleteInventoryCategory(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	if err := a.DB.Delete(&database.InventoryCategory{}, id).Error; err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type inventoryItemPayload struct {
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	VendorID   string `json:"vendor_id"`
	ProductURL string `json:"product_url"`
	Website    string `json:"website"`
	Notes      string `json:"notes"`
	CategoryID *uint  `json:"category_id"`
	Confirmed  *bool  `json:"confirmed"`
}

func (a *API) listInventoryItems(c *fiber.Ctx) error {
	var items []database.InventoryItem
	query := a.DB.Preload("Category").Preload("MatchedInventoryItem").Order("confirmed asc, name asc")
	if confirmed := strings.TrimSpace(c.Query("confirmed")); confirmed != "" {
		query = query.Where("confirmed = ?", confirmed == "true")
	}
	if categoryID := strings.TrimSpace(c.Query("category_id")); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if search := strings.TrimSpace(c.Query("q")); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("lower(name) LIKE ? OR lower(vendor_id) LIKE ?", like, like)
	}
	if err := query.Find(&items).Error; err != nil {
		return err
	}
	return c.JSON(items)
}

func (a *API) createInventoryItem(c *fiber.Ctx) error {
	payload, err := bindJSON[inventoryItemPayload](c)
	if err != nil {
		return err
	}
	item, err := inventoryItemFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Create(&item).Error; err != nil {
		return err
	}
	if err := a.DB.Preload("Category").First(&item, item.ID).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (a *API) updateInventoryItem(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	payload, err := bindJSON[inventoryItemPayload](c)
	if err != nil {
		return err
	}
	item, err := inventoryItemFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Model(&database.InventoryItem{}).Where("id = ?", id).Updates(item).Error; err != nil {
		return err
	}
	if err := a.DB.Preload("Category").First(&item, id).Error; err != nil {
		return notFoundOrError(err, "inventory item not found")
	}
	return c.JSON(item)
}

func (a *API) deleteInventoryItem(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	if err := a.DB.Delete(&database.InventoryItem{}, id).Error; err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type reorderPayload struct {
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	Notes          string `json:"notes"`
}

func (a *API) reorderInventoryItem(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	payload, err := bindJSON[reorderPayload](c)
	if err != nil {
		return err
	}
	member, err := currentMember(c)
	if err != nil {
		return err
	}
	var item database.InventoryItem
	if err := a.DB.First(&item, id).Error; err != nil {
		return notFoundOrError(err, "inventory item not found")
	}
	quantity := payload.Quantity
	if quantity < 1 {
		quantity = 1
	}
	request := database.OrderRequest{
		Name:           item.Name,
		Quantity:       quantity,
		UnitPriceCents: payload.UnitPriceCents,
		URL:            firstNonEmpty(item.ProductURL, item.Website),
		Notes:          payload.Notes,
		Status:         database.OrderRequestStatusPending,
		RequesterID:    member.ID,
	}
	applyShop(&request.ShopDomain, &request.ShopName, request.URL, "")
	if err := a.DB.Create(&request).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(request)
}

func inventoryCategoryFromPayload(payload inventoryCategoryPayload) (database.InventoryCategory, error) {
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return database.InventoryCategory{}, fail(fiber.StatusBadRequest, "category name is required")
	}
	return database.InventoryCategory{Name: name, ParentID: payload.ParentID, SortOrder: payload.SortOrder}, nil
}

func inventoryItemFromPayload(payload inventoryItemPayload) (database.InventoryItem, error) {
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return database.InventoryItem{}, fail(fiber.StatusBadRequest, "inventory item name is required")
	}
	quantity := payload.Quantity
	if quantity < 0 {
		return database.InventoryItem{}, fail(fiber.StatusBadRequest, "quantity cannot be negative")
	}
	confirmed := true
	if payload.Confirmed != nil {
		confirmed = *payload.Confirmed
	}
	return database.InventoryItem{
		Name:       name,
		Quantity:   quantity,
		VendorID:   strings.TrimSpace(payload.VendorID),
		ProductURL: strings.TrimSpace(payload.ProductURL),
		Website:    strings.TrimSpace(payload.Website),
		Notes:      strings.TrimSpace(payload.Notes),
		CategoryID: payload.CategoryID,
		Confirmed:  confirmed,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

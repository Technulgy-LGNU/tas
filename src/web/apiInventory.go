package web

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"strconv"
	"tas/src/database"
	"tas/src/util"
)

// Get all categories together with their inventory entries
func (a *API) getInventories(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 1, util.Inventory, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	var (
		inventory []database.InventoryCategory

		err error
	)

	err = a.DB.Preload("InventoryEntries").Find(&inventory).Error
	if err != nil {
		log.Printf("Error getting categories: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Unable to get categories")
	}

	return c.Status(fiber.StatusOK).JSON(inventory)
}

// Create new category
func (a *API) createCategory(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 3, util.Inventory, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	var (
		data = struct {
			Name string `json:"name"`
		}{}
	)
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid Body")
	}
	if data.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid Body")
	}
	err := a.DB.Create(&database.InventoryCategory{
		Name: data.Name,
	}).Error
	if err != nil {
		log.Printf("Error creating categories: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Unable to create categories")
	}

	var inventory []database.InventoryCategory
	err = a.DB.Preload("InventoryEntries").Find(&inventory).Error
	if err != nil {
		log.Printf("Error getting categories: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Unable to get categories")
	}

	return c.Status(fiber.StatusOK).JSON(inventory)
}

// Create a new entry in one category
func (a *API) createInventoryEntry(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 2, util.Inventory, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	var (
		data = struct {
			Name     string `json:"name"`
			Link     string `json:"link"`
			Location string `json:"location"`
			Quantity int    `json:"quantity"`
		}{}

		entry database.InventoryEntry
		err   error
	)

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid Body")
	}
	if data.Name == "" || data.Quantity == 0 {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid Body")
	}

	entry.Name = data.Name
	entry.Link = data.Link
	entry.Quantity = data.Quantity
	entry.Location = data.Location
	entry.InventoryCategoryID, err = strconv.ParseUint(c.Params("categoryID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid CategoryID")
	}

	err = a.DB.Create(&entry).Error
	if err != nil {
		log.Printf("Error creating categories: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Unable to create categories")
	}

	return c.Status(fiber.StatusOK).JSON("Entry created successfully")
}

// Delete a category
func (a *API) deleteCategory(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 3, util.Inventory, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid ID")
	}
	err = a.DB.Delete(&database.InventoryCategory{}, id).Error
	if err != nil {
		log.Printf("Error deleting category: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Unable to delete category")
	}

	return c.Status(fiber.StatusOK).JSON("Deleted category successfully")
}

// Delete an inventory entry
func (a *API) deleteInventoryEntry(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 3, util.Inventory, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid ID")
	}
	err = a.DB.Delete(&database.InventoryEntry{}, id).Error
	if err != nil {
		log.Printf("Error deleting inventory entry: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Unable to delete inventory entry")
	}

	return c.Status(fiber.StatusOK).JSON("Deleted inventory entry successfully")
}

// Update a category
func (a *API) updateCategory(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 2, util.Inventory, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	var category database.InventoryCategory

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid ID")
	}
	var (
		data = struct {
			Name string `json:"name"`
		}{}
	)
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid Body")
	}
	if data.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid Body")
	}

	err = a.DB.First(&category, id).Error
	if err != nil {
		log.Printf("Error getting category: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Unable to update category")
	}

	category.Name = data.Name
	err = a.DB.Save(&category).Error
	if err != nil {
		log.Printf("Error updating category: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Unable to update category")
	}

	return c.Status(fiber.StatusOK).JSON("Updated category successfully")
}

// Update an entry
func (a *API) updateInventoryEntry(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 2, util.Inventory, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	var entry database.InventoryEntry
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid ID")
	}
	err = a.DB.First(&entry, id).Error
	if err != nil {
		log.Printf("Error getting inventory entry: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Unable to update inventory entry")
	}

	if err := c.BodyParser(&entry); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid Body")
	}
	if entry.Name == "" || entry.Quantity == 0 {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid Body")
	}

	err = a.DB.Save(&entry).Error
	if err != nil {
		log.Printf("Error updating inventory entry: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Unable to update inventory entry")
	}

	return c.Status(fiber.StatusOK).JSON("Entry updated successfully")
}

// Update amount (For the kids while counting stuff)
func (a *API) updateInventoryAmount(c *fiber.Ctx) error {
	if !util.CheckPermissions(c.GetReqHeaders(), 1, util.Inventory, a.DB) {
		return c.Status(fiber.StatusUnauthorized).JSON("")
	}

	var entry database.InventoryEntry

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid Body")
	}
	amount, err := strconv.Atoi(c.Params("amount"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid Body")
	}

	err = a.DB.First(&entry, id).Error
	if err != nil {
		log.Printf("Error getting inventory entry: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Unable to update inventory entry")
	}
	entry.Quantity = amount
	err = a.DB.Save(&entry).Error
	if err != nil {
		log.Printf("Error updating inventory entry: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON("Unable to update inventory entry")
	}

	return c.Status(fiber.StatusOK).JSON("Updated inventory entry successfully")
}

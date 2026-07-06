package web

import (
	"strings"
	"tas/internal/database"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type orderRequestPayload struct {
	Name           string `json:"name"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	URL            string `json:"url"`
	Notes          string `json:"notes"`
	ShopName       string `json:"shop_name"`
}

func (a *API) listOrderRequests(c *fiber.Ctx) error {
	var requests []database.OrderRequest
	query := a.DB.Preload("Requester").Preload("ApprovedBy").Preload("RejectedBy").Order("shop_name asc, created_at desc")
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&requests).Error; err != nil {
		return err
	}
	return c.JSON(requests)
}

func (a *API) createOrderRequest(c *fiber.Ctx) error {
	payload, err := bindJSON[orderRequestPayload](c)
	if err != nil {
		return err
	}
	member, err := currentMember(c)
	if err != nil {
		return err
	}
	request, err := orderRequestFromPayload(payload)
	if err != nil {
		return err
	}
	request.Status = database.OrderRequestStatusPending
	request.RequesterID = member.ID
	if err := a.DB.Create(&request).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(request)
}

func (a *API) updateOrderRequest(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	payload, err := bindJSON[orderRequestPayload](c)
	if err != nil {
		return err
	}
	request, err := orderRequestFromPayload(payload)
	if err != nil {
		return err
	}
	if err := a.DB.Model(&database.OrderRequest{}).Where("id = ?", id).Updates(map[string]any{
		"name":              request.Name,
		"quantity":          request.Quantity,
		"unit_price_cents":  request.UnitPriceCents,
		"total_price_cents": int64(request.Quantity) * request.UnitPriceCents,
		"url":               request.URL,
		"notes":             request.Notes,
		"shop_domain":       request.ShopDomain,
		"shop_name":         request.ShopName,
	}).Error; err != nil {
		return err
	}
	if err := a.DB.Preload("Requester").First(&request, id).Error; err != nil {
		return notFoundOrError(err, "order request not found")
	}
	return c.JSON(request)
}

type approveOrderRequestPayload struct {
	OrderListID *uint  `json:"order_list_id"`
	ListName    string `json:"list_name"`
}

func (a *API) approveOrderRequest(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	payload, err := bindJSON[approveOrderRequestPayload](c)
	if err != nil {
		return err
	}
	member, err := currentMember(c)
	if err != nil {
		return err
	}

	var request database.OrderRequest
	if err := a.DB.First(&request, id).Error; err != nil {
		return notFoundOrError(err, "order request not found")
	}
	if request.Status != database.OrderRequestStatusPending {
		return fail(fiber.StatusConflict, "only pending requests can be approved")
	}

	var list database.OrderList
	if payload.OrderListID != nil {
		if err := a.DB.First(&list, *payload.OrderListID).Error; err != nil {
			return notFoundOrError(err, "order list not found")
		}
		if list.Status != database.OrderListStatusDraft {
			return fail(fiber.StatusConflict, "requests can only be added to draft lists")
		}
	} else {
		name := strings.TrimSpace(payload.ListName)
		if name == "" {
			name = "Order List " + time.Now().Format("2006-01-02")
		}
		list = database.OrderList{Name: name, Status: database.OrderListStatusDraft, CreatedByID: member.ID}
		if err := a.DB.Create(&list).Error; err != nil {
			return err
		}
	}

	item := database.OrderListItem{
		OrderListID:    list.ID,
		OrderRequestID: &request.ID,
		Name:           request.Name,
		Quantity:       request.Quantity,
		UnitPriceCents: request.UnitPriceCents,
		URL:            request.URL,
		Notes:          request.Notes,
		ShopDomain:     request.ShopDomain,
		ShopName:       request.ShopName,
	}
	if err := a.DB.Create(&item).Error; err != nil {
		return err
	}
	if err := a.DB.Model(&request).Updates(map[string]any{
		"status":             database.OrderRequestStatusApproved,
		"approved_by_id":     member.ID,
		"order_list_item_id": item.ID,
	}).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (a *API) rejectOrderRequest(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	member, err := currentMember(c)
	if err != nil {
		return err
	}
	var request database.OrderRequest
	if err := a.DB.First(&request, id).Error; err != nil {
		return notFoundOrError(err, "order request not found")
	}
	if err := a.DB.Model(&request).Updates(map[string]any{
		"status":         database.OrderRequestStatusRejected,
		"rejected_by_id": member.ID,
	}).Error; err != nil {
		return err
	}
	return c.JSON(request)
}

type orderListPayload struct {
	Name string `json:"name"`
}

func (a *API) listOrderLists(c *fiber.Ctx) error {
	var lists []database.OrderList
	if err := a.DB.Preload("Items").Order("created_at desc").Find(&lists).Error; err != nil {
		return err
	}
	return c.JSON(lists)
}

func (a *API) createOrderList(c *fiber.Ctx) error {
	payload, err := bindJSON[orderListPayload](c)
	if err != nil {
		return err
	}
	member, err := currentMember(c)
	if err != nil {
		return err
	}
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return fail(fiber.StatusBadRequest, "order list name is required")
	}
	list := database.OrderList{Name: name, Status: database.OrderListStatusDraft, CreatedByID: member.ID}
	if err := a.DB.Create(&list).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(list)
}

func (a *API) getOrderList(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	var list database.OrderList
	if err := a.DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("shop_name asc, name asc")
	}).First(&list, id).Error; err != nil {
		return notFoundOrError(err, "order list not found")
	}
	return c.JSON(list)
}

func (a *API) updateOrderList(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	payload, err := bindJSON[orderListPayload](c)
	if err != nil {
		return err
	}
	var list database.OrderList
	if err := a.DB.First(&list, id).Error; err != nil {
		return notFoundOrError(err, "order list not found")
	}
	if err := a.requireManageOrDraft(c, &list); err != nil {
		return err
	}
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return fail(fiber.StatusBadRequest, "order list name is required")
	}
	if err := a.DB.Model(&list).Update("name", name).Error; err != nil {
		return err
	}
	list.Name = name
	return c.JSON(list)
}

func (a *API) publishOrderList(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	now := time.Now()
	var list database.OrderList
	if err := a.DB.First(&list, id).Error; err != nil {
		return notFoundOrError(err, "order list not found")
	}
	if err := a.DB.Model(&list).Updates(map[string]any{
		"status":       database.OrderListStatusPublished,
		"published_at": &now,
	}).Error; err != nil {
		return err
	}
	list.Status = database.OrderListStatusPublished
	list.PublishedAt = &now
	return c.JSON(list)
}

func (a *API) createOrderListItem(c *fiber.Ctx) error {
	list, err := a.loadOrderListFromParam(c)
	if err != nil {
		return err
	}
	if err := a.requireManageOrDraft(c, &list); err != nil {
		return err
	}
	payload, err := bindJSON[orderRequestPayload](c)
	if err != nil {
		return err
	}
	item, err := orderListItemFromPayload(payload, list.ID)
	if err != nil {
		return err
	}
	if err := a.DB.Create(&item).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (a *API) updateOrderListItem(c *fiber.Ctx) error {
	list, item, err := a.loadOrderListAndItem(c)
	if err != nil {
		return err
	}
	if err := a.requireManageOrDraft(c, &list); err != nil {
		return err
	}
	payload, err := bindJSON[orderRequestPayload](c)
	if err != nil {
		return err
	}
	updated, err := orderListItemFromPayload(payload, list.ID)
	if err != nil {
		return err
	}
	if err := a.DB.Model(&item).Updates(map[string]any{
		"name":              updated.Name,
		"quantity":          updated.Quantity,
		"unit_price_cents":  updated.UnitPriceCents,
		"total_price_cents": int64(updated.Quantity) * updated.UnitPriceCents,
		"url":               updated.URL,
		"notes":             updated.Notes,
		"shop_domain":       updated.ShopDomain,
		"shop_name":         updated.ShopName,
	}).Error; err != nil {
		return err
	}
	if err := a.DB.First(&item, item.ID).Error; err != nil {
		return err
	}
	return c.JSON(item)
}

func (a *API) markOrderListItemOrdered(c *fiber.Ctx) error {
	_, item, err := a.loadOrderListAndItem(c)
	if err != nil {
		return err
	}
	now := time.Now()
	if err := a.DB.Model(&item).Updates(map[string]any{"ordered": true, "ordered_at": &now}).Error; err != nil {
		return err
	}
	item.Ordered = true
	item.OrderedAt = &now
	return c.JSON(item)
}

func (a *API) matchOrderListItem(c *fiber.Ctx) error {
	_, item, err := a.loadOrderListAndItem(c)
	if err != nil {
		return err
	}
	var inventory []database.InventoryItem
	if err := a.DB.Where("confirmed = ?", true).Find(&inventory).Error; err != nil {
		return err
	}
	return c.JSON(fuzzyInventoryMatches(item, inventory))
}

type receiveOrderListItemPayload struct {
	MatchedInventoryItemID *uint  `json:"matched_inventory_item_id"`
	CategoryID             *uint  `json:"category_id"`
	VendorID               string `json:"vendor_id"`
	Notes                  string `json:"notes"`
}

func (a *API) receiveOrderListItem(c *fiber.Ctx) error {
	_, item, err := a.loadOrderListAndItem(c)
	if err != nil {
		return err
	}
	payload, err := bindJSON[receiveOrderListItemPayload](c)
	if err != nil {
		return err
	}
	member, err := currentMember(c)
	if err != nil {
		return err
	}
	if item.Received {
		return fail(fiber.StatusConflict, "order list item was already received")
	}

	received := database.InventoryItem{
		Name:                   item.Name,
		Quantity:               item.Quantity,
		VendorID:               strings.TrimSpace(payload.VendorID),
		ProductURL:             item.URL,
		Website:                item.ShopDomain,
		Notes:                  strings.TrimSpace(payload.Notes),
		CategoryID:             payload.CategoryID,
		Confirmed:              false,
		MatchedInventoryItemID: payload.MatchedInventoryItemID,
		SourceOrderListItemID:  &item.ID,
	}
	if payload.MatchedInventoryItemID != nil {
		var matched database.InventoryItem
		if err := a.DB.First(&matched, *payload.MatchedInventoryItemID).Error; err != nil {
			return notFoundOrError(err, "matched inventory item not found")
		}
		if received.CategoryID == nil {
			received.CategoryID = matched.CategoryID
		}
		if received.VendorID == "" {
			received.VendorID = matched.VendorID
		}
	}
	if err := a.DB.Create(&received).Error; err != nil {
		return err
	}
	now := time.Now()
	if err := a.DB.Model(&item).Updates(map[string]any{
		"received":                   true,
		"received_at":                &now,
		"received_by_id":             member.ID,
		"received_inventory_item_id": received.ID,
	}).Error; err != nil {
		return err
	}
	item.Received = true
	item.ReceivedAt = &now
	item.ReceivedByID = &member.ID
	item.ReceivedInventoryItemID = &received.ID
	return c.JSON(fiber.Map{"item": item, "inventory_item": received})
}

func (a *API) loadOrderListFromParam(c *fiber.Ctx) (database.OrderList, error) {
	id, err := uintParam(c, "id")
	if err != nil {
		return database.OrderList{}, err
	}
	var list database.OrderList
	if err := a.DB.First(&list, id).Error; err != nil {
		return database.OrderList{}, notFoundOrError(err, "order list not found")
	}
	return list, nil
}

func (a *API) loadOrderListAndItem(c *fiber.Ctx) (database.OrderList, database.OrderListItem, error) {
	list, err := a.loadOrderListFromParam(c)
	if err != nil {
		return database.OrderList{}, database.OrderListItem{}, err
	}
	itemID, err := uintParam(c, "item_id")
	if err != nil {
		return database.OrderList{}, database.OrderListItem{}, err
	}
	var item database.OrderListItem
	if err := a.DB.Where("order_list_id = ?", list.ID).First(&item, itemID).Error; err != nil {
		return database.OrderList{}, database.OrderListItem{}, notFoundOrError(err, "order list item not found")
	}
	return list, item, nil
}

func orderRequestFromPayload(payload orderRequestPayload) (database.OrderRequest, error) {
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return database.OrderRequest{}, fail(fiber.StatusBadRequest, "order request name is required")
	}
	if payload.Quantity < 1 {
		payload.Quantity = 1
	}
	if payload.UnitPriceCents < 0 {
		return database.OrderRequest{}, fail(fiber.StatusBadRequest, "unit price cannot be negative")
	}
	request := database.OrderRequest{
		Name:           name,
		Quantity:       payload.Quantity,
		UnitPriceCents: payload.UnitPriceCents,
		URL:            strings.TrimSpace(payload.URL),
		Notes:          strings.TrimSpace(payload.Notes),
	}
	applyShop(&request.ShopDomain, &request.ShopName, request.URL, payload.ShopName)
	return request, nil
}

func orderListItemFromPayload(payload orderRequestPayload, listID uint) (database.OrderListItem, error) {
	request, err := orderRequestFromPayload(payload)
	if err != nil {
		return database.OrderListItem{}, err
	}
	return database.OrderListItem{
		OrderListID:    listID,
		Name:           request.Name,
		Quantity:       request.Quantity,
		UnitPriceCents: request.UnitPriceCents,
		URL:            request.URL,
		Notes:          request.Notes,
		ShopDomain:     request.ShopDomain,
		ShopName:       request.ShopName,
	}, nil
}

package web

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"tas/backend/database"
)

func orderRole(u User, roles ...string) bool {
	for _, role := range roles {
		if slices.Contains(u.Roles, role) {
			return true
		}
	}
	return false
}
func orderManager(u User) bool { return orderRole(u, "admin", "order_admin") }
func orderEditor(u User) bool  { return orderManager(u) || orderRole(u, "editor") }
func orderError(c fiber.Ctx, err error) error {
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(404).JSON(fiber.Map{"error": "Order content not found."})
	}
	log.Printf("Order database operation failed: %v", err)
	return c.Status(500).JSON(fiber.Map{"error": "Order content could not be saved or loaded."})
}
func orderBody(c fiber.Ctx, input any) error {
	if len(c.Body()) > 64*1024 {
		return fiber.NewError(413, "Request is too large.")
	}
	if err := json.Unmarshal(c.Body(), input); err != nil {
		return fiber.NewError(400, "Invalid JSON request.")
	}
	return nil
}
func orderText(raw string, max int, required bool) bool {
	return utf8.ValidString(raw) && !strings.ContainsAny(raw, "\x00\r\n") && utf8.RuneCountInString(raw) <= max && (!required || strings.TrimSpace(raw) != "")
}
func validateOrderPart(p *database.OrderPartFields, categories []database.OrderCategory) error {
	p.Name = strings.TrimSpace(p.Name)
	p.Shop = strings.TrimSpace(p.Shop)
	p.Link = strings.TrimSpace(p.Link)
	if !orderText(p.Name, 200, true) || !orderText(p.Shop, 200, true) {
		return fiber.NewError(400, "Provide a part name and shop, up to 200 characters each.")
	}
	if p.Amount < 1 || p.Amount > 10000 || p.UnitPriceCents < 0 || p.UnitPriceCents > 100000000 {
		return fiber.NewError(400, "Amount must be 1–10,000 whole units; unit price must be 0–1,000,000 EUR.")
	}
	if p.Link != "" {
		u, err := url.Parse(p.Link)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil || !orderText(p.Link, 2048, false) {
			return fiber.NewError(400, "Use an HTTP or HTTPS shop link without embedded credentials.")
		}
	}
	if p.CategoryID != "" && !slices.ContainsFunc(categories, func(v database.OrderCategory) bool { return v.ID == p.CategoryID }) {
		return fiber.NewError(400, "Select a category from this list.")
	}
	return nil
}
func (a *API) registerOrders(v1 fiber.Router) {
	r := v1.Group("/orders", RequireRoles("admin", "order_admin", "editor", "order_request"), func(c fiber.Ctx) error {
		if a.DB == nil {
			return c.Status(503).JSON(fiber.Map{"error": "Database unavailable."})
		}
		return c.Next()
	})
	r.Get("/stats", a.orderStats)
	r.Get("/standard-parts", a.standardParts)
	r.Post("/standard-parts", RequireRoles("admin", "order_admin"), a.saveStandardPart)
	r.Put("/standard-parts/:id", RequireRoles("admin", "order_admin"), a.saveStandardPart)
	r.Delete("/standard-parts/:id", RequireRoles("admin", "order_admin"), a.deleteStandardPart)
	r.Get("/lists", a.orderLists)
	r.Post("/lists", RequireRoles("admin", "order_admin"), a.createOrderList)
	r.Get("/lists/:id", a.getOrderList)
	r.Post("/lists/:id/actions", a.orderAction)
	r.Delete("/lists/:id", RequireRoles("admin", "order_admin"), a.deleteOrderList)
}
func visibleOrder(l database.OrderList, u User) database.OrderList {
	if !orderEditor(u) {
		own := []database.OrderRequest{}
		for _, r := range l.Content.Requests {
			if r.CreatedBy == u.ID {
				own = append(own, r)
			}
		}
		l.Content.Requests = own
	}
	return l
}
func orderSummary(l database.OrderList, u User) fiber.Map {
	l = visibleOrder(l, u)
	var total, orderedTotal int64
	ordered, pending := 0, 0
	for _, p := range l.Content.Parts {
		total += p.Amount * p.UnitPriceCents
		if p.OrderedAt != nil {
			ordered++
			orderedTotal += p.Amount * p.UnitPriceCents
		}
	}
	for _, r := range l.Content.Requests {
		if r.Status == "pending" {
			pending++
		}
	}
	return fiber.Map{"id": l.ID, "name": l.Name, "status": l.Status, "currency": l.Currency, "version": l.Version, "partCount": len(l.Content.Parts), "orderedCount": ordered, "pendingCount": pending, "totalCents": total, "orderedTotalCents": orderedTotal, "updatedAt": l.UpdatedAt, "closedAt": l.ClosedAt}
}
func (a *API) orderLists(c fiber.Ctx) error {
	lists := []database.OrderList{}
	if err := a.DB.WithContext(c.Context()).Order("created_at DESC, id DESC").Find(&lists).Error; err != nil {
		return orderError(c, err)
	}
	result := []fiber.Map{}
	u := c.Locals("user").(User)
	for _, l := range lists {
		result = append(result, orderSummary(l, u))
	}
	return c.JSON(fiber.Map{"lists": result})
}
func (a *API) orderStats(c fiber.Ctx) error {
	lists := []database.OrderList{}
	if err := a.DB.WithContext(c.Context()).Find(&lists).Error; err != nil {
		return orderError(c, err)
	}
	open, closed, pending, remaining := 0, 0, 0, 0
	var openTotal, remainingTotal int64
	u := c.Locals("user").(User)
	for _, l := range lists {
		l = visibleOrder(l, u)
		if l.Status == "open" {
			open++
		} else {
			closed++
		}
		for _, r := range l.Content.Requests {
			if r.Status == "pending" {
				pending++
			}
		}
		for _, p := range l.Content.Parts {
			if l.Status == "open" {
				openTotal += p.Amount * p.UnitPriceCents
			} else if p.OrderedAt == nil {
				remaining++
				remainingTotal += p.Amount * p.UnitPriceCents
			}
		}
	}
	return c.JSON(fiber.Map{"openLists": open, "closedLists": closed, "pendingRequests": pending, "remainingParts": remaining, "openTotalCents": openTotal, "remainingTotalCents": remainingTotal, "currency": "EUR"})
}
func (a *API) createOrderList(c fiber.Ctx) error {
	var input struct {
		Name string `json:"name"`
	}
	if err := orderBody(c, &input); err != nil {
		return orderError(c, err)
	}
	input.Name = strings.TrimSpace(input.Name)
	if !orderText(input.Name, 200, true) {
		return orderError(c, fiber.NewError(400, "Provide a list name, up to 200 characters."))
	}
	l := database.OrderList{ID: uuid.NewString(), Name: input.Name, Status: "open", Currency: "EUR", Version: 1, CreatedBy: c.Locals("user").(User).ID, Content: database.OrderContent{Categories: []database.OrderCategory{}, Parts: []database.OrderPart{}, Requests: []database.OrderRequest{}}}
	if err := a.DB.WithContext(c.Context()).Create(&l).Error; err != nil {
		return orderError(c, err)
	}
	return c.Status(201).JSON(fiber.Map{"list": l})
}
func orderListID(c fiber.Ctx) (string, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return "", fiber.NewError(400, "Invalid ID.")
	}
	return id.String(), nil
}
func (a *API) getOrderList(c fiber.Ctx) error {
	id, err := orderListID(c)
	if err != nil {
		return orderError(c, err)
	}
	var l database.OrderList
	if err := a.DB.WithContext(c.Context()).First(&l, "id = ?", id).Error; err != nil {
		return orderError(c, err)
	}
	return c.JSON(fiber.Map{"list": visibleOrder(l, c.Locals("user").(User))})
}

// Every operation uses the same row lock as closing/deleting the list.
func (a *API) changeOrder(c fiber.Ctx, version int, fn func(*gorm.DB, *database.OrderList) error) error {
	id, err := orderListID(c)
	if err != nil {
		return err
	}
	return a.DB.WithContext(c.Context()).Transaction(func(tx *gorm.DB) error {
		var l database.OrderList
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&l, "id = ?", id).Error; err != nil {
			return err
		}
		if version != l.Version {
			return fiber.NewError(409, "This list changed. Refresh it and try again; your changes have not been applied.")
		}
		return fn(tx, &l)
	})
}
func (a *API) deleteOrderList(c fiber.Ctx) error {
	var input struct {
		Version int `json:"version"`
	}
	if err := orderBody(c, &input); err != nil {
		return orderError(c, err)
	}
	err := a.changeOrder(c, input.Version, func(tx *gorm.DB, l *database.OrderList) error { return tx.Delete(l).Error })
	if err != nil {
		return orderError(c, err)
	}
	return c.SendStatus(204)
}

type orderCommand struct {
	Version  int                      `json:"version"`
	Action   string                   `json:"action"`
	TargetID string                   `json:"targetId"`
	Name     string                   `json:"name"`
	Part     database.OrderPartFields `json:"part"`
	Note     string                   `json:"note"`
	Ordered  bool                     `json:"ordered"`
}

func applyOrderCommand(l *database.OrderList, u User, cmd orderCommand) error {
	manager, editor := orderManager(u), orderEditor(u)
	// Closed lists are immutable for editors/requesters, including requests and categories.
	if l.Status == "closed" && !manager {
		return fiber.NewError(403, "Only order admins or admins can change a closed list.")
	}
	now := time.Now()
	actor := u.Username
	if actor == "" {
		actor = u.Email
	}
	if actor == "" {
		actor = u.ID
	}
	findPart := func() int {
		return slices.IndexFunc(l.Content.Parts, func(p database.OrderPart) bool { return p.ID == cmd.TargetID })
	}
	findRequest := func() int {
		return slices.IndexFunc(l.Content.Requests, func(r database.OrderRequest) bool { return r.ID == cmd.TargetID })
	}
	switch cmd.Action {
	case "rename":
		if !manager {
			return fiber.ErrForbidden
		}
		cmd.Name = strings.TrimSpace(cmd.Name)
		if !orderText(cmd.Name, 200, true) {
			return fiber.NewError(400, "Provide a list name, up to 200 characters.")
		}
		l.Name = cmd.Name
	case "close":
		if !manager {
			return fiber.ErrForbidden
		}
		if l.Status != "open" {
			return fiber.NewError(409, "List is already closed.")
		}
		for _, r := range l.Content.Requests {
			if r.Status == "pending" {
				return fiber.NewError(409, "Accept or reject pending requests before closing the list.")
			}
		}
		l.Status = "closed"
		l.ClosedAt = &now
	case "reopen":
		if !manager {
			return fiber.ErrForbidden
		}
		if l.Status != "closed" {
			return fiber.NewError(409, "List is already open.")
		}
		l.Status = "open"
		l.ClosedAt = nil
	case "add_category", "rename_category", "delete_category":
		if !manager {
			return fiber.ErrForbidden
		}
		i := slices.IndexFunc(l.Content.Categories, func(v database.OrderCategory) bool { return v.ID == cmd.TargetID })
		if cmd.Action != "add_category" && i < 0 {
			return fiber.ErrNotFound
		}
		if cmd.Action == "delete_category" {
			for _, p := range l.Content.Parts {
				if p.CategoryID == cmd.TargetID {
					return fiber.NewError(409, "Move parts out of this category before deleting it.")
				}
			}
			for _, r := range l.Content.Requests {
				if r.CategoryID == cmd.TargetID && r.Status == "pending" {
					return fiber.NewError(409, "Resolve or move pending requests before deleting this category.")
				}
			}
			l.Content.Categories = slices.Delete(l.Content.Categories, i, i+1)
		} else {
			cmd.Name = strings.TrimSpace(cmd.Name)
			if !orderText(cmd.Name, 100, true) {
				return fiber.NewError(400, "Provide a category name, up to 100 characters.")
			}
			for _, v := range l.Content.Categories {
				if (cmd.Action == "add_category" || v.ID != cmd.TargetID) && strings.EqualFold(v.Name, cmd.Name) {
					return fiber.NewError(409, "This category already exists.")
				}
			}
			if cmd.Action == "add_category" {
				if len(l.Content.Categories) >= 100 {
					return fiber.NewError(400, "Use at most 100 categories.")
				}
				l.Content.Categories = append(l.Content.Categories, database.OrderCategory{ID: uuid.NewString(), Name: cmd.Name})
			} else {
				l.Content.Categories[i].Name = cmd.Name
			}
		}
	case "add_part", "edit_part":
		if !editor {
			return fiber.ErrForbidden
		}
		if err := validateOrderPart(&cmd.Part, l.Content.Categories); err != nil {
			return err
		}
		if cmd.Action == "add_part" {
			if len(l.Content.Parts) >= 1000 {
				return fiber.NewError(400, "Use at most 1,000 parts per list.")
			}
			l.Content.Parts = append(l.Content.Parts, database.OrderPart{ID: uuid.NewString(), OrderPartFields: cmd.Part, CreatedBy: u.ID, CreatedByName: actor, CreatedAt: now})
		} else {
			i := findPart()
			if i < 0 {
				return fiber.ErrNotFound
			}
			p := &l.Content.Parts[i]
			// Changes to an ordered line need a fresh ordering check.
			if p.OrderPartFields != cmd.Part {
				p.OrderedAt = nil
				p.OrderedBy = ""
			}
			p.OrderPartFields = cmd.Part
		}
	case "delete_part":
		if !editor {
			return fiber.ErrForbidden
		}
		i := findPart()
		if i < 0 {
			return fiber.ErrNotFound
		}
		l.Content.Parts = slices.Delete(l.Content.Parts, i, i+1)
	case "request_part":
		if l.Status != "open" {
			return fiber.NewError(409, "Requests can only be submitted to open lists.")
		}
		if !orderRole(u, "order_request") && !editor {
			return fiber.ErrForbidden
		}
		if err := validateOrderPart(&cmd.Part, l.Content.Categories); err != nil {
			return err
		}
		if len(l.Content.Requests) >= 1000 {
			return fiber.NewError(400, "Use at most 1,000 requests per list.")
		}
		l.Content.Requests = append(l.Content.Requests, database.OrderRequest{ID: uuid.NewString(), OrderPartFields: cmd.Part, Status: "pending", CreatedBy: u.ID, CreatedByName: actor, CreatedAt: now})
	case "edit_request", "withdraw_request", "accept_request", "reject_request":
		i := findRequest()
		if i < 0 {
			return fiber.ErrNotFound
		}
		r := &l.Content.Requests[i]
		if r.Status != "pending" {
			return fiber.NewError(409, "This request has already been resolved.")
		}
		if cmd.Action == "edit_request" || cmd.Action == "withdraw_request" {
			if r.CreatedBy != u.ID {
				return fiber.ErrForbidden
			}
			if cmd.Action == "edit_request" {
				if err := validateOrderPart(&cmd.Part, l.Content.Categories); err != nil {
					return err
				}
				r.OrderPartFields = cmd.Part
			} else {
				r.Status = "withdrawn"
				r.ReviewedBy = u.ID
				r.ReviewedAt = &now
			}
		} else {
			if !editor {
				return fiber.ErrForbidden
			}
			if !orderText(cmd.Note, 1000, false) {
				return fiber.NewError(400, "Review notes must be at most 1,000 characters on one line.")
			}
			if cmd.Action == "accept_request" {
				if err := validateOrderPart(&cmd.Part, l.Content.Categories); err != nil {
					return err
				}
				if len(l.Content.Parts) >= 1000 {
					return fiber.NewError(400, "Use at most 1,000 parts per list.")
				}
				l.Content.Parts = append(l.Content.Parts, database.OrderPart{ID: uuid.NewString(), OrderPartFields: cmd.Part, CreatedBy: u.ID, CreatedByName: actor, CreatedAt: now, RequestID: r.ID})
				r.Status = "accepted"
			} else {
				r.Status = "rejected"
			}
			r.ReviewNote = strings.TrimSpace(cmd.Note)
			r.ReviewedBy = u.ID
			r.ReviewedAt = &now
		}
	case "set_ordered":
		if !manager {
			return fiber.ErrForbidden
		}
		if l.Status != "closed" {
			return fiber.NewError(409, "Close the list before marking parts ordered.")
		}
		i := findPart()
		if i < 0 {
			return fiber.ErrNotFound
		}
		if cmd.Ordered {
			l.Content.Parts[i].OrderedAt = &now
			l.Content.Parts[i].OrderedBy = u.ID
		} else {
			l.Content.Parts[i].OrderedAt = nil
			l.Content.Parts[i].OrderedBy = ""
		}
	default:
		return fiber.NewError(400, "Unknown order action.")
	}
	return nil
}
func (a *API) orderAction(c fiber.Ctx) error {
	var cmd orderCommand
	if err := orderBody(c, &cmd); err != nil {
		return orderError(c, err)
	}
	var result database.OrderList
	err := a.changeOrder(c, cmd.Version, func(tx *gorm.DB, l *database.OrderList) error {
		if err := applyOrderCommand(l, c.Locals("user").(User), cmd); err != nil {
			return err
		}
		l.Version++
		if err := tx.Save(l).Error; err != nil {
			return err
		}
		result = *l
		return nil
	})
	if err != nil {
		return orderError(c, err)
	}
	return c.JSON(fiber.Map{"list": visibleOrder(result, c.Locals("user").(User))})
}
func (a *API) standardParts(c fiber.Ctx) error {
	parts := []database.StandardPart{}
	if err := a.DB.WithContext(c.Context()).Order("lower(shop), lower(name), id").Find(&parts).Error; err != nil {
		return orderError(c, err)
	}
	return c.JSON(fiber.Map{"parts": parts})
}
func (a *API) saveStandardPart(c fiber.Ctx) error {
	var input struct {
		database.OrderPartFields
		Version int `json:"version"`
	}
	if err := orderBody(c, &input); err != nil {
		return orderError(c, err)
	}
	input.CategoryID = ""
	if err := validateOrderPart(&input.OrderPartFields, nil); err != nil {
		return orderError(c, err)
	}
	p := database.StandardPart{Name: input.Name, Amount: input.Amount, UnitPriceCents: input.UnitPriceCents, Shop: input.Shop, Link: input.Link, Version: 1}
	if c.Method() == "POST" {
		p.ID = uuid.NewString()
		if err := a.DB.WithContext(c.Context()).Create(&p).Error; err != nil {
			return orderError(c, err)
		}
		return c.Status(201).JSON(fiber.Map{"part": p})
	}
	id, err := orderListID(c)
	if err != nil {
		return orderError(c, err)
	}
	err = a.DB.WithContext(c.Context()).Transaction(func(tx *gorm.DB) error {
		var old database.StandardPart
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&old, "id = ?", id).Error; err != nil {
			return err
		}
		if old.Version != input.Version {
			return fiber.NewError(409, "This standard part changed. Refresh and try again.")
		}
		p.ID = id
		p.CreatedAt = old.CreatedAt
		p.Version = old.Version + 1
		return tx.Save(&p).Error
	})
	if err != nil {
		return orderError(c, err)
	}
	return c.JSON(fiber.Map{"part": p})
}
func (a *API) deleteStandardPart(c fiber.Ctx) error {
	id, err := orderListID(c)
	if err != nil {
		return orderError(c, err)
	}
	var input struct {
		Version int `json:"version"`
	}
	if err := orderBody(c, &input); err != nil {
		return orderError(c, err)
	}
	result := a.DB.WithContext(c.Context()).Where("id = ? AND version = ?", id, input.Version).Delete(&database.StandardPart{})
	if result.Error != nil {
		return orderError(c, result.Error)
	}
	if result.RowsAffected != 1 {
		return orderError(c, fiber.NewError(409, "This standard part changed or was deleted. Refresh and try again."))
	}
	return c.SendStatus(204)
}

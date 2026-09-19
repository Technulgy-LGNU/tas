package web

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"tas/backend/config"
	"tas/backend/database"
)

func orderFixture(t *testing.T) (*API, *fiber.App, *User) {
	t.Helper()
	db := websiteDatabase(t)
	if err := db.AutoMigrate(&database.OrderList{}, &database.StandardPart{}); err != nil {
		t.Fatal(err)
	}
	a := &API{DB: db}
	user := &User{ID: "manager", Username: "Order Manager", Roles: []string{"order_admin"}}
	auth, err := NewAuth(config.AuthConfig{DisableFusionAuth: true})
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New(fiber.Config{Immutable: true})
	v1 := app.Group("/api/v1", auth.CSRF, func(c fiber.Ctx) error { c.Locals("user", *user); return c.Next() })
	a.registerOrders(v1)
	return a, app, user
}
func decodeOrder(t *testing.T, body map[string]any) database.OrderList {
	t.Helper()
	raw, err := json.Marshal(body["list"])
	if err != nil {
		t.Fatal(err)
	}
	var l database.OrderList
	if err := json.Unmarshal(raw, &l); err != nil {
		t.Fatal(err)
	}
	return l
}
func TestOrderWorkflowAndPermissions(t *testing.T) {
	a, app, user := orderFixture(t)
	status, body := siteReq(t, app, "POST", "/api/v1/orders/lists", fiber.Map{"name": "German Open 2027"})
	if status != 201 {
		t.Fatalf("create: %d %v", status, body)
	}
	l := decodeOrder(t, body)
	path := "/api/v1/orders/lists/" + l.ID
	action := func(cmd orderCommand, want int) {
		t.Helper()
		cmd.Version = l.Version
		status, body = siteReq(t, app, "POST", path+"/actions", cmd)
		if status != want {
			t.Fatalf("%s status %d want %d: %v", cmd.Action, status, want, body)
		}
		if status == 200 {
			l = decodeOrder(t, body)
		}
	}
	action(orderCommand{Action: "add_category", Name: "Team Alpha"}, 200)
	categoryID := l.Content.Categories[0].ID
	action(orderCommand{Action: "add_category", TargetID: categoryID, Name: "team alpha"}, 409)
	part := database.OrderPartFields{Name: "Motor", Amount: 3, UnitPriceCents: 199, Shop: "Robot Shop", Link: "https://shop.example.org/motor", CategoryID: categoryID}
	user.ID = "team-a"
	user.Username = "Team A"
	user.Roles = []string{"order_request"}
	action(orderCommand{Action: "add_part", Part: part}, 403)
	action(orderCommand{Action: "request_part", Part: part}, 200)
	requestID := l.Content.Requests[0].ID
	if len(l.Content.Parts) != 0 {
		t.Fatal("request added a part before approval")
	}
	action(orderCommand{Action: "accept_request", TargetID: requestID, Part: part}, 403)
	action(orderCommand{Action: "rename", Name: "Unauthorized"}, 403)
	status, _ = siteReq(t, app, "POST", "/api/v1/orders/lists", fiber.Map{"name": "Not allowed"})
	if status != 403 {
		t.Fatal("requester created list")
	}
	// A different team cannot see or edit the request, even knowing its ID.
	user.ID = "team-b"
	status, body = siteReq(t, app, "GET", path, nil)
	if status != 200 || len(decodeOrder(t, body).Content.Requests) != 0 {
		t.Fatal("another team's requests leaked")
	}
	action(orderCommand{Action: "edit_request", TargetID: requestID, Part: part}, 403)
	action(orderCommand{Action: "withdraw_request", TargetID: requestID}, 403)
	status, body = siteReq(t, app, "GET", "/api/v1/orders/stats", nil)
	if status != 200 || body["pendingRequests"] != float64(0) {
		t.Fatal("other team's pending count leaked")
	}
	user.ID = "manager"
	user.Roles = []string{"order_admin"}
	action(orderCommand{Action: "close"}, 409)
	action(orderCommand{Action: "delete_category", TargetID: categoryID}, 409)
	user.ID = "editor"
	user.Roles = []string{"editor"}
	action(orderCommand{Action: "add_category", Name: "Not allowed"}, 403)
	action(orderCommand{Action: "close"}, 403)
	part.Amount = 4
	versionBeforeApproval := l.Version
	action(orderCommand{Action: "accept_request", TargetID: requestID, Part: part, Note: "Approved four motors"}, 200)
	if len(l.Content.Parts) != 1 || l.Content.Parts[0].RequestID != requestID || l.Content.Parts[0].Amount != 4 || l.Content.Requests[0].Status != "accepted" {
		t.Fatal("acceptance snapshot incorrect")
	}
	// Retrying with a fresh version must not duplicate the approved part.
	action(orderCommand{Action: "accept_request", TargetID: requestID, Part: part}, 409)
	status, _ = siteReq(t, app, "POST", path+"/actions", orderCommand{Action: "add_part", Version: versionBeforeApproval, Part: part})
	if status != 409 {
		t.Fatal("stale list save accepted")
	}
	action(orderCommand{Action: "set_ordered", TargetID: l.Content.Parts[0].ID, Ordered: true}, 403)
	action(orderCommand{Action: "add_part", Part: database.OrderPartFields{Name: "Sensor", Shop: "Another Shop", Amount: 2, UnitPriceCents: 101}}, 200)
	status, body = siteReq(t, app, "GET", "/api/v1/orders/lists", nil)
	if status != 200 || body["lists"].([]any)[0].(map[string]any)["totalCents"] != float64(998) {
		t.Fatalf("incorrect money total: %v", body)
	}
	user.Roles = []string{"order_admin"}
	action(orderCommand{Action: "set_ordered", TargetID: l.Content.Parts[0].ID, Ordered: true}, 409)
	action(orderCommand{Action: "close"}, 200)
	closedVersion := l.Version
	user.Roles = []string{"editor"}
	for _, cmd := range []orderCommand{{Action: "add_part", Part: part}, {Action: "edit_part", TargetID: l.Content.Parts[0].ID, Part: part}, {Action: "delete_part", TargetID: l.Content.Parts[0].ID}, {Action: "reopen"}, {Action: "set_ordered", TargetID: l.Content.Parts[0].ID, Ordered: true}} {
		action(cmd, 403)
	}
	user.Roles = []string{"order_request"}
	action(orderCommand{Action: "request_part", Part: part}, 403)
	user.Roles = []string{"order_admin"}
	action(orderCommand{Action: "set_ordered", TargetID: l.Content.Parts[0].ID, Ordered: true}, 200)
	if l.Content.Parts[0].OrderedAt == nil {
		t.Fatal("ordering not recorded")
	}
	status, body = siteReq(t, app, "GET", "/api/v1/orders/stats", nil)
	if status != 200 || body["remainingParts"] != float64(1) || body["remainingTotalCents"] != float64(202) || body["closedLists"] != float64(1) {
		t.Fatalf("closed stats: %v", body)
	}
	// A stale checklist must not overwrite a more recent action.
	status, _ = siteReq(t, app, "POST", path+"/actions", orderCommand{Version: closedVersion, Action: "set_ordered", TargetID: l.Content.Parts[1].ID, Ordered: true})
	if status != 409 {
		t.Fatal("stale checklist update accepted")
	}
	user.Roles = []string{"admin"}
	part.Amount = 5
	action(orderCommand{Action: "edit_part", TargetID: l.Content.Parts[0].ID, Part: part}, 200)
	if l.Content.Parts[0].OrderedAt != nil {
		t.Fatal("changed ordered part retained its check")
	}
	action(orderCommand{Action: "delete_category", TargetID: categoryID}, 409)
	action(orderCommand{Action: "reopen"}, 200)
	if l.ClosedAt != nil {
		t.Fatal("reopen left closed timestamp")
	}
	// These records are inside a rollback-only test schema.
	var persisted database.OrderList
	if err := a.DB.First(&persisted, "id = ?", l.ID).Error; err != nil || persisted.Version != l.Version {
		t.Fatal("list not persisted")
	}
	user.Roles = []string{"editor"}
	status, _ = siteReq(t, app, "DELETE", path, fiber.Map{"version": l.Version})
	if status != 403 {
		t.Fatal("editor deleted list")
	}
	user.Roles = []string{"order_admin"}
	status, _ = siteReq(t, app, "DELETE", path, fiber.Map{"version": l.Version})
	if status != 204 {
		t.Fatal("manager cannot delete list")
	}
	status, _ = siteReq(t, app, "GET", path, nil)
	if status != 404 {
		t.Fatal("deleted list still visible")
	}
}
func TestStandardPartsAndOrderAccess(t *testing.T) {
	_, app, user := orderFixture(t)
	part := database.OrderPartFields{Name: "Screw", Amount: 10, UnitPriceCents: 7, Shop: "Hardware", Link: "https://shop.example.org/screw"}
	user.Roles = []string{"editor"}
	status, _ := siteReq(t, app, "POST", "/api/v1/orders/standard-parts", part)
	if status != 403 {
		t.Fatal("editor created standard part")
	}
	user.Roles = []string{"order_admin"}
	status, body := siteReq(t, app, "POST", "/api/v1/orders/standard-parts", part)
	if status != 201 {
		t.Fatalf("create standard: %v", body)
	}
	p := body["part"].(map[string]any)
	id := p["id"].(string)
	p["unitPriceCents"] = 9
	status, body = siteReq(t, app, "PUT", "/api/v1/orders/standard-parts/"+id, p)
	if status != 200 || body["part"].(map[string]any)["version"] != float64(2) {
		t.Fatalf("standard update failed: %v", body)
	}
	status, _ = siteReq(t, app, "PUT", "/api/v1/orders/standard-parts/"+id, p)
	if status != 409 {
		t.Fatal("stale standard update accepted")
	}
	status, body = siteReq(t, app, "POST", "/api/v1/orders/lists", fiber.Map{"name": "Tools"})
	if status != 201 {
		t.Fatal(body)
	}
	l := decodeOrder(t, body)
	status, body = siteReq(t, app, "POST", "/api/v1/orders/lists/"+l.ID+"/actions", orderCommand{Version: l.Version, Action: "add_part", Part: part})
	if status != 200 {
		t.Fatal(body)
	}
	status, _ = siteReq(t, app, "DELETE", "/api/v1/orders/standard-parts/"+id, fiber.Map{"version": 99})
	if status != 409 {
		t.Fatal("stale standard delete accepted")
	}
	status, _ = siteReq(t, app, "DELETE", "/api/v1/orders/standard-parts/"+id, fiber.Map{"version": 2})
	if status != 204 {
		t.Fatal("standard delete failed")
	}
	status, body = siteReq(t, app, "GET", "/api/v1/orders/lists/"+l.ID, nil)
	if status != 200 || len(decodeOrder(t, body).Content.Parts) != 1 {
		t.Fatal("deleting standard affected copied line")
	}
	user.Roles = []string{"viewer"}
	for _, path := range []string{"/lists", "/stats", "/standard-parts", "/lists/" + l.ID} {
		status, _ = siteReq(t, app, "GET", "/api/v1/orders"+path, nil)
		if status != 403 {
			t.Fatalf("viewer access: %s", path)
		}
	}
	user.Roles = []string{"admin"}
	req := httptest.NewRequest("POST", "/api/v1/orders/lists", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 403 {
		t.Fatal("missing CSRF accepted")
	}
}
func TestOrderValidation(t *testing.T) {
	good := database.OrderPartFields{Name: "Motor", Shop: "Shop", Amount: 3, UnitPriceCents: 199, Link: "https://example.org"}
	for _, mutate := range []func(*database.OrderPartFields){func(p *database.OrderPartFields) { p.Amount = 0 }, func(p *database.OrderPartFields) { p.Amount = 10001 }, func(p *database.OrderPartFields) { p.UnitPriceCents = -1 }, func(p *database.OrderPartFields) { p.UnitPriceCents = 100000001 }, func(p *database.OrderPartFields) { p.Link = "javascript:alert(1)" }, func(p *database.OrderPartFields) { p.Link = "https://user:secret@example.org" }, func(p *database.OrderPartFields) { p.CategoryID = uuid.NewString() }, func(p *database.OrderPartFields) { p.Name = "" }} {
		p := good
		mutate(&p)
		if validateOrderPart(&p, nil) == nil {
			t.Fatalf("invalid part accepted: %+v", p)
		}
	}
	if err := validateOrderPart(&good, nil); err != nil {
		t.Fatal(err)
	}
	// Resolved requests remain immutable, even for their original requester.
	user := User{ID: "requester", Roles: []string{"order_request"}}
	l := database.OrderList{Status: "open", Content: database.OrderContent{Requests: []database.OrderRequest{{ID: "request", CreatedBy: user.ID, Status: "accepted"}}}}
	if applyOrderCommand(&l, user, orderCommand{Action: "edit_request", TargetID: "request", Part: good}) == nil {
		t.Fatal("resolved request editable")
	}
}

func TestOrderRequestResolution(t *testing.T) {
	_, app, user := orderFixture(t)
	status, body := siteReq(t, app, "POST", "/api/v1/orders/lists", fiber.Map{"name": "Requests"})
	if status != 201 {
		t.Fatal(body)
	}
	l := decodeOrder(t, body)
	path := "/api/v1/orders/lists/" + l.ID + "/actions"
	action := func(cmd orderCommand, want int) {
		t.Helper()
		cmd.Version = l.Version
		status, body = siteReq(t, app, "POST", path, cmd)
		if status != want {
			t.Fatalf("%s: %d %v", cmd.Action, status, body)
		}
		if status == 200 {
			l = decodeOrder(t, body)
		}
	}
	user.ID = "team"
	user.Roles = []string{"order_request"}
	part := database.OrderPartFields{Name: "Screw", Shop: "Hardware", Amount: 10, UnitPriceCents: 7}
	action(orderCommand{Action: "request_part", Part: part}, 200)
	id := l.Content.Requests[0].ID
	part.Amount = 20
	action(orderCommand{Action: "edit_request", TargetID: id, Part: part}, 200)
	if l.Content.Requests[0].Amount != 20 {
		t.Fatal("request edit not persisted")
	}
	action(orderCommand{Action: "withdraw_request", TargetID: id}, 200)
	action(orderCommand{Action: "edit_request", TargetID: id, Part: part}, 409)
	action(orderCommand{Action: "request_part", Part: part}, 200)
	id = l.Content.Requests[1].ID
	user.Roles = []string{"editor"}
	action(orderCommand{Action: "reject_request", TargetID: id, Note: "Already in stock"}, 200)
	if l.Content.Requests[1].Status != "rejected" || l.Content.Requests[1].ReviewNote != "Already in stock" || l.Content.Requests[1].ReviewedAt == nil || len(l.Content.Parts) != 0 {
		t.Fatal("incorrect rejection history")
	}
	action(orderCommand{Action: "accept_request", TargetID: id, Part: part}, 409)
	user.Roles = []string{"order_admin"}
	action(orderCommand{Action: "close"}, 200)
}

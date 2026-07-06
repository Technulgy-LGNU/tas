package web

import (
	"errors"
	"strconv"
	"strings"
	"tas/internal/database"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func fail(status int, message string) error {
	return fiber.NewError(status, message)
}

func bindJSON[T any](c *fiber.Ctx) (T, error) {
	var payload T
	if err := c.BodyParser(&payload); err != nil {
		return payload, fail(fiber.StatusBadRequest, "invalid JSON body")
	}
	return payload, nil
}

func uintParam(c *fiber.Ctx, name string) (uint, error) {
	raw := strings.TrimSpace(c.Params(name))
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, fail(fiber.StatusBadRequest, "invalid "+name)
	}
	return uint(id), nil
}

func currentMember(c *fiber.Ctx) (*database.Member, error) {
	member, ok := c.Locals("member").(*database.Member)
	if !ok || member == nil {
		return nil, fail(fiber.StatusUnauthorized, "missing member context")
	}
	return member, nil
}

func (a *API) requirePermission(keys ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := a.authenticate(c, false); err != nil {
			return err
		}
		member, err := currentMember(c)
		if err != nil {
			return err
		}
		if member.Status != database.MemberStatusApproved {
			return fail(fiber.StatusForbidden, "member is not approved")
		}
		if a.memberHasAnyPermission(member, keys...) {
			return c.Next()
		}
		return fail(fiber.StatusForbidden, "missing required permission")
	}
}

func (a *API) memberHasAnyPermission(member *database.Member, keys ...string) bool {
	if a.CFG.Auth.DevAllowAdmin && member.FusionAuthUserID == "dev-admin" {
		return true
	}
	required := make(map[string]bool, len(keys))
	for _, key := range keys {
		required[key] = true
	}
	for _, role := range member.Roles {
		for _, permission := range role.Permissions {
			if required[permission.Key] {
				return true
			}
		}
	}
	return false
}

func (a *API) requireManageOrDraft(c *fiber.Ctx, list *database.OrderList) error {
	member, err := currentMember(c)
	if err != nil {
		return err
	}
	if list.Status != database.OrderListStatusPublished || a.memberHasAnyPermission(member, "orders:manage") {
		return nil
	}
	return fail(fiber.StatusForbidden, "published order lists require manage permission")
}

func preloadMemberPermissions(db *gorm.DB) *gorm.DB {
	return db.Preload("Roles.Permissions")
}

func notFoundOrError(err error, message string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(fiber.StatusNotFound, message)
	}
	return err
}

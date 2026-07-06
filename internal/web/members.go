package web

import (
	"sort"
	"tas/internal/database"

	"github.com/gofiber/fiber/v2"
)

type meResponse struct {
	Member      *database.Member `json:"member"`
	Permissions []string         `json:"permissions"`
}

func (a *API) getMe(c *fiber.Ctx) error {
	member, err := currentMember(c)
	if err != nil {
		return err
	}
	return c.JSON(meResponse{Member: member, Permissions: memberPermissionKeys(member)})
}

func (a *API) listRoles(c *fiber.Ctx) error {
	var roles []database.Role
	if err := a.DB.Preload("Permissions").Order("name asc").Find(&roles).Error; err != nil {
		return err
	}
	return c.JSON(roles)
}

func (a *API) listMembers(c *fiber.Ctx) error {
	var members []database.Member
	if err := preloadMemberPermissions(a.DB).Order("created_at desc").Find(&members).Error; err != nil {
		return err
	}
	return c.JSON(members)
}

type updateMemberPayload struct {
	Status  *string `json:"status"`
	RoleIDs []uint  `json:"role_ids"`
	Name    *string `json:"name"`
	Email   *string `json:"email"`
}

func (a *API) updateMember(c *fiber.Ctx) error {
	id, err := uintParam(c, "id")
	if err != nil {
		return err
	}
	payload, err := bindJSON[updateMemberPayload](c)
	if err != nil {
		return err
	}

	var member database.Member
	if err := a.DB.First(&member, id).Error; err != nil {
		return notFoundOrError(err, "member not found")
	}

	updates := map[string]any{}
	if payload.Status != nil {
		switch *payload.Status {
		case database.MemberStatusPending, database.MemberStatusApproved, database.MemberStatusRejected:
			updates["status"] = *payload.Status
		default:
			return fail(fiber.StatusBadRequest, "invalid member status")
		}
	}
	if payload.Name != nil {
		updates["name"] = *payload.Name
	}
	if payload.Email != nil {
		updates["email"] = *payload.Email
	}
	if len(updates) > 0 {
		if err := a.DB.Model(&member).Updates(updates).Error; err != nil {
			return err
		}
	}
	if payload.RoleIDs != nil {
		var roles []database.Role
		if len(payload.RoleIDs) > 0 {
			if err := a.DB.Where("id IN ?", payload.RoleIDs).Find(&roles).Error; err != nil {
				return err
			}
			if len(roles) != len(payload.RoleIDs) {
				return fail(fiber.StatusBadRequest, "one or more roles were not found")
			}
		}
		if err := a.DB.Model(&member).Association("Roles").Replace(roles); err != nil {
			return err
		}
	}

	if err := preloadMemberPermissions(a.DB).First(&member, id).Error; err != nil {
		return err
	}
	return c.JSON(member)
}

func memberPermissionKeys(member *database.Member) []string {
	seen := map[string]bool{}
	for _, role := range member.Roles {
		for _, permission := range role.Permissions {
			seen[permission.Key] = true
		}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

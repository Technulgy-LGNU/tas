package database

import "gorm.io/gorm"

func InitPSQL(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&Permission{},
		&Role{},
		&Member{},
		&InventoryCategory{},
		&InventoryItem{},
		&OrderRequest{},
		&OrderList{},
		&OrderListItem{},
		&UploadedImage{},
		&Team{},
		&Competition{},
		&Prize{},
		&SponsorCategory{},
		&Sponsor{},
		&HomeArticle{},
	); err != nil {
		return err
	}
	return seedDefaults(db)
}

func seedDefaults(db *gorm.DB) error {
	permissions := []Permission{
		{Key: "inventory:view", Description: "View inventory"},
		{Key: "inventory:request", Description: "Request inventory-related orders"},
		{Key: "inventory:edit", Description: "Edit inventory"},
		{Key: "inventory:manage", Description: "Manage inventory"},
		{Key: "orders:view", Description: "View order requests and lists"},
		{Key: "orders:request", Description: "Create order requests"},
		{Key: "orders:edit", Description: "Edit draft order lists"},
		{Key: "orders:manage", Description: "Approve, publish, order, and receive"},
		{Key: "website:view", Description: "View website management data"},
		{Key: "website:edit", Description: "Edit website content"},
		{Key: "website:manage", Description: "Publish and manage website content"},
		{Key: "members:view", Description: "View members and roles"},
		{Key: "members:manage", Description: "Approve members and assign roles"},
	}

	for _, permission := range permissions {
		if err := db.Where(Permission{Key: permission.Key}).FirstOrCreate(&permission).Error; err != nil {
			return err
		}
	}

	var allPermissions []Permission
	if err := db.Find(&allPermissions).Error; err != nil {
		return err
	}

	roleSpecs := []struct {
		name        string
		description string
		keys        []string
	}{
		{
			name:        "admin",
			description: "Full platform access",
			keys:        permissionKeys(allPermissions),
		},
		{
			name:        "manager",
			description: "Manage inventory, orders, website content, and members",
			keys: []string{
				"inventory:view", "inventory:request", "inventory:edit", "inventory:manage",
				"orders:view", "orders:request", "orders:edit", "orders:manage",
				"website:view", "website:edit", "website:manage",
				"members:view", "members:manage",
			},
		},
		{
			name:        "editor",
			description: "Edit inventory, draft order lists, and website content",
			keys: []string{
				"inventory:view", "inventory:request", "inventory:edit",
				"orders:view", "orders:request", "orders:edit",
				"website:view", "website:edit",
			},
		},
		{
			name:        "requester",
			description: "View inventory and create order requests",
			keys: []string{
				"inventory:view", "inventory:request",
				"orders:view", "orders:request",
				"website:view",
			},
		},
		{
			name:        "viewer",
			description: "Read-only app access",
			keys: []string{
				"inventory:view",
				"orders:view",
				"website:view",
			},
		},
	}

	permissionByKey := make(map[string]Permission, len(allPermissions))
	for _, permission := range allPermissions {
		permissionByKey[permission.Key] = permission
	}

	for _, spec := range roleSpecs {
		role := Role{Name: spec.name, Description: spec.description}
		if err := db.Where(Role{Name: spec.name}).FirstOrCreate(&role).Error; err != nil {
			return err
		}
		if err := db.Model(&role).Association("Permissions").Replace(permissionsForKeys(permissionByKey, spec.keys)); err != nil {
			return err
		}
	}

	return nil
}

func permissionKeys(permissions []Permission) []string {
	keys := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		keys = append(keys, permission.Key)
	}
	return keys
}

func permissionsForKeys(permissionByKey map[string]Permission, keys []string) []Permission {
	permissions := make([]Permission, 0, len(keys))
	for _, key := range keys {
		if permission, ok := permissionByKey[key]; ok {
			permissions = append(permissions, permission)
		}
	}
	return permissions
}

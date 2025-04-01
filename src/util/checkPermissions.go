package util

import (
	"gorm.io/gorm"
	"strings"
	"tas/src/database"
)

// CheckPermissions checks if the user has the required permissions
func CheckPermissions(headers map[string][]string, level int, subPart string, db *gorm.DB) bool {
	// Check if bearer token is present
	if headers["Authorization"] == nil {
		return false
	}
	// Check if bearer token is valid
	token := strings.TrimPrefix(headers["Authorization"][0], "Bearer ")
	if token == "" {
		return false
	}

	// Check for an matching entry in the database
	var key database.BrowserToken
	result := db.Where("key = ?", token).First(&key)
	if result.Error != nil {
		return false
	}

	// Get the permissions for the user
	var perms database.Permission
	result = db.Where("user_id = ?", key.MemberID).First(&perms)
	if result.Error != nil {
		return false
	}

	// Check if the user has the required permissions
	switch subPart {
	case "members":
		if perms.Members >= level || perms.Admin {
			return true
		} else {
			return false
		}
	case "forms":
		if perms.Form >= level || perms.Admin {
			return true
		} else {
			return false
		}
	case "events":
		if perms.Events >= level || perms.Admin {
			return true
		} else {
			return false
		}
	case "teams":
		if perms.Teams >= level || perms.Admin {
			return true
		} else {
			return false
		}
	case "orders":
		if perms.Orders >= level || perms.Admin {
			return true
		} else {
			return false
		}
	case "newsletter":
		if perms.Newsletter >= level || perms.Admin {
			return true
		} else {
			return false
		}
	case "website":
		if perms.Website >= level || perms.Admin {
			return true
		} else {
			return false
		}
	case "admin":
		if perms.Admin {
			return true
		} else {
			return false
		}
	case "login":
		if perms.Login {
			return true
		} else {
			return false
		}
	default:
		if perms.Admin {
			return true
		} else {
			return false
		}
	}
}

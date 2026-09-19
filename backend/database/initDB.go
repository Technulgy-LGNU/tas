package database

import (
	"gorm.io/gorm"
	"log"
)

func InitDB(db *gorm.DB) {
	if err := db.AutoMigrate(&Image{}, &WebsiteEntry{}, &WebsiteReference{}); err != nil {
		log.Fatalf("Error migrating image library: %v", err)
	}
}

package database

import (
	"gorm.io/gorm"
	"log"
)

func InitDB(db *gorm.DB) {
	if err := db.AutoMigrate(&Image{}, &WebsiteEntry{}, &WebsiteReference{}, &OrderList{}, &StandardPart{}); err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}
}

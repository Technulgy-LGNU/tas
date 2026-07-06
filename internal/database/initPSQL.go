package database

import "gorm.io/gorm"

func InitPSQL(db *gorm.DB) error {
	return db.AutoMigrate()
}

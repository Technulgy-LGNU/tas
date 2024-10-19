package storage

import "gorm.io/gorm"

type FileUpdateData struct {
	NewName string
	NewUrl  string
}

func UpdateFile(name string, data FileUpdateData, db *gorm.DB) error {
	return nil
}

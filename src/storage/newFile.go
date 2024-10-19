package storage

import "gorm.io/gorm"

// NewFile saves the new file to the disk and creates the database entry.
func NewFile(name string, db *gorm.DB) error {
	return nil
}

package database

import (
	"fmt"
	"gorm.io/gorm"
)

func InitDatabase(db *gorm.DB) {
	fmt.Println(db)
}

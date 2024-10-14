package web

import (
	"fmt"
	"gorm.io/gorm"
	"tas/src/config"
)

func InitWeb(cfg *config.CFG, db *gorm.DB) {
	fmt.Println(cfg, db)
}

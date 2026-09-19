package main

import (
	"tas/backend/config"
	"tas/backend/database"
	"tas/backend/web"
)

func main() {
	var cfg = config.GetConfig()

	var db = database.GetDB(cfg)
	database.InitDB(db)

	web.InitWeb(cfg, db)
}

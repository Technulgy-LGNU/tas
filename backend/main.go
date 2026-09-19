package main

import (
	"fmt"
	"tas/backend/config"
	"tas/backend/database"
	"tas/backend/web"
)

func main() {
  var cfg = config.GetConfig()
  fmt.Println(cfg)

  var db = database.GetDB(cfg)
  database.InitDB(db)

  web.InitWeb(cfg, db)
}

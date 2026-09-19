package main

import (
	"log"
	"log/slog"
	"tas/backend/config"
	"tas/backend/database"
	"tas/backend/observability"
	"tas/backend/web"
)

func main() {
	var cfg = config.GetConfig()
	if err := observability.Configure(cfg.Logging.Level); err != nil {
		log.Fatal(err)
	}
	slog.Info("startup.database_connecting")

	var db = database.GetDB(cfg)
	slog.Info("startup.database_connected")
	database.InitDB(db)
	slog.Info("startup.migrations_complete")

	web.InitWeb(cfg, db)
}
